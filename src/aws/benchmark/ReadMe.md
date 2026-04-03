# AWS Benchmark

## Benchmarks

### 1. Running the `geobft.py` Benchmark

This script launches GeoBFT benchmarks with various cluster and peer configurations.

**How to run:**
```bash
python3 geobft.py
```

**What it does:**
- Creates a timestamped output folder under `out/geobft/`.
- Iterates over predefined numbers of clusters and peers.
- For each configuration, runs the benchmark by invoking `aws-spread-az.py` with appropriate arguments 
- Output logs are saved as `modubft_peers_<peers>_clusters_<clusters>.txt` in the output folder.

**Key parameters (edit in script if needed):**
- `NUMBER_OF_PEERS`: List of peer counts to test.
- `NUMBER_OF_CLUSTER`: List of cluster counts to test.
- `VALLEN`: Value sizes to test.
- `INSTANCE_TYPE`, `CLIENT_INSTANCE_TYPE`: AWS instance types.
- `batch_sizes`: Batch sizes for clusters.

### 2. Running the `modubft.py` Benchmark

This script launches ModuBFT benchmarks with various peer configurations.

**How to run:**
```bash
python3 modubft.py
```

**What it does:**
- Creates a timestamped output folder under `out/modubft/`.
- Iterates over predefined numbers of peers.
- For each configuration, runs the benchmark by invoking `aws-spread-az.py` with appropriate arguments (see script for details).
- Output logs are saved as `modubft_peers_<peers>.txt` in the output folder.

**Key parameters (edit in script if needed):**
- `NUMBER_OF_PEERS`: List of peer counts to test.
- `VALLEN`: Value sizes to test.
- `INSTANCE_TYPE`, `CLIENT_INSTANCE_TYPE`: AWS instance types.
- `USE_BUCKETS`: Whether to use buckets 
---

## Notes

- Both benchmark scripts require Python 3 and AWS credentials/configuration as needed by your deployment scripts.
- Adjust parameters in the scripts to match your experimental setup.
- Output folders and logs are created automatically for each run.
