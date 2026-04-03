# zs Folder

This folder contains scripts and documentation for orchestrating and deploying ModuBFT nodes across multiple machines in a cloud-native environment. The main script, `zs.sh`, automates the setup, distribution, and execution of ModuBFT binaries for benchmarking and consensus experiments.

## Contents

- `zs.sh`: Bash script for building, distributing, and running ModuBFT nodes.
- `README.md`: This documentation file.

---

## zs.sh Overview

`zs.sh` is designed to:

1. **Build the ModuBFT CLI binary** for Linux (amd64) using Go.
2. **Distribute the binary** to remote machines via SSH and SCP.
3. **Configure node roles and clusters** (peers, clients, clusters).
4. **Start experiments** on each node using SSH, with logging.

### Key Features

- **Cluster and Node Configuration**:  
    - `number_of_clusters`, `number_of_peers`, `number_of_clients` variables control the topology.
    - IP addresses for machines and SSH endpoints are set via variables (`machines`, `ssh_machines`, etc.).
- **Automated Build**:  
    - Uses Go cross-compilation to build the CLI for Linux.
- **Remote Execution**:  
    - Kills previous CLI processes.
    - Copies the new binary to each node.
    - Runs the experiment using `run_exp` and logs output.
- **Role Assignment**:  
    - Assigns roles: `2` for leader, `1` for peer, `0` for client.
    - Constructs comma-separated lists for node IPs, roles, and cluster IDs.

---

## Usage

### Prerequisites

- Go installed locally.
- SSH access to all target machines.
- SCP enabled for file transfer.
- `run_exp` utility available on remote machines.

### Steps

1. **Edit IP Addresses**  
     Update the `machines`, `ssh_machines`, `client_machines`, and `client_ssh_machines` variables to match your infrastructure.

2. **Configure Cluster Topology**  
     Adjust `number_of_clusters`, `number_of_peers`, and `number_of_clients` as needed.

3. **Run the Script**  
     ```bash
     ./zs.sh
     ```

     The script will:
     - Build the CLI binary.
     - Kill any running CLI processes on remote nodes.
     - Copy the new binary to each node.
     - Start the experiment and log output to `/tmp/zs_*.log`.

---

## Script Details

### Variables

- `number_of_clusters`: Number of clusters to deploy.
- `number_of_peers`: Number of peer nodes per cluster.
- `number_of_clients`: Number of client nodes per cluster.
- `machines`, `ssh_machines`: Comma-separated IPs for peers.
- `client_machines`, `client_ssh_machines`: Comma-separated IPs for clients.

### Execution Flow

1. **Build CLI**  
     - Compiles `main.go` in the `modubft` directory to `/tmp/cli`.

2. **Node Preparation**  
     - Kills any running `cli` processes on all nodes.

3. **Binary Distribution**  
     - Copies `/tmp/cli` to each node's home directory.

4. **Experiment Launch**  
     - Runs `run_exp` with appropriate arguments for each node.
     - Logs output to `/tmp/zs_*.log`.

---

## Customization

- **Change Topology**:  
    Modify cluster and node counts at the top of `zs.sh`.
- **Update IPs**:  
    Edit IP variables to match your deployment.
- **Experiment Parameters**:  
    Adjust CLI flags and `run_exp` arguments as needed.

---

## Troubleshooting

- **SSH/SCP Issues**:  
    Ensure SSH keys are set up and machines are reachable.
- **Go Build Errors**:  
    Check Go installation and environment variables.
- **Remote Execution**:  
    Verify `run_exp` and CLI binary permissions on remote nodes.

---

## References

- [ModuBFT Documentation](../modubft/README.md) (if available)
- [Go Cross Compilation](https://golang.org/doc/install)
- [SSH/SCP Usage](https://linux.die.net/man/1/ssh)

---

## License

See the main project repository for licensing details.
