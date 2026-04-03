#!/bin/bash

number_of_leaders=$1
number_of_leader_follower=$2
number_of_clients=$3

roles=""

leader="2"
follower="1"
client="0"




# Assign partitions to leaders
for _ in $(seq 0 $((number_of_leaders - 1))); do
    roles="$roles,$leader"
done


# Assign partitions to leader followers
for _ in $(seq 0 $((number_of_leaders - 1))); do
    for _ in $(seq 0 $((number_of_leader_follower - 1))); do
        roles="$roles,$follower"
    done
done


# Assign partitions to clients
for _ in $(seq 0 $((number_of_clients - 1))); do
    roles="$roles,$client"
done

# remove first comma 
roles="${roles:1}"


echo "$roles"