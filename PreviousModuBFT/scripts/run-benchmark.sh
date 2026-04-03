async_factor=32
use_pk_to_client="false"

machines="<REPLACE_WITH_IPS>"
num_machines=32

if [ ! -f ./cli-time-lat ]; then
    echo "Client file not found"
	exit
fi
if [ ! -f ./cli-time ]; then
    echo "Client file not found"
	exit
fi

echo -n "Cleaning up machines  "
for x in `seq 1 $num_machines`
do
	host=`echo $machines | cut -d ',' -f$x`
    ssh $host "killall cli-time" 2> /dev/null
    ssh $host "killall cli-time-lat" 2> /dev/null
	ssh $host "rm  /tmp/cli-*" 2> /dev/null
	echo -n "."
done
echo " done!"
sleep 2

logname=./`date -Is`.log

echo "DBG: -------------------------------------------" > $logname
echo "DBG: `date`" >> $logname
echo "DBG: Machines: $machines" >> $logname
echo "DBG: Async messages: $async_factor" >> $logname
echo "DBG: Using PK for client answer: $use_pk_to_client" >> $logname
echo "DBG: -------------------------------------------" >> $logname
echo "" >> $logname

cat $logname

echo "VAL PEERS CLIENTS TPUT NBT" >> $logname
echo "VAL PEERS CLIENTS TPUT NBT"
for value in 512 1024 2048 4096 8192
do
	for peers in 3 5 7 9 11 13 15 
	do
		for clients in 1 2 4 6 8 10 12 14 16 
		do
			./run-one-exp.sh $num_machines $machines $value $peers $clients $async_factor $use_pk_to_client > /tmp/$logname
			tput=`cat /tmp/$logname | grep "TPUT" | awk '{print $2}'`
			nbt=`cat /tmp/$logname | grep "NBT" | awk '{print $2}'`
			echo "$value $peers $clients $tput $nbt" >> $logname
			echo "$value $peers $clients $tput $nbt" 
		done
	done

done