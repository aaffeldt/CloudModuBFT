package peer

import (
	"bufio"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"io"
	cr "modubft/crypto"
	pb "modubft/proto"
	"net"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/golang/protobuf/proto"
)

type MessageWrapper struct {
	msg    *pb.PeerMessage
	sendTo []int
	hash   [32]byte
}

func setupCommunicationChannels(toInit PeerIdentifier, self *Node) {
	// channelInSize := toInit.getChannelInSize()
	self.chInData = make(chan *pb.PeerMessage, 1000)
	self.chClientInData = make(chan *pb.PeerMessage, 1000)
	self.chCoordInData = make(chan *pb.PeerMessage, 1000)
	self.chForeignData = make(chan *pb.PeerMessage, 1000)
	self.chConEvent = make(chan int)
}

const OLD_RECEIVE = false

// openAndListen is used by peers and it will spin a new goroutine to receive messages on an incoming connections in an infiitie loop. All these will be pushed into a single channel
func (p *Node) openAndListen(port int) {

	fmt.Println("Listening on " + p.myIP + ":" + strconv.Itoa(port))
	ln, err := net.Listen("tcp", ":"+strconv.Itoa(port))

	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}
	for {
		inConn, _ := ln.Accept()
		raddr := inConn.RemoteAddr().String()
		fmt.Println("Connection from " + " " + raddr)
		id_bytes := make([]byte, 4)
		n, err := inConn.Read(id_bytes[:])
		if err != nil || n != 4 {
			fmt.Printf("Error reading node ID from %s: %v\n", raddr, err)
			inConn.Close()
			continue
		}
		fmt.Printf("Read %d bytes for node ID from %s\n", n, raddr)
		id := binary.LittleEndian.Uint32(id_bytes)
		fmt.Printf("Node ID %d from %s\n", id, raddr)

		chanToUse := p.chInData

		if port == PORT_CLIENT {
			fmt.Printf("Client %s connected as ID %d\n", raddr, id)

			p.deadClients = 0
			chanToUse = p.chClientInData
			p.nodeConn[id] = inConn
			for _, client := range p.myClients {
				if strings.HasPrefix(raddr, client.MyIP) {
					// p.nodeConn[id] = inConn

				}
			}

		} else if port == PORT_COORD {
			chanToUse = p.chCoordInData
		} else {

			for _, peer := range p.foreignPeers {
				fmt.Printf("Checking foreign peer %d at %s against %s\n", peer.MyID, peer.MyIP, raddr)
				if strings.HasPrefix(raddr, peer.MyIP) {
					fmt.Printf("Foreign peer %d connected from %s\n", peer.MyID, raddr)
					// p.nodeConn[peer.MyID] = inConn
					chanToUse = p.chForeignData
				}
			}
		}

		go p.receiveOnConn(inConn, chanToUse, port)

		p.chConEvent <- port

	}
}

func distributeMessagesForUnMarshal(chIn chan []byte, unmarshChanIn []chan []byte) {
	cnt := 0
	for {
		msg := <-chIn
		// if msg.Type == pb.MsgType_ClusterPrePrepare || msg.Type == pb.MsgType_ClusterPrepare {
		// 	fmt.Printf("Received %s from %d\n", msg.Type.String(), msg.FromNodeId)
		// }
		// fmt.Printf(" Received %s from %s\n", msg.Type.String(), p.nodeAddr[msg.FromNodeId])

		unmarshChanIn[cnt%MAX_SIG_VERIF_CLIENT] <- msg
		cnt++
	}
}
func collectUnMarshaledMessages(unmarshChanOut []chan *pb.PeerMessage, chOut chan *pb.PeerMessage) {
	rc := 0
	for {
		tos := <-unmarshChanOut[rc%MAX_SIG_VERIF_CLIENT]
		// if p.nodeCount > 4 {
		// 	fmt.Printf("Node %d in collectVerifiedMessages received %s from %d\n", p.myID, tos.Type.String(), tos.FromNodeId)
		// }
		if tos != nil {
			chOut <- tos
		}
		rc++
	}
}

// receiveOnConn is an infinite loop in which a valid protobuf is attempted to be read from the TCP connection. It is unmarshalled and passed to the channel. No further validation of the message contents takes place here.
func (p *Node) receiveOnConn(conn net.Conn, ch chan *pb.PeerMessage, port int) {
	// unMarshal := make(chan []byte, 1000)
	// unmarshChanIn := make([]chan []byte, MAX_SIG_VERIF_CLIENT)
	// unmarshChanOut := make([]chan *pb.PeerMessage, MAX_SIG_VERIF_CLIENT)

	// for x := 0; x < MAX_SIG_VERIF_CLIENT; x++ {
	// 	unmarshChanIn[x] = make(chan []byte, 1000)
	// 	unmarshChanOut[x] = make(chan *pb.PeerMessage, 1000)
	// 	go p.unmarshalAndDispatch(unmarshChanIn[x], unmarshChanOut[x])

	// }

	// go distributeMessagesForUnMarshal(unMarshal, unmarshChanIn)
	// go collectUnMarshaledMessages(unmarshChanOut, ch)
	// var pbSize uint32
	// var rcvData []byte
	// lf := make([]byte, 4)
	// for {

	// 	n, err := conn.Read(lf)
	// 	if err != nil {
	// 		if err != io.EOF {
	// 			fmt.Printf("Error reading length prefix: %v\n", err)

	// 		}
	// 		continue
	// 	}

	// 	if n == 0 && p.handleClientDisconnect(conn) {
	// 		return
	// 	}
	// 	pbSizeTemp := binary.LittleEndian.Uint32(lf)
	// 	// fmt.Printf("Expecting message of size: %d bytes\n", pbSize)
	// 	if pbSizeTemp != pbSize {
	// 		pbSize = pbSizeTemp
	// 		rcvData = make([]byte, pbSize)
	// 	}
	// 	n, err = conn.Read(rcvData)
	// 	if err != nil {

	// 		fmt.Printf("Error reading message data: %v\n", err)
	// 		continue
	// 	}
	// 	if n == 0 && p.handleClientDisconnect(conn) {
	// 		return
	// 	}
	// 	// fmt.Printf("Read %d bytes for message\n", n)

	// 	msg := &pb.PeerMessage{}
	// 	err = proto.Unmarshal(rcvData[:], msg)
	// 	if err != nil {
	// 		fmt.Printf("pbSizeTemp: %d, pbSize: %d\n", pbSizeTemp, pbSize)

	// 		fmt.Printf("Error unmarshalling message: %v\n", err)
	// 		panic(err)
	// 	}
	// 	// if (msg.MsgId % 100) == 0 {
	// 	// 	fmt.Printf("Received message: %+v\n", msg)
	// 	// }

	// 	ch <- msg
	// }

	// chMarshal := make(chan []byte, 1000)
	// for range 20 {
	// 	go p.unmarshalAndDispatch(chMarshal, ch)
	// }

	if !OLD_RECEIVE {

		reader := bufio.NewReader(conn)

		for {
			lf := make([]byte, 4)
			// i := 0

			_, err := io.ReadFull(reader, lf)
			if err != nil {
				if err == io.EOF {

					if port == PORT_CLIENT {
						fmt.Println("Client disconnect ", conn.RemoteAddr())
						p.deadClients++
					}

					if p.deadClients == p.clientCount {
						fmt.Printf("All %d clients disconnected, shutting down node %d\n", p.clientCount, p.myID)
						//pprof.StopCPUProfile()
						// p.Benchmark.addToincrementTimeSpent(Measure{Kind: STOP})
						// <-p.Benchmark.stopped
						// fmt.Printf("Benchmark %+v", p.benchmark)
						// ^print time stamp^
						timestamp := time.Now().Format(time.RFC3339)
						fmt.Printf("Node %d: Before Signaling stop channel. Timestamp: %s\n", p.myID, timestamp)
						runtime.Gosched()
						fmt.Printf("Send: %p\n", p.stop)
						p.stop <- true
						// runtime.Gosched()

						afterTimestmap := time.Now().Format(time.RFC3339)
						fmt.Printf("Node %d: After Signaling stop channel. Timestamp: %s\n", p.myID, afterTimestmap)
						return
					}
					return
				} else {
					if port == PORT_CLIENT {
						fmt.Println("Client disconnect ", conn.RemoteAddr())
						p.deadClients++
					}

					if p.deadClients == p.clientCount {
						fmt.Printf("All %d clients disconnected, shutting down node %d\n", p.clientCount, p.myID)
						//pprof.StopCPUProfile()
						// p.Benchmark.addToincrementTimeSpent(Measure{Kind: STOP})
						// <-p.Benchmark.stopped
						// fmt.Printf("Benchmark %+v", p.benchmark)
						// ^print time stamp^
						timestamp := time.Now().Format(time.RFC3339)
						fmt.Printf("Node %d: Before Signaling stop channel. Timestamp: %s\n", p.myID, timestamp)
						runtime.Gosched()
						fmt.Printf("Send: %p\n", p.stop)
						p.stop <- true
						// runtime.Gosched()

						afterTimestmap := time.Now().Format(time.RFC3339)
						fmt.Printf("Node %d: After Signaling stop channel. Timestamp: %s\n", p.myID, afterTimestmap)
						return
					}

					return
				}

			}

			pbSize := binary.LittleEndian.Uint32(lf)
			rcvData := make([]byte, pbSize)
			// start := time.Now()
			_, _ = io.ReadFull(reader, rcvData)
			// read_time := time.Since(start).Nanoseconds()

			// i = 0
			// for {
			// 	p, err := reader.ReadByte()
			// 	if err == nil {
			// 		rcvData[i] = p
			// 		i++
			// 		if i == int(pbSize) {
			// 			break
			// 		}
			// 	}

			// }

			msg := &pb.PeerMessage{}
			// now := time.Now()
			// start = time.Now()
			err = proto.Unmarshal(rcvData[:], msg)
			// p.incrementTimeSpent(fmt.Sprintf("proto.Unmarshal:%s", msg.Type.String()), float64(time.Since(start).Nanoseconds()))
			// p.incrementTimeSpent(fmt.Sprintf("ReadFull:%s", msg.Type.String()), float64(read_time))

			if err != nil {
				fmt.Printf("%s\n", err)
			} else {
				//if p.VerifySig(msg) == true {
				p.incrementReceivedBytes(int(msg.FromNodeId), msg.Type, pbSize)
				ch <- msg
				//fmt.Println("TIMEReceive ", time.Since(s))
				//s = time.Now()
				//}
			}
		}
	} else {

		leftover := make([]byte, 0)
		rcvData := make([]byte, MAXBUF)
		for {

			n, _ := conn.Read(rcvData)

			//s := time.Now()
			if n == 0 {
				fmt.Println("Client disconnect ", conn.RemoteAddr())
				p.deadClients++
				if p.deadClients == p.clientCount {
					//pprof.StopCPUProfile()
					// p.Benchmark.addToincrementTimeSpent(Measure{Kind: STOP})
					// <-p.Benchmark.stopped
					// fmt.Printf("Benchmark %+v", p.benchmark)
					p.stop <- true
					return
					// os.Exit(0)
				}
				break

			}

			ll := len(leftover)
			if ll > 0 {
				n = n + ll
				smallRecv := rcvData[0 : MAXBUF-ll]
				rcvData = append(leftover, smallRecv...)
				leftover = make([]byte, 0)
				//fmt.Println("Prepended ", ll)
			}
			si := 0
			for si < n {
				lf := rcvData[si : si+4]
				pbSize := binary.LittleEndian.Uint32(lf)

				//fmt.Println(pbSize, si, n)
				if si+4+int(pbSize) <= n {

					// now := time.Now()
					// receivedDataSegment := make([]byte, int(pbSize))
					// copy(receivedDataSegment, rcvData[si+4:si+4+int(pbSize)])
					// p.unmarshalAndDispatch(receivedDataSegment, ch)
					// unMarshal <- receivedDataSegment
					p.unmarshalAndDispatch(rcvData[si+4:si+4+int(pbSize)], ch)
					si = si + 4 + int(pbSize)
				} else {
					//fmt.Println("Leftover ", n-si)
					leftover = make([]byte, n-si)
					copy(leftover, rcvData[si:n])
					break
				}
			}
		}
	}

}

func (p *Node) unmarshalAndDispatch(receivedDataSegment []byte, ch chan *pb.PeerMessage) {

	// receivedDataSegment := <-chIn
	msg := &pb.PeerMessage{}
	pbSize := uint32(len(receivedDataSegment))
	err := proto.Unmarshal(receivedDataSegment, msg)
	// p.benchmark.addToincrementTimeSpent(Measure{
	// 	Type:     msg.Type,
	// 	Kind:     UNMARSHAL,
	// 	Duration: time.Since(now),
	// })
	if err != nil {
		fmt.Printf("%s\n", err)
	} else {
		p.incrementReceivedBytes(int(msg.FromNodeId), msg.Type, pbSize)
		//if p.VerifySig(msg) == true {
		ch <- msg
		//fmt.Println("TIMEReceive ", time.Since(s))
		//s = time.Now()
		//}
	}

}

func (p *Node) initConnectionsAsPeer() {

	for _, peer := range p.myPeers {
		p.initializeTCPConnection(peer.Identifier, PORT_PEER)
	}

	for _, peer := range p.foreignPeers {
		p.initializeTCPConnection(peer.Identifier, PORT_PEER)
	}

}

func (p *Node) initializeTCPConnection(identifier Identifier, portToConn int) {
	for {
		conn, err := net.Dial("tcp", identifier.MyIP+":"+strconv.Itoa(portToConn))
		if err == nil {
			p.nodeConn[identifier.MyID] = conn
			fmt.Println("Connected to peer ", identifier.MyID, " "+identifier.MyIP+":"+strconv.Itoa(portToConn))

			var my_id_bytes [4]byte
			binary.LittleEndian.PutUint32(my_id_bytes[:], uint32(p.myID))

			n, err := conn.Write(my_id_bytes[:])
			if err != nil {
				fmt.Printf("%s %d\n", err, n)
			}
			if n == 4 {

				break
			} else {
				fmt.Printf("Error sending my ID to peer %d: sent %d bytes\n", identifier.MyID, n)
			}
		} else {
			fmt.Print("Error at node " + strconv.Itoa(p.myID) + ": ")
			fmt.Println(err)
			time.Sleep(time.Second * 1)
		}
	}
}

// verifyMessageSignature is a worker that verifies the signature of a message. It reads messages from the input channel, verifies their signatures, and sends valid messages to the output channel. Invalid messages are sent as nil.
// This function is designed to run in parallel, allowing multiple messages to be verified concurrently.
func (p *Node) verifyMessageSignature(chIn chan *pb.PeerMessage, chOut chan *pb.PeerMessage) {

	for {
		// s := time.Now()
		msg := <-chIn
		// if p.nodeCount > 4 {
		// 	fmt.Printf("Node %d in verifyMessageSignature received %s from %d\n", p.myID, msg.Type.String(), msg.FromNodeId)
		// }
		// fmt.Printf("TIMEWait_MSG_TOVerif %s-(%s)\n", time.Since(s), msg.Type.String())

		// for _, auth := range msg.Auth {
		// 	fmt.Printf("Sigtype %s, FromNodeId %d, MsgId %d, EpochId %d\n", auth.SigType.String(), msg.FromNodeId, msg.MsgId, msg.EpochId)
		// }
		// s = time.Now()
		// now := time.Now()
		if p.VerifySig(msg) {

			// p.benchmark.addToincrementTimeSpent(Measure{
			// 	Type:     msg.Type,
			// 	Kind:     VERIFY,
			// 	Duration: time.Since(now),
			// })
			// if msg.Type == pb.MsgType_PrePrepare {
			// 	shaOfClientMsg := p.computeClientMessageDigest(msg)
			// 	msg.CompressedCertificates = shaOfClientMsg[:]
			// }

			chOut <- msg
			// if p.nodeCount > 4 {
			// 	fmt.Printf("Node %d: Valid signature for message %d from %d\n", p.myID, msg.MsgId, msg.FromNodeId)
			// }
		} else {
			chOut <- nil
			panic(fmt.Sprintf("Node %d: Invalid signature for message %d from %d %v \n", p.myID, msg.MsgId, msg.FromNodeId, msg))

		}

		// fmt.Printf("TIMEVerif %s-(%s)\n", time.Since(s), msg.Type.String())
	}

}

func distributeMessagesForSignatureVerification(chIn chan *pb.PeerMessage, verifChanIn []chan *pb.PeerMessage) {
	cnt := 0
	for {
		msg := <-chIn
		// if msg.Type == pb.MsgType_ClusterPrePrepare || msg.Type == pb.MsgType_ClusterPrepare {
		// 	fmt.Printf("Received %s from %d\n", msg.Type.String(), msg.FromNodeId)
		// }
		// fmt.Printf(" Received %s from %s\n", msg.Type.String(), p.nodeAddr[msg.FromNodeId])

		verifChanIn[cnt%MAX_SIG_VERIF_CLIENT] <- msg
		cnt++
	}
}

// peerParallelVerifierStep distributes messages from the input channel in a round-robin fashion to several parallel verifiers. Results are collected the same way into a single channel.
func (p *Node) peerParallelVerifierStep(chIn chan *pb.PeerMessage, chOut chan *pb.PeerMessage) {

	verifChanIn := make([]chan *pb.PeerMessage, MAX_SIG_VERIF_CLIENT)
	verifChanOut := make([]chan *pb.PeerMessage, MAX_SIG_VERIF_CLIENT)

	for x := 0; x < MAX_SIG_VERIF_CLIENT; x++ {
		verifChanIn[x] = make(chan *pb.PeerMessage, 1000)
		verifChanOut[x] = make(chan *pb.PeerMessage, 1000)

		go p.verifyMessageSignature(verifChanIn[x], verifChanOut[x])

	}

	go distributeMessagesForSignatureVerification(chIn, verifChanIn)

	p.collectVerifiedMessages(verifChanOut, chOut)
}

// collectVerifiedMessages collects messages from multiple verification channels and sends them to a single output channel.
// It uses a round-robin approach to read from the verification channels and forward valid messages to the output channel.
// If a verification channel returns nil, it simply skips sending that message.
// This function runs indefinitely, continuously collecting messages from the verification channels.
// Alternatively, you probably could start one goroutine per channel and write to the output channel directly. (REVIEW)
func (p *Node) collectVerifiedMessages(verifChanOut []chan *pb.PeerMessage, chOut chan *pb.PeerMessage) {
	rc := 0
	for {
		tos := <-verifChanOut[rc%MAX_SIG_VERIF_CLIENT]
		// if p.nodeCount > 4 {
		// 	fmt.Printf("Node %d in collectVerifiedMessages received %s from %d\n", p.myID, tos.Type.String(), tos.FromNodeId)
		// }
		if tos != nil {
			chOut <- tos
		}
		rc++
	}
}

// peerParallelSenderStep computes signatures on several threads (round robin distribution from chIn to signQueues). The results og the signatures (sendQueues) are read round-robin and enqueued to their corresponding peer queu (peerQueue). Each peer queue has its associates sending thread.
func (p *Node) peerParallelSenderStep(chIn chan *MessageWrapper, step int) {

	certificateInQueues := make([]chan *MessageWrapper, MAX_SIG_CREATE)
	certificateOutQueues := make([]chan *MessageWrapper, MAX_SIG_CREATE)

	hashInQueues := make([]chan *MessageWrapper, MAX_SIG_CREATE)
	hashOutQueues := make([]chan *MessageWrapper, MAX_SIG_CREATE)
	signQueues := make([]chan *MessageWrapper, MAX_SIG_CREATE)
	sendQueues := make([]chan *MessageWrapper, MAX_SIG_CREATE)
	peerQueues := make([]chan *MessageWrapper, p.nodeCount)

	for x := 0; x < MAX_SIG_CREATE; x++ {
		certificateInQueues[x] = make(chan *MessageWrapper, 1000)
		certificateOutQueues[x] = make(chan *MessageWrapper, 1000)
		hashInQueues[x] = make(chan *MessageWrapper, 1000)
		hashOutQueues[x] = make(chan *MessageWrapper, 1000)
		signQueues[x] = make(chan *MessageWrapper, 1000)
		sendQueues[x] = make(chan *MessageWrapper, 1000)
		sendQueues[x] = make(chan *MessageWrapper, 1000)
	}

	for x := 0; x < p.nodeCount; x++ {
		fmt.Printf("Creating peerQueue for peer %d\n", x)
		peerQueues[x] = make(chan *MessageWrapper, 1000)
	}

	// fmt.Println("peerQueues: ", peerQueues)
	// fmt.Println("hashInQueues: ", hashInQueues)
	// fmt.Println("hashOutQueues: ", hashOutQueues)
	// fmt.Println("signQueues: ", signQueues)
	// fmt.Println("sendQueues: ", sendQueues)
	// fmt.Println("MAX_SIG_CREATE: ", MAX_SIG_CREATE)
	// fmt.Println("p.nodeCount: ", p.nodeCount)
	// fmt.Println("p.nodeConn: ", p.nodeConn)
	// fmt.Println("p.nodeConn length: ", len(p.nodeConn))
	// fmt.Println("p.myID: ", p.myID)

	//server := make([]*sha256.Avx512Server, MAX_SIG_CREATE)

	for x := 0; x < MAX_SIG_CREATE; x++ {

		go p.computeCertificate(certificateInQueues[x], certificateOutQueues[x])

		go p.computeMessageHash(hashInQueues[x], hashOutQueues[x]) //, server[x])

		//server[x] = sha256.NewAvx512Server()

		go p.createMessageWithSignature(signQueues[x], sendQueues[x]) //, server[x])

	}

	// fmt.Println("p.nodeConn length: ", len(p.nodeConn))
	// fmt.Println("p.nodeConn: ", p.nodeConn)
	for x := 0; x < p.nodeCount; x++ {
		fmt.Printf("Starting sendMessageToPeer for peer %d\n", x)
		go p.sendMessageToPeer(peerQueues[x], x, x+p.allCount*step)

	}
	go p.distributeMessageToCertificateInQueue(certificateInQueues, chIn)

	go p.distributeMessageToHashInQueue(certificateOutQueues, hashInQueues)

	go p.distributeHashesForSignatures(hashOutQueues, signQueues)

	p.forwardMessagesToSendToPeerQueues(sendQueues, peerQueues)
}

// forwardMessagesToSendToPeerQueues forwards messages from the send queues to the peer queues in a round-robin fashion.
func (p *Node) forwardMessagesToSendToPeerQueues(sendQueues []chan *MessageWrapper, peerQueues []chan *MessageWrapper) {
	qCnt := 0
	for {
		clone := <-sendQueues[qCnt%MAX_SIG_CREATE]
		// if p.nodeCount > 4 {
		// 	// fmt.Printf("Distributing message with hash %x to queue %d\n", wrap.hash, hashCnt%MAX_SIG_CREATE)
		// 	fmt.Printf("Forwarding message from %d to peer %d with type %s and auth type %s\n", clone.msg.FromNodeId, clone.sendTo[0], clone.msg.Type.String(), clone.msg.Auth[0].SigType.String())
		// }
		//s := time.Now()
		qCnt++
		// if clone.msg.MsgId == 0 {
		// 	fmt.Printf("Will send to peer %d, msgId %d, epoch %d\n", clone.sendTo[0], clone.msg.MsgId, clone.msg.EpochId)
		// }

		peerQueues[clone.sendTo[0]] <- clone

		//fmt.Println("TIMESend ", time.Since(s))

	}
}

func (p *Node) computeMsgHash(pm *pb.PeerMessage) [32]byte {
	// start := time.Now()
	authSave := pm.Auth
	pm.Auth = nil

	msgBytes, err := proto.Marshal(pm)
	if err != nil {
		fmt.Printf("%s\n", err)
	}

	msgHash := sha256.Sum256(msgBytes)

	pm.Auth = authSave

	// p.incrementTimeSpent(fmt.Sprintf("computeMsgHash:%s", pm.Type.String()), float64(time.Since(start).Nanoseconds()))

	return msgHash
}

func (p *Node) CreateSigForHash(hash [32]byte, sigType pb.SigType, toNodeId int) *pb.Authenticator {

	sig := make([]byte, 0)
	dest := toNodeId

	if toNodeId == p.myID {
		return &pb.Authenticator{}
	}

	if sigType == pb.SigType_PKSig {
		sig = p.createPKSig(hash)
		dest = -1
	} else {
		sig = p.createMACSig(int32(toNodeId), hash)
	}

	auth := &pb.Authenticator{FromNodeId: int32(p.myID), ToNodeId: int32(dest), SigType: sigType, Sig: sig}

	return auth

}

// func (p *Node) clonePeerMessage(req *pb.PeerMessage) *pb.PeerMessage {
// 	reqAClone := make([]byte, len(req.AttachedData))
// 	copy(reqAClone, req.AttachedData)
// 	retv := &pb.PeerMessage{FromNodeId: req.FromNodeId, MsgId: req.MsgId, EpochId: req.EpochId, Type: req.Type, AttachedData: reqAClone, Auth: req.Auth, ClusterId: req.ClusterId, Certificate: req.Certificate}
// 	return retv
// }

func (p *Node) clonePeerMessage(req *pb.PeerMessage) *pb.PeerMessage {
	return proto.Clone(req).(*pb.PeerMessage)

	// reqAClone := make([]byte, len(req.AttachedData))
	// copy(reqAClone, req.AttachedData)
	// certsPtr := make([]*pb.ClusterCertificate, len(req.Certificate))
	// if p.clustering {
	// 	copy(certsPtr, req.Certificate)
	// }

	// // for i := range req.Certificate {
	// // 	certsPtr[i] = &pb.ClusterCertificate{
	// // 		Certificate: make([]byte, len(req.Certificate[i].Certificate)),
	// // 		FromNodeId:  req.Certificate[i].FromNodeId,
	// // 		Sig:         make([]byte, len(req.Certificate[i].Sig)),
	// // 	}
	// // 	copy(certsPtr[i].Certificate, req.Certificate[i].Certificate)
	// // 	copy(certsPtr[i].Sig, req.Certificate[i].Sig)
	// // }
	// retv := &pb.PeerMessage{FromNodeId: req.FromNodeId, MsgId: req.MsgId,
	// 	EpochId: req.EpochId, Type: req.Type, AttachedData: reqAClone, Auth: req.Auth,
	// 	ClusterId: req.ClusterId, Certificate: certsPtr, CompressedCertificates: req.CompressedCertificates}
	// return retv
}

// func (p *Node) clonePeerMessageNoAuth(req *pb.PeerMessage) *pb.PeerMessage {
// 	reqAClone := make([]byte, len(req.AttachedData))
// 	copy(reqAClone, req.AttachedData)
// 	retv := &pb.PeerMessage{FromNodeId: req.FromNodeId, MsgId: req.MsgId, EpochId: req.EpochId, Type: req.Type, AttachedData: reqAClone, ClusterId: req.ClusterId, Certificate: req.Certificate}
// 	return retv
// }

func (p *Node) clonePeerMessageNoAuth(req *pb.PeerMessage) *pb.PeerMessage {
	authSave := req.Auth
	req.Auth = nil
	clone := proto.Clone(req).(*pb.PeerMessage)

	req.Auth = authSave
	return clone
	// reqAClone := make([]byte, len(req.AttachedData))
	// copy(reqAClone, req.AttachedData)

	// // certs := cloneClusterCertificates(req.Certificate)

	// certsPtr := make([]*pb.ClusterCertificate, len(req.Certificate))
	// if p.clustering {
	// 	copy(certsPtr, req.Certificate)
	// }

	// // for i := range req.Certificate {
	// // 	certsPtr[i] = &pb.ClusterCertificate{
	// // 		Certificate: make([]byte, len(req.Certificate[i].Certificate)),
	// // 		FromNodeId:  req.Certificate[i].FromNodeId,
	// // 		Sig:         make([]byte, len(req.Certificate[i].Sig)),
	// // 	}
	// // 	copy(certsPtr[i].Certificate, req.Certificate[i].Certificate)
	// // 	copy(certsPtr[i].Sig, req.Certificate[i].Sig)
	// // }
	// // certsPtr := make([]*pb.ClusterCertificate, len(req.Certificate))
	// // var cert []*pb.ClusterCertificate = make([]*pb.ClusterCertificate, len(req.Certificate))

	// // deepcopy.Copy(&cert, &req.Certificate)

	// // if req.MsgId == 0 {
	// // 	fmt.Printf("cert: %+v, req.Certificate: %+v, cert addr: %p, req.Certificate addr: %p\n", cert, req.Certificate, cert, req.Certificate)
	// // }
	// retv := &pb.PeerMessage{FromNodeId: req.FromNodeId, MsgId: req.MsgId, EpochId: req.EpochId, Type: req.Type,
	// 	AttachedData: reqAClone, ClusterId: req.ClusterId, Certificate: certsPtr, CompressedCertificates: req.CompressedCertificates}
	// return retv
}

func (p *Node) sendMessageToPeer(inChan chan *MessageWrapper, cid int, idx int) {
	fmt.Printf("Starting sender for peer %d\n", cid)
	sendBuf := make([]byte, 0)
	pendingBytes := 0
	clone := &MessageWrapper{}
	nextMsg := &MessageWrapper{}
	nextMsg = nil
	// p.myClients
	// check whether cid is a client id
	is_client_recipient := false
	for _, client := range p.myClients {
		if cid == client.MyID {
			is_client_recipient = true
		}
	}
	// print ,p.nodeConn[cid]
	fmt.Printf("p.nodeConn[%d]: %v, %v \n", cid, p.nodeConn[cid], p.nodeConn)
	for {

		if nextMsg != nil {
			clone = nextMsg
			nextMsg = nil
		} else {
			clone = <-inChan
		}

		// if p.nodeCount > 4 {
		// 	// fmt.Printf("Distributing message with hash %x to queue %d\n", wrap.hash, hashCnt%MAX_SIG_CREATE)
		// 	fmt.Printf("Sending message from %d to %d with type %s and auth type %s\n", clone.msg.FromNodeId, clone.sendTo[0], clone.msg.Type.String(), clone.msg.Auth[0].SigType.String())
		// }

		if clone.msg.MsgId == 0 {
			// fmt.Printf("Send msg to %d \n", cid)
		}
		// now := time.Now()
		// start := time.Now()
		raw, err := proto.Marshal(clone.msg)
		// p.incrementTimeSpent(fmt.Sprintf("proto.Marshal:%s", clone.msg.Type.String()), float64(time.Since(start).Nanoseconds()))

		// p.benchmark.addToincrementTimeSpent(Measure{
		// 	Type:     clone.msg.Type,
		// 	Kind:     MARSHAL,
		// 	Duration: time.Since(now),
		// })
		if err != nil {
			fmt.Printf("%s\n", err)
		} else {

			// print message from to and the auth sig type
			// fmt.Printf("Sending message from %d to %d with type %s and auth type %s\n", clone.msg.FromNodeId, clone.sendTo[0], clone.msg.Type.String(), clone.msg.Auth[0].SigType.String())
			lba := make([]byte, 4)
			binary.LittleEndian.PutUint32(lba, uint32(len(raw)))
			if clone.msg.MsgId == 0 {
				fmt.Printf("Size of message to %d is %d bytes\n", cid, len(raw)+len(lba))
			}

			// fmt.Printf("Node %d sending %d bytes to node %d for msg type %s\n", p.myID, len(raw)+len(lba), cid, clone.msg.Type.String())

			p.incrementBytesSent(idx, cid, clone.msg.Type.String(), raw)

			sendBuf = append(sendBuf, lba...)
			sendBuf = append(sendBuf, raw...)
			pendingBytes += len(lba) + len(raw)

			if is_client_recipient && clone.msg.MsgId < 1000 {
				timestamp := time.Now().UnixNano()
				fmt.Printf("Response client request %d from %d at time %d\n", clone.msg.MsgId, clone.msg.FromNodeId, timestamp)
			}

			select {
			case nextMsg = <-inChan:
				// will do an other pass

				if pendingBytes < 64000 {
					continue
				}
			default:
				nextMsg = nil
			}

			// if p.logSentBytes {
			// p.bytesSent[cid] += pendingBytes
			// }

			// . += len(sendBuf)
			// start = time.Now()
			n, err := p.nodeConn[cid].Write(sendBuf)
			// p.incrementTimeSpent(fmt.Sprintf("conn.Write:%s", clone.msg.Type.String()), float64(time.Since(start).Nanoseconds()))
			// if p.nodeCount > 4 {
			// 	fmt.Printf("Sent %d bytes to peer %d\n", n, cid)
			// }
			sendBuf = make([]byte, 0)
			pendingBytes = 0

			if err != nil {
				fmt.Printf("%s\n", err)
			}

			if n == 0 {
				fmt.Printf(" Sent %d bytes\n", n)
			}

		}
	}
}

func (p *Node) distributeMessageToCertificateInQueue(certificateInQueues []chan *MessageWrapper, chIn chan *MessageWrapper) {
	certificateCnt := 0
	// if p is instance of ClusterNode

	for {
		wrap := <-chIn
		// if p.nodeCount > 4 {
		// 	fmt.Printf("Distributing message with hash %x to queue %d\n", wrap.hash, hashCnt%MAX_SIG_CREATE)
		// }

		certificateInQueues[certificateCnt%MAX_SIG_CREATE] <- wrap
		certificateCnt++

	}
}

func (p *Node) distributeMessageToHashInQueue(certificateOutQueues []chan *MessageWrapper, hashInQueues []chan *MessageWrapper) {
	certCnt := 0
	hashCnt := 0
	// if p is instance of ClusterNode

	for {
		wrap := <-certificateOutQueues[certCnt%MAX_SIG_CREATE]
		certCnt++
		// if p.nodeCount > 4 {
		// 	fmt.Printf("Distributing message with hash %x to queue %d\n", wrap.hash, hashCnt%MAX_SIG_CREATE)
		// }

		hashInQueues[hashCnt%MAX_SIG_CREATE] <- wrap
		hashCnt++
	}
}

// distributeHashesForSignatures distributes the hashes to the signing queues.
func (p *Node) distributeHashesForSignatures(hashOutQueues []chan *MessageWrapper, signQueues []chan *MessageWrapper) {
	sendCnt := 0
	hashCnt := 0
	for {
		wrap := <-hashOutQueues[hashCnt%MAX_SIG_CREATE]
		// if p.nodeCount > 4 {
		// 	// fmt.Printf("Distributing message with hash %x to queue %d\n", wrap.hash, hashCnt%MAX_SIG_CREATE)
		// 	fmt.Printf("Distributing message with hash %x to signing queue %d\n", wrap.hash, hashCnt%MAX_SIG_CREATE)
		// }
		hashCnt++

		for _, elem := range wrap.sendTo {

			newWrap := &MessageWrapper{}
			newWrap.msg = p.clonePeerMessageNoAuth(wrap.msg)
			newWrap.sendTo = make([]int, 1)
			newWrap.hash = wrap.hash
			newWrap.sendTo[0] = elem
			signQueues[sendCnt%MAX_SIG_CREATE] <- newWrap
			sendCnt++
		}
	}
}

func (p *Node) computeCertificate(chIn <-chan *MessageWrapper, chOut chan<- *MessageWrapper) {

	for {
		wrap := <-chIn
		msg := wrap.msg

		switch msg.Type {

		case pb.MsgType_ClusterPrepare:
			// c := make([]byte, len(msg.Certificate[0].Certificate))
			// copy(c, msg.Certificate[0].Certificate)
			// buf := &bytes.Buffer{}
			// gzWriter, _ := gzip.NewWriterLevel(buf, COMPRESSION)
			// gzWriter.Write(c)
			// gzWriter.Flush()
			// gzWriter.Close()
			// msg.Certificate[0].Certificate = buf.Bytes()

		// case pb.MsgType_ClusterPrepare:
		// 	if p.useCertificateCompression {
		// 		var b bytes.Buffer
		// 		compressed := make([]byte, 0)
		// 		for _, cert := range msg.Certificate {
		// 			lba := make([]byte, 4)
		// 			raw, _ := proto.Marshal(cert)
		// 			binary.LittleEndian.PutUint32(lba, uint32(len(raw)))
		// 			// fmt.Printf("Certificate size %d for node %d\n", len(raw), cert.FromNodeId)
		// 			compressed = append(compressed, lba...)
		// 			compressed = append(compressed, raw...)

		// 		}

		// 		// fmt.Printf("Total certificate size %d\n", len(compressed))
		// 		w := zlib.NewWriter(&b)

		// 		w.Write(compressed)

		// 		w.Close()
		// 		// b.Bytes()
		// 		wrap.msg = proto.Clone(msg).(*pb.PeerMessage)
		// 		wrap.msg.CompressedCertificates = b.Bytes()
		// 		// fmt.Printf("Compressed certificate size %d\n", len(wrap.msg.CompressedCertificates))

		// 		wrap.msg.Certificate = nil
		// 	}
		case pb.MsgType_ClusterPrePrepare:
			// start := time.Now()
			var hash [32]byte
			if p.sendTransaction {
				state := p.generateStateFromCert(msg)
				hash = sha256.Sum256(state)
			} else {
				hash = sha256.Sum256(msg.AttachedData)
			}

			// p.incrementTimeSpent(fmt.Sprintf("CertificateHash:%s", msg.Type.String()), float64(time.Since(start).Nanoseconds()))
			msg.Certificate = make([]*pb.ClusterCertificate, 1)
			msg.Certificate[0] = &pb.ClusterCertificate{
				// Certificate: hash[:],
				FromNodeId: msg.FromNodeId,
			}

			if p.useCryptoForCertificates {
				// start := time.Now()
				if p.PKSigner == nil {
					p.PKSigner, _ = cr.LoadPrivateKey(p.myPrivKey)
				}

				sig, _ := p.PKSigner.Sign(hash[:])

				msg.Certificate[0].Sig = sig
				// p.incrementTimeSpent(fmt.Sprintf("CertificateSign:%s", msg.Type.String()), float64(time.Since(start).Nanoseconds()))
			}

			if p.sendTransaction {

				wrap.msg = &pb.PeerMessage{
					FromNodeId:   msg.FromNodeId,
					MsgId:        msg.MsgId,
					EpochId:      msg.EpochId,
					Type:         msg.Type,
					Certificate:  msg.Certificate,
					ClusterId:    msg.ClusterId,
					MessageSizes: msg.MessageSizes,
					AttachedData: msg.AttachedData,
				}
			} else {

				wrap.msg = &pb.PeerMessage{
					FromNodeId:   msg.FromNodeId,
					MsgId:        msg.MsgId,
					EpochId:      msg.EpochId,
					Type:         msg.Type,
					Certificate:  msg.Certificate,
					ClusterId:    msg.ClusterId,
					MessageSizes: msg.MessageSizes,
				}
			}

			if wrap.msg.MsgId == 0 {
				// print messageSIzes
				// fmt.Printf("MessageSizes: %+v\n", msg.MessageSizes)
				// fmt.Printf("ClusterPrePrepare: MsgId=%d, EpochId=%d, FromNodeId=%d, CertificateHash=%x\n", msg.MsgId, msg.EpochId, msg.FromNodeId, msg.AttachedData)
			}

			// fmt.Printf("Send message of type %s and size %d from %d to %+v\n", wrap.msg.Type.String(), len(wrap.msg.AttachedData), wrap.msg.FromNodeId, wrap.sendTo)

		}
		chOut <- wrap
	}
}

func (p *Node) computeMessageHash(chIn <-chan *MessageWrapper, chOut chan<- *MessageWrapper) { //, srv *sha256.Avx512Server) {

	for {
		wrap := <-chIn
		// if p.nodeCount > 4 {
		// 	// fmt.Printf("Distributing message with hash %x to queue %d\n", wrap.hash, hashCnt%MAX_SIG_CREATE)
		// 	fmt.Printf("Computing hash for message from %d to %d\n", wrap.msg.FromNodeId, wrap.sendTo[0])
		// }
		//s := time.Now()
		// now := time.Now()
		msg := wrap.msg

		h := p.computeMsgHash(msg)
		for e := 0; e < 32; e++ {
			wrap.hash[e] = h[e]
		}

		// p.benchmark.addToincrementTimeSpent(Measure{
		// 	Type:     msg.Type,
		// 	Kind:     HASH,
		// 	Duration: time.Since(now),
		// })
		chOut <- wrap
		//fmt.Println("TIMESign ", time.Since(s))

	}
}

func (p *Node) createMessageWithSignature(chIn <-chan *MessageWrapper, chOut chan<- *MessageWrapper) { //, srv *sha256.Avx512Server) {

	for {
		wrap := <-chIn
		//s := time.Now()
		// now := time.Now()
		// if p.nodeCount > 4 {
		// 	// fmt.Printf("Distributing message with hash %x to queue %d\n", wrap.hash, hashCnt%MAX_SIG_CREATE)
		// 	fmt.Printf("Creating signature for message from %d to %d\n", wrap.msg.FromNodeId, wrap.sendTo[0])
		// }

		msg := wrap.msg

		hash := wrap.hash

		////s := time.Now()
		auth := &pb.Authenticator{}

		// check whether  wrap.msg.Certificate is set
		if p.nodeRole[wrap.sendTo[0]] == ROLE_CLIENT && p.usePKToClient == true {
			if msg.MsgId == 0 {
				fmt.Printf("Creating PK signature for client %d\n", wrap.sendTo[0])
			}
			// start := time.Now()
			auth = p.CreateSigForHash(hash, pb.SigType_PKSig, wrap.sendTo[0])
			// p.incrementTimeSpent(fmt.Sprintf("CreateSigForHash-%s:%s", pb.SigType_PKSig.String(), msg.Type.String()), float64(time.Since(start).Nanoseconds()))

			// p.benchmark.addToincrementTimeSpent(Measure{
			// 	Type:     msg.Type,
			// 	Kind:     SIGN,
			// 	Duration: time.Since(now),
			// })
			// fmt.Printf("Created PK signature for peer %d with hash %x and sig type %s\n", wrap.sendTo[0], hash, auth.SigType.String())
			// print auth
			// fmt.Println("Created PK for peer ", wrap.sendTo[0], " with hash ", hash, " and sig ", auth.SigType.String())

		} else {
			// start := time.Now()
			auth = p.CreateSigForHash(hash, pb.SigType_MAC, wrap.sendTo[0])
			// p.incrementTimeSpent(fmt.Sprintf("CreateSigForHash-%s:%s", pb.SigType_MAC.String(), msg.Type.String()), float64(time.Since(start).Nanoseconds()))
			// p.benchmark.addToincrementTimeSpent(Measure{
			// 	Type:     msg.Type,
			// 	Kind:     MAC,
			// 	Duration: time.Since(now),
			// })
			// fmt.Println("Created MAC for peer ", wrap.sendTo[0], " with hash ", hash, " and sig ", auth.SigType.String())
		}

		//extra sig for the coordinator
		authCoord := &pb.Authenticator{}
		if p.coordPresent {
			authCoord = p.CreateSigForHash(hash, pb.SigType_MAC, p.coordId)
		}

		//fmt.Println("OutSign ", time.Since(s))
		////s = time.Now()

		msg.Auth = make([]*pb.Authenticator, 2)
		msg.Auth[0] = auth
		msg.Auth[1] = authCoord

		dest := make([]int, 1)
		dest[0] = wrap.sendTo[0]
		newWrap := &MessageWrapper{msg: msg, sendTo: dest}

		chOut <- newWrap
		//fmt.Println("TIMESign ", time.Since(s))

	}
}

func (p *Node) CloseAllConnections() {
	for _, conn := range p.nodeConn {
		if conn != nil {
			conn.Close()
		}

	}
}
