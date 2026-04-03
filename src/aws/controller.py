import subprocess

import argparse
import os
import sys
import threading
# username=$1
# total_container=$2
# ssh_ips=$3
# ips=$4
# roles=$5
# cluster=$6
# dont_verify_cluster_commit=$7
# use_pk_to_client=$8
# profile=$9
# runtime=${10}
# vallen=${11}
# cluster_batch_size=${12}
# use_certificate_compression=${13}
# use_dool=${14}
# alternateClusterPrepare=${15}
# buckets=${16}


def main():
    parser = argparse.ArgumentParser(
        description="Generate list of buckets for nodes based on their roles and clusters.")
    parser.add_argument("--username", type=str, help="Username")
    parser.add_argument("--total_container", type=str,
                        help="Total number of containers")
    parser.add_argument("--ssh_ips", type=str,
                        help="Comma-separated list of SSH IPs")
    parser.add_argument("--ips", type=str, help="Comma-separated list of IPs")
    parser.add_argument("--roles", type=str,
                        help="Comma-separated list of roles")
    parser.add_argument("--cluster", type=str,
                        help="Comma-separated list of clusters")
    parser.add_argument("--dont_verify_cluster_commit",
                        type=str, help="Flag to not verify cluster commit")
    parser.add_argument("--use_pk_to_client", type=str,
                        help="Flag to use PK to client")
    parser.add_argument("--profile", type=str, help="Profile name")
    parser.add_argument("--runtime", type=str, help="Runtime value")
    parser.add_argument("--vallen", type=str, help="Value length")
    parser.add_argument("--cluster_batch_size", type=str,
                        help="Cluster batch size")
    parser.add_argument("--use_certificate_compression",
                        type=str, help="Flag to use certificate compression")
    parser.add_argument("--use_dool", type=str, help="Flag to use dool")
    parser.add_argument("--alternateClusterPrepare", type=str,
                        help="Flag for alternate cluster prepare")
    parser.add_argument("--buckets", type=str,
                        help="Comma-separated list of buckets")

    parser.add_argument("--send_transaction", type=str, default="true")
    parser.add_argument("--cluster_prepare_to_all_and_enable_hash",
                        type=str, default="false")

    args = parser.parse_args()

    print(args)

    setup_dool_environment()

    shutdown_cli_services(args)
    # # exit()
    setup_node_environment(args)

    # # rm /tmp/aws_*.log
    # # rm /tmp/aws_*.json

    subprocess.run(["rm", "-f", "/tmp/aws_*.log"])
    subprocess.run(["rm", "-f", "/tmp/benchmark*.json"])
    subprocess.run(["rm", "-f", "/tmp/aws_*.json"])
    subprocess.run(["rm", "-f", "/tmp/*.prof"])

    run_on_nodes(args, " ".join(["rm", "-f", "/tmp/aws_*.log"]))
    run_on_nodes(args, " ".join(["rm", "-f", "~/benchmark*.json"]))

    run_experiments_on_nodes(args)
    transfer_files_from_nodes(args, "~/benchmark_*.json",
                              "/tmp/", recursive=False)

    transfer_files_from_nodes(args, "/tmp/aws_*.log", "/tmp/", recursive=False)
    transfer_files_from_nodes(args, "/tmp/*.prof", "/tmp/")
    transfer_files_from_nodes(args, "/tmp/*.prof", "/tmp/",add_idx=True)


def run_experiments_on_nodes(args):
    total_container = int(args.total_container)
    ips_list = args.ips.split(",")
    ssh_ips_list = args.ssh_ips.split(",")

    cmds = {ssh_ip: "\n" for ssh_ip in ssh_ips_list}

    for i in range(total_container):
        idx_ip = ips_list[i].strip()
        idx_ssh_ip = ssh_ips_list[i].strip()
        print(f"Node {i} IP: {idx_ip}")
        #     ssh -o StrictHostKeyChecking=no "$username"@"$idx_ssh_ip" "run_exp -m 'modubft' -n 0 --  ./cli  --id $i --buckets $buckets --alternateClusterPrepare $alternateClusterPrepare  --nodes $ips   --roles $roles   --clusters $cluster   --dont_verify_cluster_commit $dont_verify_cluster_commit   --use_pk_to_client $use_pk_to_client   --profile $profile   --runtime $runtime   --vallen $vallen   --cluster_batch_size $cluster_batch_size --use_dool $use_dool --use_certificate_compression $use_certificate_compression" > "/tmp/aws_$i.log" 2>"/tmp/aws_$i.err" &

        cmd = f"./cli --send_transaction '{args.send_transaction}' --cluster_prepare_to_all_and_enable_hash '{args.cluster_prepare_to_all_and_enable_hash}' --id {i} --buckets {args.buckets} --alternateClusterPrepare {args.alternateClusterPrepare}  --nodes {args.ips}   --roles {args.roles}   --clusters {args.cluster}   --dont_verify_cluster_commit '{args.dont_verify_cluster_commit}'   --use_pk_to_client '{args.use_pk_to_client}'   --profile '{args.profile}'   --runtime {args.runtime}   --vallen {args.vallen}   --cluster_batch_size {args.cluster_batch_size} --use_dool '{args.use_dool}' --use_certificate_compression '{args.use_certificate_compression}' 2>&1 | tee -a  /tmp/aws_{i}.log &"
        # accumulate commands for the same ssh_ip
        cmds[idx_ssh_ip] += cmd + "\n"

    pids = []
    for idx_ssh_ip, cmd in cmds.items():
        cmds[idx_ssh_ip] += "\nwait"
        print(f"Executing on {idx_ssh_ip}: {cmds[idx_ssh_ip]}", flush=True)
        ssh_command = f"ssh -i {os.path.expanduser("~/ssh_key")}  -o StrictHostKeyChecking=no {args.username}@{idx_ssh_ip} 'bash -s' <<EOF\n{cmds[idx_ssh_ip]}\nEOF"
        proc = subprocess.Popen(ssh_command,
                                stdout=subprocess.PIPE, stderr=subprocess.PIPE, shell=True, text=True, bufsize=1)

        pids.append(proc)

    wait_for_process_completion(pids)

    print("All processes completed.", flush=True)


def setup_node_environment(args):
    transfer_files_to_nodes(args, "~/cli", "~/")
    transfer_files_to_nodes(args, "/tmp/dool", "~/", recursive=True)
    for cmd in ["chmod +x ~/dool/bin/dool",
                # "chmod +x ~/iperf3",
                "sed -i '1s|^|export PATH=$PATH:~/dool/bin\\n|' ~/.bashrc",
                "sed -i '1s|^|export PATH=$PATH:~/\\n|' ~/.bashrc",
                "source ~/.bashrc"]:
        run_on_nodes(args, cmd)


def run_on_nodes(args, cmd):
    pids = []
    total_container = int(args.total_container)
    ssh_ips_list = args.ssh_ips.split(",")
    username = args.username

    done_ssh_ips = set()

    for i in range(total_container):
        idx_ip = ssh_ips_list[i].strip()
        if idx_ip in done_ssh_ips:
            continue

        print(f"Node {i} IP: {idx_ip}")

        ssh_key_path = os.path.expanduser("~/ssh_key")
        proc = subprocess.Popen([
            "ssh", "-i", ssh_key_path,
            "-o", "StrictHostKeyChecking=no",
            f"{username}@{idx_ip}", cmd
        ])
        pids.append(proc)
        done_ssh_ips.add(idx_ip)
    wait_for_process_completion(pids)


def transfer_files_to_nodes(args, file_to_copy_src, file_to_copy_dst, recursive=False):
    pids = []
    total_container = int(args.total_container)
    ssh_ips_list = args.ssh_ips.split(",")
    username = args.username

    done_ssh_ips = set()
    print(
        f"Transferring {file_to_copy_src} to {file_to_copy_dst} on all nodes...")

    for i in range(total_container):
        idx_ip = ssh_ips_list[i].strip()
        if idx_ip in done_ssh_ips:
            continue

        print(f"Node {i} IP: {idx_ip}")

        proc_args = [
            "scp"]
        if recursive:
            proc_args.append("-r")
        proc_args.extend([
            "-o", "StrictHostKeyChecking=no", "-i", os.path.expanduser(
                "~/ssh_key"),
            os.path.expanduser(file_to_copy_src),
            f"{username}@{idx_ip}:{file_to_copy_dst}"
        ])
        proc = subprocess.Popen(proc_args)
        pids.append(proc)
        done_ssh_ips.add(idx_ip)
    wait_for_process_completion(pids)


def transfer_files_from_nodes(args, file_to_copy_src, file_to_copy_dst, recursive=False,add_idx=False):
    pids = []
    total_container = int(args.total_container)
    ssh_ips_list = args.ssh_ips.split(",")
    username = args.username

    done_ssh_ips = set()
    print(
        f"Transferring {file_to_copy_src} from nodes to {file_to_copy_dst}...")

    for i in range(total_container):
        idx_ip = ssh_ips_list[i].strip()
        if idx_ip in done_ssh_ips:
            continue

        print(f"Node {i} IP: {idx_ip}")

        proc_args = [
            "scp"]
        if recursive:
            proc_args.append("-r")
        
        if add_idx:
            proc_args.extend([
                "-o", "StrictHostKeyChecking=no", "-i", os.path.expanduser(
                    "~/ssh_key"),
                f"{username}@{idx_ip}:{file_to_copy_src}",
                os.path.expanduser(f"{file_to_copy_dst}cpu_{i}.prof")
            ])
        else: 
            proc_args.extend([
                "-o", "StrictHostKeyChecking=no", "-i", os.path.expanduser(
                    "~/ssh_key"),
                f"{username}@{idx_ip}:{file_to_copy_src}",
                os.path.expanduser(file_to_copy_dst)
            ])
        proc = subprocess.Popen(proc_args)

        pids.append(proc)
        done_ssh_ips.add(idx_ip)
    wait_for_process_completion(pids)


def shutdown_cli_services(args):
    pids = []
    total_container = int(args.total_container)
    ssh_ips_list = args.ssh_ips.split(",")
    username = args.username

    done_ssh_ips = set()

    for i in range(total_container):

        idx_ip = ssh_ips_list[i].strip()
        if idx_ip in done_ssh_ips:
            continue
        print(f"Node {i} IP: {idx_ip}")

        ssh_key_path = os.path.expanduser("~/ssh_key")
        proc = subprocess.Popen([
            "ssh", "-i", ssh_key_path,
            "-o", "StrictHostKeyChecking=no",
            f"{username}@{idx_ip}",
            "killall cli"
        ])
        pids.append(proc)
        done_ssh_ips.add(idx_ip)

    wait_for_process_completion(pids)


def stream_output(proc, key):
    if proc.stdout is not None:
        for line in proc.stdout:
            print(f"[{key}] {line.strip()}", flush=True)

    if proc.stderr is not None:
        for line in proc.stderr:
            print(f"[{key} ERROR] {line.strip()}", flush=True)


def wait_for_process_completion(pids):
    threads = []
    for proc in (pids):
        t = threading.Thread(target=stream_output, args=(proc, proc.pid))
        t.daemon = True
        t.start()
        threads.append(t)

    for proc in pids:
        print(f"Waiting for process {proc.pid} to complete...")

        proc.wait()

        if proc.returncode != 0:

            print(
                f"Process {proc.pid} failed with return code {proc.returncode}.")

            print(f"Process {proc.pid} completed.")


def setup_dool_environment():
    # Create /tmp/dool/bin directory if it doesn't exist
    os.makedirs("/tmp/dool/bin", exist_ok=True)

    # Download dool script
    subprocess.run([
        "curl",
        "https://raw.githubusercontent.com/scottchiefbaker/dool/refs/heads/next/dool?ref=v1.3.4",
        "-o",
        "/tmp/dool/bin/dool"
    ], check=True)

    # Make the dool script executable
    subprocess.run([
        "chmod",
        "+x",
        "/tmp/dool/bin/dool"
    ], check=True)

    # Add /tmp/dool/bin to PATH for the current process
    os.environ["PATH"] = f"/tmp/dool/bin:{os.environ['PATH']}"


if __name__ == "__main__":
    main()
