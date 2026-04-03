package peer

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"io"
	pb "modubft/proto"
	"sync"
	"time"

	"google.golang.org/protobuf/proto"
)

func (me *Client) RunClient(vallen int, local_clients int, runTime int) {
	raws := make([][]byte, 500)
	for i := 0; i < 500; i++ {
		_, _, raws[i], _ = me.generateMessageWithSignature(generateRandomBytes(vallen), pb.SigType_PKSig)
	}

	chResult := make(chan int, 1000)

	go me.ClientReceiveFromPeers(chResult)
	<-chResult
	fmt.Printf("Sending message  to leader %d\n",  me.leaderId)

	me.ClientSendToLeader(raws[0], pb.SigType_PKSig)

	

	realReceived := 0
	stop_chan := make(chan bool, 1)

	finished_chan := make(chan int, local_clients)
	allowed_to_send := make(chan bool, local_clients)
	sent := new(sync.Map)

	for i := 0; i < local_clients; i++ {
		allowed_to_send <- true
		sent.Store(i, 0)

	}
	<- allowed_to_send 

	for i := 0; i < 100; i++ {

		go func(stop_chan chan bool, rawB []byte, cid int, finished_chan chan int, allowed_to_send chan bool) {
			sent_out := 0
			for notStop := range allowed_to_send {
				if !notStop {
					break
				}
				me.ClientSendToLeaderCachedRaw(rawB)
				sent_out++
								// fmt.Printf("Node%d: Client %d sent message %d to leader %d \n", me.myID, cid, sent_out, me.leaderId)

				// sent.Store(cid, sent_out)
			}
		}(stop_chan, raws[i%500], i, finished_chan, allowed_to_send)

	}

	start := time.Now()

	// for int(elapsed) < (int(time.Second) * runTime)

	realReceived = 0
	notStop := true
	lastReceived := -1
	warmup := false
benchmark_for:
	for {
		if !warmup && time.Since(start).Seconds() >= 5 {
			start = time.Now()
			warmup = true
			me.Benchmark.SetStartTime()
			go func(stop_chan chan bool) {
				time.Sleep(time.Second * time.Duration(runTime))
				stop_chan <- true
			}(stop_chan)

		}
		select {
		case <-stop_chan:
			notStop = false
			for {
				select {
				case msgId := <-chResult:
					fmt.Printf("Node%d: received response for message %d\n", me.myID, msgId)
					if warmup {
						realReceived++
					}
					// allowed_to_send <- notStop
				default:
					fmt.Printf("Node%d: Stopping client after %d messages\n", me.myID, realReceived)
					break benchmark_for

				}
			}
		case msgId := <-chResult:
			if msgId <= lastReceived {
	
				panic(fmt.Sprintf("Node%d: Received message %d, but last received was %d", me.myID, msgId, lastReceived))
			}
			lastReceived = msgId
			if warmup {
				realReceived++
				
			}

			// fmt.Printf("Node%d: received response for message %d\n", me.myID, msgId)
			allowed_to_send <- notStop
			// time_spent_receiving += int(time.Since(now).Nanoseconds())

		}
	}

	// <-stop_chan

	// fmt.Printf("Node%d: Stopping clients after %d seconds\n", myId, time.Since(start).Milliseconds()/1000)
	// realReceived = 0
	// for i := 0; i < local_clients; i++ {
	// 	realReceived += <-finished_chan
	// 	// fmt.Printf("Node%d: Client %d finished with %d messages\n", myId, i, realReceived)
	// }
	sent_out := 0
	for i := 0; i < local_clients; i++ {
		// time_spent_receiving += time_spent_receiving_local_clients[i]

		s, ok := sent.Load(i)
		if ok {
			if v, ok := s.(int); ok {
				sent_out += v
			}
		}
	}

	// fmt.Printf("Node%d: Total time spent sending: %d ns", myId, time_spent_waiting_to_send_total)
	// fmt.Printf("Node%d: Total time spent receiving: %d ns", myId, time_spent_waiting_receiving)
	// fmt.Printf("Node%d: Total time spent: %d ns", myId, time_spent_waiting_to_send_total+time_spent_waiting_receiving)
	// fmt.Printf("Node%d: Percentage of time spent sending: %f%%", myId, float64(time_spent_waiting_to_send_total)/float64(time_spent_waiting_to_send_total+time_spent_waiting_receiving)*100)

	// benchmark.SendNS = time_spent_waiting_to_send_total
	// benchmark.ReceiveNS = time_spent_waiting_receiving
	// benchmark.TotalNS = time_spent_waiting_to_send_total + time_spent_waiting_receiving
	fmt.Printf("Sent Out %d and Received %d\n", sent_out, realReceived)
	elapsed := time.Since(start)
	tput := float32(realReceived) / (float32(elapsed) / 1000000000.0)
	fmt.Printf("\nThroughput [ops/s] %d\n", int(tput))
	me.Benchmark.RunTime = (float64(elapsed) / 1000000000.0)
	me.Benchmark.Throughput = float64(tput)
	// bytes_sent := realReceived * (len(rawB) + 4)
	// benchmark.BytesSent = bytes_sent

	// fmt.Printf("Node%d:Total bytes sent: %d\n", myId, bytes_sent)
	// datarate := float64(bytes_sent) / (float64(elapsed) / 1000000000.0)
	// fmt.Printf("Node%d:Data rate [B/s] %f\n", myId, datarate)
	// fmt.Printf("Node%d:Data rate [GB/s] %f\n", myId, datarate/math.Pow(10, 9))

	// benchmark.Throughput = tput
	// benchmark.DataRate = (datarate)

	// benchmark.BytesReceived = realReceived * vallen
	// benchmark.WriteToFile()
}

func generateRandomBytes(vallen int) []byte {
	rtxt := make([]byte, vallen)
	if _, err := io.ReadFull(rand.Reader, rtxt); err != nil {
		fmt.Println(err)
	}
	return rtxt
}

// ClientSendToLeader is used to send a message from a client node to the leader node. It computes signatures, etc. inside. It does not wait for an answer and returns once the data is sent.
func (p *Client) ClientSendToLeader(rawData []byte, sigType pb.SigType) ([32]byte, []*pb.Authenticator, []byte) {

	msg, hash, rawBytes, err := p.generateMessageWithSignature(rawData, sigType)
	if err != nil {
		fmt.Printf("%s\n", err)
	}
	lba := make([]byte, 4)
	binary.LittleEndian.PutUint32(lba, uint32(len(rawBytes)))

	p.nodeConn[p.leaderId].Write(append(lba, rawBytes...))
	return hash, msg.Auth, rawBytes
}

func (p *Client) generateMessageWithSignature(rawData []byte, sigType pb.SigType) (*pb.PeerMessage, [32]byte, []byte, error) {
	msg := &pb.PeerMessage{FromNodeId: int32(p.myID), Type: pb.MsgType_ClientRequest, AttachedData: rawData}
	sendTo := p.leaderId
	hash := p.computeMsgHash(msg)
	sign := p.CreateSigForHash(hash, sigType, sendTo)
	msg.Auth = make([]*pb.Authenticator, 1)
	msg.Auth[0] = sign
	// fmt.Printf("Node %d sending message to leader %d with hash %x\n", p.myID, p.leaderId, hash)
	// fmt.Println("Raw message size: ", len(rawData), "msg size: ", proto.Size(msg))
	rawBytes, err := proto.Marshal(msg)
	return msg, hash, rawBytes, err
}

// ClientSendToLeaderCached does the same as ClientSendToLeader, but instead of computing the signatures for the message, it reuses an already computed one. Of course, the message contents have to remain identical.
func (p *Client) ClientSendToLeaderCached(rawData []byte, sigType pb.SigType, cachedSign []*pb.Authenticator) bool {

	msg := &pb.PeerMessage{FromNodeId: int32(p.myID), Type: pb.MsgType_ClientRequest, AttachedData: rawData}
	msg.Auth = cachedSign

	rawBytes, err := proto.Marshal(msg)
	if err != nil {
		fmt.Printf("%s\n", err)
	}
	lba := make([]byte, 4)
	binary.LittleEndian.PutUint32(lba, uint32(len(rawBytes)))
	n, err := p.nodeConn[p.leaderId].Write(append(lba, rawBytes...))
	if err != nil {
		fmt.Printf("%s %d\n", err, n)
		return false
	}

	return true
}

// ClientSendToLeaderCached does the same as ClientSendToLeader, but sends an already serialized message.
func (p *Client) ClientSendToLeaderCachedRaw(rawBytes []byte) bool {

	lba := make([]byte, 4)
	binary.LittleEndian.PutUint32(lba, uint32(len(rawBytes)))
	n, err := p.nodeConn[p.leaderId].Write(append(lba, rawBytes...))
	if err != nil {
		fmt.Printf("%s %d\n", err, n)
		return false
	}

	return true
}

// func (p *Client) SendTimeOut(msgId chan int, slot int) {

// 	time.Sleep(TIMEOUT_CLIENT * time.Millisecond)
// 	msgId <- -slot
// }

// ClientReceiveFromPeers starts a thread for each peer of the consensus and collects messages into a channel that is fed into a verification step. Finally, the method returns messageIDs as they are committed through a channel.
func (p *Client) ClientReceiveFromPeers(msgId chan int) {

	answers := make([]int, CIRCULAR_BUFFER_SIZE)
	epoch := make([]int, CIRCULAR_BUFFER_SIZE)
	last := -1

	verifMsgChan := make(chan *pb.PeerMessage, 1000)
	validMsgChan := make(chan *pb.PeerMessage, 1000)

	for _, peer := range p.myPeers {
		go p.receiveOnConn(p.nodeConn[peer.MyID], verifMsgChan, -1)
	}
	for _, peer := range p.foreignPeers {
		go p.receiveOnConn(p.nodeConn[peer.MyID], verifMsgChan, -1)
	}

	go p.peerParallelVerifierStep(verifMsgChan, validMsgChan)

	for {
		msg := <-validMsgChan

		if msg == nil {
			fmt.Println("Received msg with invalid sig!")
			continue
		}
		if msg.MsgId == 0 {
			fmt.Printf("Received msg from %d \n ", msg.FromNodeId)
		}
		if msg.MsgId == -2 {
			msgId <- int(msg.MsgId)
			continue
		}
		if msg.MsgId == -1 {
			if p.leaderId != int(msg.FromNodeId) {
				fmt.Printf("Changing leader from %d to %d\n", p.leaderId, int(msg.FromNodeId))
				p.leaderId = int(msg.FromNodeId)
				msgId <- int(msg.MsgId)
				p.reconfigured = true
				//answers = make([]int, CIRCULAR_BUFFER_SIZE)
			}
		} else {
			if epoch[msg.MsgId%CIRCULAR_BUFFER_SIZE] < int(msg.EpochId) {
				answers[msg.MsgId%CIRCULAR_BUFFER_SIZE] = 1
				epoch[msg.MsgId%CIRCULAR_BUFFER_SIZE] = int(msg.EpochId)
			} else {
				answers[msg.MsgId%CIRCULAR_BUFFER_SIZE]++
			}

			//if p.reconfigured && int(msg.MsgId)%5000 == 1 {
			//	fmt.Printf("Received reply after reconfiguration %d from %d with last %d, total %d\n", msg.MsgId, int(msg.FromNodeId), last, answers[int(msg.MsgId)%CIRCULAR_BUFFER_SIZE])
			//}

			//fmt.Println("Answer state of ", msg.MsgId, " is ", answers[msg.MsgId%1000])

			// need to collect ALL answers before we consider the message committed.
			if answers[msg.MsgId%CIRCULAR_BUFFER_SIZE] == p.peerCount {

				//if p.reconfigured && int(msg.MsgId)%5000 == 1 {
					// fmt.Printf("Received  reply for %d from %d with last %d\n", msg.MsgId, int(msg.FromNodeId), last)
				//}
				if last < int(msg.MsgId) {
					last = int(msg.MsgId)
					msgId <- int(msg.MsgId)
				}
				answers[msg.MsgId%CIRCULAR_BUFFER_SIZE] = 0

			}
		}

	}
}
