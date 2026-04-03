n=32
machines="<REPLACE_WITH_IPS>"

num_peer=$2
num_client=$3
total_nodes=$((num_peer+num_client+1))
total_nodes_nocoord=$((num_peer+num_client))

ASYNC_MSGS=8

size=$1

role_list="2"

for r in `seq 2 $num_peer`
do
	role_list="$role_list,1"
done

for r in `seq 1 $num_client`
do
	role_list="$role_list,0"
done

role_list="$role_list,3"

ip_list=`echo $machines | cut -d ',' -f1`

for id in `seq 2 $total_nodes` 
do
	nth=`echo $machines | cut -d ',' -f$id`
	ip_list="$ip_list,$nth"
done

echo "NODES: $ip_list"
echo "ROLES: $role_list"

echo -n "COPY: "
for x in `seq 1 $total_nodes`
do
	host=`echo $machines | cut -d ',' -f$x`	
    scp ./cli-timeout $host:/tmp/ > /dev/null
    echo -n "$x "
done
echo ""


clientend=$((num_peer+num_client))
clientbegin=$((num_peer+1))
coordloc=$((num_peer+num_client+1))

for x in `seq 1 $num_peer`
do
	host=`echo $machines | cut -d ',' -f$x`
    nodeId=$((x-1))    
    (ssh $host "/tmp/cli-timeout $nodeId $ip_list $role_list $size" &)
done

for x in `seq $clientbegin $clientend`
do
	host=`echo $machines | cut -d ',' -f$x`
    nodeId=$((x-1))    
    (ssh $host "/tmp/cli-timeout $nodeId $ip_list $role_list $size $ASYNC_MSGS 20 1" &)
done

host=`echo $machines | cut -d ',' -f$coordloc`
ssh $host "/tmp/cli-timeout $clientend $ip_list $role_list $size"

sleep 1

echo -n "CLEANUP: " 

for x in `seq 1 $total_nodes`
do
	host=`echo $machines | cut -d ',' -f$x`
    ssh $host "killall cli-timeout" 2> /dev/null
    echo -n "$x "
done
echo ""



