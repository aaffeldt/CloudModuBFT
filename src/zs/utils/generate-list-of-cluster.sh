#!/bin/bash

number_of_leaders=$1
number_of_leader_follower=$2
number_of_clients=$3

cluster=""




# Assign partitions to leaders
for c in $(seq 0 $((number_of_leaders - 1))); do
    cluster="$cluster,$c"
done



# Assign partitions to leader followers
for c in $(seq 0 $((number_of_leaders - 1))); do
    for _ in $(seq 0 $((number_of_leader_follower - 1))); do
        cluster="$cluster,$c"
    done
done

# Assign partitions to clients
for c in $(seq 0 $((number_of_clients - 1))); do
    cluster="$cluster,$c"
done


# remove first comma 
cluster="${cluster:1}"


echo "$cluster"