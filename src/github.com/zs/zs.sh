#!/bin/bash

number_of_clusters=1
number_of_peers=5
number_of_clients=1
total_nodes_per_cluster=$((number_of_peers +  number_of_clients ))
total_container=$((total_nodes_per_cluster * number_of_clusters ))


# --peer_machines "10.10.2.183,10.10.2.184,10.10.2.185,10.10.2.186,10.10.2.187,10.10.2.188" --peer_ssh_machines "10.0.2.83,10.0.2.84,10.0.2.85,10.0.2.86,10.0.2.87,10.0.2.88" \
# --client_machines "10.10.2.61,10.10.2.62" --client_ssh_machines "10.0.2.61,10.0.2.62"
ssh_machines="10.0.2.81,10.0.2.82,10.0.2.83,10.0.2.84,10.0.2.85,10.0.2.87,10.0.2.88"
machines="10.10.2.181,10.10.2.182,10.10.2.183,10.10.2.184,10.10.2.185,10.10.2.187,10.10.2.188"
client_machines="10.10.2.62,10.10.2.63"
client_ssh_machines="10.0.2.62,10.0.2.63"

current_dir=$(pwd)


cd "github.com/zistvan" || exit
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

for i in $(seq 0 $((total_container -1)) ); do
    
    idx_ip=$(echo "$ssh_ips" | cut -d',' -f $((i + 1)))
    echo "Node $i IP: $idx_ip"

    ssh aaffeldt@"$idx_ip" "killall cli"



done


for i in $(seq 0 $((total_container -1)) ); do
    
    idx_ip=$(echo "$ssh_ips" | cut -d',' -f $((i + 1)))
    echo "Node $i IP: $idx_ip"

    
    scp /tmp/cli aaffeldt@"$idx_ip":~/ 

done
# exit 
rm /tmp/zs_*.log
for i in $(seq 0 $((total_container -1)) ); do
    
    idx_ip=$(echo "$ips" | cut -d',' -f $((i + 1)))
    idx_ssh_ip=$(echo "$ssh_ips" | cut -d',' -f $((i + 1)))
    echo "Node $i IP: $idx_ip"

    if (( i == $((total_container -1)) )); then
         
        ssh aaffeldt@"$idx_ssh_ip" "run_exp -m 'benchmark modubft' -n 0  -- ./cli --id $i --nodes $ips --roles $roles --clusters $cluster " &> "/tmp/zs_$i.log"
    else
      ( ssh aaffeldt@"$idx_ssh_ip" "run_exp -m 'benchmark modubft' -n 0  -- ./cli --id $i --nodes $ips --roles $roles --clusters $cluster " &) &> "/tmp/zs_$i.log"
    fi
    
done



