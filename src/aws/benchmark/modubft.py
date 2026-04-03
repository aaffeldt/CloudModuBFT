

import subprocess
import os

import numpy as np

#      python3 aws-spread-az.py - -number_of_clusters 1 - -number_of_peers 3 - -vallen_arr "512" - -cluster_batch_s


# ize_arr "1" - -number_of_buckets 1 - -number_of_clients 2 - -basefolder ../.
# ./client_out - -instance_type "c5a.xlarge" - -client_instance_type "c5a.xla
# rge"  2>&1 | tee  "out.txt"

NUMBER_OF_PEERS = [3, 6, 12, 18, 21]
VALLEN = [512, 4096]
NUMBER_OF_CLIENTS = 1
INSTANCE_TYPE = "c5a.large"
CLIENT_INSTANCE_TYPE = "c5a.large"
USE_BUCKETS=True

def main():
    timestamp = subprocess.check_output(
        ["date", "+%Y%m%d-%H%M%S"]).decode().strip()

    basefolder = os.path.join(
        "out/modubft/", timestamp)
    basefolder = os.path.realpath(basefolder)

    print(f"Basefolder: {basefolder}")
    os.makedirs(basefolder, exist_ok=True)

    for number_of_peers in (NUMBER_OF_PEERS):
        run_benchmark(basefolder, number_of_peers)
        


def run_benchmark(basefolder, number_of_peers):
    print(f"Starting benchmark with {number_of_peers} peers.")
    output_file = generate_peer_output_filepath(basefolder, number_of_peers)

    script_dir = get_script_directory()
    print(f"Script dir: {script_dir}")
    cmd = create_script_execution_command(
        basefolder, number_of_peers, script_dir)

    print(f"Running command: {' '.join(cmd)}")
    run_subprocess_and_log(output_file, cmd)

    print(f"Benchmark with {number_of_peers} peers completed.")


def create_script_execution_command(basefolder, number_of_peers, script_dir):
    cmd = ["pushd", f"'{script_dir}'", "&&"] + \
        generate_benchmark_command(
            basefolder, number_of_peers) + ["&&", "popd"]
    cmd = " ".join(map(str, cmd))
    cmd = ["bash", "-c", cmd]
    return cmd


def generate_benchmark_command(basefolder, number_of_peers):
    
    f = int(np.floor((number_of_peers - 1)/3))
    buckets = number_of_peers - f 
    if not USE_BUCKETS:
        buckets = 1
    else:
        print(f"Using buckets: {buckets} for {number_of_peers} peers with f={f}")    
    
    return ["python3", "aws-spread-az.py",
            "--number_of_clusters", str(1),
            "--number_of_peers", str(
                number_of_peers),
            "--vallen_arr", f"'{",".join(
                map(str, VALLEN))}'",
            "--cluster_batch_size_arr", str(64),
            "--number_of_buckets", str(buckets),
            "--number_of_clients", str(
                NUMBER_OF_CLIENTS),
            "--basefolder", f"'{basefolder}'",
            "--instance_type", INSTANCE_TYPE,
            "--client_instance_type", CLIENT_INSTANCE_TYPE
            ]


def get_script_directory():
    return os.path.abspath(os.path.join(os.path.dirname(__file__), ".."))


def generate_peer_output_filepath(basefolder, number_of_peers):
    return os.path.realpath(os.path.join(
        basefolder, f"modubft_peers_{number_of_peers}.txt"))


def run_subprocess_and_log(output_file, cmd):
    proc = subprocess.Popen(cmd, stdout=subprocess.PIPE,
                            stderr=subprocess.PIPE, text=True, bufsize=1)
    with open(output_file, "w") as f:
        if proc.stdout is not None:
            for line in proc.stdout:
                print(line, end="")
                f.write(line)
                f.flush()
        if proc.stderr is not None:
            for line in proc.stderr:
                print(line, end="")
                f.write(line)
                f.flush()
    print("Waiting for process to complete...")
    proc.wait()
    if proc.returncode != 0:
        raise Exception(
            f" Command failed with {proc.returncode}")


if __name__ == "__main__":
    main()
