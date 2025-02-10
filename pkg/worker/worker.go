package worker

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"

	nats "github.com/nats-io/nats.go"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type NATSConfig struct {
	NATSURL     string
	NATSSubject string
}

type Config struct {
	Nconfig   NATSConfig
	DonorUUID string
}

type consume struct {
	config Config
	cli    *kubernetes.Clientset
}

type Results struct {
	stealerUUID    string
	donorUUID      string
	message        string
	status         string
	startTime      metav1.Time
	completionTime metav1.Time
	duration       string
}

type PodResults struct {
	Results Results     `json:"results"`
	Pod     *corev1.Pod `json:"pod"`
}

func getClientSet() (*kubernetes.Clientset, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return clientset, err
}

func New(config Config) (*consume, error) {
	clientset, err := getClientSet()
	if err != nil {
		return nil, err
	}
	return &consume{
		cli:    clientset,
		config: config,
	}, nil
}

func (n *consume) Start(stopChan chan<- bool) error {
	defer func() { stopChan <- true }()

	// Configure structured logging with slog
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// Connect to NATS server
	natsConnect, err := nats.Connect(n.config.Nconfig.NATSURL)
	if err != nil {
		slog.Error("Failed to connect to NATS server: ", "error", err)
		return err
	}
	defer natsConnect.Close()
	slog.Info("Connected to NATS server", "server", n.config.Nconfig.NATSURL)

	slog.Info("Subscribe to stolen Pod results messages...", "subject", n.config.Nconfig.NATSSubject)
	// // Subscribe to the subject
	natsConnect.Subscribe(n.config.Nconfig.NATSSubject, func(msg *nats.Msg) {
		var pod corev1.Pod
		var podResults PodResults
		slog.Info("Received message", "subject", n.config.Nconfig.NATSSubject, "data", string(msg.Data))

		// Deserialize the entire pod metadata to JSON
		err := json.Unmarshal(msg.Data, &podResults)
		if err != nil {
			slog.Error("Failed to Unmarshal podResults from rawData", "error", err, "rawData", string(msg.Data))
			return
		}
		pod = *podResults.Pod
		slog.Info("Deserialized Pod", "pod", pod)
		if pod.Labels["donorUUID"] != n.config.DonorUUID {
			slog.Error("Pod is not intended for this donor", "podName", pod.Name, "podNamespace", pod.Namespace)
			return

		}
		results := podResults.Results
		slog.Info("Stolen Pod Execution Results", "results", results)

		// Check if the Pod exists or not
		currentPod, err := n.cli.CoreV1().Pods(pod.Namespace).Get(context.TODO(), pod.Name, metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			slog.Warn("Pod not found. It might have been deleted before acknowledging the message, possibly due to a worker restart",
				"podName", pod.Name, "podNamespace", pod.Namespace)
			msg.Ack()
			return
		} else if pod.Labels["stealerUUID"] != currentPod.Labels["stealerUUID"] ||
			pod.Labels["donorUUID"] != currentPod.Labels["donorUUID"] {
			slog.Error("Stolen Pod Lables are not matching with the current Pod", "podName", pod.Name,
				"podNamespace", pod.Namespace, "currentPodLables", currentPod.Labels, "stolenPodLables", pod.Labels)
			return
		} else if err != nil {
			slog.Error("Error retrieving Pod", "podName", pod.Name, "podNamespace", pod.Namespace, "error", err)
			return
		}

		if currentPod.Status.Phase != corev1.PodPending {
			slog.Error("Pod is not in Pending state", "podName", pod.Name, "podNamespace", pod.Namespace,
				"phase", currentPod.Status.Phase)
			return
		}

		// Delete the pending Pod here
		err = n.cli.CoreV1().Pods(pod.Namespace).Delete(context.TODO(), pod.Name, metav1.DeleteOptions{})
		if err != nil {
			slog.Error("Failed to delete Pod", "podName", pod.Name, "podNamespace", pod.Namespace, "error", err)
			return
		}
		slog.Info("Deleted the pending Pod", "podName", pod.Name, "podNamespace", pod.Namespace)

		// Acknowledge message. This will remove the message from the NATS server
		msg.Ack()
	})
	select {}
}
