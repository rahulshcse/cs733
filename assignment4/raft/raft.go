package raftnode
import (
	clusterorg "github.com/cs733-iitb/cluster"
	cluster "github.com/cs733-iitb/cluster/mock"
	"encoding/gob"
	log "github.com/cs733-iitb/log"
	"time"
	"os"
	"encoding/json"
	//"path/filepath"
	"errors"
	"strconv"
	"fmt"
	"bufio"
)

// Returns a Node object
// func raft.New(config Config) Node

type Event interface{}
var MsgId int64 = 1

// data goes in via Append, comes out as CommitInfo from the node's CommitChannel
// Index is valId only if err == nil
type CommitInfo struct {
	Data			[]byte
	Index			int64
	Err				error // Err can be errred
}
type Node interface {
	// Client's message to Raft node
	Append([]byte)
	// Last known committed index in the log. This could be -1 until the system stabilizes.
	CommittedIndex() int
	// Returns the data at a log index, or an error.
	Get(index int64) ([]byte, error)
	// Node's Id
	Id() int
	// Id of leader. -1 if unknown
	LeaderId() int
	// Signal to shut down all goroutines, stop sockets, flush log and close it, cancel Timers.
	Shutdown()
	IsClosedRaft() bool
	ProcessEvents()
	CommitCh() chan CommitInfo
	Up()
}
type RaftNode struct { // implements Node interface
	Server			*cluster.MockServer
	Sm			StateMachine
	EventCh 		chan Event
	// A channel for client to listen on. What goes into Append must come out of here at some point.
	CommitChannel 		chan CommitInfo
	ShutDown 		chan int
	AppEvent		chan Event
	Timer			*time.Timer
	ElectionTimeout		int
	HeartbeatTimeout	int
	Closed 				int
	Isdown				int
}

// 
type TimeoutInt struct{
	Id			int
	ElectionTimeout		int
	HeartbeatTimeout	int
}
type TimeoutArr struct {
	Tm			[]TimeoutInt
}
func init() {
	gob.Register(VoteReq{})
	gob.Register(VoteResp{})
	gob.Register(AppendEntriesReq{})
	gob.Register(AppendEntriesResp{})
}
  

//_____________________________________________________________________________________________________________________

func DeleteOldLogs() { 
	os.RemoveAll("statelog1")
	os.RemoveAll("statelog2")
	os.RemoveAll("statelog3")
	os.RemoveAll("statelog4")
	os.RemoveAll("statelog5")
	os.RemoveAll("log1")
	os.RemoveAll("log2")
	os.RemoveAll("log3")
	os.RemoveAll("log4")
	os.RemoveAll("log5")
	os.RemoveAll("testlog1")
	os.RemoveAll("testlog2")
	os.RemoveAll("testlog3")
	os.RemoveAll("testlog4")
	os.RemoveAll("testlog5")
 }

func ToTimeoutArr(intf interface{}) (TimeoutArr, error){
	var tm TimeoutArr
	var ok bool
	var configFile string
	var err error
	if configFile, ok = intf.(string); ok {
		var f *os.File
		if f, err = os.Open(configFile); err != nil {
			return TimeoutArr{}, err
		}
		defer f.Close()
		dec := json.NewDecoder(f)
		if err = dec.Decode(&tm); err != nil {
			return TimeoutArr{}, err
		}
	} else if tm, ok = intf.(TimeoutArr); !ok {
		return TimeoutArr{}, errors.New("Expected a configuration.json file or a Config structure")
	}
	return tm, nil
}

// Create new Raft Node
func NewNode(myId int)(Node) {
	rf := new(RaftNode)
	//rf.Sm.InitState()
	rf.EventCh = make(chan Event, 1000)
	rf.AppEvent = make(chan Event, 1000)
	rf.CommitChannel = make(chan CommitInfo, 1000)
	rf.ShutDown = make(chan int, 1)
	rf.Closed = 0
	rf.Isdown = 0
	tmArr,_:= ToTimeoutArr("timeout_config.json")
	for _, tmint:= range tmArr.Tm {
		if myId==tmint.Id {
			rf.ElectionTimeout = tmint.ElectionTimeout
			rf.HeartbeatTimeout= tmint.HeartbeatTimeout
		}
	}
	rf.Timer = time.NewTimer(time.Millisecond*time.Duration(rf.ElectionTimeout))
	return rf
}

// Initialize State Machine
func (rn *RaftNode) NewSm(Id int, peerIds []int) StateMachine {
	var Sm StateMachine
	Sm.RaftNode=rn
	Sm.Id=Id
	Sm.LeaderId=-1
	Sm.ServerIds=peerIds
	Sm.State=follower
	Sm.Term=0
	Sm.is_not_sm_test=1
	Sm.VoteCount=0
	Sm.NegVoteCount=0
	Sm.VotedFor=-1
	Sm.Log=[]LogEntry{}
	Sm.InitLog()
	Sm.LastLogIndex=0
	Sm.CommitIndex=0
	Sm.NextIndex=nil
	Sm.MatchIndex=nil
	Sm.SendEvents=nil
	return Sm
}

// Handling append request of clients
func (rn *RaftNode) Append(data []byte) {
	rn.AppEvent <- Append{Data: data}
}

// Get latest committed index of a node
func (rn *RaftNode) CommittedIndex() int {
	return rn.Sm.CommitIndex
}

// Get log from storage corresponding to an index
func (rn *RaftNode) Get(index int64) ([]byte, error) {
	rn.Sm.Lock()
	defer rn.Sm.Unlock()
	logname:="log"
	logname+=strconv.Itoa(rn.Sm.Id)
	lg,err1 := log.Open(logname)
	if err1!=nil {
		return nil, err1
	}
	defer lg.Close()
	logentry,err2 := lg.Get(index)
	if err2!=nil {
		return nil, err2
	}
	return logentry.(LogEntry).Data, nil
}

// Access Commit Channel, which contain replicated and committed Append data
func (rn *RaftNode) CommitCh() chan CommitInfo  { return rn.CommitChannel }

// Get node Id
func (rn *RaftNode) Id() int {
	return rn.Sm.Id
}

// Get Leader Id
func (rn *RaftNode) LeaderId() int {
	return rn.Sm.LeaderId
}

// to shut down individual nodes
func (rn *RaftNode) Shutdown() {
	fmt.Printf("Shutting Down:%d\n", rn.Sm.Id)
	rn.ShutDown <- 1
	for rn.Closed != 1 {
		time.Sleep(500*time.Millisecond)
	}
	rn.Server.Close()
	close(rn.EventCh)
	close(rn.CommitChannel)
	close(rn.AppEvent)
	rn.Timer.Stop()	
	rn.Sm.LeaderId = -1
	rn.Sm.State = follower
	fmt.Printf("\nRaft Node %d closed\n",rn.Id())
	rn.Closed = 2
}

func (rn *RaftNode) Up() {
	rn.Closed = 0
	rn.EventCh = make(chan Event, 1000)
	rn.AppEvent = make(chan Event, 1000)
	rn.CommitChannel = make(chan CommitInfo, 1000)
	rn.Timer = time.NewTimer(time.Millisecond*time.Duration(rn.ElectionTimeout))
}

func (rn *RaftNode) IsClosedRaft() bool {
	if rn.Closed == 2 {
		return true
	} else {
		return false
	}
}

//_____________________________________________________________________________________________________________________

// creates raft node, initializes state machines in each node
// and run goroutines corresponding to each node
func MakeRafts() map[int]Node {
	cluster, _ := cluster.NewCluster("cluster_test_config.json")
	raftNode1 := NewNode(1)
	raftNode1.(*RaftNode).Server = cluster.Servers[1]
	peerIds:=raftNode1.(*RaftNode).Server.Peers()
	raftNode1.(*RaftNode).Sm=raftNode1.(*RaftNode).NewSm(1, peerIds)
	raftNodes := make(map[int]Node)
	raftNodes[1] = raftNode1
	for _,raftNodeId:= range raftNode1.(*RaftNode).Server.Peers() {
		newNode := NewNode(raftNodeId)
		newNode.(*RaftNode).Server = cluster.Servers[raftNodeId]
		peerIds:=newNode.(*RaftNode).Server.Peers()
		newNode.(*RaftNode).Sm=newNode.(*RaftNode).NewSm(raftNodeId, peerIds)
		raftNodes[raftNodeId] = newNode
	}
	return raftNodes
}

// Process All Events
func (rn *RaftNode) ProcessEvents() {
	for {
		if len(rn.Server.Inbox())!=0 {
		}
		select {
			case <-rn.ShutDown:// Shutdown request
				rn.Closed = 1
				fmt.Printf("Raft %d Closed\n\n",rn.Sm.Id)
				return;
			case <-rn.Timer.C:// Timeout Event
				rn.EventCh <- Timeout{}
			case msg := <- rn.Server.Inbox():// Meassage from other node
				rn.EventCh <- msg.Msg
			case appev := <- rn.AppEvent:// Append request from client
				rn.EventCh <- appev
		}
		ev := <- rn.EventCh
		if len(rn.ShutDown) >0 {
			rn.Closed = 1
			<-rn.ShutDown
			fmt.Printf("Raft %d Closed\n\n",rn.Sm.Id)
			return;
		}
		if ev !=nil {// if there is an event to be processed
			// Process particular event
			//fmt.Printf("TT:%d\n",rn.Sm.Id)
			rn.Sm.ProcessEvent(ev)
			actions := rn.Sm.SendEvents
			rn.Sm.SendEvents = []SendEvent{}
			if actions !=nil && len(actions) !=0 {
				filename:="testlog"
				filename+=strconv.Itoa(rn.Sm.Id)
				filename+=".dat"
				f, err := os.OpenFile(filename, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
				if err != nil {
				}
				defer f.Close()
				b := bufio.NewWriter(f)
				defer func() {
				if err = b.Flush(); err != nil {
				}
				}()
				fmt.Fprintf(b,"No of Sendevents for %d:%d\nSendevents:%v\n",rn.Sm.Id,len(actions),actions)
				for _, sendEvent := range actions {
					env := &clusterorg.Envelope{Pid: sendEvent.DestId, MsgId: int64(MsgId), Msg: sendEvent.Event}
					MsgId++
					if rn.Server.IsClosed()==false {
						fmt.Fprintf(b,"Outbox max length:%d\n",cap(rn.Server.Outbox()))
						rn.Server.Outbox() <- env
						fmt.Fprintf(b,"Outbox current length\n",cap(rn.Server.Outbox()))
					} else {
						fmt.Fprintf(b,"Outbox closed\n")
					}
				}
			}
		}
		if len(rn.ShutDown) >0 {
			rn.Closed = 1
			<-rn.ShutDown
			fmt.Printf("Raft %d Closed\n\n",rn.Sm.Id)
			return;
		}
	}
}

// Get leader based on majority node state, nil if majority do not agree on particular leader
func GetLeader(rfs map[int]Node) Node{
	m := make(map[int]int)
	nodeId := -1
	for _,rf:= range rfs {
		//fmt.Printf("LeaderId for %d is= %d\n",rf.Id(), rf.LeaderId())
		m[rf.LeaderId()]++
	}
	for Id,count:= range m {
		if count > len(rfs)/2 {
			nodeId=Id
		}	
	}
	if nodeId != -1 {
		for _,rf:= range rfs {
			if rf.Id()==nodeId {
				return rf
			}
		}
		return nil
	}
	return nil
}

//Partitions servers into 2 or more partitions

func Partition(mc map[int]Node,partitions ...[]int) error {
	index := make(map[int]int)         // Pid -> index of partitions array
	for ip, part := range partitions { // for each partition
		for _, pid := range part { // for each pid in that partition
			// Associate pid to partition if not already assigned
			if partid, ok := index[pid]; ok {
				if partid != ip {
					return errors.New(fmt.Sprintf("Server id %d in different partitions: %+v", partid, partitions))
				}
			} else {
				index[pid] = ip // Add server to partition ip
			}
		}
	}

	// Create default partition for pids not accounted for.
	defaultPartition := []int{}
	idefaultPartition := len(partitions)
	for pid, _ := range mc {
		if _, ok := index[pid]; !ok {
			defaultPartition = append(defaultPartition, pid)
			index[pid] = idefaultPartition
		}
	}
	if len(defaultPartition) > 0 {
		partitions = append(partitions, defaultPartition)
	}

	// Inform servers of the partitions they belong to
	for pid, _ := range mc {
		mc[pid].(*RaftNode).Server.Partition(partitions[index[pid]])
	}

	return nil
}

//Heals multiple partitions

func Heal(mc map[int]Node) {
	for _, srv := range mc {
		srv.(*RaftNode).Server.Heal()
	}
}
