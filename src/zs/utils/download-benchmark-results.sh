#!/bin/bash

folder=$1

mv /tmp/zs_*.log "$folder/"
mv /tmp/zs_*.err "$folder/"
mv /tmp/zs_*.json "$folder/"
echo "will move benchmark files"
# ls /tmp/ 

mv /tmp/benchmark*.json "$folder/"
mv /tmp/*.prof "$folder/"l