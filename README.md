# CloudModuBFT: A Modular BFT Framework for Cloud-Native Consensus

CloudModuBFT is a high-performance, research-oriented Byzantine Fault Tolerance (BFT) framework implemented in Go. It is specifically engineered to evaluate consensus protocols in geo-distributed and cloud-native environments (AWS). The project features a modular architecture that separates the consensus logic from the underlying network and infrastructure, allowing for rapid experimentation with different topologies, batching strategies, and cryptographic optimizations.

---

## 🚀 Key Features

- **Multi-Cluster Support:** Deploy nodes across multiple clusters to simulate geo-distributed environments or edge computing scenarios.
- **Dynamic Role Assignment:** Nodes can be configured as **Leaders**, **Peers**, or **Clients** via command-line flags.
- **Infrastructure as Code (IaC):** Automated provisioning on AWS using Terraform, with support for **Spread Placement Groups** and **Partition Groups** to ensure hardware isolation.
- **Advanced Optimizations:**
    - **Certificate Compression:** Uses Gzip compression for consensus certificates to reduce bandwidth usage.
    - **Batching:** Configurable cluster-level batch sizes to optimize throughput.
    - **Crypto Flexibility:** Support for various signature verification and creation strategies (e.g., Public Key, MACs).
- **Integrated Benchmarking:** Automated collection of latency, throughput, and system resource metrics (CPU, Memory) using `dool` and `iperf`.

---

## 🏗 Detailed Architecture

The system is built on a modular foundation, allowing for flexible configuration of node behaviors and network topologies.

### 1. Core Consensus Engine (`src/modubft`)
The heart of the system is the Go-based BFT implementation.
- **Node State Machine (`node/node.go`):** Manages the lifecycle of a consensus node, including message handling (Prepare, Commit, etc.), checkpointing, and view changes.
- **Client Implementation (`node/client.go`):** A sophisticated request generator that:
    - Pre-generates signed requests to minimize latency impact.
    - Uses parallel verification for responses from peers.
    - Measures throughput and latency with warmup and cooldown periods.
- **Identification & Topology (`node/identifier.go`):** Defines the `PeerIdentifier` and `ClientIdentifier` structures. It handles the mapping of nodes to clusters, buckets, and roles.
- **Communication Layer (`node/communication.go`):** Manages TCP connections between nodes and handles the serialization/deserialization of Protobuf messages.
- **Protocol Constants (`node/const.go`):** Defines critical parameters such as `CHECKPOINT_INTERVAL` (default 500), `CIRCULAR_BUFFER_SIZE` (default 2000), and communication ports.

### 2. Orchestration & Deployment (`src/zs`)
A comprehensive toolset for managing large-scale experiments:
- **`zs.sh`:** The main entry point for developers. It:
    - Cross-compiles the Go binary for Linux (`GOOS=linux GOARCH=amd64`).
    - Distributes the binary and configuration files to remote AWS instances via SCP.
    - Orchestrates the execution of nodes and clients using SSH.
- **`controller.py`:** A Python-based coordinator that synchronizes the start of all nodes across the cluster and manages the duration of the experiment.
- **Utility Scripts:**
    - `build-modubft.sh`: Handles the Go build process.
    - `generate-list-of-cluster.py`: Dynamically creates cluster mapping configurations.
    - `download-benchmark-results.sh`: Automatically retrieves log and benchmark files from remote nodes after an experiment finishes.

### 3. AWS Infrastructure (`src/aws`)
Terraform-based provisioning for reproducible cloud environments:
- **Placement Strategies:** Uses `aws_placement_group` with `strategy = "spread"` or `strategy = "partition"` to ensure nodes are placed on distinct hardware racks, minimizing correlated failures.
- **Networking:** Automated VPC and Subnet creation with Security Groups tailored for BFT communication.
- **Provisioning Scenarios:**
    - `one-availability-zone-using-partitions/`: Optimized for low-latency testing within a single AZ while maintaining hardware isolation.
    - `spread-across-availability-zone/`: Simulates higher latency environments by spreading nodes across multiple AWS AZs.

### 4. Evaluation & Metrics (`src/evaluation`)
Post-experiment data analysis:
- **Jupyter Notebooks:** Provide visual insights into performance characteristics (e.g., `0-bandwidth.ipynb`, `05-mulitleader.ipynb`).
- **Data Parsing (`utils/benchmark_utils.py`):** Python library for cleaning and processing the JSON output from nodes, including integration with `dool` resource data.

---

## 🛠 Prerequisites

- **Go:** 1.25 or later.
- **Terraform:** For infrastructure management.
- **Python 3.x:** With `pandas`, `matplotlib`, and `seaborn` for data analysis.
- **AWS CLI:** Configured with credentials for the target region.
- **Dool:** (Optional) Installed on target nodes for detailed resource monitoring.

---

## ⚙ Command-Line Interface (CLI) Flags

The `modubft` binary supports a wide range of configuration flags for fine-tuning:

| Flag | Description | Default |
|------|-------------|---------|
| `--id` | Unique integer ID for the node | `-1` |
| `--nodes` | Comma-separated list of IP addresses for all nodes | `""` |
| `--roles` | Comma-separated roles (2: Leader, 1: Peer, 0: Client) | `""` |
| `--clusters` | Comma-separated cluster IDs for each node | `""` |
| `--vallen` | Size of the payload in each request (bytes) | `512` |
| `--cluster_batch_size` | Number of requests to batch for cluster operations | `300` |
| `--runtime` | Total duration of the experiment in seconds | `10` |
| `--use_certificate_compression` | Enable Gzip compression for certificates | `false` |
| `--use_pk_to_client` | Use public keys for client-to-leader communication | `false` |
| `--profile` | Enable CPU and Heap profiling | `false` |
| `--dont_verify_cluster_commit` | Disable crypto verification (for baseline testing) | `false` |

---

## 🔐 Post-Cloning Setup (Required)

This repository has been sanitized for public use. Before running experiments, you must replace the following placeholders with your actual infrastructure details:

### 1. Network Configuration
In `src/zs/zs.sh` and `src/zs/machines.json`, replace the following placeholders with your EC2 instance IPs:
- `<REPLACE_WITH_SSH_IPS>`: Public/Elastic IPs used for SSH access.
- `<REPLACE_WITH_PRIVATE_IPS>`: Internal VPC IPs used for inter-node BFT communication.
- `0.0.0.0`: Replace these entries in `machines.json` with your cluster's IP list.

### 2. Cryptographic Keys
For the consensus protocol to function, you must provide valid RSA private keys:
- **Files:** `src/modubft/crypto/signer.go` and `src/github.com/zistvan/crypto/signer.go`
- **Action:** Replace `REPLACE_WITH_PRIVATE_KEY` within the `LoadPrivateKey` function with your actual PEM-encoded RSA private key.

### 3. AWS SSH Key
- **Location:** `src/aws/aws_ec2_key`
- **Action:** Place your private SSH key in this file (it is ignored by Git via `.gitignore`) to allow the orchestration scripts to communicate with your AWS instances.

---

## 🏃 Workflow Example

### 1. Provision the Cluster
```bash
cd src/aws/infrastructure/one-availability-zone-using-partitions
terraform init
terraform apply -var="number_of_nodes=6" -var="number_of_clients=2"
```

### 2. Run the Benchmark
Use `zs.sh` to build, deploy, and execute:
```bash
cd src/zs
# Syntax: ./zs.sh [clusters] [peers] [val_len] [batch_size] [clients] [out_folder] [runtime]
./zs.sh 2 3 1024 128 2 ./bench_results 30
```

### 3. Review Performance
After completion, check the `bench_results` directory for `benchmark_*.json` files. Use the Jupyter notebooks in `src/evaluation/notebooks` to plot the results.

---

## 📜 Development & Contribution

- **Communication:** Node messages are defined using Protocol Buffers in `src/modubft/proto/`.
- **Benchmarking:** Every node exports its performance data to a JSON file upon completion.
- **Profiling:** If `--profile` is enabled, `.prof` files are generated in `/tmp/` on the target nodes.

