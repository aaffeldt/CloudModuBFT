# simple.py

## Overview

`simple.py` is a command-line tool designed to help automate the setup and configuration of ModuBFT clusters for benchmarking and consensus experiments. It generates the necessary cluster, role, and bucket mappings, builds the ModuBFT CLI binary, and prepares the environment for deployment.


## Features

- **Cluster Topology Generation**:  
    Computes cluster, role, and bucket assignments for nodes and clients.
- **Automated Build**:  
    Builds the ModuBFT CLI binary for Linux (amd64) using Go.
- **Binary Distribution**:  
    Copies the built binary to the working directory.
- **Flexible Configuration**:  
    Supports various cluster sizes, client counts, batch sizes, and other experiment parameters via command-line arguments.

## Usage

```bash
python3 simple.py --basefolder <output_folder> [options]
```

### Key Arguments

- `--number_of_clusters`: Number of clusters (default: 1)
- `--number_of_peers`: Number of peers per cluster (default: 3)
- `--number_of_clients`: Number of clients per leader (default: 1)
- `--number_of_buckets`: Number of buckets per cluster (default: 1)
- `--vallen_arr`: Comma-separated value sizes (default: "512,4096")
- `--cluster_batch_size_arr`: Comma-separated batch sizes (default: "1,4,16,64,128,256")
- `--basefolder`: **(Required)** Output folder for experiment results
- `--runtime`: Experiment runtime in seconds (default: 10)
- `--use_dool`: Enable/disable resource monitoring (default: "true")
- `--main_file`: Path to the ModuBFT main Go file (default: "../modubft/main.go")
- ...and more (see `--help` for all options)

---

## Example

```bash
python3 simple.py --basefolder ./results --number_of_clusters 2 --number_of_peers 4 --number_of_clients 2
```

---

## Output

- Prints generated cluster, role, and bucket mappings.
- Builds the CLI binary and copies it to the working directory.

---

## Functions

- `build_binary(main_file)`: Builds the Go binary for ModuBFT.
- `copy_cli_binary(folder)`: Copies the built binary to the specified folder.
- `generate_cluster_mapping(...)`: Generates cluster assignments.
- `generate_role_mapping(...)`: Assigns roles (leader, peer, client).
- `generate_bucket_mapping(...)`: Assigns buckets to nodes/clients.

---

## Requirements

- Python 3.x
- Go toolchain (for building the binary)
- ModuBFT source code

---

## See Also

- [`simple.sh`](./simple.sh): Bash script for Docker-based cluster setup.
- [`compose.yaml`](./compose.yaml): Docker Compose configuration for multi-node deployment.

---

## License

See the main project repository for licensing details.
