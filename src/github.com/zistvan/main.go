package main

import (
	"crypto/rand"
	"flag"
	"fmt"
	"io"
	"math"

	"os"
	"strconv"
	"strings"
	"time"

	n "github.com/zistvan/node"
	pb "github.com/zistvan/proto"
)

func main() {

	var myId int
	var runTime int
	var nodeListS, nodeRoleS, clusterListS string
	var profile bool
	var vallen int
	var usePkToClient bool
	var dontVerifyClusterCommit bool
	var clusterBatchSize int

	flag.BoolVar(&usePkToClient, "use_pk_to_client", false, "Use public key to client communication")
	flag.IntVar(&myId, "id", -1, "Node ID")
	flag.StringVar(&nodeListS, "nodes", "", "Comma-separated list of node addresses")
	flag.StringVar(&nodeRoleS, "roles", "", "Comma-separated list of node roles (0: leader, 1: peer, 2: client)")
	flag.StringVar(&clusterListS, "clusters", "", "Comma-separated list of cluster IDs for each node (0: cluster 0, 1: cluster 1, etc.)")
	flag.BoolVar(&profile, "profile", false, "Enable profiling (CPU and trace)")

	flag.IntVar(&runTime, "runtime", 10, "Runtime in seconds for the benchmark")
	flag.IntVar(&vallen, "vallen", 512, "Length of the value to be sent in bytes")
	flag.BoolVar(&dontVerifyClusterCommit, "dont_verify_cluster_commit", false, "Don't verify cluster commit (for testing purposes)")
	flag.IntVar(&clusterBatchSize, "cluster_batch_size", 64, "Batch size for cluster operations (default: 64)")

	flag.Parse()

	fmt.Printf("Node ID: %d\n", myId)
	fmt.Printf("Node list: %s\n", nodeListS)
	fmt.Printf("Node roles: %s\n", nodeRoleS)
	fmt.Printf("Cluster list: %s\n", clusterListS)

	nodeList := strings.Split(nodeListS, ",")

	nodeRole := make([]int, len(nodeList))
	for ind, elem := range strings.Split(nodeRoleS, ",") {
		nodeRole[ind], _ = strconv.Atoi(elem)
	}

	clusterList := make([]int, len(nodeList))
	for ind, elem := range strings.Split(clusterListS, ",") {
		clusterList[ind], _ = strconv.Atoi(elem)
	}

	// batchF := 64
	// runTime := 10
	// showLiveTput := 0

	// if len(os.Args) > 5 {
	// 	batchF, _ = strconv.Atoi(os.Args[5])
	// }

	// if len(os.Args) > 6 {
	// 	runTime, _ = strconv.Atoi(os.Args[6])
	// }

	// if len(os.Args) > 7 {
	// 	showLiveTput, _ = strconv.Atoi(os.Args[7])
	// }

	myCluster := clusterList[myId]
	nodesInMyCluster := make([]int, 0)

	// cluster_mode := false
	for i := 0; i < len(clusterList); i++ {
		if clusterList[i] == myCluster {
			nodesInMyCluster = append(nodesInMyCluster, i)
		} else {
			// cluster_mode = true
		}
	}
	// fmt.Printf("Node%d: Nodes in my cluster: %v\n", myId, nodesInMyCluster)

	myInnerClusterID := -1
	innerNodeList := make([]string, len(nodesInMyCluster))
	innerNodeRole := make([]int, len(nodesInMyCluster))

	for c, nodeInMyCluster := range nodesInMyCluster {

		if nodeInMyCluster == myId {
			myInnerClusterID = c
		}
		innerNodeList[c] = nodeList[nodeInMyCluster]
		innerNodeRole[c] = nodeRole[nodeInMyCluster]
	}

	// fmt.Printf("Node%d: My inner cluster ID: %d\n", myId, myInnerClusterID)
	// fmt.Printf("Node%d: Inner node list: %v\n", myId, innerNodeList)
	// fmt.Printf("Node%d: Inner node role: %v\n", myId, innerNodeRole)
	// fmt.Printf("Node%d: Cluster list: %v\n", myId, clusterList)
	// fmt.Printf("Node%d: My cluster: %d\n", myId, myCluster)
	// fmt.Printf("Node%d: My role: %d\n", myId, nodeRole[myId])

	// os.Exit(0)

	// runtime.SetBlockProfileRate(1)
	// f, err := os.Create(fmt.Sprintf("%d_cpu.prof", myId))
	// if err != nil {
	// 	fmt.Println("Error creating CPU profile:", err)
	// 	os.Exit(1)
	// }
	// defer f.Close()

	// f_t, _ := os.Create(fmt.Sprintf("%d_trace.prof", myId))
	// trace.Start(f_t)
	// defer trace.Stop()

	// pprof.StartCPUProfile(f)
	// defer pprof.StopCPUProfile()

	me := &n.Node{}

	me.Initialize(myInnerClusterID, innerNodeList, innerNodeRole)

	local_clients := 499

	me.Run()

	if nodeRole[myId] == n.ROLE_CLIENT {
		time.Sleep(time.Millisecond * 5000)

		// vallen, _ := strconv.Atoi(os.Args[4])
		rtxt := make([]byte, vallen)
		if _, err := io.ReadFull(rand.Reader, rtxt); err != nil {
			fmt.Println(err)
		}

		// rtxt = me.CompressRequest(rtxt)
		// rtxt = me.DecompressRequest(rtxt)

		// if vallen != len(rtxt) {
		// 	fmt.Printf("Node%d: Error: Compressed request length %d does not match original length %d\n", myId, len(rtxt), vallen)
		// 	os.Exit(1)
		// }

		chResult := make(chan int, 1000)

		go me.ClientReceiveFromPeers(chResult)
		_, _, rawB := me.ClientSendToLeader(rtxt, pb.SigType_PKSig)
		// <-chResult

		// issued := 1
		// received := 1

		start := time.Now()
		elapsed := time.Since(start)
		// mustStop := false

		realReceived := 0
		//msgId := 0

		// lastCheckReceived := 0
		// lastStepReceived := 0
		// lastStepBegin := time.Now()

		// log_bytes_sent := os.Getenv("BFT_LOG_SENT_BYTES") != "false"

		stop_chan := make(chan bool, 1)
		finished_chan := make(chan int, local_clients)
		allowed_to_send := make(chan bool, local_clients)

		for i := 0; i < local_clients; i++ {
			allowed_to_send <- true
		}

		time_spent_receiving_local_clients := make([]int, local_clients)
		time_spent_sending_local_clients := make([]int, local_clients)
		time_spent_waiting_to_send := make([]int, local_clients)
		for i := 0; i < local_clients; i++ {
			time_spent_receiving_local_clients[i] = 0
			time_spent_sending_local_clients[i] = 0
			time_spent_waiting_to_send[i] = 0
		}
		time_spent_waiting_receiving := 0

		// for int(elapsed) < (int(time.Second) * runTime)
		go func(stop_chan chan bool, chResult chan int) {
			realReceived = 0
			notStop := true
			lastReceived := -1
			for {

				select {
				case <-stop_chan:
					notStop = false
					for {
						select {
						case msgId := <-chResult:
							fmt.Printf("Node%d: received response for message %d\n", myId, msgId)
							realReceived++
							allowed_to_send <- notStop
						case <-time.After(time.Second * 30):
							break
						default:
							fmt.Printf("Node%d: Stopping client after %d messages\n", myId, realReceived)
							stop_chan <- notStop
							for {

							}
						}
					}
				case msgId := <-chResult:
					if msgId <= lastReceived {
						panic(fmt.Sprintf("Node%d: Received message %d, but last received was %d", myId, msgId, lastReceived))
					}
					lastReceived = msgId
					realReceived++
					allowed_to_send <- notStop
					// time_spent_receiving += int(time.Since(now).Nanoseconds())
				case <-time.After(time.Millisecond * 1):
					time_spent_waiting_receiving += int(time.Millisecond * 1)
				}
			}

		}(stop_chan, chResult)

		for i := 0; i < local_clients; i++ {

			go func(stop_chan chan bool, rawB []byte, cid int, finished_chan chan int, allowed_to_send chan bool) {
				sent_out := 0
				for {
					select {
					case notStop := <-allowed_to_send:
						if !notStop {
							break
						}
						me.ClientSendToLeaderCachedRaw(rawB)
						sent_out++
					case <-time.After(time.Millisecond * 1):
						time_spent_waiting_to_send[cid] += int(time.Millisecond * 1)
					}
				}
			}(stop_chan, rawB, i, finished_chan, allowed_to_send)

		}

		time.Sleep(time.Second * time.Duration(runTime))

		stop_chan <- true
		<-stop_chan

		fmt.Printf("Node%d: Stopping clients after %d seconds\n", myId, time.Since(start).Milliseconds()/1000)
		// realReceived = 0
		// for i := 0; i < local_clients; i++ {
		// 	realReceived += <-finished_chan
		// 	// fmt.Printf("Node%d: Client %d finished with %d messages\n", myId, i, realReceived)
		// }
		time_spent_waiting_to_send_total := 0
		for i := 0; i < local_clients; i++ {
			// time_spent_receiving += time_spent_receiving_local_clients[i]

			time_spent_waiting_to_send_total += time_spent_waiting_to_send[i]
		}

		fmt.Printf("Node%d: Total time spent sending: %d ns", myId, time_spent_waiting_to_send_total)
		fmt.Printf("Node%d: Total time spent receiving: %d ns", myId, time_spent_waiting_receiving)
		fmt.Printf("Node%d: Total time spent: %d ns", myId, time_spent_waiting_to_send_total+time_spent_waiting_receiving)
		fmt.Printf("Node%d: Percentage of time spent sending: %f%%", myId, float64(time_spent_waiting_to_send_total)/float64(time_spent_waiting_to_send_total+time_spent_waiting_receiving)*100)

		elapsed = time.Since(start)
		tput := float32(realReceived) / (float32(elapsed) / 1000000000.0)
		fmt.Printf("\nThroughput [ops/s] %d\n", int(tput))

		bytes_sent := realReceived * (len(rawB) + 4)

		fmt.Printf("Node%d:Total bytes sent: %d\n", myId, bytes_sent)
		datarate := float64(bytes_sent) / (float64(elapsed) / 1000000000.0)
		fmt.Printf("Node%d:Data rate [B/s] %f\n", myId, datarate)
		fmt.Printf("Node%d:Data rate [GB/s] %f\n", myId, datarate/math.Pow(10, 9))

	}

	os.Exit(0)
}
