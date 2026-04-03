#!/bin/bash

number_of_leaders=$1
number_of_leader_follower=$2
number_of_clients=$3
number_of_azs=${4:-3}

region="eu-central-1"

azs=("a" "b" "c" "d" "e" "f" "g" "h" "i" "j" "k" "l" "m" "n" "o" "p" "q" "r" "s" "t" "u" "v" "w" "x" "y" "z")
#add region to each entry
for i in "${!azs[@]}"; do
    azs[$i]="${region}${azs[$i]}"
done

# get first number_of_azs out of azs array 

azs=("${azs[@]:0:number_of_azs}")





az_list=""






# Assign az_list to leaders
for i in $(seq 0 $((number_of_leaders - 1))); do

    c="${azs[$((i % number_of_azs))]}"
    az_list="$az_list,$c"
done


# Assign az_list to leader followers
for c in $(seq 0 $((number_of_leaders - 1))); do
    for i in $(seq 0 $((number_of_leader_follower - 1))); do
        idx=$((((c + 1  + i) % number_of_leader_follower % number_of_azs)   ))
        az="${azs[$idx]}"
        az_list="$az_list,$az"
    done
done

# Assign azs to clients
for i in $(seq 0 $((number_of_clients - 1))); do
    # c = azs[0]
    c="${azs[$((i % number_of_azs))]}"
    az_list="$az_list,$c"
done

# remove first comma 
az_list="${az_list:1}"


echo "$az_list"