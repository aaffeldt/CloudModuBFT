package peer

import (
	"crypto/cipher"
	"fmt"
	"slices"
	"time"

	cr "modubft/crypto"
	"net"

	pb "modubft/proto"
)

type PeerState struct {
	nodeId             int
	initialized        bool
	maxPrepared        int
	digestState        []byte
	digestCertificates []byte
	maxEpoch           int
}

type Node struct {
	MACKeys   [][]byte
	MACCipher []cipher.AEAD
	MACNonce  [][]byte

	PubKey     []string
	PKVerifier []cr.Unsigner
	myPrivKey  string
	PKSigner   cr.Signer

	peerCount        int
	nodeCount        int
	foreignPeerCount int
	allCount         int
	coordPresent     bool
	nodeAddr         []string
	nodeRole         []int
	nodeConn         []net.Conn
	clientCount      int
	deadClients      int
	leaderId         int
	coordId          int
	sendTransaction  bool

	//debugging
	reconfigured bool

	cntPrepare           []int
	cntCommit            []int
	cntClusterPrePrepare []int
	cntClusterCommit     [CIRCULAR_BUFFER_SIZE][]bool

	senderOf     []int
	lastPrepared int

	myRole            int
	myEpoch           int
	nextPropSeq       int
	myCommitSeq       int
	myID              int
	myIP              string
	nextCommit        int
	nextClusterCommit int
	wedged            bool
	f                 int

	chInData                     chan *pb.PeerMessage
	chClientInData               chan *pb.PeerMessage
	chCoordInData                chan *pb.PeerMessage
	chForeignData                chan *pb.PeerMessage
	communicateLastCheckpointSeq chan int
	chConEvent                   chan int

	clientMsgDigests           [][]byte
	clientMsgData              [][]byte
	commitLog                  []pb.PeerMessage
	preparedLog                [CIRCULAR_BUFFER_SIZE][]pb.PeerMessage
	myClusterCommits			[CIRCULAR_BUFFER_SIZE]*pb.PeerMessage
	clusterCertificates        [CIRCULAR_BUFFER_SIZE][]*pb.ClusterCertificate
	foreignClusterCertificates [CIRCULAR_BUFFER_SIZE][][]byte
	checkpointMsgs             [2][]pb.PeerMessage

	lastCheckpointSeq    int
	cntPendingCheckpoint int
	lastCheckpointHash   [32]byte
	runningStateHash     [32]byte
	clusterStateHash     [32]byte
	usePKToClient        bool

	myPeers      []PeerIdentifier
	myClients    []ClientIdentifier
	foreignPeers []PeerIdentifier

	peerIDs                                          []int
	peerIDsWithOutMe                                 []int
	foreignPeerIds                                   []int
	foreignPeerIdsPF                                 []int
	clustering                                       bool
	myCluster                                        int
	numberOfClusters                                 int
	ClusterBatchSize                                 int
	useCryptoForCertificates                         bool
	useCertificateCompression                        bool
	Benchmark                                        Benchmark
	stop                                             chan bool
	Stopped                                          bool
	msgSize                                          int
	alternatingClusterPrepare                        bool
	mySortedClusterPosition                          int
	clusterPrepareMessagesReceived                   int
	batchesProposed                                  int
	myBucket                                         int
	numberOfBuckets                                  int
	ClusterPrepareToAllAndEnableHashForClusterCommit bool
	// benchmark BenchmarkData
}

func (p *Node) GetAllCount() int {
	return p.allCount
}
func (p *Node) GetBenchmark() Benchmark {
	return p.Benchmark
}

type Leader struct {
	Node
}

type Follower struct {
	Node
}
type Client struct {
	Node
}

type Coordinator struct {
	Node
}

func NewPeer(toInit PeerIdentifier) Node {

	self := Node{
		myID:                      toInit.MyID,
		myRole:                    toInit.MyRole,
		myIP:                      toInit.MyIP,
		myEpoch:                   0,
		nextPropSeq:               0,
		wedged:                    false,
		reconfigured:              false,
		coordPresent:              false,
		myPeers:                   toInit.MyPeers,
		foreignPeers:              toInit.ForeignPeers,
		myClients:                 toInit.MyClients,
		ClusterBatchSize:          toInit.ClusterBatchSize,
		useCryptoForCertificates:  toInit.UseCryptoForCertificates,
		myCluster:                 toInit.MyCluster,
		stop:                      make(chan bool, 1),
		useCertificateCompression: toInit.UseCertificateCompression,
		msgSize:                   -1,
		alternatingClusterPrepare: toInit.AlternatingClusterPrepare,
		myBucket:                  toInit.MyBucket,
		numberOfBuckets:           toInit.NumberOfBuckets,
		ClusterPrepareToAllAndEnableHashForClusterCommit: toInit.ClusterPrepareToAllAndEnableHashForClusterCommit,
		sendTransaction:              false,
		communicateLastCheckpointSeq: make(chan int, 10),
	}

	self.nextPropSeq = self.myBucket * self.ClusterBatchSize

	self.clustering = len(self.foreignPeers) > 0

	setCounts(&self, toInit)

	self.nodeRole = make([]int, self.allCount)
	initializeNodeRoles(toInit, &self)

	self.peerIDs = self.generatePeerIDList(self.myPeers)

	self.mySortedClusterPosition = self.findClusterPosition()

	self.peerIDsWithOutMe = self.getPeerIDsExcludingSelf()
	self.foreignPeerIds = self.generatePeerIDList(self.foreignPeers)

	fmt.Printf("Peer IDs %v", self.peerIDs)
	identifiers := toInit.collectIdentifiers()

	self.f = (self.peerCount - 1) / 2

	self.nodeConn = make([]net.Conn, self.allCount)

	self.PubKey, self.PKVerifier = generatePublicKeys(identifiers.getIDs())

	self.MACKeys, self.MACCipher, self.MACNonce = generateMACKeys(self.myID, identifiers.getIDs())

	setupCommunicationChannels(toInit, &self)

	self.clientMsgData = make([][]byte, LOG_SIZE)
	self.clientMsgDigests = make([][]byte, LOG_SIZE)
	self.commitLog = make([]pb.PeerMessage, LOG_SIZE)

	totalConnectedPeers := self.peerCount + self.foreignPeerCount
	fmt.Printf("Total connected peers: %d\n", totalConnectedPeers)
	self.checkpointMsgs[1] = make([]pb.PeerMessage, self.allCount)
	for x := 0; x < CIRCULAR_BUFFER_SIZE; x++ {
		self.preparedLog[x] = make([]pb.PeerMessage, self.allCount)
		self.myClusterCommits[x] = nil
		self.clusterCertificates[x] = make([]*pb.ClusterCertificate, 0)
		self.cntClusterCommit[x] = make([]bool, self.numberOfClusters)
		self.foreignClusterCertificates[x] = make([][]byte, self.numberOfClusters)
		for j := 0; j < self.numberOfClusters; j++ {
			self.foreignClusterCertificates[x][j] = make([]byte, 0)
		}
	}

	self.lastCheckpointSeq = 0
	self.cntPendingCheckpoint = 0
	// self.calcPFpeers()
	self.logRoleInitialization()
	self.Benchmark = NewBenchmark(&self)

	return self
}

func (p *Node) findClusterPosition() int {

	peerIDArray := make([]int, len(p.peerIDs))
	copy(peerIDArray, p.peerIDs)
	slices.Sort(peerIDArray)
	for idx, id := range peerIDArray {
		fmt.Printf("Sorted peer ID %d at index %d\n", id, idx)
		if id == p.myID {
			return idx

		}
	}
	if p.myRole != ROLE_CLIENT {
		panic("Could not find my ID in the sorted peer ID list")
	} else {
		return -1
	}

}

func (p *Node) getPeerIDsExcludingSelf() []int {
	p.peerIDsWithOutMe = make([]int, 0)
	for _, peer := range p.peerIDs {
		if peer != p.myID {
			p.peerIDsWithOutMe = append(p.peerIDsWithOutMe, peer)
		}
	}
	return p.peerIDsWithOutMe
}

func (c *Client) Run() {
	time.Sleep(time.Millisecond * 3000)
	c.initConnectionsAsClient()

}

func (c *Client) initConnectionsAsClient() {

	for _, peer := range c.myPeers {
		c.initializeTCPConnection(peer.Identifier, PORT_CLIENT)
	}

	for _, peer := range c.foreignPeers {
		c.initializeTCPConnection(peer.Identifier, PORT_CLIENT)
	}

}

func (co Coordinator) Run() {
	go co.openAndListen(PORT_COORD)

	time.Sleep(time.Millisecond * 2000)

	co.initConnectionsAsPeer()

	waitingFor := co.peerCount + 1
	for waitingFor > 0 {
		<-co.chConEvent
		//fmt.Printf("Connection on port %d\n", conPort)
		waitingFor--
	}

	co.RunCoordLoop()
}

func (co Coordinator) RunCoordLoop() {
	panic("unimplemented")
}

func (p *Node) Run() {
	go p.openAndListen(PORT_CLIENT)
	go p.openAndListen(PORT_PEER)
	go p.openAndListen(PORT_COORD)

	time.Sleep(time.Millisecond * 2000)

	p.initConnectionsAsPeer()

	waitingFor := p.allCount

	if p.coordPresent {
		waitingFor++
	}
	for waitingFor > 0 {
		<-p.chConEvent
		//fmt.Printf("Connection on port %d\n", conPort)
		fmt.Printf("Node %d: Remaining connections to establish: %d\n", p.myID, waitingFor-1)
		waitingFor--
	}

	
	p.RunPeerLoop()
}

/*
RunPeerLoop executes the peer pipeline in an infinite loop. It has the following stages:
- parallel receive, one thread per peer, feeding into a single channel from all other peers (chInData) and a single channel for all clients (chClientData)
- parallel verifier step takes maeesages from a channel and does a FIFO round-robin parallelization -- there's two instances, one for peers and one for clients. Both put verified messages into 'verifOut'.
- decision step -- single thread, defines what happens to each message
- parallel sender step -- sends a given message in parallel to different recipients. This means also computing signatures in parallel.
*/
func (p *Node) RunPeerLoop() {

	/*
		f, err := os.Create("/tmp/abft/cpuprof.prof")
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
	*/

	//runtime.GOMAXPROCS(20)

	peerVerifOut := make(chan *pb.PeerMessage, 1000)
	clientVerifOut := make(chan *pb.PeerMessage, 1000)
	coordVerifOut := make(chan *pb.PeerMessage, 1000)
	foreignPeerVerifOut := make(chan *pb.PeerMessage, 1000)
	decOut := make(chan *MessageWrapper, 1000)
	// preprepateOut := make(chan *MessageWrapper, 1000)

	p.cntPrepare = make([]int, CIRCULAR_BUFFER_SIZE)
	p.cntCommit = make([]int, CIRCULAR_BUFFER_SIZE)

	p.cntClusterPrePrepare = make([]int, CIRCULAR_BUFFER_SIZE)
	p.senderOf = make([]int, CIRCULAR_BUFFER_SIZE)

	go p.peerParallelVerifierStep(p.chInData, peerVerifOut)
	go p.peerParallelVerifierStep(p.chClientInData, clientVerifOut)
	go p.peerParallelVerifierStep(p.chCoordInData, coordVerifOut)
	go p.peerParallelVerifierStep(p.chForeignData, foreignPeerVerifOut)

	go p.peerDecisionStep(peerVerifOut, clientVerifOut, coordVerifOut, foreignPeerVerifOut, decOut)

	// go p.handleClientRequests(clientVerifOut, preprepateOut)

	go p.peerParallelSenderStep(decOut, 0)

	for c,cl := range p.myClients {
		if cl.MyBucket == p.myBucket && cl.MyCluster == p.myCluster &&  p.myRole == ROLE_PEER_LEADER {
			
	
		
		fmt.Printf("Node %d: Connected to client %d\n", p.myID, p.myClients[c].Identifier.MyID)
		// send client response to client 
	
		wrap := MessageWrapper{
			msg:    &pb.PeerMessage{
				FromNodeId:             int32(p.myID),
				MsgId:                  -2,
				EpochId:                0,
				Type:                   pb.MsgType_ClientResponse,
				AttachedData:           []byte{},
			},
			sendTo: []int{p.myClients[c].Identifier.MyID},
		}
		decOut <- &wrap
			
			}
	}

	// go p.peerParallelSenderStep(preprepateOut, 1)
	// In the goroutine waiting:
	fmt.Printf("Recv: %p\n", p.stop)
	<-p.stop
	timestamp := time.Now().Format(time.RFC3339)
	fmt.Printf("Node %d: Received stop signal. Timestamp: %s\n", p.myID, timestamp)
	p.Benchmark.CalculateRunTime()
	aftercalulateTimestamp := time.Now().Format(time.RFC3339)
	fmt.Printf("Node %d: After calculating runtime. Timestamp: %s\n", p.myID, aftercalulateTimestamp)
	p.Benchmark.GetBytesSentSum()
	p.Benchmark.GetBytesReceivedSum()
	p.Benchmark.GetTimeSpent()

}
