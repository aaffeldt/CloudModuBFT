#!/bin/bash

number_of_leaders=$1
number_of_leader_follower=$2
number_of_clients=$3
number_of_partitions=$4

partition=""



# Assign partitions to leaders
for i in $(seq 0 $((number_of_leaders - 1))); do
    partition="$partition,$((i % number_of_partitions + 1))"
done


# Assign partitions to leader followers
for c in $(seq 0 $((number_of_leaders - 1))); do
    for i in $(seq 0 $((number_of_leader_follower - 1))); do
        partition="$partition,$((((c + i) % number_of_leader_follower % number_of_partitions) + 1))"
    done
done

# Assign partitions to clients
for i in $(seq 0 $((number_of_clients - 1))); do
    partition="$partition,$((i % number_of_partitions + 1))"
done


# remove first comma 
partition="${partition:1}"


echo "$partition"