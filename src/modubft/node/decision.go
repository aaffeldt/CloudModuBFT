package peer

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"math"
	"math/rand/v2"
	pb "modubft/proto"
	"time"

	"google.golang.org/protobuf/proto"
)

func (p *Node) readNextMessage(chPeerIn chan *pb.PeerMessage, chClientIn chan *pb.PeerMessage, chCoordIn chan *pb.PeerMessage, foreignChIn chan *pb.PeerMessage, wait_for_checkpoint bool, wait_for_preprepare bool) (msg *pb.PeerMessage) {

readNextMessageStart:
	msg = &pb.PeerMessage{}
	if p.wedged == true {
		select {
		case msg = <-chPeerIn:
		case msg = <-chCoordIn:
		}
	} else {

		select {
		case msg = <-chPeerIn:
			return
		case msg = <-chCoordIn:
			return
		case msg = <-foreignChIn:
			return
		default:

		}

		if p.hasReachedCheckpointThreshold() {
			select {
			case msg = <-chPeerIn:

			case msg = <-chCoordIn:

			case msg = <-foreignChIn:

			case <-time.After(time.Millisecond * 1):

				goto readNextMessageStart
			}
		} else {
			select {
			case msg = <-chPeerIn:

			case msg = <-chClientIn:

			case msg = <-chCoordIn:

			case msg = <-foreignChIn:
			}
		}

	}
	return
}

// func (p *Node) handleClientRequests(chClientIn chan *pb.PeerMessage, chOut chan *MessageWrapper) {
// 	msg := &pb.PeerMessage{}
// 	wait_for_checkpoint := false
// 	lastCheckpointSeq := 0

// 	for {

// 		msg = <-chClientIn
// 		// start := time.Now()
// 		if p.myRole == ROLE_PEER_LEADER {
// 			if p.nextPropSeq < 1000 {
// 				timestamp := time.Now().Nanosecond()
// 				fmt.Printf("Received client request %d from %d at time %d\n", p.nextPropSeq, msg.FromNodeId, timestamp)
// 			}
// 			if p.nextPropSeq == p.ClusterBatchSize*p.myBucket {
// 				p.Benchmark.SetStartTime()
// 				fmt.Printf("Received seq: %d Message from %d at time: %s\n", msg.MsgId, msg.FromNodeId, time.Since(p.Benchmark.GetStartTime()))
// 			}
// 			handleClientRequest(p, msg, chOut)
// 			wait_for_checkpoint = (p.nextPropSeq - lastCheckpointSeq) >= (CIRCULAR_BUFFER_SIZE - 1)

// 			select {
// 			case lastCheckpointSeq = <-p.communicateLastCheckpointSeq:
// 				wait_for_checkpoint = (p.nextPropSeq - lastCheckpointSeq) >= (CIRCULAR_BUFFER_SIZE - 1)
// 			default:
// 			}

// 			if wait_for_checkpoint {
// 				lastCheckpointSeq = <-p.communicateLastCheckpointSeq
// 				// fmt.Printf("Received last checkpoint sequence: %d\n", lastCheckpointSeq)
// 			}
// 			// wait_for_preprepare = (p.nextPropSeq - p.nextCommit) >= (10)

// 		} else {
// 			p.notifyClientOfLeader(msg, chOut)
// 		}

// 	}

// 	// p.incrementTimeSpent(fmt.Sprintf("handleClientRequest:%s", msg.Type.String()), float64(time.Since(start).Nanoseconds()))

// }

func (p *Node) peerDecisionStep(chPeerIn chan *pb.PeerMessage, chClientIn chan *pb.PeerMessage, chCoordIn chan *pb.PeerMessage, foreignChIn chan *pb.PeerMessage, chOut chan *MessageWrapper) {
	msg := &pb.PeerMessage{}

	wait_for_checkpoint := false
	wait_for_preprepare := false
	// wait_for_incoming_checkpoint := false
	// preprepareChan := make(chan *pb.PeerMessage, 1000)
	// prepareChan := make(chan *pb.PeerMessage, 1000)
	// commitChan := make(chan *pb.PeerMessage, 1000)
	// clusterpreprepareChan := make(chan *pb.PeerMessage, 1000)
	// clusterprepareChan := make(chan *pb.PeerMessage, 1000)
	// clustercommitChan := make(chan *pb.PeerMessage, 1000)

	for {

		if p.clustering {
			p.handleClusterCommitment(AllTrue(p.cntClusterCommit[p.nextClusterCommit%CIRCULAR_BUFFER_SIZE]), int32(p.nextClusterCommit), chOut)

			p.handleClusterPrePrepare(p.cntClusterPrePrepare[p.nextClusterCommit%CIRCULAR_BUFFER_SIZE] == p.peerCount, int32(p.nextClusterCommit), chOut)
		}

		msg = p.readNextMessage(chPeerIn, chClientIn, chCoordIn, foreignChIn, wait_for_checkpoint, wait_for_preprepare)

		if msg == nil {

			if p.clustering {
				p.handleClusterPrePrepare(p.cntClusterPrePrepare[p.nextClusterCommit%CIRCULAR_BUFFER_SIZE] == p.peerCount, int32(p.nextClusterCommit), chOut)
			}
			continue
		}
		// wait_for_preprepare = (p.nextPropSeq - p.nextCommit) >= (10)

		// now := time.Now()
		/*

			s := time.Now()
		*/
		// msgCnt++
		// s := time.Now()
		rejected := !(p.myEpoch == int(msg.EpochId) && p.wedged == false)
		switch msg.Type {
		case pb.MsgType_ClientRequest:
			// start := time.Now()
			// throw not implemented

			// panic(fmt.Sprintf("Peer received client request message, which is not implemented : %s , from %d , msgId %d, epoch %d , %+v", msg.Type, msg.FromNodeId, msg.MsgId, msg.EpochId, msg))

			if p.myRole == ROLE_PEER_LEADER {
				if wait_for_checkpoint {
					panic(fmt.Sprintf("Waiting for checkpoint. Last checkpoint was at %d, next proposal is at %d\n", p.lastCheckpointSeq, p.nextPropSeq))
				}
				if p.nextPropSeq < 1000 {
					timestamp := time.Now().UnixNano()
					fmt.Printf("Received client request %d from %d at time %d\n", p.nextPropSeq, msg.FromNodeId, timestamp)
				}
				if p.nextPropSeq == p.ClusterBatchSize*p.myBucket {
					p.Benchmark.SetStartTime()
					fmt.Printf("Received seq: %d Message from %d at time: %s\n", msg.MsgId, msg.FromNodeId, time.Since(p.Benchmark.GetStartTime()))
				}
				handleClientRequest(p, msg, chOut)

				wait_for_checkpoint = p.hasReachedCheckpointThreshold()

				// wait_for_preprepare = (p.nextPropSeq - p.nextCommit) >= (10)

			} else {
				p.notifyClientOfLeader(msg, chOut)
			}
			// p.incrementTimeSpent(fmt.Sprintf("handleClientRequest:%s", msg.Type.String()), float64(time.Since(start).Nanoseconds()))

		case pb.MsgType_PrePrepare:
			// start := time.Now()
			if rejected {

				break
			}

			if msg.MsgId == 0 || msg.MsgId == int32(p.ClusterBatchSize) {
				p.Benchmark.SetStartTime()
				// fmt.Printf("Last Checkpoint at node %d for MsgId %d: %d\n", p.myID, msg.MsgId, p.lastCheckpointSeq)
				// fmt.Printf("Received seq: %d MsgType_PrePrepare from %d at time: %s\n", msg.MsgId, msg.FromNodeId, time.Since(p.Benchmark.GetStartTime()))
			}

			if (msg.MsgId - int32(p.lastCheckpointSeq)) >= CIRCULAR_BUFFER_SIZE {
				panic(fmt.Sprintf("ERROR: Too early pre-prepare message received at node %d for MsgId %d last checkpoint at %d", p.myID, msg.MsgId, p.lastCheckpointSeq))
			} else {

				p.handleValidPrePrepare(msg, chOut)
			}
			// p.incrementTimeSpent(fmt.Sprintf("handleValidPrePrepare:%s", msg.Type.String()), float64(time.Since(start).Nanoseconds()))

		case pb.MsgType_Prepare:
			// start := time.Now()
			if rejected {

				break
			}
			if msg.MsgId%1000 == 0 {
				// fmt.Printf("Received seq: %d MsgType_Prepare from %d at time: %s\n", msg.MsgId, msg.FromNodeId, time.Since(p.Benchmark.GetStartTime()))
			}

			if (msg.MsgId - int32(p.lastCheckpointSeq)) >= CIRCULAR_BUFFER_SIZE {
				panic(fmt.Sprintf("ERROR: Too early prepare message received at node %d for MsgId %d last checkpoint at %d", p.myID, msg.MsgId, p.lastCheckpointSeq))
			}
			messageBufferIndex := msg.MsgId % CIRCULAR_BUFFER_SIZE

			p.cntPrepare[messageBufferIndex]++
			p.preparedLog[messageBufferIndex][msg.FromNodeId] = *p.clonePeerMessage(msg)

			isFullyPrepared := p.cntPrepare[messageBufferIndex] == p.peerCount
			// isFullyPrepared := p.cntPrepare[messageBufferIndex] >= (2*p.f + 1)
			if isFullyPrepared {

				if isDigestMismatch(msg, p) {
					fmt.Println("PREPARE ERROR in message digest! Expected \n", msg.AttachedData, "\nand got \n", p.clientMsgDigests[msg.MsgId%LOG_SIZE][:])
					continue
				}

				resp := p.createCommitMessage(msg)

				p.cntPrepare[messageBufferIndex] = 0
				p.lastPrepared = int(msg.MsgId)

				chOut <- resp
			}
			// p.incrementTimeSpent(fmt.Sprintf("MsgType_Prepare:%s", msg.Type.String()), float64(time.Since(start).Nanoseconds()))

		case pb.MsgType_Commit:
			// start := time.Now()
			if rejected {
				break
			}

			// fmt.Printf("Received commit message at node %d from %d for SEQ%d\n", p.myID, msg.FromNodeId, msg.MsgId)

			if isDigestMismatch(msg, p) {
				fmt.Println("COMMIT ERROR in message digest! Expected \n", msg.AttachedData, "\nand got \n", p.clientMsgDigests[msg.MsgId%LOG_SIZE][:])
				continue
			}

			if msg.MsgId%1000 == 0 {
				// fmt.Printf("Received seq: %d MsgType_Commit from %d at time: %s\n", msg.MsgId, msg.FromNodeId, time.Since(p.Benchmark.GetStartTime()))
			}

			if (msg.MsgId - int32(p.lastCheckpointSeq)) >= CIRCULAR_BUFFER_SIZE {
				panic(fmt.Sprintf("ERROR: Too early commit message received at node %d for MsgId %d last checkpoint at %d", p.myID, msg.MsgId, p.lastCheckpointSeq))
			}

			p.cntCommit[msg.MsgId%CIRCULAR_BUFFER_SIZE]++

			// if (msg.MsgId % CIRCULAR_BUFFER_SIZE) == 0 {
			// 	fmt.Printf("MSGID: %d - COUNT: %d\n", msg.MsgId, p.cntCommit[msg.MsgId%CIRCULAR_BUFFER_SIZE])
			// }
			hasReachedConsensus := p.cntCommit[msg.MsgId%CIRCULAR_BUFFER_SIZE] == p.peerCount-1
			// hasReachedConsensus := p.cntCommit[msg.MsgId%CIRCULAR_BUFFER_SIZE] >= (2*p.f + 1)

			if hasReachedConsensus && p.nextCommit == int(msg.MsgId) {

				if p.clustering {
					if msg.MsgId%1000 == 0 {
						// fmt.Printf("Starting ClusterPrePreparing from %d with MsgId %d to %v \n", p.myID, msg.MsgId, p.foreignPeerIdsPF)
					}
					p.startClusterPrePreparing(chOut)
				} else {
					p.startCommitting(msg, chOut)
				}

				// os.Exit(0)

			}

			// p.incrementTimeSpent(fmt.Sprintf("MsgType_Commit:%s", msg.Type.String()), float64(time.Since(start).Nanoseconds()))
		case pb.MsgType_ClusterPrePrepare:
			if (msg.MsgId - int32(p.lastCheckpointSeq)) >= CIRCULAR_BUFFER_SIZE {
				panic(fmt.Sprintf("ERROR: Too early ClusterPrePrepare message received at node %d for MsgId %d last checkpoint at %d", p.myID, msg.MsgId, p.lastCheckpointSeq))
			}
			// start := time.Now()
			p.cntClusterPrePrepare[msg.MsgId%CIRCULAR_BUFFER_SIZE]++

			// p.appendClusterCertificateToChain(msg)

			p.clusterCertificates[msg.MsgId%CIRCULAR_BUFFER_SIZE] = append(p.clusterCertificates[msg.MsgId%CIRCULAR_BUFFER_SIZE], msg.Certificate...)

			// if (msg.MsgId % CIRCULAR_BUFFER_SIZE) == 0 {
			// 	fmt.Printf("cntClusterPrePrepare MSGID: %d - COUNT: %d\n", msg.MsgId, p.cntClusterPrePrepare[msg.MsgId%CIRCULAR_BUFFER_SIZE])
			// }

			msgId := msg.MsgId
			allPeersCommitted := p.cntClusterPrePrepare[msg.MsgId%CIRCULAR_BUFFER_SIZE] == p.peerCount
			p.handleClusterPrePrepare(allPeersCommitted, msgId, chOut)

			// if allPeersCommitted && p.nextClusterCommit == int(msg.MsgId) {
			// 	for {
			// 		p.cntClusterPrePrepare[p.nextClusterCommit%CIRCULAR_BUFFER_SIZE] = 0

			// 		// p.executeClusterCommits(msg, chOut)

			// 		certificateCopy := p.copyClusterCertificate(p.nextClusterCommit)
			// 		if p.myID == p.leaderId {
			// 			msgCP := &pb.PeerMessage{
			// 				FromNodeId:  int32(p.myID),
			// 				MsgId:       int32(p.nextClusterCommit),
			// 				EpochId:     int32(p.myEpoch),
			// 				Type:        pb.MsgType_ClusterPrepare,
			// 				Certificate: certificateCopy,
			// 				ClusterId:   int32(p.myCluster),
			// 			}
			// 			resp := &MessageWrapper{
			// 				msg:    msgCP,
			// 				sendTo: p.foreignPeerIds,
			// 			}
			// 			chOut <- resp

			// 		}

			// 		p.clusterCertificates[p.nextClusterCommit%CIRCULAR_BUFFER_SIZE] = make([]byte, 0)
			// 		p.nextClusterCommit = p.nextClusterCommit + p.ClusterBatchSize
			// 		if !(p.cntClusterPrePrepare[(p.nextClusterCommit)%CIRCULAR_BUFFER_SIZE] == p.peerCount) {
			// 			break
			// 		}
			// 	}
			// }
			// p.incrementTimeSpent(fmt.Sprintf("MsgType_ClusterPrePrepare:%s", msg.Type.String()), float64(time.Since(start).Nanoseconds()))
		case pb.MsgType_ClusterPrepare:
			// start := time.Now()
			if int(msg.MsgId) < p.nextClusterCommit || msg.EpochId != int32(p.myEpoch) {
				break
			}

			if msg.MsgId%1000 == 0 {
				// fmt.Printf("Sending ClusterPrepare from %d with MsgId %d to %v\n", p.myID, msg.MsgId, p.peerIDs)
			}
			if (msg.MsgId - int32(p.lastCheckpointSeq)) >= CIRCULAR_BUFFER_SIZE {
				panic(fmt.Sprintf("ERROR: Too early ClusterPrepare message received at node %d for MsgId %d last checkpoint at %d", p.myID, msg.MsgId, p.lastCheckpointSeq))
			}

			// if (msg.MsgId % CIRCULAR_BUFFER_SIZE) == 0 {
			// fmt.Printf("Received ClusterPrepare MSGID: %d -  cluster %d - number of clusters %d peer count %d = %d - nextClusterCommit %d: \n", msg.MsgId, msg.ClusterId, p.numberOfClusters, p.peerCount, ((p.numberOfClusters) * p.peerCount), p.nextClusterCommit)
			// }

			if p.ClusterPrepareToAllAndEnableHashForClusterCommit {
				p.cntClusterCommit[msg.MsgId%CIRCULAR_BUFFER_SIZE][msg.ClusterId] = true
				if len(p.foreignClusterCertificates[msg.MsgId%CIRCULAR_BUFFER_SIZE][msg.ClusterId]) == 0 {
					p.foreignClusterCertificates[msg.MsgId%CIRCULAR_BUFFER_SIZE][msg.ClusterId] = make([]byte, len(msg.AttachedData))
					copy(p.foreignClusterCertificates[msg.MsgId%CIRCULAR_BUFFER_SIZE][msg.ClusterId], msg.AttachedData)

					// fmt.Printf("Node %d: Stored foreign ClusterCommit certificates for MsgId %d for cluster %d: %v\n", p.myID, msg.MsgId, msg.ClusterId, p.foreignClusterCertificates[msg.MsgId%CIRCULAR_BUFFER_SIZE][msg.ClusterId])
				}

				if msg.ClusterId == int32(p.myCluster) {
					p.myClusterCommits[msg.MsgId%CIRCULAR_BUFFER_SIZE] = p.clonePeerMessageNoAuth(msg)
				}

				allTrue := AllTrue(p.cntClusterCommit[msg.MsgId%CIRCULAR_BUFFER_SIZE])

				p.handleClusterCommitment(allTrue, (msg.MsgId), chOut)
				// if allTrue && p.nextClusterCommit == int(msg.MsgId) {
				// 	if msg.MsgId%1000 == 0 {
				// 		// fmt.Printf("Certificate msg %v\n", msg)
				// 	}

				// 	for {
				// 		for c := 0; c < p.numberOfClusters; c++ {
				// 			committed := false
				// 			if c == p.myCluster {
				// 				p.executeClusterCommits(msg, chOut)
				// 				committed = true
				// 			} else {

				// 				certs := p.foreignClusterCertificates[(p.nextClusterCommit)%CIRCULAR_BUFFER_SIZE][c]
				// 				// fmt.Printf("Foreign certs for cluster %d: %v\n", c, certs)
				// 				p.runningStateHash = p.nextStateHash(p.runningStateHash, certs)
				// 				committed = true
				// 				p.foreignClusterCertificates[(p.nextClusterCommit)%CIRCULAR_BUFFER_SIZE][c] = make([]byte, 0)
				// 			}
				// 			if !committed {
				// 				panic("Could not find certificate for cluster " + fmt.Sprint(c))
				// 			}
				// 		}
				// 		p.cntClusterCommit[p.nextClusterCommit%CIRCULAR_BUFFER_SIZE] = make([]bool, p.numberOfClusters)
				// 		p.clusterCertificates[p.nextClusterCommit%CIRCULAR_BUFFER_SIZE] = make([]*pb.ClusterCertificate, 0)
				// 		p.nextClusterCommit = p.nextClusterCommit + p.ClusterBatchSize

				// 		AllTrue(p.cntClusterCommit[(p.nextClusterCommit)%CIRCULAR_BUFFER_SIZE])
				// 		if !(AllTrue(p.cntClusterCommit[(p.nextClusterCommit)%CIRCULAR_BUFFER_SIZE])) {
				// 			break
				// 		}
				// 	}

				// }
			} else {

				msgCC := &pb.PeerMessage{
					FromNodeId:             int32(p.myID),
					MsgId:                  msg.MsgId,
					EpochId:                int32(p.myEpoch),
					Type:                   pb.MsgType_ClusterCommit,
					Certificate:            msg.Certificate,
					ClusterId:              msg.ClusterId,
					CompressedCertificates: msg.CompressedCertificates,
					MessageSizes:           msg.MessageSizes,
					AttachedData:           msg.AttachedData,
				}
				resp := &MessageWrapper{
					msg:    msgCC,
					sendTo: p.peerIDs,
				}
				// fmt.Printf("Node %d: Sending ClusterCommit with size %d from MsgId %d to %v \n", p.myID, len(msgCC.AttachedData), msg.MsgId, resp.sendTo)

				chOut <- resp
			}

			// p.incrementTimeSpent(fmt.Sprintf("MsgType_ClusterPrepare:%s", msg.Type.String()), float64(time.Since(start).Nanoseconds()))
		case pb.MsgType_ClusterCommit:
			// start := time.Now()
			if int(msg.MsgId) < p.nextClusterCommit {
				break
			}
			if (msg.MsgId - int32(p.lastCheckpointSeq)) >= CIRCULAR_BUFFER_SIZE {
				panic(fmt.Sprintf("ERROR: Too early ClusterCommit message received at node %d for MsgId %d last checkpoint at %d", p.myID, msg.MsgId, p.lastCheckpointSeq))
			}

			p.cntClusterCommit[msg.MsgId%CIRCULAR_BUFFER_SIZE][msg.ClusterId] = true
			// fmt.Printf("Node %d: Received ClusterCommit from %d with MsgId %d for cluster %d, %v \n", p.myID, msg.FromNodeId, msg.MsgId, msg.ClusterId, msg)
			if len(p.foreignClusterCertificates[msg.MsgId%CIRCULAR_BUFFER_SIZE][msg.ClusterId]) == 0 {
				p.foreignClusterCertificates[msg.MsgId%CIRCULAR_BUFFER_SIZE][msg.ClusterId] = make([]byte, len(msg.AttachedData))
				copy(p.foreignClusterCertificates[msg.MsgId%CIRCULAR_BUFFER_SIZE][msg.ClusterId], msg.AttachedData)
			}

			if msg.ClusterId == int32(p.myCluster) {
				p.myClusterCommits[msg.MsgId%CIRCULAR_BUFFER_SIZE] = p.clonePeerMessageNoAuth(msg)
			}

			// if (msg.MsgId % CIRCULAR_BUFFER_SIZE) == 0 {
			// 	fmt.Printf("cntClusterCommit MSGID: %d - COUNT %v - cluster %d - number of clusters %d peer count %d = %d - nextClusterCommit %d: \n", msg.MsgId, p.cntClusterCommit[msg.MsgId%CIRCULAR_BUFFER_SIZE], msg.ClusterId, p.numberOfClusters, p.peerCount, ((p.numberOfClusters) * p.peerCount), p.nextClusterCommit)
			// }
			// allValid := true

			allTrue := AllTrue(p.cntClusterCommit[msg.MsgId%CIRCULAR_BUFFER_SIZE])

			msgId := msg.MsgId

			p.handleClusterCommitment(allTrue, msgId, chOut)

			// p.incrementTimeSpent(fmt.Sprintf("MsgType_ClusterCommit:%s", msg.Type.String()), float64(time.Since(start).Nanoseconds()))
		case pb.MsgType_Checkpoint:
			// start := time.Now()
			if rejected {
				break
			}
			// fmt.Printf("Got checkpoint at peer %d from peer %d at SEQ%d\n", p.myID, int(msg.FromNodeId), int(msg.MsgId))
			if int(msg.MsgId) == p.lastCheckpointSeq+CHECKPOINT_INTERVAL {
				p.cntPendingCheckpoint++
				p.checkpointMsgs[1][msg.FromNodeId] = *p.clonePeerMessage(msg)
				if p.cntPendingCheckpoint == p.peerCount {
					p.lastCheckpointSeq = int(msg.MsgId)
					// fmt.Printf("Node %d: Reached checkpoint at sequence %d\n", p.myID, p.lastCheckpointSeq)
					// fmt.Printf("Node %d: Checkpoint at sequence %d with state hash %x\n", p.myID, p.lastCheckpointSeq, sha256.Sum256(msg.AttachedData))
					if p.myRole == ROLE_PEER_LEADER {
						// p.communicateLastCheckpointSeq <- int(msg.MsgId)

					}
					// fmt.Printf("Peer %d checkpointed at %d\n", p.myID, p.lastCheckpointSeq)
					p.cntPendingCheckpoint = 0
					p.checkpointMsgs[0] = p.checkpointMsgs[1]
					wait_for_checkpoint = p.hasReachedCheckpointThreshold()
					// wait_for_checkpoint = ((1 + p.nextPropSeq) - p.lastCheckpointSeq) > CIRCULAR_BUFFER_SIZE
				}
			}
			// p.incrementTimeSpent(fmt.Sprintf("MsgType_Checkpoint:%s", msg.Type.String()), float64(time.Since(start).Nanoseconds()))
		case pb.MsgType_Probe:

			p.handleProbeMessage(msg, chOut)
		case pb.MsgType_NewConfig:

			if !p.wedged {
				rejected = true
				break
			}
			p.handleNewConfigMessage(msg, chOut)

		case pb.MsgType_SyncState:
			if !p.wedged {
				rejected = true
				break
			}
			p.handleSyncStateMessage(msg, chOut)

		}

		if rejected {
			fmt.Printf("Rejected %d from %d, msgId %d, epoch %d\n", msg.Type, msg.FromNodeId, msg.MsgId, msg.EpochId)
		} else {

			// p.benchmark.addToincrementTimeSpent(Measure{
			// 	Type:     msg.Type,
			// 	Kind:     DECIDE,
			// 	Duration: time.Since(now),
			// })
		}

	}
}

func (p *Node) handleClusterCommitment(allTrue bool, msgId int32, chOut chan *MessageWrapper) {
	if allTrue && p.nextClusterCommit == int(msgId) {

		for {
			for c := 0; c < p.numberOfClusters; c++ {
				committed := false
				if c == p.myCluster {
					p.executeClusterCommits(p.myClusterCommits[msgId%CIRCULAR_BUFFER_SIZE], chOut)
					committed = true
				} else {

					certs := p.foreignClusterCertificates[(p.nextClusterCommit)%CIRCULAR_BUFFER_SIZE][c]
					// fmt.Printf("Foreign certs for cluster %d: %v\n", c, certs)
					p.runningStateHash = p.nextStateHash(p.runningStateHash, certs)
					committed = true
					p.foreignClusterCertificates[(p.nextClusterCommit)%CIRCULAR_BUFFER_SIZE][c] = make([]byte, 0)
				}
				if !committed {
					panic("Could not find certificate for cluster " + fmt.Sprint(c))
				}
			}
			p.cntClusterCommit[p.nextClusterCommit%CIRCULAR_BUFFER_SIZE] = make([]bool, p.numberOfClusters)
			p.clusterCertificates[p.nextClusterCommit%CIRCULAR_BUFFER_SIZE] = make([]*pb.ClusterCertificate, 0)
			p.nextClusterCommit = p.nextClusterCommit + p.ClusterBatchSize

			AllTrue(p.cntClusterCommit[(p.nextClusterCommit)%CIRCULAR_BUFFER_SIZE])
			if !(AllTrue(p.cntClusterCommit[(p.nextClusterCommit)%CIRCULAR_BUFFER_SIZE])) {
				break
			}
		}

	}
}

func (p *Node) handleClusterPrePrepare(allPeersCommitted bool, msgId int32, chOut chan *MessageWrapper) {

	allowedToSend := p.clusterPrepareAllowedToSend(msgId)
	
	if allPeersCommitted && allowedToSend  {
		p.cntClusterPrePrepare[msgId%CIRCULAR_BUFFER_SIZE] = 0

		// p.executeClusterCommits(msg, chOut)
		// certificateCopy := p.copyClusterCertificate(int(msg.MsgId))
		// certificateCopy := p.clusterCertificates[msg.MsgId%CIRCULAR_BUFFER_SIZE]
		var state []byte
		var certificateSizes []int32
		startId := msgId

		if p.alternatingClusterPrepare {
			if int(math.Floor(float64(startId % (int32(p.peerCount) * int32(p.ClusterBatchSize)))/float64(p.ClusterBatchSize) )) == (p.mySortedClusterPosition) {

				state = make([]byte, 0)
				certificateSizes = make([]int32, p.ClusterBatchSize)
				for i := 0; i < p.ClusterBatchSize; i++ {
					clientMessage := p.clientMsgData[(int(startId)+i)%LOG_SIZE]
					state = append(state, clientMessage...)[:]
					certificateSizes[i] = int32(len(clientMessage))
				}

				resp := p.createClusterPrepareMessage(msgId, state, certificateSizes)
				chOut <- resp
			}
		} else {
			if p.myID == p.leaderId {
				bucketIndex := msgId % (int32(p.numberOfBuckets) * int32(p.ClusterBatchSize))
				bucketGroupIndex := int(math.Floor(float64(bucketIndex) / float64(p.ClusterBatchSize)))

				if bucketGroupIndex == p.myBucket {

					state = make([]byte, 0)
					certificateSizes = make([]int32, p.ClusterBatchSize)
					for i := 0; i < p.ClusterBatchSize; i++ {
						clientMessage := p.clientMsgData[(int(startId)+i)%LOG_SIZE]
						state = append(state, clientMessage...)[:]
						certificateSizes[i] = int32(len(clientMessage))
					}

					// fmt.Printf("Proposed %d: %d : %d \n", msg.MsgId, bucketGroupIndex, bucketIndex)
					resp := p.createClusterPrepareMessage(msgId, state, certificateSizes)
					// proposed[msg.MsgId%CIRCULAR_BUFFER_SIZE] = false
					chOut <- resp
				}

			}
		}
		p.clusterPrepareMessagesReceived++

		if p.ClusterPrepareToAllAndEnableHashForClusterCommit {

			msgCC := &pb.PeerMessage{
				FromNodeId: int32(p.myID),
				MsgId:      int32(msgId),
				EpochId:    int32(p.myEpoch),
				Type:       pb.MsgType_ClusterPrepare,
				ClusterId:  int32(p.myCluster),
			}
			resp := &MessageWrapper{
				msg:    msgCC,
				sendTo: make([]int, 1),
			}
			resp.sendTo[0] = p.myID
			chOut <- resp
		} else {

			msgCC := &pb.PeerMessage{
				FromNodeId: int32(p.myID),
				MsgId:      int32(msgId),
				EpochId:    int32(p.myEpoch),
				Type:       pb.MsgType_ClusterCommit,
				ClusterId:  int32(p.myCluster),
			}
			resp := &MessageWrapper{
				msg:    msgCC,
				sendTo: make([]int, 1),
			}
			resp.sendTo[0] = p.myID
			chOut <- resp
		}

		// p.clusterCertificates[p.nextClusterCommit%CIRCULAR_BUFFER_SIZE] = make([]byte, 0)
		// p.nextClusterCommit = p.nextClusterCommit + p.ClusterBatchSize
		// if !(p.cntClusterPrePrepare[(p.nextClusterCommit)%CIRCULAR_BUFFER_SIZE] == p.peerCount) {
		// 	break
		// }

	}
}

func (p *Node) clusterPrepareAllowedToSend(msgId int32) bool {
	allowedToSend := false

	for i := 0; i < ALLOWED_PARALLEL_CLUSTER_PREPARES; i++ {
		allowedToSend = (((i*p.ClusterBatchSize) +  p.nextClusterCommit) == int(msgId))
		if allowedToSend {
			break
		}
	}
	return allowedToSend
}

func (p *Node) hasReachedCheckpointThreshold() bool {

	return (p.nextPropSeq - p.lastCheckpointSeq) >= (CHECKPOINT_INTERVAL + p.ClusterBatchSize)
}

func (p *Node) createClusterPrepareMessage(msgId int32, state []byte, certificateSizes []int32) *MessageWrapper {
	msgCP := &pb.PeerMessage{
		FromNodeId:   int32(p.myID),
		MsgId:        int32(msgId),
		EpochId:      int32(p.myEpoch),
		Type:         pb.MsgType_ClusterPrepare,
		Certificate:  p.clusterCertificates[msgId%CIRCULAR_BUFFER_SIZE],
		AttachedData: state,
		MessageSizes: certificateSizes,
		ClusterId:    int32(p.myCluster),
	}

	selectedForeignPeers := p.randomForeignPeerIdsPF()

	resp := &MessageWrapper{
		msg:    msgCP,
		sendTo: selectedForeignPeers,
	}
	return resp
}

func (p *Node) randomForeignPeerIdsPF() []int {
	if p.ClusterPrepareToAllAndEnableHashForClusterCommit {
		return p.foreignPeerIds
	}

	ids := make([]int, (p.numberOfClusters-1)*(p.f+1))
	randomized_peer_order := rand.Perm(len(p.foreignPeers))

	i := 0
	for c := 0; c < p.numberOfClusters; c++ {
		count := 0

		for _, idx := range randomized_peer_order {
			peer := p.foreignPeers[idx]
			if peer.MyCluster == c && count < (p.f+1) {

				ids[i] = peer.MyID
				i++
				count++
			}
		}
	}
	return ids
}

// newMethod selects up to (f+1) peer IDs from each cluster and stores them in foreignPeerIdsPF.
// It iterates through all clusters and, for each cluster, collects the IDs of peers belonging to that cluster
// until (f+1) peers are found. The resulting slice has a length of numberOfClusters * (f+1).
func (p *Node) calcPFpeers() {
	panic("Deprecated, use randomForeignPeerIdsPF instead")
	ids := make([]int, p.numberOfClusters*(p.f+1))
	i := 0
	for c := 0; c < p.numberOfClusters; c++ {
		count := 0
		for _, peer := range p.foreignPeers {
			if peer.MyCluster == c && count < (p.f+1) {

				ids[i] = peer.MyID
				i++
				count++
			}
		}
	}
	p.foreignPeerIdsPF = ids
}

func AllTrue(arr []bool) bool {
	allTrue := true
	for _, v := range arr {
		if !v {
			allTrue = false
			break
		}
	}
	return allTrue
}

func (p *Node) startClusterPrePreparing(chOut chan *MessageWrapper) {
	for {

		// p.clusterStateHash = p.nextStateHash(p.clusterStateHash, p.clientMsgDigests[p.nextCommit%LOG_SIZE])
		// p.clusterStateHash = p.clientMsgDigests[p.nextCommit%LOG_SIZE]

		isCommitThresholdReached := ((p.nextCommit + 1) % p.ClusterBatchSize) == 0
		if isCommitThresholdReached {
			// fmt.Printf("p.clusterStateHash %x , %x  \n", p.clusterStateHash, p.clientMsgDigests[p.nextCommit%LOG_SIZE])
			// currentClusterState := p.getCurrentClusterState()
			startId := int32(p.nextCommit) - int32(p.ClusterBatchSize-1)
			// p.cntClusterCommit[startId%CIRCULAR_BUFFER_SIZE] = make([]bool, p.numberOfClusters)
			// fmt.Printf("currentClusterState %x \n", currentClusterState)
			// state := make([]byte, 0),
			// for i := 0; i < p.ClusterBatchSize; i++ {
			// 	state = append(state, p.clientMsgDigests[(int(startId)+i)%LOG_SIZE]...)[:]
			// }

			digestLen := len(p.clientMsgDigests[0])
			state := make([]byte, p.ClusterBatchSize*digestLen)
			if p.sendTransaction {
				// state := make([]byte, 0)
				// startId := msg.MsgId
				// certificateSizes := make([]int32, p.ClusterBatchSize)
				// for i := 0; i < p.ClusterBatchSize; i++ {
				// 	clientMessage := p.clientMsgData[(int(startId)+i)%LOG_SIZE]
				// 	state = append(state, clientMessage...)[:]
				// 	certificateSizes[i] = int32(len(clientMessage))
				// }
				transactions := make([]byte, 0)
				messageSizes := make([]int32, p.ClusterBatchSize)
				for i := 0; i < p.ClusterBatchSize; i++ {
					// copy(state[i*digestLen:], p.clientMsgDigests[(int(startId)+i)%LOG_SIZE])
					clientMessage := p.clientMsgData[(int(startId)+i)%LOG_SIZE]
					transactions = append(transactions, clientMessage...)[:]
					messageSizes[i] = int32(len(clientMessage))
				}
				p.sendClusterPrePrepareMessageWithTransactions(startId, transactions, chOut, messageSizes)
			} else {
				for i := 0; i < p.ClusterBatchSize; i++ {
					copy(state[i*digestLen:], p.clientMsgDigests[(int(startId)+i)%LOG_SIZE])
				}
				p.sendClusterPrePrepareMessage(startId, state, chOut)
			}

		}

		p.nextCommit++
		if !(p.cntCommit[(p.nextCommit)%CIRCULAR_BUFFER_SIZE] == p.peerCount-1) {
			break
		}
	}
}

func (p *Node) sendClusterPrePrepareMessageWithTransactions(startId int32, transactions []byte, chOut chan *MessageWrapper, messageSizes []int32) {
	msgCPP := &pb.PeerMessage{
		FromNodeId: int32(p.myID),
		MsgId:      startId,
		EpochId:    int32(p.myEpoch),
		Type:       pb.MsgType_ClusterPrePrepare,
		// Certificate:  certificates,
		MessageSizes: messageSizes,
		AttachedData: transactions,
	}
	resp := &MessageWrapper{
		msg:    msgCPP,
		sendTo: p.peerIDs,
	}
	// fmt.Printf("Sending ClusterPrePrepare with transactions of size %d from %d with MsgId %d to %v at time: %s\n", len(transactions), p.myID, startId, p.peerIDs, time.Since(p.Benchmark.GetStartTime()))

	chOut <- resp

}

func (p *Node) getCurrentClusterState() []byte {
	currentClusterState := make([]byte, len(p.clusterStateHash))
	copy(currentClusterState, p.clusterStateHash[:])
	return currentClusterState
}

func (p *Node) sendClusterPrePrepareMessage(startId int32, currentClusterState []byte, chOut chan *MessageWrapper) {
	// fmt.Printf("sendClusterPrePrepareMessage currentClusterState %x \n", currentClusterState)
	// copyState := make([]byte, len(currentClusterState))
	// copy(copyState, currentClusterState)
	copyState := currentClusterState

	// certificate := pb.ClusterCertificate{
	// 	Certificate: copyState,
	// 	FromNodeId:  int32(p.myID),
	// }
	// certificates := make([]*pb.ClusterCertificate, 1)
	// certificates[0] = &certificate

	msgCPP := &pb.PeerMessage{
		FromNodeId: int32(p.myID),
		MsgId:      startId,
		EpochId:    int32(p.myEpoch),
		Type:       pb.MsgType_ClusterPrePrepare,
		// Certificate:  certificates,
		AttachedData: copyState,
	}

	resp := &MessageWrapper{
		msg:    msgCPP,
		sendTo: p.peerIDs,
	}
	if (startId) == 0 {
		// nextCommitCopy := p.nextCommit
		// fmt.Printf("Sending ClusterPrePrepare message from node %d for commit %d to %v at time: %s\n", p.myID, nextCommitCopy, p.peerIDs, time.Since(p.Benchmark.GetStartTime()))
	}
	chOut <- resp
}

func (p *Node) startCommitting(msg *pb.PeerMessage, chOut chan *MessageWrapper) {

	for {

		p.commit(p.nextCommit, msg, chOut)

		p.nextCommit++
		if !(p.cntCommit[(p.nextCommit)%CIRCULAR_BUFFER_SIZE] == p.peerCount-1) {
			break
		}
	}

}

func (p *Node) commit(msgId int, msg *pb.PeerMessage, chOut chan *MessageWrapper) {
	// fmt.Printf("Commiting %d\n", msgId)
	if msgId%1000 == 0 {
		// elapsed_time := time.Since(p.Benchmark.GetStartTime())
		// elapsed_time_in_seconds := elapsed_time.Seconds()
		// fmt.Printf("Seq: %d commit at node %d after %f seconds at time: %s\n", msgId, p.myID, elapsed_time_in_seconds, elapsed_time)
	}
	if msgId-p.lastCheckpointSeq >= CIRCULAR_BUFFER_SIZE {
		panic(fmt.Sprintf("ERROR: Too early commit execution at node %d for MsgId %d last checkpoint at %d", p.myID, msgId, p.lastCheckpointSeq))
	}
	resp := p.createClientResponse(msgId)
	//if p.reconfigured && msg.MsgId%5000 == 1 {
	//	fmt.Printf("Committed %d after reconfigured. MyId: %d, from %d\n", int32(msg.MsgId), int32(p.myID), msg.FromNodeId)
	//}
	wrappedMsgId := msgId % CIRCULAR_BUFFER_SIZE
	p.cntCommit[wrappedMsgId] = 0
	p.commitLog[wrappedMsgId] = *p.clonePeerMessageNoAuth(msg)

	/** STATS
	commitedMessages++
	if commitedMessages%2000 == 0 {
		fmt.Println("Throughput at node ", p.myID, ": ", (p.ClusterBatchSize0000000 * float32(commitedMessages) / float32(time.Since(startTime))), " (", (time.Since(startTime)), ")")
		commitedMessages = 0
		startTime = time.Now()
	}
	**/
	chOut <- resp

	// now := time.Now()
	p.runningStateHash = p.nextStateHash(p.runningStateHash, p.clientMsgDigests[msgId%LOG_SIZE])
	// p.benchmark.addToincrementTimeSpent(Measure{
	// 	Type:     msg.Type,
	// 	Kind:     HASH,
	// 	Duration: time.Since(now),
	// })

	nextCheckpointOffset := msgId - p.lastCheckpointSeq
	if nextCheckpointOffset == CHECKPOINT_INTERVAL {
		p.sendCheckpointToPeers(msgId, chOut)
	}

}

func (p *Node) handleSyncStateMessage(msg *pb.PeerMessage, chOut chan *MessageWrapper) {
	fmt.Printf("Received SyncState. Peer: %d at time: %s\n", p.myID, time.Since(p.Benchmark.GetStartTime()))

	p.reconfigured = true

	payload := &pb.PayloadSyncState{}
	err1 := proto.Unmarshal(msg.AttachedData, payload)
	if err1 != nil {
		fmt.Printf("%s\n", err1)
	}

	payloadNewConfig := &pb.PayloadNewConfig{}
	err1 = proto.Unmarshal(payload.NewConfigMessage.AttachedData, payloadNewConfig)
	if err1 != nil {
		fmt.Printf("%s\n", err1)
	}

	msgBytes, err2 := proto.Marshal(payload.Certificate)
	if err2 != nil {
		fmt.Printf("%s\n", err2)
	}
	msgHash := sha256.Sum256(msgBytes)

	//if p.VerifySig(payload.NewConfigMessage) && bytes.Compare(msgHash[:], payloadNewConfig.DigestP) == 0 {
	if bytes.Compare(msgHash[:], payloadNewConfig.DigestP) == 0 {

		leaderCheckpointSeq := int(payload.Certificate.Certificates[0].MsgId)
		leaderPreparedSeq := int(payload.Certificate.CntPrepared) + leaderCheckpointSeq
		if p.lastCheckpointSeq != leaderCheckpointSeq {
			fmt.Printf("Different snapshot. Leader: %d, Mine: %d\n", leaderCheckpointSeq, p.lastCheckpointSeq)
		}
		if p.lastPrepared+1 == leaderPreparedSeq {
			p.lastPrepared++
			index := int(payload.Certificate.CntPrepared) * p.peerCount
			//fmt.Printf("client %d for %d\n", int(payload.Clients[len(payload.Clients)-1]), leaderPreparedSeq)
			p.senderOf[leaderPreparedSeq%CIRCULAR_BUFFER_SIZE] = int(payload.Clients[len(payload.Clients)-1])
			for x := 0; x < p.peerCount; x++ {
				p.preparedLog[leaderPreparedSeq%(CIRCULAR_BUFFER_SIZE)][x] = *p.clonePeerMessage(payload.Certificate.Certificates[index+x])
			}

		} else if p.lastPrepared+1 < leaderPreparedSeq {
			fmt.Printf("WEIRD: Different number of prepared. Leader up to: %d, Mine: %d\n", leaderPreparedSeq, p.lastPrepared)
		} else {
			fmt.Printf("Same number of prepared. Leader up to: %d, Mine: %d\n", leaderPreparedSeq, p.lastPrepared)
		}
		//Resetting the internal state
		p.cntPrepare = make([]int, CIRCULAR_BUFFER_SIZE)
		p.cntCommit = make([]int, CIRCULAR_BUFFER_SIZE)
		p.cntPendingCheckpoint = 0
		p.nextCommit = leaderCheckpointSeq + 1
		//commitedMessages = 0

		//Setting new role and unwedging
		p.wedged = false
		p.myRole = ROLE_PEER_FOLLOWER
		p.myEpoch = int(msg.EpochId)
		p.leaderId = int(payloadNewConfig.Leader)

		//send commit after SyncState message
		for x := p.lastCheckpointSeq + 1; x <= p.lastPrepared; x++ {

			msgP := &pb.PeerMessage{}
			msgP.Type = pb.MsgType_Commit
			msgP.FromNodeId = int32(p.myID)
			msgP.MsgId = int32(x)
			msgP.EpochId = int32(p.myEpoch)

			msgP.AttachedData = p.preparedLog[x%(CIRCULAR_BUFFER_SIZE)][0].AttachedData

			resp := &MessageWrapper{sendTo: make([]int, p.peerCount-1), msg: msgP}
			c := 0
			for x := 0; x < p.peerCount; x++ {
				if x != p.myID {
					resp.sendTo[c] = x
					c++
				}
			}

			chOut <- resp
		}

	}
}

func (p *Node) handleNewConfigMessage(msg *pb.PeerMessage, chOut chan *MessageWrapper) {
	msgP := &pb.PayloadNewConfig{}
	err := proto.Unmarshal(msg.AttachedData, msgP)
	if err != nil {
		fmt.Printf("%s\n", err)
	}

	if int(msgP.Leader) == p.myID {

		fmt.Printf("Received NewConfig. Peer: %d at time: %s\n", p.myID, time.Since(p.Benchmark.GetStartTime()))

		p.reconfigured = true

		//Resetting the internal state
		p.nextPropSeq = p.lastPrepared + 1
		p.nextCommit = p.lastCheckpointSeq + 1

		p.cntPrepare = make([]int, CIRCULAR_BUFFER_SIZE)
		p.cntCommit = make([]int, CIRCULAR_BUFFER_SIZE)

		p.cntPendingCheckpoint = 0

		//commitedMessages = 0

		//Setting new role and unwedging
		p.wedged = false
		p.myRole = ROLE_PEER_LEADER
		p.myEpoch = int(msg.EpochId)
		p.leaderId = p.myID

		//Sending SyncState to new followers
		msgR := &pb.PeerMessage{}
		msgR.Type = pb.MsgType_SyncState
		msgR.FromNodeId = int32(p.myID)
		msgR.MsgId = 0
		msgR.EpochId = int32(p.myEpoch)

		//Creating the message payload, including the NewConfig message and the payload of the probe_ack msg that the peer sent to the reconfiguration service
		payload := &pb.PayloadSyncState{}
		payload.NewConfigMessage = msg
		payload.Certificate = &pb.CertificateMessage{}

		payload.Certificate.Initialized = true

		cntPreparedCertificates := p.lastPrepared - p.lastCheckpointSeq

		payload.Certificate.CntPrepared = int32(cntPreparedCertificates)
		payload.Certificate.Certificates = make([]*pb.PeerMessage, cntPreparedCertificates*p.peerCount+p.peerCount)
		payload.Clients = make([]int32, cntPreparedCertificates)

		for x := 0; x < p.peerCount; x++ {
			payload.Certificate.Certificates[x] = &p.checkpointMsgs[0][x]
		}
		//fmt.Printf("checkpoint sequence %d, starting prepared %d, end prepared %d\n", p.lastCheckpointSeq, p.lastCheckpointSeq+1, p.lastCheckpointSeq+cntPreparedCertificates+1)
		index := 1
		for x := p.lastCheckpointSeq + 1; x < p.lastCheckpointSeq+cntPreparedCertificates+1; x++ {
			for y := 0; y < p.peerCount; y++ {
				payload.Certificate.Certificates[index*p.peerCount+y] = &p.preparedLog[x%(CIRCULAR_BUFFER_SIZE)][y]
				//fmt.Printf("prepared msg: %d, peer %d, index %d, size %d\n", x, y, index, cntPreparedCertificates*p.peerCount+p.peerCount)
			}
			payload.Clients[index-1] = int32(p.senderOf[x%CIRCULAR_BUFFER_SIZE])
			index++
		}

		rawMsg, err := proto.Marshal(payload)
		if err != nil {
			fmt.Printf("Marshalling certificates error: %s\n", err)
		}
		msgR.AttachedData = rawMsg

		resp := &MessageWrapper{sendTo: make([]int, p.peerCount-1), msg: msgR}
		c := 0
		for x := 0; x < p.peerCount; x++ {
			if x != p.myID {
				resp.sendTo[c] = x
				c++
			}
		}

		chOut <- resp

		//send commit after SyncState message
		for x := p.lastCheckpointSeq + 1; x <= p.lastPrepared; x++ {

			msgP := &pb.PeerMessage{}
			msgP.Type = pb.MsgType_Commit
			msgP.FromNodeId = int32(p.myID)
			msgP.MsgId = int32(x)
			msgP.EpochId = int32(p.myEpoch)

			msgP.AttachedData = p.preparedLog[x%(CIRCULAR_BUFFER_SIZE)][0].AttachedData

			resp := &MessageWrapper{sendTo: make([]int, p.peerCount-1), msg: msgP}
			c := 0
			for x := 0; x < p.peerCount; x++ {
				if x != p.myID {
					resp.sendTo[c] = x
					c++
				}
			}

			chOut <- resp
		}
	}
}

func (p *Node) handleProbeMessage(msg *pb.PeerMessage, chOut chan *MessageWrapper) {
	fmt.Printf("Coordinator says: %s at time: %s\n", string(msg.AttachedData), time.Since(p.Benchmark.GetStartTime()))

	if p.nodeRole[msg.FromNodeId] != ROLE_COORDINATOR {
		fmt.Println("ERROR!!!!")
	}

	msgR := &pb.PeerMessage{}
	msgR.Type = pb.MsgType_ProbeAck
	msgR.FromNodeId = int32(p.myID)
	msgR.MsgId = msg.MsgId
	msgR.EpochId = msg.EpochId

	cntPreparedCertificates := p.lastPrepared - p.lastCheckpointSeq

	msgC := &pb.CertificateMessage{}
	msgC.Initialized = true
	msgC.CntPrepared = int32(cntPreparedCertificates)
	msgC.Certificates = make([]*pb.PeerMessage, cntPreparedCertificates*p.peerCount+p.peerCount)

	for x := 0; x < p.peerCount; x++ {
		msgC.Certificates[x] = &p.checkpointMsgs[0][x]
	}
	//fmt.Printf("checkpoint sequence %d, starting prepared %d, end prepared %d\n", p.lastCheckpointSeq, p.lastCheckpointSeq+1, p.lastCheckpointSeq+cntPreparedCertificates+1)
	index := 1
	for x := p.lastCheckpointSeq + 1; x < p.lastCheckpointSeq+cntPreparedCertificates+1; x++ {
		for y := 0; y < p.peerCount; y++ {
			msgC.Certificates[index*p.peerCount+y] = &p.preparedLog[x%(CIRCULAR_BUFFER_SIZE)][y]
			//fmt.Printf("prepared msg: %d, peer %d, index %d, size %d\n", x, y, index, cntPreparedCertificates*p.peerCount+p.peerCount)
		}
		index++
	}
	rawMsg, err := proto.Marshal(msgC)
	if err != nil {
		fmt.Printf("Marshalling certificates error: %s\n", err)
	}

	msgR.AttachedData = rawMsg

	resp := &MessageWrapper{sendTo: make([]int, 1), msg: msgR}
	resp.sendTo[0] = int(msg.FromNodeId)

	chOut <- resp

	p.wedged = true
}

func (p *Node) sendCheckpointToPeers(msgId int, chOut chan *MessageWrapper) {
	p.lastCheckpointHash = p.runningStateHash
	msgP := &pb.PeerMessage{}
	msgP.Type = pb.MsgType_Checkpoint
	msgP.FromNodeId = int32(p.myID)
	msgP.MsgId = int32(msgId)
	msgP.EpochId = int32(p.myEpoch)

	msgP.AttachedData = p.lastCheckpointHash[:]

	// fmt.Printf("Attempting checkpoint on peer %d at SEQ%d\n", p.myID, p.nextCommit)

	resp := &MessageWrapper{sendTo: make([]int, p.peerCount), msg: msgP}
	resp.sendTo = p.peerIDs

	chOut <- resp
}

func (p *Node) createCommitMessage(msg *pb.PeerMessage) *MessageWrapper {
	msgP := &pb.PeerMessage{}
	msgP.Type = pb.MsgType_Commit
	msgP.FromNodeId = int32(p.myID)
	msgP.MsgId = msg.MsgId
	msgP.EpochId = msg.EpochId
	msgP.AttachedData = msg.AttachedData

	resp := &MessageWrapper{sendTo: make([]int, p.peerCount-1), msg: msgP}
	c := 0
	for _, id := range p.peerIDs {
		if id != p.myID {
			resp.sendTo[c] = id
			c++
		}
	}

	return resp
}

func isDigestMismatch(msg *pb.PeerMessage, p *Node) bool {
	isDigestMismatch := bytes.Compare(msg.AttachedData, p.clientMsgDigests[msg.MsgId%LOG_SIZE][:]) != 0
	return isDigestMismatch
}

func (p *Node) handleValidPrePrepare(msg *pb.PeerMessage, chOut chan *MessageWrapper) {

	shaOfClientMsg := p.computeClientMessageDigest(msg)

	p.storeClientMessageDigest(msg.MsgId, shaOfClientMsg)

	//h512.Write(rawMsg)
	//sha := h512.Sum([]byte{})

	//saving the message to the log
	p.storeClientMessageData(msg)

	resp := p.createPrepareMessage(msg, shaOfClientMsg)

	chOut <- resp
}

func (p *Node) notifyClientOfLeader(msg *pb.PeerMessage, chOut chan *MessageWrapper) {
	fmt.Printf("I am no the ledaer. MyId is %d, and the leader is %d at time: %s\n", p.myID, p.leaderId, time.Since(p.Benchmark.GetStartTime()))

	msgR := &pb.PeerMessage{}
	msgR.Type = pb.MsgType_ClientResponse
	msgR.FromNodeId = int32(p.leaderId)
	msgR.MsgId = int32(-1)
	msgR.EpochId = msg.EpochId

	resp := &MessageWrapper{sendTo: make([]int, 1), msg: msgR}
	resp.sendTo[0] = int(msg.FromNodeId)

	chOut <- resp
}

func handleClientRequest(p *Node, msg *pb.PeerMessage, chOut chan *MessageWrapper) {
	msgPP := &pb.PeerMessage{}
	msgPP.Type = pb.MsgType_PrePrepare
	msgPP.FromNodeId = int32(p.myID)
	msgPP.MsgId = int32(p.nextPropSeq)
	msgPP.EpochId = int32(p.myEpoch)

	// now := time.Now()
	rawMsg, err := proto.Marshal(msg)
	// p.benchmark.addToincrementTimeSpent(Measure{
	// 	Type:     msg.Type,
	// 	Kind:     MARSHAL,
	// 	Duration: time.Since(now),
	// })

	// fmt.Printf("Node %d received client request %d from %d\n", p.myID, msg.MsgId, msg.FromNodeId)
	// fmt.Println(proto.Size(msg), " Received Client Request in bytes in raw message")

	if err != nil {
		fmt.Printf("%s\n", err)
	}

	msgPP.AttachedData = rawMsg
	// fmt.Println(proto.Size(msgPP), " PrePrepare msg bytes in raw message")

	resp := &MessageWrapper{sendTo: p.peerIDs, msg: msgPP}
	// c := 0

	// for x := 0; x < p.peerCount; x++ {
	// 	resp.sendTo[c] = x
	// 	c++
	// }

	p.nextPropSeq++

	if (p.nextPropSeq % p.ClusterBatchSize) == 0 {
		p.nextPropSeq = p.nextPropSeq + (p.numberOfBuckets-1)*p.ClusterBatchSize
		// fmt.Printf("Next proposal sequence set to %d\n", p.nextPropSeq)
	}

	// fmt.Printf("Node %d sending PrePrepare message for proposal %d to %v, with size %d bytes at time: %s\n", p.myID, msgPP.MsgId, resp.sendTo, proto.Size(msgPP), time.Since(p.Benchmark.GetStartTime()))
	chOut <- resp
}

func (p *Node) createPrepareMessage(msg *pb.PeerMessage, shaOfClientMsg [32]byte) *MessageWrapper {
	msgP := &pb.PeerMessage{}
	msgP.Type = pb.MsgType_Prepare
	msgP.FromNodeId = int32(p.myID)
	msgP.MsgId = msg.MsgId
	msgP.EpochId = msg.EpochId

	msgP.AttachedData = shaOfClientMsg[:]
	resp := &MessageWrapper{sendTo: p.peerIDs, msg: msgP}

	return resp
}

func (p *Node) storeClientMessageData(msg *pb.PeerMessage) {
	clMsg := &pb.PeerMessage{}
	// now := time.Now()
	proto.Unmarshal(msg.AttachedData, clMsg)
	// p.benchmark.addToincrementTimeSpent(Measure{
	// 	Type:     msg.Type,
	// 	Kind:     UNMARSHAL,
	// 	Duration: time.Since(now),
	// })
	// fmt.Printf("Storing client message %d from %d\n", msg.MsgId, clMsg.FromNodeId)
	p.senderOf[msg.MsgId%CIRCULAR_BUFFER_SIZE] = int(clMsg.FromNodeId)
	p.clientMsgData[msg.MsgId%LOG_SIZE] = make([]byte, len(msg.AttachedData))
	copy(p.clientMsgData[msg.MsgId%LOG_SIZE], msg.AttachedData)
	// if p.msgSize < 0 {
	// 	p.msgSize = len(msg.AttachedData)

	// 	fmt.Printf("Message size set to %d bytes\n", p.msgSize)
	// }

}

func (p *Node) storeClientMessageDigest(msgID int32, shaOfClientMsg [32]byte) {
	if (msgID - int32(p.lastCheckpointSeq)) >= CIRCULAR_BUFFER_SIZE {
		panic(fmt.Sprintf("ERROR: Trying to store client message digest beyond log size. msgID: %d, lastCheckpointSeq: %d\n", msgID, p.lastCheckpointSeq))
	}
	p.clientMsgDigests[msgID%LOG_SIZE] = make([]byte, len(shaOfClientMsg))
	copy(p.clientMsgDigests[msgID%LOG_SIZE], shaOfClientMsg[:])
}

func (p *Node) computeClientMessageDigest(msg *pb.PeerMessage) [32]byte {
	// now := time.Now()
	shaOfClientMsg := sha256.Sum256(msg.AttachedData)
	// p.benchmark.addToincrementTimeSpent(Measure{
	// 	Type:     msg.Type,
	// 	Kind:     HASH,
	// 	Duration: time.Since(now),
	// })
	return shaOfClientMsg
}

func (p *Node) nextStateHash(curHash [32]byte, nextMsg []byte) [32]byte {

	aux := append(nextMsg[:], curHash[:]...)
	hash := sha256.Sum256(aux)

	return hash

}

func (p *Node) createClientResponse(msgId int) *MessageWrapper {

	msgR := &pb.PeerMessage{}
	msgR.Type = pb.MsgType_ClientResponse
	msgR.FromNodeId = int32(p.myID)
	msgR.MsgId = int32(msgId)
	// msgR.EpochId = msg.EpochId
	resp := &MessageWrapper{sendTo: make([]int, 1), msg: msgR}
	resp.sendTo[0] = p.senderOf[msgId%CIRCULAR_BUFFER_SIZE]
	if msgId%1000 == 0 {
		// fmt.Printf("Sending frist Client Response to %d at time: %s\n", resp.sendTo[0], time.Since(p.Benchmark.GetStartTime()))
	}

	return resp

}

func convertToPtrSlice(certs []pb.ClusterCertificate) []*pb.ClusterCertificate {
	ptrSlice := make([]*pb.ClusterCertificate, len(certs))
	for i := range certs {
		ptrSlice[i] = &certs[i]
	}
	return ptrSlice
}

func (p *Node) executeClusterCommits(msg *pb.PeerMessage, chOut chan *MessageWrapper) {
	for i := p.nextClusterCommit; i < p.nextClusterCommit+p.ClusterBatchSize; i++ {
		p.commit(int(i), msg, chOut)
	}
}
