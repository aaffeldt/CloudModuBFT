#!/bin/bash
set -e
set -x 
controller_ip=$1
ssh_key=${2}
username=${3:-"ubuntu"}


ssh-keygen -f "$HOME/.ssh/known_hosts" -R "$controller_ip"
scp -o StrictHostKeyChecking=no -i "$ssh_key" /tmp/cli "$username"@"$controller_ip":~/cli
scp -o StrictHostKeyChecking=no -i "$ssh_key" ./controller.sh "$username"@"$controller_ip":~/controller.sh  # || destroy_terraform_with_cd
scp -o StrictHostKeyChecking=no -i "$ssh_key" "$ssh_key" "$username"@"$controller_ip":~/ssh_key  #  || destroy_terraform_with_cd
