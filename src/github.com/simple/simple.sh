#!/bin/bash


number_of_clusters=1
number_of_peers=3
number_of_clients=1
total_nodes_per_cluster=$((number_of_peers +  number_of_clients ))
total_container=$((total_nodes_per_cluster * number_of_clusters ))

echo $total_container

docker ps -q --filter "network=simple-net" | xargs -r docker stop
docker ps -a -q --filter "network=simple-net" | xargs -r docker rm

docker build -t simple:zistvan -f github.com/simple/Dockerfile . || exit 1


docker network create --subnet=10.0.0.0/24 simple-net

roles=""
ips=""
cluster=""

for c in $(seq 0 $((number_of_clusters -1 ))); do
    idx=$((c * total_nodes_per_cluster  + 2))
    
    ips="$ips,10.0.0.$idx"
    roles="$roles,2"
    cluster="$cluster,$c"
    
    
    for i in $(seq 1 $((number_of_peers -1 ))); do
        idx=$((c * total_nodes_per_cluster + i + 2))
        ips="$ips,10.0.0.$idx"
        roles="$roles,1"
        cluster="$cluster,$c"
    done
    for i in $(seq 0 $((number_of_clients -1 ))); do
        idx=$((c * total_nodes_per_cluster + i + number_of_peers + 2))
        ips="$ips,10.0.0.$idx"
        roles="$roles,0"
        cluster="$cluster,$c"
    done
done

ips="${ips:1}"
roles="${roles:1}"
cluster="${cluster:1}"
echo "$ips"
echo "$roles"
echo "$cluster"


for i in $(seq 0 $((total_container -1)) ); do
    echo "docker run --net simple-net --ip 10.0.0.$((i + 2)) simple:zistvan --id $i --nodes \"$ips\" --roles \"$roles\" --clusters \"$cluster\""
    idx=$((i + 2 ))
    if (( i == $((total_container -1)) )); then
        docker run  --net simple-net --ip "10.0.0.$idx"  simple:zistvan  --id "$i" --nodes "$ips" --roles "$roles" --clusters "$cluster"
    else
        docker run -d --net simple-net --ip "10.0.0.$idx"  simple:zistvan  --id "$i" --nodes "$ips" --roles "$roles" --clusters "$cluster"
        
        
    fi
    
done


