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


INFRASTRUCTURE_FOLDER = os.path.realpath(
    "infrastructure/one-availability-zone-using-partitions")


class Infrastructure():

    def __init__(self,
                 region,
                 availability_zone_full,
                 public_key_path,
                 number_of_partitions,
                 total_number_of_nodes,
                 number_of_clients,
                 node_partitions_list,
                 client_partitions_list,
                 ami,
                 instance_type,
                 client_instance_type):
        self.region = region
        self.availability_zone_full = availability_zone_full
        self.public_key_path = public_key_path
        self.number_of_partitions = number_of_partitions
        self.total_number_of_nodes = total_number_of_nodes
        self.number_of_clients = number_of_clients
        self.node_partitions_list = node_partitions_list
        self.client_partitions_list = client_partitions_list
        self.ami = ami
        self.instance_type = instance_type
        self.client_instance_type = client_instance_type

    def __enter__(self):
        cmd = ["bash", "-c",
               f"pushd \"{INFRASTRUCTURE_FOLDER}\" && terraform init && popd"]
        subprocess.run(cmd, check=True)
        cmd = [
            "bash", "-c",
            f'pushd "{INFRASTRUCTURE_FOLDER}" && terraform plan -out "plan" -var=region="{self.region}" -var=availability_zone="{self.availability_zone_full}" -var=public_key_path="{self.public_key_path}" -var=number_of_partitions="{self.number_of_partitions}" -var=number_of_nodes="{self.total_number_of_nodes}" -var=number_of_clients="{self.number_of_clients}" -var=list_of_node_partitions=[{self.node_partitions_list}] -var=list_of_client_partitions=[{self.client_partitions_list}] -var=ami="{self.ami}" -var=instance_type="{self.instance_type}" -var=client_instance_type="{self.client_instance_type}" && popd'
        ]
        subprocess.run(cmd, check=True, stdout=subprocess.DEVNULL)

        cmd = ["bash", "-c",
               f"pushd \"{INFRASTRUCTURE_FOLDER}\" && terraform apply   \"plan\" && popd"]
        subprocess.run(cmd, check=True)

        controller_ip = subprocess.check_output(
            ["bash", "-c",
             f'pushd "{INFRASTRUCTURE_FOLDER}" > /dev/null && terraform output -raw controller_ip && popd > /dev/null'],
            text=True
        ).strip()
        peer_machines = subprocess.check_output(
            ["bash", "-c",
             f'pushd "{INFRASTRUCTURE_FOLDER}" > /dev/null && terraform output -raw node_ip_addresses && popd > /dev/null'],
            text=True
        ).strip()
        client_machines = subprocess.check_output(
            ["bash", "-c",
             f'pushd "{INFRASTRUCTURE_FOLDER}" > /dev/null && terraform output -raw client_ip_addresses && popd > /dev/null'],
            text=True
        ).strip()

        self.controller_ip = controller_ip
        self.peer_machines = peer_machines.split(",")
        self.client_machines = client_machines.split(",")
        return self

    def __exit__(self, exc_type, exc_value, traceback):
        cmd = [
            "bash", "-c",
            f'pushd "{INFRASTRUCTURE_FOLDER}" && terraform plan -destroy -out destroyplan  -var=region="{self.region}" -var=availability_zone="{self.availability_zone_full}" -var=public_key_path="{self.public_key_path}" -var=number_of_partitions="{self.number_of_partitions}" -var=number_of_nodes="{self.total_number_of_nodes}" -var=number_of_clients="{self.number_of_clients}" -var=list_of_node_partitions=[{self.node_partitions_list}] -var=list_of_client_partitions=[{self.client_partitions_list}] -var=ami="{self.ami}" -var=instance_type="{self.instance_type}" -var=client_instance_type="{self.client_instance_type}"  && terraform apply   destroyplan && popd'
        ]
        subprocess.run(cmd, check=True, stdout=subprocess.DEVNULL)
        print("Infrastructure destroyed.")


def copy_setup_to_controller(controller_ip, ssh_key, username="ubuntu"):

    # Remove controller_ip from known_hosts
    known_hosts = os.path.expanduser("~/.ssh/known_hosts")
    subprocess.run([
        "ssh-keygen", "-f", known_hosts, "-R", controller_ip
    ], check=True)

    # Copy /tmp/cli to controller
    if not os.path.isfile(ssh_key):
        raise FileNotFoundError(f"SSH key not found: {ssh_key}")

    def _run_scp(args_list):
        res = subprocess.run(args_list, stdout=subprocess.PIPE,
                             stderr=subprocess.PIPE, text=True)
        if res.returncode != 0:
            print("\n--- SCP command failed ---")
            print("Command:", " ".join(args_list))
            print("Return code:", res.returncode)
            if res.stdout:
                print("STDOUT:\n", res.stdout)
            if res.stderr:
                print("STDERR:\n", res.stderr)
            raise CalledProcessError(
                res.returncode, args_list, output=res.stdout, stderr=res.stderr)
        return res

    _run_scp([
        "scp", "-o", "StrictHostKeyChecking=no", "-i", ssh_key,
        "/tmp/cli", f"{username}@{controller_ip}:~/cli"
    ])

    # Copy controller.py to controller
    _run_scp([
        "scp", "-o", "StrictHostKeyChecking=no", "-i", ssh_key,
        "./controller.py", f"{username}@{controller_ip}:~/controller.py"
    ])

    # Copy ssh_key to controller
    _run_scp([
        "scp", "-o", "StrictHostKeyChecking=no", "-i", ssh_key,
        ssh_key, f"{username}@{controller_ip}:~/ssh_key"
    ])


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
    parser.add_argument("--username", type=str, default="ubuntu")
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
    parser.add_argument("--max_partitions", type=int, default=6)

    parser.add_argument("--region", type=str, default="eu-central-1")
    parser.add_argument("--availability_zone", type=str, default="a")
    parser.add_argument("--ami", type=str, default="ami-0a116fa7c861dd5f9")
    parser.add_argument("--instance_type", type=str, default="t2.micro")
    parser.add_argument("--client_instance_type",
                        type=str, default="t2.micro")

    parser.add_argument("--send_transaction", type=str, default="true")
    parser.add_argument("--cluster_prepare_to_all_and_enable_hash",
                        type=str, default="false")

    args = parser.parse_args()

    build_binary(main_file=args.main_file)
    number_of_nodes = int(args.number_of_clusters) * int(args.number_of_peers)
    number_of_clients_per_leader = int(args.number_of_clients)
    real_client_count = int(args.number_of_buckets) * \
        int(args.number_of_clusters)
    number_of_clients = number_of_clients_per_leader * real_client_count

    number_of_machines = number_of_nodes + number_of_clients

    vallen_arr = list(map(int, args.vallen_arr.split(",")))
    cluster_batch_size_arr = list(
        map(int, args.cluster_batch_size_arr.split(",")))

    cluster = generate_cluster_mapping(int(args.number_of_clusters), int(args.number_of_buckets), int(
        args.number_of_peers), number_of_clients_per_leader)

    roles = generate_role_mapping(cluster.split(","), int(
        args.number_of_buckets), number_of_clients_per_leader)

    buckets = generate_bucket_mapping(cluster.split(","), list(
        map(int, roles.split(","))), number_of_clients_per_leader)

    partions = generate_partition_mapping(cluster.split(","), list(
        map(int, roles.split(","))), list(map(int, buckets.split(","))), number_of_clients_per_leader, args.max_partitions)

    number_of_partitions = len(set(partions.split(",")))

    node_partitions_list = partions.split(",")[:number_of_nodes]
    client_partitions_list = partions.split(",")[number_of_nodes:]
    print(f"Cluster: {cluster}")
    print(f"Roles: {roles}")
    print(f"Buckets: {buckets}")
    print(f"Partitions: {partions}")
    print(f"Number of partitions: {number_of_partitions}")
    print(f"Node partitions list: {node_partitions_list}")
    print(f"Client partitions list: {client_partitions_list}")

    assert_validity(args, number_of_nodes, number_of_clients, roles,
                    partions, node_partitions_list, client_partitions_list)

    absolute_public_key_path = os.path.realpath(args.public_key_path)
    absolute_private_key_path = os.path.realpath(args.private_key_path)
    assert os.path.isfile(absolute_public_key_path)
    assert os.path.isfile(absolute_private_key_path)

    with Infrastructure(
        region=args.region,
        availability_zone_full=f"{args.region}{args.availability_zone}",
        public_key_path=absolute_public_key_path,
        number_of_partitions=number_of_partitions,
        total_number_of_nodes=number_of_nodes,
        number_of_clients=real_client_count,
        node_partitions_list=",".join(
            node_partitions_list),  # all nodes in all partitions
        client_partitions_list=",".join(
            client_partitions_list),  # all clients in all partitions
        ami=args.ami,
        instance_type=args.instance_type,
        client_instance_type=args.client_instance_type
    ) as infra:
        print("Infrastructure created.")
        controller_ip = infra.controller_ip
        peer_machines = infra.peer_machines
        client_machines = infra.client_machines

        client_machines = expand_client_machines(
            number_of_clients_per_leader, client_machines)

        ips = ",".join(peer_machines)+"," + ",".join(client_machines)
        ssh_ips = ips

        print(f"Controller IP: {controller_ip}")
        print(f"Peer machines: {peer_machines}")
        print(f"Client machines: {client_machines}")
        print(f"SSH IPs: {ssh_ips}")

        assert len(peer_machines) == number_of_nodes
        assert len(client_machines) == number_of_clients
        assert len(ssh_ips.split(",")) == number_of_machines

        copy_setup_to_controller(controller_ip, absolute_private_key_path,
                                 username=args.username)

        for cluster_batch_size in cluster_batch_size_arr:
            for vallen in vallen_arr:
                print(
                    f"Running experiment with cluster_batch_size={cluster_batch_size}, vallen={vallen}")

                timestamp, folder = create_experiment_folder(
                    args, cluster_batch_size, vallen)

                print(f"Experiment folder: {folder}")
                print(f"Timestamp: {timestamp}")

                run_experiment_on_controller(args, number_of_machines, cluster, roles, buckets,
                                             absolute_private_key_path, controller_ip, ips, ssh_ips, cluster_batch_size, vallen)

                to_copy = [
                    "/tmp/aws_*.log",
                    "/tmp/aws_*.err",
                    "/tmp/aws_*.json",
                    "/tmp/benchmark*.json",
                    "/tmp/*.prof"
                ]
                for file in to_copy:
                    subprocess.run([
                        "scp", "-o", "StrictHostKeyChecking=no", "-i", absolute_private_key_path,
                        f"{args.username}@{controller_ip}:{file}", folder
                    ], check=False)

                print(folder, flush=True)
                extract_throughput_from_logs(folder)
                save_dool_results(folder)


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
        print(line, flush=True)


def save_dool_results(folder):
    with open(f"{folder}/dool_output.txt", "w") as f:
        f.write(str(clean_dool_data(assign_cluster_and_role(
            load_json_benchmark_dfs(folder+"/"))).Dool.to_frame().to_string()))


def create_experiment_folder(args, cluster_batch_size, vallen):
    timestamp = subprocess.check_output(
        ["date", "+%Y%m%d%H%M%S"], text=True).strip()
    folder = os.path.join(
        args.basefolder, f"{timestamp}-cb{cluster_batch_size}-v{vallen}")
    os.makedirs(folder, exist_ok=True)
    copy_cli_binary(folder)
    return timestamp, folder


def run_experiment_on_controller(args, number_of_machines, cluster, roles, buckets, absolute_private_key_path, controller_ip, ips, ssh_ips, cluster_batch_size, vallen):
    controller_args = [
        "python3", "controller.py",
        "--username", args.username,
        "--total_container", str(number_of_machines),
        "--ssh_ips", ssh_ips,
        "--ips", ips,
        "--roles", roles,
        "--cluster", cluster,
        "--dont_verify_cluster_commit", args.dont_verify_cluster_commit,
        "--use_pk_to_client", args.use_pk_to_client,
        "--profile", args.profile,
        "--runtime", args.runtime,
        "--vallen", str(vallen),
        "--cluster_batch_size", str(cluster_batch_size),
        "--use_certificate_compression", args.use_certificate_compression,
        "--use_dool", args.use_dool,
        "--alternateClusterPrepare", args.alternateClusterPrepare,
        "--buckets", buckets
    ]

    subprocess.run(
        ["ssh", "-o", "StrictHostKeyChecking=no", "-i", absolute_private_key_path,
         f"{args.username}@{controller_ip}"] + controller_args,
        check=True
    )


def copy_cli_binary(folder):
    cli_src = "/tmp/cli"
    cli_dst = os.path.join(folder, "cli")
    cp_result = subprocess.run(
        ["cp", cli_src, cli_dst], stderr=PIPE, text=True)
    if cp_result.returncode != 0:
        print(cp_result.stderr)
        raise RuntimeError("Failed to copy binary")


def expand_client_machines(number_of_clients_per_leader, client_machines):
    client_machines_expanded = []
    for c in client_machines:
        for _ in range(number_of_clients_per_leader):
            client_machines_expanded.append(c)
    return client_machines_expanded


def assert_validity(args, number_of_nodes, number_of_clients, roles, partions, node_partitions_list, client_partitions_list):
    assert len(node_partitions_list) == number_of_nodes
    assert len(
        client_partitions_list) == roles.count("2")
    assert len(set(node_partitions_list)) <= args.max_partitions
    assert len(set(client_partitions_list)) <= args.max_partitions
    assert len(set(partions.split(","))) <= args.max_partitions
    assert len(set(roles.split(","))) == 3
    assert roles.count("2") == int(args.number_of_clusters) * \
        int(args.number_of_buckets)
    assert roles.count("0") == number_of_clients
    assert roles.count("1") == number_of_nodes - \
        (int(args.number_of_clusters)*int(args.number_of_buckets))


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


def generate_partition_mapping(list_of_cluster, list_of_roles, list_of_buckets, number_of_clients_per_bucket, max_partitions):
    partition_count = {}
    partition_leader_follower = []

    compact_client_partitions_list = {}

    for r, c in zip(list_of_roles, list_of_cluster):
        if c not in partition_count:
            partition_count[c] = 1

        if r == 2:
            partition_leader_follower.append(
                partition_count[c] % max_partitions)
            partition_count[c] += 1
        elif r == 1:
            partition_leader_follower.append(
                partition_count[c] % max_partitions)
            partition_count[c] += 1
        else:
            partition_leader_follower.append(-1)

    partition = []
    for p, r, c, b in zip(partition_leader_follower, list_of_roles, list_of_cluster, list_of_buckets):
        if p == -1:

            for fr, fc, fb, fp in zip(list_of_roles, list_of_cluster, list_of_buckets, partition_leader_follower):
                if fr == 2 and fc == c and fb == b:
                    if (p, r, c, b) not in compact_client_partitions_list:
                        compact_client_partitions_list[(p, r, c, b)] = fp
                        partition.append(fp)
                    break
        else:
            partition.append(p)

    return ",".join(map(str, partition))


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
