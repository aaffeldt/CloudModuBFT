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
    copy_cli_binary("./")

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
    
    
    print("Cluster:", cluster)
    print("Roles:", roles)
    print("Buckets:", buckets)
    
    


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

    for i in range(number_of_cluster):
        for _ in range(number_of_buckets*number_of_clients_per_bucket):
            cluster.append(i % number_of_cluster)

    return ",".join(map(str, cluster))


if __name__ == "__main__":
    main()
    