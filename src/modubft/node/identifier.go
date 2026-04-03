package peer

type Identifier struct {
	MyID      int
	MyIP      string
	MyRole    int
	MyCluster int
	MyBucket  int
}
type IdentifierList []Identifier

type ClientIdentifier struct {
	Identifier
	MyPeers       []PeerIdentifier
	UsePkToClient bool
}
type PeerIdentifier struct {
	Identifier

	MyPeers      []PeerIdentifier
	ForeignPeers []PeerIdentifier
	MyClients    []ClientIdentifier

	ChInSize                                         int
	ClusterBatchSize                                 int
	UseCryptoForCertificates                         bool
	UseCertificateCompression                        bool
	AlternatingClusterPrepare                        bool
	NumberOfBuckets                                  int
	ClusterPrepareToAllAndEnableHashForClusterCommit bool
	SendTransaction                                  bool
}

func setCounts(self *Node, toInit PeerIdentifier) {
	self.clientCount = len(toInit.MyClients)
	self.peerCount = len(toInit.MyPeers)
	self.foreignPeerCount = len(toInit.ForeignPeers)

	calculateClusterCount(toInit, self)

	self.allCount = self.clientCount + self.peerCount + self.foreignPeerCount
	self.nodeCount = self.allCount
}

func calculateClusterCount(toInit PeerIdentifier, self *Node) {
	unique := make(map[int]bool)

	for _, peer := range toInit.ForeignPeers {
		unique[peer.MyCluster] = true
	}
	self.numberOfClusters = 1
	for range unique {
		self.numberOfClusters++
	}

}

func (peers IdentifierList) getIDs() (ids []int) {
	ids = make([]int, len(peers))
	for i, peer := range peers {
		ids[i] = peer.MyID
	}
	return

}

func (toInit PeerIdentifier) getChannelInSize() int {
	switch toInit.ChInSize {
	case 0:
		return 1000
	default:
		return toInit.ChInSize
	}
}

func (toInit PeerIdentifier) collectIdentifiers() IdentifierList {
	var count int = len(toInit.MyClients) + len(toInit.MyPeers) + len(toInit.ForeignPeers)
	var Identifiers IdentifierList = make(IdentifierList, count)
	i := 0
	for _, peer := range toInit.MyPeers {
		Identifiers[i] = peer.Identifier
		i++
	}
	for _, client := range toInit.MyClients {
		Identifiers[i] = client.Identifier
		i++
	}
	for _, foreignPeer := range toInit.ForeignPeers {
		Identifiers[i] = foreignPeer.Identifier
		i++
	}
	return Identifiers
}

func initializeNodeRoles(toInit PeerIdentifier, self *Node) {
	i := 0
	self.leaderId = self.myID

	for _, peer := range toInit.MyPeers {
		if self.myRole != ROLE_PEER_LEADER {
			if peer.MyRole == ROLE_PEER_LEADER {
				if self.myRole != ROLE_CLIENT {
					self.leaderId = peer.MyID
				} else {
					if peer.MyBucket == self.myBucket {
						self.leaderId = peer.MyID
					}
				}

			}
		}
		self.nodeRole[i] = peer.MyRole
		i++
	}
	for range toInit.MyClients {
		self.nodeRole[i] = ROLE_CLIENT
		i++
	}
	for _, foreignPeer := range toInit.ForeignPeers {
		self.nodeRole[i] = foreignPeer.MyRole
		i++
	}
}

func (p *Node) generatePeerIDList(peers []PeerIdentifier) []int {
	var IDs []int = make([]int, len(peers))
	for i, peer := range peers {
		IDs[i] = peer.MyID
	}
	return IDs
}
