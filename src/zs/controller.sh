#!/bin/bash

username=$1
total_container=$2
ssh_ips=$3
ips=$4
roles=$5
cluster=$6
dont_verify_cluster_commit=$7
use_pk_to_client=$8
profile=$9
runtime=${10}
vallen=${11}
cluster_batch_size=${12}
use_certificate_compression=${13}
use_dool=${14}
alternateClusterPrepare=${15}
buckets=${16}

# Print received arguments for debugging
echo "use_pk_to_client: $use_pk_to_client"
echo "profile: $profile"
echo "runtime: $runtime"
echo "vallen: $vallen"
echo "cluster_batch_size: $cluster_batch_size"


mkdir -p /tmp/dool/bin
curl "https://raw.githubusercontent.com/scottchiefbaker/dool/refs/heads/next/dool?ref=v1.3.4" -o /tmp/dool/bin/dool
chmod +x -R /tmp/dool/bin/dool
# add dool to path
export PATH=$PATH:/tmp/dool/bin/

# wget https://iperf.fr/download/ubuntu/libiperf0_3.1.3-1_amd64.deb
# wget https://iperf.fr/download/ubuntu/iperf3_3.1.3-1_amd64.deb 



# dpkg  -x libiperf0_3.1.3-1_amd64.deb libperf3
# dpkg  -x  iperf3_3.1.3-1_amd64.deb iperf3






pids=()
for i in $(seq 0 $((total_container -1)) ); do
    
    idx_ip=$(echo "$ssh_ips" | cut -d',' -f $((i + 1)))
    echo "Node $i IP: $idx_ip"
    
    ssh -o StrictHostKeyChecking=no "$username"@"$idx_ip" "killall cli" &
    pids+=($!)
done

for pid in "${pids[@]}"; do
   wait "$pid"
   # do something when a job completes
done




setup_machine(){
    local i=$1

    idx_ip=$(echo "$ssh_ips" | cut -d',' -f $((i + 1)))
    echo "Node $i IP: $idx_ip"


    scp -o StrictHostKeyChecking=no  "/tmp/cli" "$username"@"$idx_ip":~/
    scp -r -o StrictHostKeyChecking=no  /tmp/dool "$username"@"$idx_ip":~/
    # scp -r -o StrictHostKeyChecking=no   ~/iperf3/usr/bin/iperf3 "$username"@"$idx_ip":~/
    # scp -r -o StrictHostKeyChecking=no  ~/libperf3/usr/lib/x86_64-linux-gnu/ "$username"@"$idx_ip":~/

    ssh -o StrictHostKeyChecking=no  "$username"@"$idx_ip" 'chmod +x ~/dool/bin/dool'
    ssh -o StrictHostKeyChecking=no  "$username"@"$idx_ip" 'chmod +x ~/iperf3'

    ssh -o StrictHostKeyChecking=no  "$username"@"$idx_ip" "sed -i '1s|^|export PATH=\$PATH:~/dool/bin/\\n|' ~/.bashrc"
    # ssh -o StrictHostKeyChecking=no  "$username"@"$idx_ip" "sed -i '1s|^|export LD_LIBRARY_PATH=\$LD_LIBRARY_PATH:~/\\n|' ~/.bashrc"
    ssh -o StrictHostKeyChecking=no  "$username"@"$idx_ip" "sed -i '1s|^|export PATH=\$PATH:~/\\n|' ~/.bashrc"

    ssh -o StrictHostKeyChecking=no  "$username"@"$idx_ip" 'source ~/.bashrc'
    ssh -o StrictHostKeyChecking=no  "$username"@"$idx_ip" 'echo $PATH' 


}
pids=()
for i in $(seq 0 $((total_container -1)) ); do
    
    setup_machine $i & 
    pids+=($!)
done

for pid in "${pids[@]}"; do
   wait "$pid"
done



# exit
rm /tmp/zs_*.log
rm /tmp/zs_*.json



pids=()
for i in $(seq 0 $((total_container -1)) ); do
    idx_ip=$(echo "$ips" | cut -d',' -f $((i + 1)))
    idx_ssh_ip=$(echo "$ssh_ips" | cut -d',' -f $((i + 1)))
    echo "Node $i IP: $idx_ip"


    ssh -o StrictHostKeyChecking=no "$username"@"$idx_ssh_ip" "run_exp -m 'modubft' -n 0 --  ./cli  --id $i --buckets $buckets --alternateClusterPrepare $alternateClusterPrepare  --nodes $ips   --roles $roles   --clusters $cluster   --dont_verify_cluster_commit $dont_verify_cluster_commit   --use_pk_to_client $use_pk_to_client   --profile $profile   --runtime $runtime   --vallen $vallen   --cluster_batch_size $cluster_batch_size --use_dool $use_dool --use_certificate_compression $use_certificate_compression" > "/tmp/zs_$i.log" 2>"/tmp/zs_$i.err" &
pids+=($!)

done

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
    cat /tmp/zs_*.log
    tail -q -n+2 /tmp/zs_*.err
    sleep 1
done


echo "Finished execution on all nodes, waiting a bit for finalization..."

sleep 1
pids=()
for i in $(seq $((total_container -1)) -1 0); do
    
    idx_ip=$(echo "$ips" | cut -d',' -f $((i + 1)))
    idx_ssh_ip=$(echo "$ssh_ips" | cut -d',' -f $((i + 1)))
    echo "Node $i IP: $idx_ip"
    
    scp -o StrictHostKeyChecking=no  "$username"@"$idx_ssh_ip":~/benchmark.json /tmp/zs_$i.json & 
     pids+=($!)

    # if i == 0 
 
    scp -o StrictHostKeyChecking=no  "$username"@"$idx_ssh_ip":/tmp/cpu.prof /tmp/cpu_$i.prof &
    scp -o StrictHostKeyChecking=no  "$username"@"$idx_ssh_ip":/tmp/trace.prof /tmp/trace_$i.prof &

   
    echo "Benchmark succeeded for Node $i ($idx_ssh_ip)"
    
    
done


for pid in "${pids[@]}"; do
   wait "$pid"
   # do something when a job completes
done
