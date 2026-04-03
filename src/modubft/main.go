package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	peer "modubft/node"
	"os"
	"os/exec"
	"runtime/pprof"
	"strconv"
	"strings"
)

var myId int
var runTime int
var nodeListS, nodeRoleS, clusterListS string
var profile bool
var vallen int
var usePkToClient bool
var dontVerifyClusterCommit bool
var clusterBatchSize int
var useCertificateCompression bool
var alternatingClusterPrepareSender bool

var done chan bool

func main() {

	done = make(chan bool, 1000)
	// f_b, _ := os.Create("/tmp/block.prof")
	// runtime.SetBlockProfileRate(1)
	// p := pprof.Lookup("block")
	// defer func() {
	// 	p.WriteTo(f_b, 0)

	// }()

	// f_heap, _ := os.Create("/tmp/heap.prof")
	// defer pprof.WriteHeapProfile(f_heap)

	f, _ := os.Create("/tmp/cpu.prof")
	pprof.StartCPUProfile(f)
	defer pprof.StopCPUProfile()

	// f_t, _ := os.Create("/tmp/trace.prof")
	// trace.Start(f_t)
	// defer trace.Stop()
	// iperf

	// if myId == 0 {
	// 	http.HandleFunc("/Done", func(w http.ResponseWriter, r *http.Request) {
	// 		fmt.Printf("Received done from node \n")
	// 		done <- true
	// 	})
	// 	go func() {
	// 		log.Fatal(http.ListenAndServe(":8080", nil))
	// 	}()

	// }

	toInit, setup := initializePeerIdentifier()
	//  print to init
	fmt.Printf("Initializing Node: %+v \n", toInit)

	node := peer.NewPeer(toInit)

	node.Benchmark.Setup = setup
	fmt.Printf("Node: %+v ", toInit.Identifier)

	var dool *exec.Cmd
	var buffer bytes.Buffer
	if setup.UseDool {
		dool = runDool(&buffer)
	}

	// node.Benchmark.IperfBefore = launchIperfTesting(toInit)
	// if toInit.MyID == 0 {
	// 	go startIperfServer()
	// }

	if toInit.MyRole == peer.ROLE_CLIENT {
		client := peer.Client{Node: node}
		client.Run()
		// time.Sleep(time.Millisecond * 1)

		// clients := (peer.CHECKPOINT_INTERVAL / toInit.NumberOfBuckets) - toInit.NumberOfBuckets
		clients :=  (peer.CHECKPOINT_INTERVAL / toInit.NumberOfBuckets) - toInit.NumberOfBuckets 
		fmt.Printf("Client%d: Running with %d clients for %d seconds\n", toInit.MyID, clients, setup.RunTime)

	

		client.RunClient(vallen, clients, 10)
		client.CloseAllConnections()

		if setup.UseDool {
			terminateDoolProcess(dool, &buffer, &client.Node)
		}

		// client.Benchmark.IperfAfter = launchIperfTesting(toInit)

		exportBenchmarkToFile(client.Node)

	} else {

		// }
		node.Run()
		//finished
		fmt.Printf("Node%d: Finished execution\n", toInit.MyID)
		if setup.UseDool {
			terminateDoolProcess(dool, &buffer, &node)
		}
		// node.Benchmark.IperfAfter = launchIperfTesting(toInit)

		exportBenchmarkToFile(node)
		// fmt.Println("Benchmark:", node.GetBenchmark())

	}

	// if myId == 0 {
	// 	i := 0
	// 	for range done {
	// 		fmt.Printf("Node0: Received done %d/%d\n", i+1, node.GetAllCount()-1)
	// 		i++
	// 		if i == (node.GetAllCount() - 1) {
	// 			break
	// 		}
	// 	}
	// } else {
	// 	for {
	// 		_, err := http.Get("http://" + getPeerWithId0(toInit).MyIP + ":8080/Done")
	// 		if err == nil {
	// 			fmt.Printf("Node%d: Exiting\n", toInit.MyID)
	// 			break
	// 		} else {
	// 			fmt.Println(err)
	// 		}
	// 	}

	// }

}

func getPeerWithId0(toInit peer.PeerIdentifier) peer.PeerIdentifier {
	for _, peer := range toInit.MyPeers {
		if peer.MyID == 0 {
			return peer
		}
	}
	for _, peer := range toInit.ForeignPeers {
		if peer.MyID == 0 {
			return peer
		}
	}
	return peer.PeerIdentifier{}
}

// func launchIperfTesting(toInit peer.PeerIdentifier) iperf.TestReport {
// 	if toInit.MyID != 0 {
// 		time.Sleep(1 * time.Second)
// 		if s := runIperfClientOnPeers(toInit.MyPeers); s != nil {
// 			return *s
// 		}
// 		if s := runIperfClientOnPeers(toInit.ForeignPeers); s != nil {
// 			return *s
// 		}
// 	}
// 	return iperf.TestReport{}
// }

// func startIperfServer() {
// 	s := iperf.NewServer()
// 	err := s.Start()
// 	if err != nil {
// 		fmt.Printf("failed to start server: %v\n", err)
// 		// os.Exit(-1)
// 	}

// 	for s.Running {
// 		time.Sleep(100 * time.Millisecond)
// 	}

// 	fmt.Println("server finished")
// }

// func runIperfClientOnPeers(peers []peer.PeerIdentifier) *iperf.TestReport {
// 	tries := 0
// 	for _, peer := range peers {
// 		if peer.MyID == 0 {
// 			c := iperf.NewClient(peer.MyIP)
// 			for {
// 				err := c.Start()
// 				if err != nil {
// 					fmt.Printf("failed to start client: %v\n", err)
// 					return nil
// 				} else {
// 					<-c.Done
// 					report := c.Report()
// 					if !strings.Contains(report.String(), "the server is busy running a test") && !strings.Contains(report.String(), "Connection refused") {
// 						fmt.Println(report)
// 						return report
// 					} else {
// 						fmt.Println("Server busy, retrying...")
// 						tries++
// 						if tries >= 10 {
// 							return nil
// 						}
// 					}

// 				}
// 				time.Sleep(10 * time.Second)
// 			}

// 		}
// 	}
// 	return nil
// }

// terminateDoolProcess terminates the specified 'dool' process, logs a fatal error if the process cannot be killed,
// and if successful, retrieves the output from the provided buffer and sets it as the Dool output in the node's benchmark.
func terminateDoolProcess(dool *exec.Cmd, buffer *bytes.Buffer, node *peer.Node) {
	if err := dool.Process.Kill(); err != nil {
		log.Fatal("failed to kill process: ", err)
	} else {
		output := buffer.String()
		// fmt.Printf("Dool Output: %s\n", output)
		node.Benchmark.SetDoolOutput(output)
	}
}

func runDool(buffer *bytes.Buffer) *exec.Cmd {
	fmt.Println("Running DOOL...")
	cmd := exec.Command("dool")
	cmd.Stdout = buffer
	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}

	return cmd
}

func exportBenchmarkToFile(node peer.Node) {
	benchmarkJSONStr := node.GetBenchmark().String()
	filename := fmt.Sprintf("benchmark_%d.json", node.Benchmark.Setup.MyId)
	f, err := os.Create(filename)
	if err != nil {
		fmt.Println("Error creating benchmark.json:", err)
		return
	}
	defer f.Close()
	_, err = f.WriteString(benchmarkJSONStr)
	if err != nil {
		fmt.Println("Error writing to benchmark.json:", err)
	}
}

func initializePeerIdentifier() (toInit peer.PeerIdentifier, F peer.Flags) {
	var sendTransactionS, usePkToClientS, profileS, dontVerifyClusterCommitS, useCertificateCompressionS, useDoolS, alternatingClusterPrepareSenderS, bucketListS, ClusterPrepareToAllAndEnableHashForClusterCommitS string

	flag.StringVar(&usePkToClientS, "use_pk_to_client", "false", "Use public key to client communication")
	flag.IntVar(&myId, "id", -1, "Node ID")
	flag.StringVar(&nodeListS, "nodes", "", "Comma-separated list of node addresses")
	flag.StringVar(&nodeRoleS, "roles", "", "Comma-separated list of node roles (2: leader, 1: peer, 0: client)")
	flag.StringVar(&clusterListS, "clusters", "", "Comma-separated list of cluster IDs for each node (0: cluster 0, 1: cluster 1, etc.)")
	flag.StringVar(&profileS, "profile", "false", "Enable profiling (CPU and trace)")
	flag.IntVar(&runTime, "runtime", 10, "Runtime in seconds for the benchmark")
	flag.IntVar(&vallen, "vallen", 512, "Length of the value to be sent in bytes")
	flag.StringVar(&dontVerifyClusterCommitS, "dont_verify_cluster_commit", "false", "Don't verify cluster commit (for testing purposes)")
	flag.IntVar(&clusterBatchSize, "cluster_batch_size", 300, "Batch size for cluster operations (default: 64)")
	flag.StringVar(&useCertificateCompressionS, "use_certificate_compression", "false", "Use certificate compression (true/false)")
	flag.StringVar(&useDoolS, "use_dool", "false", "Use DOOL (true/false)")
	flag.StringVar(&alternatingClusterPrepareSenderS, "alternateClusterPrepare", "false", "Alternating Cluster Prepare Sender (true/false)")
	flag.StringVar(&bucketListS, "buckets", "", "Comma-separated list of bucket IDs for each node (0: bucket 0, 1: bucket 1, etc.)")
	flag.StringVar(&ClusterPrepareToAllAndEnableHashForClusterCommitS, "cluster_prepare_to_all_and_enable_hash", "false", "Cluster Prepare to all and enable hash for cluster commit (true/false)")
	flag.StringVar(&sendTransactionS, "send_transaction", "true", "Send transactions in cluster pre-prepare messages (true/false)")
	flag.Parse()

	useDool := useDoolS == "true"
	useCertificateCompression = useCertificateCompressionS == "true"
	usePkToClient = usePkToClientS == "true"
	profile = profileS == "true"
	dontVerifyClusterCommit = dontVerifyClusterCommitS == "true"
	alternatingClusterPrepareSender = alternatingClusterPrepareSenderS == "true"
	ClusterPrepareToAllAndEnableHashForClusterCommit := ClusterPrepareToAllAndEnableHashForClusterCommitS == "true"
	sendTransaction := sendTransactionS == "true"

	fmt.Println(os.Args)
	F = peer.Flags{
		MyId:                      myId,
		RunTime:                   runTime,
		NodeListS:                 nodeListS,
		NodeRoleS:                 nodeRoleS,
		ClusterListS:              clusterListS,
		Profile:                   profile,
		Vallen:                    vallen,
		UsePkToClient:             usePkToClient,
		DontVerifyClusterCommit:   dontVerifyClusterCommit,
		ClusterBatchSize:          clusterBatchSize,
		UseCertificateCompression: useCertificateCompression,
		UseDool:                   useDool,
		AlternatingClusterPrepare: alternatingClusterPrepareSender,
		ClusterPrepareToAllAndEnableHashForClusterCommit: ClusterPrepareToAllAndEnableHashForClusterCommit,
		SendTransaction: sendTransaction,
	}

	fmt.Println(F)
	nodeList := strings.Split(nodeListS, ",")
	nodeRoles := strings.Split(nodeRoleS, ",")
	clusterList := strings.Split(clusterListS, ",")
	bucketList := strings.Split(bucketListS, ",")
	myCluster := convertStringToInt(clusterList[myId])
	myRole := convertStringToInt(nodeRoles[myId])
	myIP := nodeList[myId]
	myBucket := convertStringToInt(bucketList[myId])

	numberOfBuckets := 0
	for _, bucket := range bucketList {
		intBucket := convertStringToInt(bucket)
		numberOfBuckets = max(intBucket, numberOfBuckets)
	}
	numberOfBuckets++

	toInit = peer.PeerIdentifier{
		Identifier: peer.Identifier{
			MyID:      myId,
			MyIP:      myIP,
			MyRole:    myRole,
			MyCluster: myCluster,
			MyBucket:  myBucket,
		},
		ClusterBatchSize:          clusterBatchSize,
		MyPeers:                   make([]peer.PeerIdentifier, 0),
		ForeignPeers:              make([]peer.PeerIdentifier, 0),
		MyClients:                 make([]peer.ClientIdentifier, 0),
		ChInSize:                  0,
		UseCryptoForCertificates:  !dontVerifyClusterCommit,
		UseCertificateCompression: useCertificateCompression,
		AlternatingClusterPrepare: alternatingClusterPrepareSender,
		NumberOfBuckets:           numberOfBuckets,
		ClusterPrepareToAllAndEnableHashForClusterCommit: ClusterPrepareToAllAndEnableHashForClusterCommit,
		SendTransaction: sendTransaction,
	}

	for i, _ := range nodeList {

		ip := nodeList[i]

		role := convertStringToInt(nodeRoles[i])
		bucket := convertStringToInt(bucketList[i])

		cluster := convertStringToInt(clusterList[i])

		identifier := peer.Identifier{
			MyID:      i,
			MyIP:      ip,
			MyRole:    role,
			MyCluster: cluster,
			MyBucket:  bucket,
		}

		if cluster == myCluster {
			if role != peer.ROLE_CLIENT {
				toInit.MyPeers = append(toInit.MyPeers, peer.PeerIdentifier{Identifier: identifier})
			} else {
				toInit.MyClients = append(toInit.MyClients, peer.ClientIdentifier{
					Identifier:    identifier,
					UsePkToClient: usePkToClient,
				})
			}
		} else {
			if role != peer.ROLE_CLIENT {
				toInit.ForeignPeers = append(toInit.ForeignPeers, peer.PeerIdentifier{Identifier: identifier})
			} else {
				toInit.MyClients = append(toInit.MyClients, peer.ClientIdentifier{
					Identifier:    identifier,
					UsePkToClient: usePkToClient,
				})
			}
		}

	}
	return
}

func convertStringToInt(s string) int {
	v, err := strconv.Atoi(s)
	if err != nil {
		panic(fmt.Sprintf("Could not convert %s to int \n", s))
	}
	return v
}
