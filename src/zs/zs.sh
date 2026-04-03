#!/bin/bash
set -e 

function print_help() {
    echo "Usage: $0 [number_of_clusters] [number_of_peers] [vallen_arr] [cluster_batch_size_arr] [number_of_clients] [basefolder] [runtime] [username] [dont_verify_cluster_commit] [use_pk_to_client] [profile] [use_certificate_compression] [use_dool]"
    echo
    echo "Arguments:"
    echo "  number_of_clusters           Number of clusters (default: 2)"
    echo "  number_of_peers              Number of peers per cluster (default: 3)"
    echo "  vallen_arr                   Comma-separated list of value lengths (default: 512,4096)"
    echo "  cluster_batch_size_arr       Comma-separated list of cluster batch sizes (default: 1,4,16,64,128,256)"
    echo "  number_of_clients            Number of clients (default: 1)"
    echo "  basefolder                   Base folder for output files"
    echo "  runtime                      Runtime duration in seconds (default: 10)"
    echo "  username                     SSH username for instances (default: ubuntu)"
    echo "  dont_verify_cluster_commit   Disable cluster commit verification (default: false)"
    echo "  use_pk_to_client             Use public key for client (default: false)"
    echo "  profile                      Enable profiling (default: false)"
    echo "  use_certificate_compression  Enable certificate compression (default: false)"
    echo "  use_dool                     Use dool (default: true)"
    echo
    echo "Example:"
    echo "  $0 2 3 512,4096 1,4,16,64,128,256 1 /path/to/basefolder 10 ubuntu false false false false true"
    exit 0
}

if [[ "$1" == "--help" || "$1" == "-h" ]]; then
    print_help
fi


number_of_clusters=${1:-2}
number_of_leader=$number_of_clusters
number_of_peers=${2:-3}
vallen_arr=${3:-"512,4096"}
cluster_batch_size_arr=${4:-"1,4,16,64,128,256"}
number_of_buckets=${5:-1}
basefolder=${6}
runtime=${7:-10}
# username=${7:-"ubuntu"}
# dont_verify_cluster_commit=${8:-"false"}
# use_pk_to_client=${9:-"false"}
# profile=${10:-"false"}
username=${8:-"aaffeldt"}
dont_verify_cluster_commit=${9:-"false"}
use_pk_to_client=${10:-"false"}
profile=${11:-"false"}
use_certificate_compression=${12:-"false"}
use_dool=${13:-"true"}
alternateClusterPrepare=${14:-"false"}
# number_of_buckets=2
number_of_clients=$((number_of_clusters*number_of_buckets))
number_of_clients=5

total_number_of_nodes=$((number_of_clusters * (number_of_peers)))
echo "Total number of nodes: $total_number_of_nodes"
total_number_of_machines=$((number_of_clusters * (number_of_peers) + number_of_clients))
echo "Total number of machines (nodes + clients): $total_number_of_machines"



public_key_path="./aws_ec2_key.pub"
private_key_path="./aws_ec2_key"

#get absolut path 
public_key_path=$(realpath "$public_key_path")
private_key_path=$(realpath "$private_key_path")

number_of_partitions=$(( 6 < ((total_number_of_nodes/number_of_clusters)) ? 6 : ((total_number_of_nodes/number_of_clusters)) ))
echo "Number of partitions: $number_of_partitions"
infrastructure_folder="infrastructure/one-availability-zone-using-partitions"

# split vallen_array by , 
IFS=',' read -ra vallen_array <<< "$vallen_arr"
IFS=',' read -ra cluster_batch_size_array <<< "$cluster_batch_size_arr"

echo "cluster_batch_size_array: ${cluster_batch_size_array[*]}"
echo "vallen_array: ${vallen_array[*]}"





if [ -f "zs.sh" ]
then
    echo "File is in current working directory"
else 
    echo "File is NOT in current working directory"
    exit 1 
fi



# check if basefolder set 
if [ -z "$basefolder" ]; then
    echo "basefolder not set, using default aws/out/real_bench"
    exit 1
fi


partition=$(bash ./utils/generate-partition-list.sh "$number_of_clusters" "$number_of_peers" "$number_of_clients" 6)
roles=$(bash ./utils/generate-list-of-roles.sh "$number_of_clusters" "$number_of_peers" "$number_of_clients")
cluster=$(bash ./utils/generate-list-of-cluster.sh "$number_of_clusters" "$number_of_peers" "$number_of_clients")
cluster=$(python3 ./utils/generate-list-of-cluster.py "$number_of_clusters" "$number_of_buckets" "$number_of_peers" "$number_of_clients")
echo "cluster: $cluster"
roles=$(bash ./utils/generate-list-of-roles.sh "$number_of_clusters" "$number_of_peers" "$number_of_clients")
roles=$(python3 ./utils/generate-list-of-roles.py "$cluster" "$number_of_buckets" "$number_of_clients")
echo "roles: $roles"
# exit 
partition=$(python3 ./utils/generate-partition-list.py "$cluster" "$roles" 6 )
buckets=$(python3 ./utils/generate-list-of-buckets.py "$cluster" "$roles" "$number_of_clients")

#split partition list by comma and get the four first entries 
partition_node=$(echo "$partition" | cut -d',' -f1-"$total_number_of_nodes")
partition_client=$(echo "$partition" | cut -d',' -f$((total_number_of_nodes + 1))-)




echo "roles: $roles"
echo "partition: $partition"
echo "partition_client: $partition_client"
echo "partition_node: $partition_node"
echo "cluster: $cluster"
echo "buckets: $buckets"

# exit 
 
rm /tmp/cli || true 
binary=$(bash ./utils/build-modubft.sh ./../modubft/main.go)

echo "binary: $binary"




peer_ssh_machines="<REPLACE_WITH_SSH_IPS>"
peer_machines="<REPLACE_WITH_PRIVATE_IPS>"
client_machine="<REPLACE_WITH_CLIENT_PRIVATE_IPS>"
client_ssh_machine="<REPLACE_WITH_CLIENT_SSH_IPS>"
split_client_machine=$(echo "$client_machine" | tr ',' '\n')
number_of_client_machines=$(echo "$client_machine" | tr ',' '\n' | wc -l)

split_ssh_client_machine=$(echo "$client_ssh_machine" | tr ',' '\n')
number_of_ssh_client_machines=$(echo "$client_ssh_machine" | tr ',' '\n' | wc -l)


for ((i=0; i<number_of_clients; i++)); do
    machine=$(echo "$split_client_machine" | sed -n $((i % number_of_client_machines + 1))p)
    ssh_machine=$(echo "$split_ssh_client_machine" | sed -n $((i % number_of_ssh_client_machines + 1))p)
    client_machines="$client_machines,$machine"
    client_ssh_machines="$client_ssh_machines,$ssh_machine"

done
#remove first comma of machiens 
client_machines=${client_machines:1}
client_ssh_machines=${client_ssh_machines:1}
echo "client_machines $client_machines"
# exit 

#get only the total_number_of_nodes from peer_machines 
peer_machines=$(echo "$peer_machines" | cut -d',' -f1-"$total_number_of_nodes")
peer_ssh_machines=$(echo "$peer_ssh_machines" | cut -d',' -f1-"$total_number_of_nodes")

client_machines=$(echo "$client_machines" | cut -d',' -f1-"$number_of_clients")
client_ssh_machines=$(echo "$client_ssh_machines" | cut -d',' -f1-"$number_of_clients")

echo "peer_machines: $peer_machines"
echo "client_machines: $client_machines"

ips="${peer_machines},${client_machines}"
ssh_ips="${peer_ssh_machines},${client_ssh_machines}"
echo "All IPs: $ips"










set +e 

for cluster_batch_size in "${cluster_batch_size_array[@]}"; do
    for vallen in "${vallen_array[@]}"; do
        
        #  if number of clusters is 1 only run with cluster_batch_size 1 
        # if [ "$number_of_clusters" -eq 1 ] && [ "$cluster_batch_size" -ne 1 ]; then
        #     echo "Skipping cluster_batch_size $cluster_batch_size for number_of_clusters 1"
        #     continue
        # fi
      


        timestamp=$(date +"%Y%m%d_%H%M%S")
        folder="${basefolder}/${timestamp}"
        mkdir -p "$folder"

        cp /tmp/cli "$folder/cli"

        
        # com="bash controller.sh $username $total_number_of_machines $ssh_ips $ips $roles $cluster $dont_verify_cluster_commit $use_pk_to_client $profile $runtime $vallen $cluster_batch_size $use_certificate_compression $use_dool $alternateClusterPrepare $buckets" 
        com="python3 controller.py --username $username --total_container $total_number_of_machines --ssh_ips $ssh_ips --ips $ips --roles $roles --cluster $cluster --dont_verify_cluster_commit $dont_verify_cluster_commit --use_pk_to_client $use_pk_to_client --profile $profile --runtime $runtime --vallen $vallen --cluster_batch_size $cluster_batch_size --use_certificate_compression $use_certificate_compression --use_dool $use_dool --alternateClusterPrepare $alternateClusterPrepare --buckets $buckets"
        echo "folder: $folder, com: $com"

        $com  2>&1 | tee -a  "$folder/command.txt"  

         

        bash ./utils/download-benchmark-results.sh "$folder"
        
        cat "$folder"/*.log | grep "Throughput"

        cpu=$(python3 -c "import sys; sys.path.append('../evaluation/utils'); from benchmark_utils import * ; print(clean_dool_data(assign_cluster_and_role(load_json_benchmark_dfs(\"$folder/\"))).Dool.to_frame().to_string())")
        # echo "cpu: $cpu"
        echo -e "$cpu" > "$folder/cpu.txt" 
        echo "folder: $folder"

        
         
    done
    
done



