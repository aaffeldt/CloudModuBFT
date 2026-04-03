#!/bin/bash

    # parser.add_argument("--number_of_clusters", type=str, default="1")
    # parser.add_argument("--number_of_clients", type=str, default="1")
    # parser.add_argument("--number_of_peers", type=str, default="3")
    # parser.add_argument("--vallen_arr", type=str, default="512,4096")
    # parser.add_argument("--cluster_batch_size_arr",
    #                     type=str, default="1,4,16,64,128,256")
    # parser.add_argument("--number_of_buckets", type=str, default="1")
    # parser.add_argument("--basefolder", type=str, required=True)
    # parser.add_argument("--runtime", type=str, default="10")
    # parser.add_argument("--username", type=str, default="ubuntu")
    # parser.add_argument("--dont_verify_cluster_commit",
    #                     type=str, default="false")
    # parser.add_argument("--use_pk_to_client", type=str, default="false")
    # parser.add_argument("--profile", type=str, default="false")
    # parser.add_argument("--use_certificate_compression",
    #                     type=str, default="false")
    # parser.add_argument("--use_dool", type=str, default="true")
    # parser.add_argument("--alternateClusterPrepare", type=str, default="false")
    # parser.add_argument("--public_key_path", type=str,
    #                     default="./aws_ec2_key.pub")
    # parser.add_argument("--private_key_path", type=str,
    #                     default="./aws_ec2_key")
    # parser.add_argument("--machines", type=str, default="machines.json")
    # parser.add_argument("--main_file", type=str,
    #                     default="../modubft/main.go")
    # parser.add_argument("--max_partitions", type=int, default=6)

    # parser.add_argument("--region", type=str, default="eu-central-1")
    # parser.add_argument("--availability_zone", type=str, default="a")
    # parser.add_argument("--ami", type=str, default="ami-0a116fa7c861dd5f9")
    # parser.add_argument("--instance_type", type=str, default="t2.micro")
    # parser.add_argument("--client_instance_type",
    #                     type=str, default="t2.micro")

    # parser.add_argument("--send_transaction", type=str, default="true")
    # parser.add_argument("--cluster_prepare_to_all_and_enable_hash",
    #                     type=str, default="false")

number_of_clusters=1
number_of_peers=21
number_of_clients=1
number_of_buckets=1
cluster_prepare_to_all_and_enable_hash="false"
alternate_cluster_prepare="false"
cluster_batch_size=1
vallen=512

lists=$(python3 simple.py  --basefolder "./out" \
    --number_of_clusters "$number_of_clusters" \
    --number_of_peers "$number_of_peers" \
    --number_of_clients "$number_of_clients" \
    --number_of_buckets "$number_of_buckets" )
# outputs
# Cluster: 0,0,0,0
# Roles: 2,1,1,0
# Buckets: 0,-1,-1,0
cluster=$(echo "$lists" | sed -n 's/^Cluster: //p')
roles=$(echo "$lists" | sed -n 's/^Roles: //p')
buckets=$(echo "$lists" | sed -n 's/^Buckets: //p')
echo "Cluster: $cluster"
echo "Roles: $roles"
echo "Buckets: $buckets"

total_container=$(( number_of_clusters * (number_of_peers + number_of_clients * number_of_buckets ) ))

echo "Total containers: $total_container"

docker ps -q --filter "network=simple-net" | xargs -r docker stop
docker ps -a -q --filter "network=simple-net" | xargs -r docker rm

docker build --no-cache -t simple . || exit 1


docker network create --subnet=10.0.0.0/24 simple-net

# CMD ./cli --send_transaction "${SEND_TRANSACTION}" --cluster_prepare_to_all_and_enable_hash "${CLUSTER_PREPARE_TO_ALL_AND_ENABLE_HASH}" --id ${ID} --buckets ${BUCKETS} --alternateClusterPrepare ${ALTERNATE_CLUSTER_PREPARE}  --nodes ${NODES}   --roles ${ROLES}   --clusters ${CLUSTERS}   --dont_verify_cluster_commit "${DONT_VERIFY_CLUSTER_COMMIT}"   --use_pk_to_client "${USE_PK_TO_CLIENT}"   --profile "${PROFILE}"   --runtime ${RUNTIME}   --vallen ${VALLEN}   --cluster_batch_size ${CLUSTER_BATCH_SIZE} --use_dool "${USE_DOOL}" --use_certificate_compression "${USE_CERTIFICATE_COMPRESSION}" 2>&1 | tee -a  /tmp/aws_${ID}.log

ips=""
for i in $(seq 0 $((total_container -1)) ); do
    idx=$((i + 2 ))
    ips+="10.0.0.$idx"
    if (( i != $((total_container -1)) )); then
        ips+=","
    fi
done
echo "IPs: $ips"


container_ids=""
for i in $(seq 0 $((total_container -1)) ); do
    echo "docker run --net simple-net --ip 10.0.0.$((i + 2)) simple" 
    idx=$((i + 2 ))

    #write to out/simple_"$i".log 
    if (( i == $((total_container -1)) )); then 
            # docker run --net simple-net   simple  "(./cli --id \"$i\" --nodes \"$ips\" --roles \"$roles\" --clusters \"$cluster\"  --buckets ${buckets} --cluster_prepare_to_all_and_enable_hash \"${cluster_prepare_to_all_and_enable_hash}\"  --alternateClusterPrepare ${alternate_cluster_prepare}   --vallen ${vallen}   --cluster_batch_size ${cluster_batch_size} --use_dool \"false\")"

    container_id=$(docker run  -d  --net simple-net  simple  "(./cli --id \"$i\" --nodes \"$ips\" --roles \"$roles\" --clusters \"$cluster\"  --buckets ${buckets} --cluster_prepare_to_all_and_enable_hash \"${cluster_prepare_to_all_and_enable_hash}\"  --alternateClusterPrepare ${alternate_cluster_prepare}   --vallen ${vallen}   --cluster_batch_size ${cluster_batch_size} --use_dool \"false\")")
            docker logs -f "$container_id"  | tee ./out/simple_"$i".log &
    else
    container_id=$( docker run  -d  --net simple-net  simple  "(./cli --id \"$i\" --nodes \"$ips\" --roles \"$roles\" --clusters \"$cluster\"  --buckets ${buckets} --cluster_prepare_to_all_and_enable_hash \"${cluster_prepare_to_all_and_enable_hash}\"  --alternateClusterPrepare ${alternate_cluster_prepare}   --vallen ${vallen}   --cluster_batch_size ${cluster_batch_size} --use_dool \"false\")")
     container_ids+="$container_id,"   
        
    fi
    
done

for i in $(seq 0 $((total_container -2)) ); do
    cid=$(echo "$container_ids" | cut -d',' -f$((i +1)) )
    echo "Following logs for container id: $cid"
    docker logs -f "$cid"  | tee ./out/simple_"$i".log &
done


