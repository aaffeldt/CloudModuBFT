#!/bin/bash

number_of_clusters=1
number_of_peers=3
number_of_clients=1
total_nodes_per_cluster=$((number_of_peers +  number_of_clients ))
total_container=$((total_nodes_per_cluster * number_of_clusters ))
dont_verify_cluster_commit=false
username="aaffeldt"

dont_verify_cluster_commit=false
profile=false
use_pk_to_client=false
runtime=10
vallen=4096
cluster_batch_size=64
use_certificate_compression="false"
use_dool="true"


basefolder="zs/out/dool_and_iperf"
timestamp=$(date +"%Y%m%d_%H%M%S")

folder="${basefolder}/${timestamp}_clusters${number_of_clusters}_peers${number_of_peers}_clients${number_of_clients}_nodes${total_nodes_per_cluster}_dontverify${dont_verify_cluster_commit}_profile${profile}_pkclient${use_pk_to_client}_runtime${runtime}_vallen${vallen}_batch${cluster_batch_size}"
mkdir -p "$folder"

# --peer_machines "10.10.2.183,10.10.2.184,10.10.2.185,10.10.2.186,10.10.2.187,10.10.2.188" --peer_ssh_machines "10.0.2.83,10.0.2.84,10.0.2.85,10.0.2.86,10.0.2.87,10.0.2.88" \
# --client_machines "10.10.2.61,10.10.2.62" --client_ssh_machines "10.0.2.61,10.0.2.62"
ssh_machines="10.0.2.81,10.0.2.87,10.0.2.84,10.0.2.87,10.0.2.88"
machines="10.10.2.181,10.10.2.187,10.10.2.184,10.10.2.187,10.10.2.188"
client_machines="10.10.2.62,10.10.2.63"
client_ssh_machines="10.0.2.62,10.0.2.63"

current_dir=$(pwd)

CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go version

cd "modubft" || exit
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /tmp/cli main.go || exit 1
cd "$current_dir" || exit






roles=""
ips=""
ssh_ips=""
cluster=""

for c in $(seq 0 $((number_of_clusters -1 ))); do
    idx=$((c * total_nodes_per_cluster ))
    
    ip=$(echo $machines | cut -d',' -f $((idx + 1)))
    ips="$ips,$ip"
    
    ssh_ip=$(echo $ssh_machines | cut -d',' -f $((idx + 1)))
    ssh_ips="$ssh_ips,$ssh_ip"
    
    
    
    
    roles="$roles,2"
    cluster="$cluster,$c"
    
    
    for i in $(seq 1 $((number_of_peers -1 ))); do
        idx=$((c * total_nodes_per_cluster + i ))
        ip=$(echo $machines | cut -d',' -f $((idx + 1)))
        ips="$ips,$ip"
        
        ssh_ip=$(echo $ssh_machines | cut -d',' -f $((idx + 1)))
        ssh_ips="$ssh_ips,$ssh_ip"
        
        
        roles="$roles,1"
        cluster="$cluster,$c"
    done
    
done


for i in $(seq 0 $((number_of_clients * number_of_clusters - 1  ))); do
    c=$((i / number_of_clusters + i ))
    
    ip=$(echo $client_machines | cut -d',' -f $((i + 1)))
    ips="$ips,$ip"
    
    ssh_ip=$(echo $client_ssh_machines | cut -d',' -f $((i + 1)))
    ssh_ips="$ssh_ips,$ssh_ip"
    
    
    roles="$roles,0"
    cluster="$cluster,$c"
done

ips="${ips:1}"
ssh_ips="${ssh_ips:1}"
roles="${roles:1}"
cluster="${cluster:1}"

echo "$ips"
echo "$ssh_ips"
echo "$roles"
echo "$cluster"



mkdir -p /tmp/dool/bin
wget "https://raw.githubusercontent.com/scottchiefbaker/dool/refs/heads/next/dool?ref=v1.3.4" -O /tmp/dool/bin/dool
chmod +x -R /tmp/dool/bin/dool
# add dool to path

# wget https://iperf.fr/download/ubuntu/libiperf0_3.1.3-1_amd64.deb -O /tmp/libiperf0_3.1.3-1_amd64.deb
# wget https://iperf.fr/download/ubuntu/iperf3_3.1.3-1_amd64.deb -O /tmp/iperf3_3.1.3-1_amd64.deb



# dpkg  -x /tmp/libiperf0_3.1.3-1_amd64.deb /tmp/libperf3
# dpkg  -x  /tmp/iperf3_3.1.3-1_amd64.deb /tmp/iperf3


for i in $(seq 0 $((total_container -1)) ); do
    
    idx_ip=$(echo "$ssh_ips" | cut -d',' -f $((i + 1)))
    echo "Node $i IP: $idx_ip"
    
    ssh "$username"@"$idx_ip" "killall cli"
    
    
    
done


for i in $(seq 0 $((total_container -1)) ); do
    
    idx_ip=$(echo "$ssh_ips" | cut -d',' -f $((i + 1)))
    echo "Node $i IP: $idx_ip"
    
    
    scp /tmp/cli "$username"@"$idx_ip":~/
    ssh -o StrictHostKeyChecking=no  "$username"@"$idx_ip" 'rm -rf ~/dool'
    scp -r -o StrictHostKeyChecking=no  /tmp/dool/ "$username"@"$idx_ip":~/
    
    # scp -r -o StrictHostKeyChecking=no   /tmp/iperf3/usr/bin/iperf3 "$username"@"$idx_ip":~/
    
    # scp -r -o StrictHostKeyChecking=no  /tmp/libperf3/usr/lib/x86_64-linux-gnu/ "$username"@"$idx_ip":~/
    
    
    
    ssh -o StrictHostKeyChecking=no  "$username"@"$idx_ip" 'chmod +x ~/dool/bin/dool'
    # ssh -o StrictHostKeyChecking=no  "$username"@"$idx_ip" 'chmod +x ~/iperf3'
    
    ssh -o StrictHostKeyChecking=no  "$username"@"$idx_ip" "sed -i '1s|^|export PATH=\$PATH:~/dool/bin/\\n|' ~/.bashrc"
    ssh -o StrictHostKeyChecking=no  "$username"@"$idx_ip" "sed -i '1s|^|export LD_LIBRARY_PATH=\$LD_LIBRARY_PATH:~/\\n|' ~/.bashrc"
    ssh -o StrictHostKeyChecking=no  "$username"@"$idx_ip" "sed -i '1s|^|export PATH=\$PATH:~/\\n|' ~/.bashrc"
    
    ssh -o StrictHostKeyChecking=no  "$username"@"$idx_ip" 'source ~/.bashrc'
    ssh -o StrictHostKeyChecking=no  "$username"@"$idx_ip" 'echo $PATH'
    
done
# exit
rm /tmp/zs_*.log
rm /tmp/zs_*.err
rm /tmp/zs_*.json
pids=()
for i in $(seq 0 $((total_container -1)) ); do
    
    idx_ip=$(echo "$ips" | cut -d',' -f $((i + 1)))
    idx_ssh_ip=$(echo "$ssh_ips" | cut -d',' -f $((i + 1)))
    echo "Node $i IP: $idx_ip"
    
    if (( i == $((total_container -1)) )); then
        
        ssh "$username"@"$idx_ssh_ip" "run_exp -m 'benchmark modubft' -n 0  -- ./cli   --id $i   --nodes $ips   --roles $roles   --clusters $cluster   --dont_verify_cluster_commit $dont_verify_cluster_commit   --use_pk_to_client $use_pk_to_client   --profile $profile   --runtime $runtime   --vallen $vallen   --cluster_batch_size $cluster_batch_size --use_dool $use_dool --use_certificate_compression $use_certificate_compression "> "/tmp/zs_$i.log" 2>"/tmp/zs_$i.err" &
        pids+=($!)
    else
        ssh "$username"@"$idx_ssh_ip" "run_exp -m 'benchmark modubft' -n 0  -- ./cli   --id $i   --nodes $ips   --roles $roles   --clusters $cluster   --dont_verify_cluster_commit $dont_verify_cluster_commit   --use_pk_to_client $use_pk_to_client   --profile $profile   --runtime $runtime   --vallen $vallen   --cluster_batch_size $cluster_batch_size --use_dool $use_dool --use_certificate_compression $use_certificate_compression" > "/tmp/zs_$i.log" 2>"/tmp/zs_$i.err" &
        pids+=($!)
    fi
    
done


# Example: Print a message every second while waiting for background processes
while :; do
    running=0
    for pid in "${pids[@]}"; do
        if kill -0 "$pid" 2>/dev/null; then
            running=1
            break
        fi
    done
    if (( running == 0 )); then
        break
    fi
    echo "Waiting for background processes to finish..."
    tail -q /tmp/zs_*.log
    tail -q -n+2 /tmp/zs_*.err
    sleep 1
done




echo "Finished execution on all nodes, waiting a bit for finalization..."

sleep 1

for i in $(seq 0 $((total_container -1)) ); do
    
    idx_ip=$(echo "$ips" | cut -d',' -f $((i + 1)))
    idx_ssh_ip=$(echo "$ssh_ips" | cut -d',' -f $((i + 1)))
    echo "Node $i IP: $idx_ip"
    
    scp -o StrictHostKeyChecking=no  "$username"@"$idx_ssh_ip":~/benchmark.json /tmp/zs_$i.json

    cp "/tmp/zs_$i.json" "$folder/zs_$i.json"
    cp "/tmp/zs_$i.log" "$folder/zs_$i.log"
    cp "/tmp/zs_$i.err" "$folder/zs_$i.err"

    echo "Benchmark succeeded for Node $i ($idx_ip)"
    
    
done