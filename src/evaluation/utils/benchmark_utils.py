import pandas as pd
import json
import os
import numpy as np
import re
from glob import glob

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

def derive_benchmark_parameters(cluster_list, role_list):
    """
    Derives benchmark parameters from cluster and role lists.

    Args:
        cluster_list (str): Comma-separated cluster IDs.
        role_list (str): Comma-separated node roles (0=Client, 1=Follower, 2=Leader).

    Returns:
        tuple: (cluster_list, role_list, number_of_clusters, peers_per_cluster, partition)
    """
    cluster_list = cluster_list.split(",")
    cluster_list = [int(x) for x in cluster_list]

    role_list = role_list.split(",")
    role_list = [int(x) for x in role_list]

    number_of_clusters = max(cluster_list) + 1
    number_of_peers = int(np.count_nonzero(np.array(role_list) == 1))
    number_of_clients = int(np.count_nonzero(np.array(role_list) == 0))

    partition = []
    for c in range(1, number_of_clusters + 1):
        partition.append(c)
    for c in range(1, number_of_clusters + 1):
        for i in range(0, number_of_peers):
            partition.append((c + i) % (number_of_peers + 1) + 1)
    for c in range(1, number_of_clusters + 1):
        partition.append(c)
    return cluster_list, role_list, number_of_clusters, (number_of_peers + number_of_clients) / number_of_clusters, partition

def apply_derive_benchmark_parameters(row):
    """
    Applies derive_benchmark_parameters to a DataFrame row.

    Args:
        row (pd.Series): DataFrame row containing 'Setup.ClusterListS' and 'Setup.NodeRoleS'.

    Returns:
        pd.Series: Derived benchmark parameters.
    """
    cluster_list, role_list, number_of_clusters, number_of_peers, partition = derive_benchmark_parameters(
        row["Setup.ClusterListS"], row["Setup.NodeRoleS"]
    )
    return pd.Series({
        "cluster_list": cluster_list,
        "role_list": role_list,
        "number_of_clusters": number_of_clusters,
        "number_of_peers": number_of_peers,
        "partition": partition
    })

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
def clean_dool_data(df):
    df["Dool"] =df["Dool"].apply(clean_dool_output)
    return df 
    
def extract_max_cpu_usage(dool):
    """
    Extracts the maximum CPU usage from cleaned 'dool' output.

    Args:
        dool (str): Cleaned output string from 'dool'.

    Returns:
        int: Maximum CPU usage value.
    """
    cpu = []
    for line in dool.split("\n")[5:]:
        usr = (' '.join(line.split())).split(" ")[0]
        cpu.append(int(usr))
    return max(cpu)

def BytesSentTo(row):
    """
    Processes the 'BytesSentTo' field in a benchmark row, annotating with cluster and role info.

    Args:
        row (pd.Series): DataFrame row containing benchmark data.

    Returns:
        pd.DataFrame or None: DataFrame with annotated bytes sent info, or None if node is a client.
    """
    my_id = row["Setup.MyId"]

    dic = {0: "Client", 1: "Follower", 2: "Leader"}
    map_idx_to_role = {i: dic[int(role)] for (i, role) in enumerate(row["Setup.NodeRoleS"].split(","))}
    map_idx_to_cluster = {i: int(role) for (i, role) in enumerate(row["Setup.ClusterListS"].split(","))}
    if map_idx_to_role[my_id] == "Client":
        return

    f = pd.DataFrame(row["BytesSentTo"])

    f["Cluster"] = f["CID"].apply(lambda i: "intra" if map_idx_to_cluster[i] == map_idx_to_cluster[my_id] else "inter")
    f["ToRole"] = f["CID"].apply(lambda i: map_idx_to_role[i])

    f[["number_of_clusters", "number_of_peers", "Setup.Vallen", "Setup.ClusterBatchSize", "MyRole"]] = row[
        ["number_of_clusters", "number_of_peers", "Setup.Vallen", "Setup.ClusterBatchSize", "MyRole"]]

    return f

def aggregate_benchmark_results(evaluate_benchmarks, folder):
    """
    Aggregates benchmark results from a folder using a provided evaluation function.

    Args:
        evaluate_benchmarks (callable): Function to evaluate individual benchmark files.
        folder (str): Path to the folder containing benchmark files.

    Returns:
        tuple: Aggregated DataFrames (dfs, throughputs, bandwidths, max_cpu_usages, bytes_sent_summaries).
    """
    setup_cols = ['Setup.RunTime', 'Setup.NodeListS', 'Setup.NodeRoleS',
                  'Setup.ClusterListS', 'Setup.Profile', 'Setup.Vallen',
                  'Setup.UsePkToClient', 'Setup.DontVerifyClusterCommit',
                  'Setup.ClusterBatchSize', 'Setup.UseCertificateCompression',
                  'Setup.UseDool', 'number_of_clusters',
                  'number_of_peers', "Total Peers"]
    node_cols = setup_cols + ['Setup.MyId', 'MyPartition', 'MyRole']

    throughputs = []
    bandwidths = []
    max_cpu_usages = []
    bytes_sent_summaries = []
    dfs = []
    for f in os.listdir(folder):
        try:
            df, cumulative_throughput_role_zero, bytes_sent_summary, max_cpu_usage_by_role, mean_bandwidth_by_role = evaluate_benchmarks(folder + f)
            throughputs.append(cumulative_throughput_role_zero)
            bandwidths.append(mean_bandwidth_by_role)
            max_cpu_usages.append(max_cpu_usage_by_role)
            bytes_sent_summaries.append(bytes_sent_summary)
            dfs.append(df)
        except Exception as e:
            print(e)

    dfs = pd.concat(dfs).drop_duplicates(node_cols, ignore_index=True)
    throughputs = pd.concat(throughputs).drop_duplicates(setup_cols, ignore_index=True)
    max_cpu_usages = pd.concat(max_cpu_usages).drop_duplicates(setup_cols + ["MyRole"], ignore_index=True)
    bytes_sent_summaries = pd.concat(bytes_sent_summaries)
    bandwidths = pd.concat(bandwidths).drop_duplicates(setup_cols + ["MyRole"], ignore_index=True)
    return dfs, throughputs, bandwidths, max_cpu_usages, bytes_sent_summaries

def load_json_benchmark_dfs(folder):
    """
    Loads and concatenates all JSON benchmark files in a folder into a single DataFrame.

    Args:
        folder (str): Path to the folder containing JSON files.

    Returns:
        pd.DataFrame: Concatenated benchmark data.
    """
    
    dfs = [load_and_normalize_benchmark_data(json_file) for json_file in glob(folder + "*.json")]
    df = pd.concat(dfs).reset_index()
    df["folder"] = folder
    return df

def merge_benchmark_results(folders):
    """
    Merges benchmark results from multiple folders into a single DataFrame.

    Args:
        folders (list): List of folder paths.
        load_json_benchmark_dfs (callable): Function to load benchmark DataFrames from a folder.

    Returns:
        pd.DataFrame: Merged benchmark data.
    """
    df = pd.concat([load_json_benchmark_dfs(folder) for folder in folders]).reset_index()
    return df

def assign_cluster_and_role(frame):
    df = frame.copy()
    df["IDtoRole"] = df.apply(lambda r : { i: role for i,role in enumerate(r["Setup.NodeRoleS"].split(","))}  , axis=1)
    df["Cluster"]= df.apply(lambda r : r["Setup.ClusterListS"].split(",")[r["Setup.MyId"]]  , axis=1).astype(int)
    df["Role"] = df.apply(lambda r : r["Setup.NodeRoleS"].split(",")[r["Setup.MyId"]]  , axis=1).astype(int)
    return df 