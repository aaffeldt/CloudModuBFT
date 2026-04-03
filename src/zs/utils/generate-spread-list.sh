#!/bin/bash

number_of_leaders=$1
number_of_leader_follower=$2
number_of_clients=$3

spread=""




# Assign spread to leaders
for c in $(seq 0 $((number_of_leaders - 1))); do
    spread="$spread,$c"
done


# Assign spread to leader followers
for c in $(seq 0 $((number_of_leaders - 1))); do
    for _ in $(seq 0 $((number_of_leader_follower - 1))); do
        spread="$spread,$c"
    done
done

# Assign spread to clients
for c in $(seq 0 $((number_of_clients - 1))); do
    spread="$spread,$c"
done

# remove first comma 
spread="${spread:1}"


echo "$spread"