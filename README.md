# hiro-workload-donor

## Overview

The `hiro-workload-donor` project is designed to mark Kubernetes pods as donors, allowing a stealer cluster to identify and steal these pods. This project includes a mutating admission webhook that modifies pod specifications to indicate they are eligible for workload stealing.

# hiro-workload-donor

## Overview

The `hiro-workload-donor` project is designed to mark Kubernetes pods as donors, allowing a stealer cluster to identify and steal these pods. This project includes a mutating admission webhook that modifies pod specifications to indicate they are eligible for workload stealing.

## Installation Steps

1. **Clone the repository**:
    ```sh
    git clone https://github.com/HIRO-MicroDataCenters-BV/hiro-workload-donor.git
    cd hiro-workload-donor
    chmod +x scripts/*
    ```

2. **Run the `start_donor.sh` script**:
    The `start_donor.sh` script initializes the Kind cluster, builds and installs the application, and sets up the necessary configurations to mark the pods as donors.
    ```sh
    ./start_donor.sh
    ```

Once the above component is installed, test it by running `scripts/test_pod_steal.sh`
```sh
cd script
./test_pod_steal.sh
```

