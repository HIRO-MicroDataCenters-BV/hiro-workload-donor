package main

import (
	"flag"
	"fmt"
	"log"
	"log/slog"
	"strings"

	"workloadstealagent/pkg/controller"
	"workloadstealagent/pkg/validate"
	"workloadstealagent/pkg/worker"

	"github.com/google/uuid"
	"github.com/spf13/viper"
)

var (
	tlsKeyPath  string
	tlsCertPath string
)

func main() {
	stopChan := make(chan bool)

	// To Do: Every resrat of the donor will have a new UUID.
	// This logic has to be changed so that the UUID is persisted.
	donorUUID := uuid.New().String()
	// donorUUID := "b458a190-4744-46f4-b16b-2739cf9fccb8"

	slog.Info("Configuring Validator")
	natsConfig := validate.NATSConfig{
		NATSURL:     getENVValue("NATS_URL"),
		NATSSubject: getENVValue("NATS_WORKLOAD_SUBJECT"),
	}
	validatorConfig := validate.Config{
		Nconfig:          natsConfig,
		DonorUUID:        donorUUID,
		IgnoreNamespaces: strings.Split(getENVValue("IGNORE_NAMESPACES"), ","),
		LableToFilter:    getENVValue("NO_WORK_LOAD_STEAL_LABLE"),
	}
	validator, err := validate.New(validatorConfig)
	if err != nil {
		log.Fatal(err)
	}

	slog.Info("Configuring Controller(Admission WebHook)")
	controllerConfig := controller.Config{
		MPort:       8443,
		VPort:       8444,
		TLSKeyPath:  tlsKeyPath,
		TLSCertPath: tlsCertPath,
	}
	server := controller.New(controllerConfig, validator)

	slog.Info("Configuring Worker")
	workerNATSConfig := worker.NATSConfig{
		NATSURL:     getENVValue("NATS_URL"),
		NATSSubject: getENVValue("NATS_RETURN_WORKLOAD_SUBJECT"),
	}
	workerConfig := worker.Config{
		Nconfig:   workerNATSConfig,
		DonorUUID: donorUUID,
	}
	consumer, err := worker.New(workerConfig)
	if err != nil {
		log.Fatal(err)
	}
	//go log.Fatal(server.Start(stopChan))
	go func() {
		server.StartMutate(stopChan)
	}()

	go func() {
		server.StartValidate(stopChan)
	}()

	go func() {
		consumer.Start(stopChan)
	}()

	<-stopChan
}

func init() {
	flag.StringVar(&tlsKeyPath, "tlsKeyPath", "/etc/certs/tls.key", "Absolute path to the TLS key")
	flag.StringVar(&tlsCertPath, "tlsCertPath", "/etc/certs/tls.crt", "Absolute path to the TLS certificate")

	flag.Parse()
	// Initialize Viper
	viper.SetConfigType("env") // Use environment variables
	viper.AutomaticEnv()       // Automatically read environment variables
}

func getENVValue(envKey string) string {
	// Read environment variables
	value := viper.GetString(envKey)
	if value == "" {
		message := fmt.Sprintf("%s environment variable is not set", envKey)
		log.Fatal(message)
	}
	return value
}
