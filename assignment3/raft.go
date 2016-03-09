package raftnode
import (
	import "github.com/cs733-iitb/cluster"
	import "github.com/cs733-iitb/log"
)

// Returns a Node object
func raft.New(config Config) Node

type Event interface{}

type Node interface {
	// Client's message to Raft node
	Append([]byte)
	// Last known committed index in the log. This could be -1 until the system stabilizes.
	CommittedIndex() int
	// Returns the data at a log index, or an error.
	Get(index int) (err, []byte)
	// Node's id
	Id()
	// Id of leader. -1 if unknown
	LeaderId() int
	// Signal to shut down all goroutines, stop sockets, flush log and close it, cancel timers.
	Shutdown()
}
// data goes in via Append, comes out as CommitInfo from the node's CommitChannel
// Index is valid only if err == nil
type struct CommitInfo {
	Data			[]byte
	Index			int64 // or int .. whatever you have in your code
	Err			error // Err can be errred
}
// This is an example structure for Config .. change it to your convenience.
type struct Config {
	cluster			cluster.Config// Information about all servers, including this.
	srv			ServerImpl
	Id			int // this node's id. One of the cluster's entries should match.
	lg			*log.Log// Log file directory for this node
	ElectionTimeout		int
	HeartbeatTimeout	int
}
type RaftNode struct { // implements Node interface
	config 			*Config
	sm			*StateMachine
	eventCh 		chan Event
	// A channel for client to listen on. What goes into Append must come out of here at some point.
	CommitChannel 		chan CommitInfo
	timeoutCh 		chan Time
}

func NewConf(myid int, configuration interface{}) (config *Config, err error) {
	var config *Config
	if config, err = ToConfig(configuration); err != nil {
		return nil, err
	}
	if config.InboxSize == 0 {
		config.InboxSize = 10
	}
	if config.OutboxSize == 0 {
		config.OutboxSize = 10
	}
	srvr := new(serverImpl)
	server = srvr
	srvr.inbox = make(chan *Envelope, config.InboxSize)
	srvr.outbox = make(chan *Envelope, config.OutboxSize)
	srvr.peers = []*Peer{}
	err = srvr.initEndPoints(myid, config)
	return srvr, err
}

func NewSM(myid int, configuration interface{}) (sm *StateMachine, err error) {
	var sm *StateMachine
	if config, err = ToSM(configuration); err != nil {
		return nil, err
	}
	if config.InboxSize == 0 {
		config.InboxSize = 10
	}
	if config.OutboxSize == 0 {
		config.OutboxSize = 10
	}
	srvr := new(StateMachine)
	server = srvr
	srvr.inbox = make(chan *Envelope, config.InboxSize)
	srvr.outbox = make(chan *Envelope, config.OutboxSize)
	srvr.peers = []*Peer{}
	err = srvr.initEndPoints(myid, config)
	return srvr, err
}

func ToConfig(configuration interface{}) (config *Config, err error){
	var cfg Config
	var ok bool
	var configFile string
	if configFile, ok = configuration.(string); ok {
		var f *os.File
		if f, err = os.Open(configFile); err != nil {
			return nil, err
		}
		defer f.Close()
		dec := json.NewDecoder(f)
		if err = dec.Decode(&cfg); err != nil {
			return nil, err
		}
	} else if cfg, ok = configuration.(Config); !ok {
		return nil, errors.New("Expected a configuration.json file or a Config structure")
	}
	return &sm, nil
}

func ToSM(configuration interface{}) (sm *StateMachine, err error){
	var sm StateMachine
	var ok bool
	var configFile string
	if configFile, ok = configuration.(string); ok {
		var f *os.File
		if f, err = os.Open(configFile); err != nil {
			return nil, err
		}
		defer f.Close()
		dec := json.NewDecoder(f)
		if err = dec.Decode(&cfg); err != nil {
			return nil, err
		}
	} else if cfg, ok = configuration.(StateMachine); !ok {
		return nil, errors.New("Expected a sm.json file or a StateMachine structure")
	}
	return &cfg, nil
}
func (rn *RaftNode) Append(data) {
	rn.eventCh <- Append{data: data}
}
func (rn *RaftNode) CommittedIndex() int {
	
}
func (rn *RaftNode) Get() int {
	
}
func (rn *RaftNode) Id() int {
	return rn.sm.id
}
func (rn *RaftNode) LeaderId() int {
	return rn.sm.leaderId
}
func (rn *RaftNode) Shutdown() int {
	
}
func (rn *RaftNode) processEvents {
	for {
		var ev Event
		select {
			case msg := <- rn.eventCh {ev = msg}
			case <- rn.timeoutCh {ev = Timeout{}}
			case env := <-rn.srv.Inbox() {rn.env.Msg}
		}
	}
	actions := sm.ProcessEvent(ev)
	doActions(actions)
}
func (rn *RaftNode) processInbox {
	for {
		select {
			case env := <-rn.srv.Inbox() {rn.eventCh<-env.Msg}
		}
	}
}
