package peer

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand/v2"
	pb "modubft/proto"
	"sync"
	"time"
)

const (
	MARSHAL   = "MARSHAL"
	SIGN      = "SIGN"
	MAC       = "MAC"
	HASH      = "HASH"
	DECIDE    = "DECIDE"
	VERIFY    = "VERIFY"
	UNMARSHAL = "UNMARSHAL"
	STOP      = "STOP"
)

type Measure struct {
	Type     pb.MsgType
	Kind     string
	Duration time.Duration
}

type Flags struct {
	MyId                                             int
	RunTime                                          int
	NodeListS                                        string
	NodeRoleS                                        string
	ClusterListS                                     string
	Profile                                          bool
	Vallen                                           int
	UsePkToClient                                    bool
	DontVerifyClusterCommit                          bool
	ClusterBatchSize                                 int
	UseCertificateCompression                        bool
	UseDool                                          bool
	AlternatingClusterPrepare                        bool
	ClusterPrepareToAllAndEnableHashForClusterCommit bool
	SendTransaction                                  bool
}

type Benchmark struct {
	TimeSpentForDecision   map[pb.MsgType]float64
	TimeSpentFor           map[pb.MsgType]map[string]float64
	TimeSpent              []TimeKeyValue
	incrementTimeSpentChan chan Measure
	stop                   bool
	stopped                chan bool
	bytesSentTo            [][]BytesKeyValue
	timeSpent              *sync.Map
	bytesReceivedFrom      *sync.Map
	BytesSentTo            []BytesKeyValue
	BytesSentSum           int
	BytesReceivedFrom      []BytesKeyValue
	BytesReceivedSum       int
	StartTime              time.Time
	RunTime                float64
	Throughput             float64
	Setup                  Flags
	Dool                   string
	alreadySetStartTime    bool
	bytesSentToLocks       []sync.Mutex
}

// Getters for Benchmark fields
func (b *Benchmark) GetTimeSpentForDecision() map[pb.MsgType]float64 {
	return b.TimeSpentForDecision
}

func (b *Benchmark) GetTimeSpentFor() map[pb.MsgType]map[string]float64 {
	return b.TimeSpentFor
}

func (b *Benchmark) GetIncrementTimeSpentChan() chan Measure {
	return b.incrementTimeSpentChan
}

func (b *Benchmark) GetStop() bool {
	return b.stop
}

type BytesKey struct {
	CID int
	Typ string
}

func (p *Node) incrementBytesSent(idx int, cid int, typ string, sendBuf []byte) {

	p.Benchmark.bytesSentToLocks[idx].Lock()
	defer p.Benchmark.bytesSentToLocks[idx].Unlock()

	messageTypeBytesSent := p.Benchmark.bytesSentTo[idx][pb.MsgType_value[typ]]

	p.Benchmark.bytesSentTo[idx][pb.MsgType_value[typ]] = BytesKeyValue{
		BytesKey: BytesKey{
			CID: cid,
			Typ: typ,
		},
		Value: messageTypeBytesSent.Value + len(sendBuf) + 4,
	}

}

func (p *Node) incrementTimeSpent(typ string, timespent float64) {
	if rand.Float64() < benchRate {
		fmt.Printf("TIME SPENT: type %s, time spent %f\n", typ, timespent)
	}
	return

	m := p.Benchmark.getTimeSpentSyncMap()
	value, _ := m.Load(typ)
	floatValue := 0.0
	if v, ok := value.(float64); ok {
		floatValue = v
	}
	m.Store(typ, floatValue+(timespent))

}

const benchRate = -0.1

func (p *Node) incrementReceivedBytes(from int, typ pb.MsgType, pbSize uint32) {

	if rand.Float64() < benchRate {
		fmt.Printf("RECEIVED BYTES: from %d, type %s, size %d\n", from, typ.String(), pbSize)
	}
	return
	m := p.Benchmark.GetBytesReceivedFromSyncMap()
	key := BytesKey{
		CID: from,
		Typ: typ.String(),
	}
	val, _ := m.Load(key)
	var value int
	if v, ok := val.(int); ok {
		value = v
	} else {
		value = 0
	}

	m.Store(key, value+int(pbSize))
}

func (b *Benchmark) GetStopped() chan bool {
	return b.stopped
}

func (b *Benchmark) GetBytesSentToSyncMap() [][]BytesKeyValue {
	return b.bytesSentTo
}

func (b *Benchmark) getTimeSpentSyncMap() *sync.Map {
	return b.timeSpent

}
func (b *Benchmark) GetBytesReceivedFromSyncMap() *sync.Map {
	return b.bytesReceivedFrom
}

func (b *Benchmark) GetStartTime() time.Time {
	return b.StartTime
}

func (b *Benchmark) GetRunTime() float64 {
	return b.RunTime
}

func (b *Benchmark) GetThroughput() float64 {
	return b.Throughput
}

func (b *Benchmark) GetSetup() Flags {
	return b.Setup
}

func (b *Benchmark) GetDool() string {
	return b.Dool
}

// func (b *Benchmark) GetIperfBefore() iperf.TestReport {
// 	return b.IperfBefore
// }

// func (b *Benchmark) GetIperfAfter() iperf.TestReport {
// 	return b.IperfAfter
// }

func (b Benchmark) String() string {
	out, err := json.Marshal(b)
	if err != nil {
		panic(err)
	}

	return string(out)
}
func (b *Benchmark) GetBytesSentSum() int {
	sum := 0
	fmt.Printf("Getting Bytes Sent Sum from sync.Map\n")
	fmt.Printf("Current sync.Map: %v\n", b.bytesSentTo)
	temp := make([]BytesKeyValue, 0)

	for idx, bytes := range b.bytesSentTo {
		b.bytesSentToLocks[idx].Lock()
		for _, v := range bytes {
			if v.CID != -1 {
				sum += v.Value
				fmt.Printf("Appending to BytesSentTo from map: %v\n", v)
				temp = append(temp, BytesKeyValue{
					BytesKey: BytesKey{
						CID: v.CID,
						Typ: v.Typ,
					},
					Value: v.Value,
				})
			}
		}
		b.bytesSentToLocks[idx].Unlock()
	}
	fmt.Printf("Setting BytesSentTo to: %v\n", temp)
	b.BytesSentTo = temp
	b.BytesSentSum = sum
	return b.BytesSentSum
}

type TimeKeyValue struct {
	Typ   string
	Value float64
}

func (b *Benchmark) GetTimeSpent() {
	temp := make([]TimeKeyValue, 0)
	b.timeSpent.Range(
		func(key, value any) bool {
			if rbFloat, ok := value.(float64); ok {
				if k, ok := key.(string); ok {
					v := TimeKeyValue{
						Typ:   k,
						Value: rbFloat,
					}
					temp = append(temp, v)
				}
			}
			return true
		})
	b.TimeSpent = temp

}

func (b *Benchmark) GetBytesReceivedSum() int {

	sum := 0
	fmt.Printf("Getting Bytes Sent Sum from sync.Map\n")
	fmt.Printf("Current sync.Map: %v\n", b.bytesReceivedFrom)
	temp := make([]BytesKeyValue, 0)
	b.bytesReceivedFrom.Range(
		func(key, receivedBytes any) bool {

			if rbInt, ok := receivedBytes.(int); ok {
				fmt.Printf("Appending to BytesReceivedFrom: %v , %v \n", rbInt, key)
				sum += rbInt
				if k, ok := key.(BytesKey); ok {
					v := BytesKeyValue{
						BytesKey: k,
						Value:    rbInt,
					}

					temp = append(temp, v)
				} else {
					fmt.Printf("Benchmark: GetBytesReceivedSum: key is not BytesKey: %v\n", key)
				}
			} else {
				fmt.Printf("Benchmark: GetBytesReceivedSum: receivedBytes is not int: %v\n", receivedBytes)
			}

			return true
		})
	fmt.Printf("Setting BytesReceivedFrom to: %v\n", temp)
	b.BytesReceivedFrom = temp
	b.BytesReceivedSum = sum
	return b.BytesReceivedSum

}
func (b *Benchmark) SetStartTime() {
	if !b.alreadySetStartTime {
		b.alreadySetStartTime = true
		timestamp := time.Now()
		fmt.Printf("Benchmark: Setting start time to %s\n", timestamp.Format(time.RFC3339))
		b.StartTime = timestamp
	}

}
func (b *Benchmark) CalculateRunTime() {
	b.RunTime = float64(time.Since(b.StartTime).Nanoseconds()) * math.Pow10(-9)
}

func (b *Benchmark) SetDoolOutput(output string) {
	b.Dool = output
}

type BytesKeyValue struct {
	BytesKey
	Value int
}

func NewBenchmark(node *Node) Benchmark {
	bytesSentTo := make([][]BytesKeyValue, 2*node.allCount)

	number_of_msg_types := len(pb.MsgType_value)
	for i := 0; i < 2*node.allCount; i++ {
		bytesSentTo[i] = make([]BytesKeyValue, number_of_msg_types)
		for j := 0; j < number_of_msg_types; j++ {
			bytesSentTo[i][j] = BytesKeyValue{
				BytesKey: BytesKey{
					CID: -1,
					Typ: pb.MsgType(j).String(),
				},
				Value: 0,
			}
		}

	}
	timeSpent := new(sync.Map)
	bytesReceivedFrom := new(sync.Map)
	// for i := 0; i < node.allCount; i++ {
	// 	bytesSentTo.Store(i, 0)
	// 	bytesReceivedFrom.Store(i, 0)
	// }
	// create per-cid locks

	bytesSentToLocks := make([]sync.Mutex, 2*node.allCount)

	benchmark := Benchmark{
		TimeSpentForDecision:   make(map[pb.MsgType]float64),
		TimeSpentFor:           make(map[pb.MsgType]map[string]float64),
		incrementTimeSpentChan: make(chan Measure, 1000),
		stop:                   false,
		stopped:                make(chan bool),
		BytesSentTo:            make([]BytesKeyValue, 0),
		BytesReceivedFrom:      make([]BytesKeyValue, 0),
		Throughput:             -1,
		bytesSentTo:            bytesSentTo,
		bytesReceivedFrom:      bytesReceivedFrom,
		timeSpent:              timeSpent,
		bytesSentToLocks:       bytesSentToLocks,
	}

	return benchmark
}

func (b *Benchmark) getTimeSpentFor(t pb.MsgType, k string) float64 {
	_, ok := b.TimeSpentFor[t]
	if ok {
		val, ok := b.TimeSpentFor[t][k]
		if ok {
			return val
		} else {
			return 0
		}
	}
	return 0
}

func (b *Benchmark) setTimeSpentFor(t pb.MsgType, k string, val float64) {
	_, ok := b.TimeSpentFor[t]
	if ok {
		b.TimeSpentFor[t][k] = val
	} else {
		b.TimeSpentFor[t] = make(map[string]float64)
		b.TimeSpentFor[t][k] = val
	}
}

func (b *Benchmark) getTimeSpentForDecision(t pb.MsgType) float64 {
	val, ok := b.TimeSpentForDecision[t]
	if ok {
		return val
	}
	return 0
}

func (b *Benchmark) setTimeSpentForDecision(t pb.MsgType, val float64) {
	b.TimeSpentForDecision[t] = val
}

func elapsedTimeInSeconds(duration time.Duration) float64 {
	newVar := (float64(duration.Nanoseconds()) * math.Pow10(-9))
	return newVar
}

// func (b *Benchmark) addToincrementTimeSpent(measure Measure) {
// 	if measure.Kind == STOP {
// 		b.incrementTimeSpentChan <- measure
// 	} else {
// 		return
// 	}

// }

// func (b *Benchmark) incrementTimeSpent(measure Measure) {
// 	t := measure.Type
// 	kind := measure.Kind
// 	duration := measure.Duration
// 	b.setTimeSpentFor(t, kind, b.getTimeSpentFor(t, kind)+elapsedTimeInSeconds(duration))
// }
