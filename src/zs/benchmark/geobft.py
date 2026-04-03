

import subprocess
import os



MACHINES = "modubft.json"
MACHINES = os.path.realpath(MACHINES)
NUMBER_OF_PEERS = [3]
NUMBER_OF_CLUSTER = [2]
VALLEN = [512, 4096]
BATCH_SIZES = [128]
NUMBER_OF_CLIENTS = 1
CLUSTER_PREPARE_TO_ALL_AND_ENABLE_HASH = "true"
CLUSTER_PREPARE_ALTERNATE ="false"

def main():

    print(f"Running on machines defined in: {MACHINES}")
    timestamp = subprocess.check_output(
        ["date", "+%Y%m%d-%H%M%S"]).decode().strip()

    basefolder = os.path.join(
        "out/modubft/", timestamp)
    basefolder = os.path.realpath(basefolder)

    print(f"Basefolder: {basefolder}")
    os.makedirs(basefolder, exist_ok=True)

    # copy machine file into basefolder
    subprocess.run(["cp", MACHINES, basefolder], check=True)

    for number_of_cluster in NUMBER_OF_CLUSTER:
        for number_of_peers in NUMBER_OF_PEERS:
            run_benchmark(basefolder, number_of_peers, number_of_cluster)


def run_benchmark(basefolder, number_of_peers, number_of_cluster):
    print(
        f"Starting benchmark with {number_of_peers} peers and {number_of_cluster} clusters...")

    output_file = generate_peer_output_filepath(
        basefolder, number_of_peers, number_of_cluster)

    script_dir = get_script_directory()
    print(f"Script dir: {script_dir}")
    cmd = create_script_execution_command(
        basefolder, number_of_peers, number_of_cluster, script_dir)

    print(f"Running command: {' '.join(cmd)}")
    run_subprocess_and_log(output_file, cmd)

    print(f"Benchmark with {number_of_peers} peers completed.")


def create_script_execution_command(basefolder, number_of_peers, number_of_cluster, script_dir):
    cmd = ["pushd", f"'{script_dir}'", "&&"] + generate_benchmark_command(
        basefolder, number_of_peers, number_of_cluster) + ["&&", "popd"]
    cmd = " ".join(map(str, cmd))
    cmd = ["bash", "-c", cmd]
    return cmd


def generate_benchmark_command(basefolder, number_of_peers, number_of_cluster):
    return ["python3", "zs.py",
            "--number_of_clusters", str(number_of_cluster),
            "--number_of_peers", str(
                number_of_peers),
            "--vallen_arr", f"'{",".join(
                map(str, VALLEN))}'",
            "--cluster_batch_size_arr", f"'{",".join(
                map(str, BATCH_SIZES))}'",
            "--number_of_buckets", str(2),
            "--number_of_clients", str(
                NUMBER_OF_CLIENTS),
            "--basefolder", f"'{basefolder}'",
            "--machines", f"'{MACHINES}'", 
            "--cluster_prepare_to_all_and_enable_hash", CLUSTER_PREPARE_TO_ALL_AND_ENABLE_HASH,
            "--alternateClusterPrepare",CLUSTER_PREPARE_ALTERNATE
            ]


def get_script_directory():
    return os.path.abspath(os.path.join(os.path.dirname(__file__), ".."))


def generate_peer_output_filepath(basefolder, number_of_peers, number_of_cluster):
    return os.path.realpath(os.path.join(
        basefolder, f"modubft_peers_{number_of_peers}_clusters_{number_of_cluster}.log"))


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
