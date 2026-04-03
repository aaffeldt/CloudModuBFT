# zs/benchmark

This folder contains scripts and configuration files for running ModuBFT and GeoBFT benchmarks in your environment.

## Folder Contents

- **modubft.py**: Script to run ModuBFT benchmarks 
- **geobft.py**: Script to run GeoBFT benchmarks 
- **modubft.json**: Example machine configuration for ModuBFT runs.
- **geobft.json**: Example machine configuration for GeoBFT runs.

---

## Running the Benchmarks

### 1. Running the GeoBFT Benchmark (`geobft.py`)

This script launches GeoBFT benchmarks with various cluster and peer configurations.

**How to run:**
```bash
python3 geobft.py
```

**What it does:**
- Creates a timestamped output folder under `out/modubft/`.
- Copies the `modubft.json` machine file into the output folder.
- Iterates over the cluster and peer configurations defined in the script.
- For each configuration, runs the benchmark by invoking `zs.py` with the appropriate arguments.
- Output logs are saved as `modubft_peers_<peers>_clusters_<clusters>.log` in the output folder.

**Key parameters (edit in script if needed):**
- `NUMBER_OF_PEERS`: List of peer counts to test.
- `NUMBER_OF_CLUSTER`: List of cluster counts to test.
- `VALLEN`: Value sizes to test.
- `BATCH_SIZES`: Batch sizes for clusters.
- `NUMBER_OF_CLIENTS`: Number of clients.
- `CLUSTER_PREPARE_TO_ALL_AND_ENABLE_HASH`: non-optimistic global sharing
, `CLUSTER_PREPARE_ALTERNATE`: alternating cluster prepare 

---

### 2. Running the ModuBFT Benchmark (`modubft.py`)

This script launches ModuBFT benchmarks with various peer configurations

**How to run:**
```bash
python3 modubft.py
```

**What it does:**
- Creates a timestamped output folder under `out/modubft/`.
- Copies the `modubft.json` machine file into the output folder.
- Iterates over the peer configurations defined in the script.
- For each configuration, runs the benchmark by invoking `zs.py` with the appropriate arguments.
- Output logs are saved as `modubft_peers_<peers>.txt` in the output folder.

**Key parameters (edit in script if needed):**
- `NUMBER_OF_PEERS`: List of peer counts to test.
- `VALLEN`: Value sizes to test.
- `NUMBER_OF_CLIENTS`: Number of clients.


