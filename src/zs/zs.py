import argparse

import os
import re
import subprocess
from subprocess import CalledProcessError, Popen, PIPE
import json
import pandas as pd
import json
import os
import re
from glob import glob
# python3 zs.py --number_of_clusters 2 --number_of_peers 3 --vallen_arr "512" --cluster_batch_size_arr "64" --number_of_buckets 1 --number_of_clients 5 --basefolder ../../client_out


def main():
    parser = argparse.ArgumentParser()

    parser.add_argument("--number_of_clusters", type=str, default="1")
    parser.add_argument("--number_of_clients", type=str, default="1")
    parser.add_argument("--number_of_peers", type=str, default="3")
    parser.add_argument("--vallen_arr", type=str, default="512,4096")
    parser.add_argument("--cluster_batch_size_arr",
                        type=str, default="1,4,16,64,128,256")
    parser.add_argument("--number_of_buckets", type=str, default="1")
    parser.add_argument("--basefolder", type=str, required=True)
    parser.add_argument("--runtime", type=str, default="10")
    parser.add_argument("--username", type=str, default="aaffeldt")
    parser.add_argument("--dont_verify_cluster_commit",
                        type=str, default="false")
    parser.add_argument("--use_pk_to_client", type=str, default="false")
    parser.add_argument("--profile", type=str, default="false")
    parser.add_argument("--use_certificate_compression",
                        type=str, default="false")
    parser.add_argument("--use_dool", type=str, default="true")
    parser.add_argument("--alternateClusterPrepare", type=str, default="false")
    parser.add_argument("--public_key_path", type=str,
                        default="./aws_ec2_key.pub")
    parser.add_argument("--private_key_path", type=str,
                        default="./aws_ec2_key")
    parser.add_argument("--machines", type=str, default="machines.json")
    parser.add_argument("--main_file", type=str,
                        default="../modubft/main.go")

    parser.add_argument("--send_transaction", type=str, default="true")
    parser.add_argument("--cluster_prepare_to_all_and_enable_hash",
                        type=str, default="false")

    args = parser.parse_args()

    build_binary(main_file=args.main_file)
    number_of_nodes = int(args.number_of_clusters) * int(args.number_of_peers)
    number_of_clients_per_leader = int(args.number_of_clients)
    number_of_clients = number_of_clients_per_leader * \
        int(args.number_of_buckets) * int(args.number_of_clusters)

    number_of_machines = number_of_nodes + number_of_clients

    # public_key_path = os.path.realpath(args.public_key_path)
    # private_key_path = os.path.realpath(args.private_key_path)

    vallen_arr = list(map(int, args.vallen_arr.split(",")))
    cluster_batch_size_arr = list(
        map(int, args.cluster_batch_size_arr.split(",")))

    cluster = generate_cluster_mapping(int(args.number_of_clusters), int(args.number_of_buckets), int(
        args.number_of_peers), number_of_clients_per_leader)

    roles = generate_role_mapping(cluster.split(","), int(
        args.number_of_buckets), number_of_clients_per_leader)

    buckets = generate_bucket_mapping(cluster.split(","), list(
        map(int, roles.split(","))), number_of_clients_per_leader)

    print(f"Cluster: {cluster}")
    print(f"Roles: {roles}")
    print(f"Buckets: {buckets}")

    machines = read_machine_definitions(args)

    client_machines_str, client_ssh_machines_str = assemble_client_machines(
        machines, number_of_clients)
    print(f"Client machines: {client_machines_str}")
    print(f"Client ssh machines: {client_ssh_machines_str}")

    peer_machines_str, peer_ssh_machines_str = extract_peer_machines(
        machines, number_of_nodes)

    print(f"Peer machines: {peer_machines_str}")
    print(f"Peer ssh machines: {peer_ssh_machines_str}")

    ips = f"{peer_machines_str},{client_machines_str}"
    ssh_ips = f"{peer_ssh_machines_str},{client_ssh_machines_str}"
    print(f"All machines: {ips}")

    for cluster_batch_size in cluster_batch_size_arr:
        for vallen in vallen_arr:
            print(
                f"Running experiment with cluster_batch_size={cluster_batch_size}, vallen={vallen}")

            folder = execute_experiment(
                args, number_of_machines, cluster, roles, buckets, ips, ssh_ips, cluster_batch_size, vallen)

            print(folder)
            extract_throughput_from_logs(folder)
            save_dool_results(folder)


def save_dool_results(folder):
    with open(f"{folder}/dool_output.txt", "w") as f:
        f.write((clean_dool_data(assign_cluster_and_role(
            load_json_benchmark_dfs(folder+"/"))).Dool.to_frame().to_string()))


def execute_experiment(args, number_of_machines, cluster, roles, buckets, ips, ssh_ips, cluster_batch_size, vallen):
    timestamp = get_current_timestamp()
    folder = create_timestamped_folder(args, timestamp)

    copy_cli_binary(folder)

    com = (
        f"python3 controller.py "
        f"--username {args.username} "
        f"--total_container {number_of_machines} "
        f"--ssh_ips {ssh_ips} "
        f"--ips {ips} "
        f"--roles {roles} "
        f"--cluster {cluster} "
        f"--dont_verify_cluster_commit {args.dont_verify_cluster_commit} "
        f"--use_pk_to_client {args.use_pk_to_client} "
        f"--profile {args.profile} "
        f"--runtime {args.runtime} "
        f"--vallen {vallen} "
        f"--cluster_batch_size {cluster_batch_size} "
        f"--use_certificate_compression {args.use_certificate_compression} "
        f"--use_dool {args.use_dool} "
        f"--alternateClusterPrepare {args.alternateClusterPrepare} "
        f"--buckets {buckets} "
        f"--send_transaction {args.send_transaction} "
        f"--cluster_prepare_to_all_and_enable_hash {args.cluster_prepare_to_all_and_enable_hash} "
    )
    print(f"Executing: {com}")
    execute_command(com)

    move_log_files_to_folder(folder, [
        "/tmp/zs_*.log",
        "/tmp/zs_*.err",
        "/tmp/zs_*.json",
        "/tmp/benchmark*.json",
        "/tmp/*.prof"
    ])

    return folder


def execute_command(com):
    with Popen(
            com, shell=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE, bufsize=1, universal_newlines=True) as p:
        for line in p.stdout:
            print(line, end='')  # process line here
        if p.returncode != 0:
            print(p.stderr.read())
            # raise CalledProcessError(p.returncode, com)


def extract_throughput_from_logs(folder):
    throughput_lines = []
    log_files = [f for f in os.listdir(folder) if f.endswith(".log")]
    for log_file in log_files:
        log_path = os.path.join(folder, log_file)
        with open(log_path, "r") as f:
            for line in f:
                if "Throughput" in line:
                    throughput_lines.append(line.strip())
    for line in throughput_lines:
        print(line)


def move_log_files_to_folder(folder, log_files):
    for pattern in log_files:
        mv_cmd = f'mv {pattern} "{folder}/"'
        print(f"Moving files with: {mv_cmd}")
        subprocess.run(mv_cmd, shell=True)


def copy_cli_binary(folder):
    cli_src = "/tmp/cli"
    cli_dst = os.path.join(folder, "cli")
    cp_result = subprocess.run(
        ["cp", cli_src, cli_dst], stderr=PIPE, text=True)
    if cp_result.returncode != 0:
        print(cp_result.stderr)
        raise RuntimeError("Failed to copy binary")


def create_timestamped_folder(args, timestamp):
    folder = os.path.join(args.basefolder, timestamp)
    os.makedirs(folder, exist_ok=True)
    return folder


def get_current_timestamp():
    return Popen(["date", "+%Y%m%d_%H%M%S"],
                 stdout=PIPE).communicate()[0].decode().strip()


def build_binary(main_file=" ../modubft/main.go"):
    main_file_folder = os.path.dirname(main_file)
    main_file_basename = os.path.basename(main_file)
    binary = "/tmp/cli"

    build_cmd = [
        "bash", "-c",
        f'pushd "{main_file_folder}" || exit 1; '
        'CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go version; '
        f'echo "Building {main_file_basename} to {binary}"; '
        f'CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o "{binary}" "{main_file_basename}" || exit 1; '
        'popd || exit 1; '
        f'echo "{binary}"'
    ]

    result = subprocess.run(
        build_cmd, stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True)
    print(result.stdout)
    if result.returncode != 0:
        print(result.stderr)
        raise RuntimeError("Failed to build binary")


def extract_peer_machines(machines, number_of_nodes):

    peer_machines = machines.get("peer_machines", [])
    peer_ssh_machines = machines.get("peer_ssh_machines", [])

    assert len(peer_machines) > 0 and len(
        peer_ssh_machines) > 0, "Peer machines and ssh peer machines cannot be empty"
    assert len(
        peer_machines) >= number_of_nodes, f"Not enough peer machines in machines.json, required {number_of_nodes}, found {len(peer_machines)}"
    assert len(
        peer_ssh_machines) >= number_of_nodes, f"Not enough peer ssh machines in machines.json, required {number_of_nodes}, found {len(peer_ssh_machines)}"

    peer_machines = peer_machines[:number_of_nodes]
    peer_ssh_machines = peer_ssh_machines[:number_of_nodes]

    assert len(peer_machines) == len(
        peer_ssh_machines), "Peer machines and ssh peer machines must have the same length"

    return ",".join(peer_machines), ",".join(peer_ssh_machines)


def read_machine_definitions(args):
    with open(args.machines, "r") as f:
        machines = json.load(f)
    return machines


def assemble_client_machines(machines, number_of_clients):

    client_machines = []
    client_ssh_machines = []

    split_client_machine = machines.get("client_machine", [])
    split_ssh_client_machine = machines.get("client_ssh_machine", [])

    print(f"Client machines from json: {split_client_machine}")
    print(f"Client ssh machines from json: {split_ssh_client_machine}")
    number_of_client_machines = len(split_client_machine)
    number_of_ssh_client_machines = len(split_ssh_client_machine)

    assert number_of_client_machines > 0 and number_of_ssh_client_machines > 0, "Client machines and ssh client machines cannot be empty"

    for i in range(number_of_clients):
        machine = split_client_machine[i % number_of_client_machines]
        ssh_machine = split_ssh_client_machine[i %
                                               number_of_ssh_client_machines]
        client_machines.append(machine)
        client_ssh_machines.append(ssh_machine)

    client_machines_str = ",".join(client_machines)
    client_ssh_machines_str = ",".join(client_ssh_machines)
    return client_machines_str, client_ssh_machines_str


def generate_bucket_mapping(list_of_cluster, list_of_roles, number_of_clients_per_bucket):

    leader_bucket_count = {}
    client_bucket_count = {}
    buckets = []
    for c, r in zip(list_of_cluster, list_of_roles):

        if c not in leader_bucket_count:
            leader_bucket_count[c] = 0

        if c not in client_bucket_count:
            client_bucket_count[c] = 0

        if r == 2:
            buckets.append(leader_bucket_count[c])
            leader_bucket_count[c] += 1
        elif r == 1:
            buckets.append(-1)
        else:
            buckets.append(
                client_bucket_count[c]//number_of_clients_per_bucket)
            client_bucket_count[c] += 1

    return ",".join(map(str, buckets))


def generate_role_mapping(list_of_cluster, number_of_buckets, number_of_clients_per_bucket):

    client_count = {}
    follower_count = {}

    number_of_cluster = len(set(list_of_cluster))
    number_of_nodes = len(list_of_cluster)

    number_of_clients = (
        number_of_cluster*number_of_buckets*number_of_clients_per_bucket)
    number_of_leader = (number_of_cluster*number_of_buckets)

    number_of_follower_per_cluster = (
        number_of_nodes - number_of_clients - number_of_leader)/number_of_cluster

    # print(number_of_follower_per_cluster)

    list_of_roles = []
    for c in reversed(list_of_cluster):
        if c not in client_count:
            client_count[c] = 0
        if c not in follower_count:
            follower_count[c] = 0

        if client_count[c] < number_of_buckets*number_of_clients_per_bucket:
            list_of_roles.append(0)
            client_count[c] += 1
        elif follower_count[c] < number_of_follower_per_cluster:
            list_of_roles.append(1)
            follower_count[c] += 1
        else:
            list_of_roles.append(2)

    return ",".join(map(str, reversed(list_of_roles)))


def generate_cluster_mapping(number_of_cluster, number_of_buckets, number_of_peers_in_a_cluster, number_of_clients_per_bucket):
    cluster = []

    for i in range(number_of_cluster*number_of_peers_in_a_cluster):
        cluster.append(i % number_of_cluster)

    for i in range(number_of_buckets*number_of_cluster*number_of_clients_per_bucket):
        cluster.append(i % number_of_cluster)

    return ",".join(map(str, cluster))


def clean_dool_output(dool_str):
    """
    Cleans the output from the 'dool' tool by removing ANSI escape sequences and non-ASCII characters.

    Args:
        dool_str (str): Raw output string from 'dool'.

    Returns:
        str: Cleaned output string.
    """
    # Remove ANSI escape sequences
    ansi_escape = re.compile(r'\x1b\[[0-9;]*[a-zA-Z]')
    cleaned = ansi_escape.sub('', dool_str)
    # Replace non-ASCII characters with space
    cleaned = re.sub(r'[^\x00-\x7F]+', ' ', cleaned)
    return cleaned


def load_json_benchmark_dfs(folder):
    """
    Loads and concatenates all JSON benchmark files in a folder into a single DataFrame.

    Args:
        folder (str): Path to the folder containing JSON files.

    Returns:
        pd.DataFrame: Concatenated benchmark data.
    """

    dfs = [load_and_normalize_benchmark_data(
        json_file) for json_file in glob(folder + "*.json")]
    df = pd.concat(dfs).reset_index()
    df["folder"] = folder
    return df


def load_and_normalize_benchmark_data(file):
    """
    Loads benchmark data from a JSON file and normalizes it into a pandas DataFrame.

    Args:
        file (str): Path to the JSON file.

    Returns:
        pd.DataFrame: Normalized benchmark data.
    """
    data = json.load(open(file))
    df = pd.json_normalize(data)
    return df


def assign_cluster_and_role(frame):
    df = frame.copy()
    df["IDtoRole"] = df.apply(lambda r: {i: role for i, role in enumerate(
        r["Setup.NodeRoleS"].split(","))}, axis=1)
    df["Cluster"] = df.apply(lambda r: r["Setup.ClusterListS"].split(",")[
                             r["Setup.MyId"]], axis=1).astype(int)
    df["Role"] = df.apply(lambda r: r["Setup.NodeRoleS"].split(",")[
                          r["Setup.MyId"]], axis=1).astype(int)
    return df


def clean_dool_data(df):
    df["Dool"] = df["Dool"].apply(clean_dool_output)
    return df


if __name__ == "__main__":
    main()
