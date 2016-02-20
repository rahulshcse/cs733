package statemachine
import (
	"sort"
)

type Append struct {
	data			[]byte
}

type LogEntry struct {
	term			int
	data			[]byte
}

type VoteReq struct {
	term			int
	candidateId		int
	lastLogIndex		int
	lastLogTerm		int
}

type VoteResp struct {
	fromId			int
	term			int
	voteGranted		bool
}

type AppendEntriesReq struct {
	term			int
	leaderId		int
	prevLogIndex		int
	prevLogTerm		int
	entries			[]LogEntry
	leaderCommit 		int
}

type AppendEntriesResp struct {
	msg			AppendEntriesReq
	fromId			int
	term			int
	success			bool
}

type Timeout struct {
}

type SendEvent struct {
	destId			int
	event			interface{}
}

type StateMachine struct {
	id 			int
	leaderId		int
	serverIds		[]int
	state			int
	term			int
	voteCount		int
	negVoteCount		int
	votedFor		int
	log			[]LogEntry
	lastLogIndex		int
	commitIndex		int
	nextIndex		map[int]int
	matchIndex		map[int]int
	sendEvents		[]SendEvent
}
//Constants to be used to determine state
const (
	follower  		int = 1
	candidate 		int = 2
	leader    		int = 3
)
  

//_____________________________________________________________________________________________________________________



func (sm *StateMachine) Alarm() {
	// Timer reset. Functionality not included as part of this assignment
}
  

//_____________________________________________________________________________________________________________________


func (sm *StateMachine) StateStore() {
	// Store commitIndex, term and votedFor in persistent storage. Functionality not included as part of this assignment 
}
  

//_____________________________________________________________________________________________________________________


// Append log, increment last log index, and update log on disk
func (sm *StateMachine) LogStore(index int, term int, data []byte) {
	sm.log=append(sm.log,LogEntry{term,data})
	sm.lastLogIndex++
	//Update log on disk. Functionality not included as part of this assignment 
}
  

//_____________________________________________________________________________________________________________________


// Delete log entries starting from index, update last log index, update log on disk
func (sm *StateMachine) LogDelete(index int) {
	sm.log=sm.log[:index]
	sm.lastLogIndex=index-1
	//Update log on disk. Functionality not included as part of this assignment 
}
  

//_____________________________________________________________________________________________________________________


// return a send event
func (sm *StateMachine) Send(destId int,event interface{}) SendEvent {
	return SendEvent{destId,event}
}
  

//_____________________________________________________________________________________________________________________


// convert to follower state by changing parameters of state machine
func (sm *StateMachine) toFollower() []SendEvent {
	sm.state=follower
	sm.leaderId=-1
	sm.voteCount=0
	sm.negVoteCount=0
	sm.nextIndex=nil
	sm.matchIndex=nil
	return nil
}
  

//_____________________________________________________________________________________________________________________


// convert to candidate state by changing parameters of state machine
func (sm *StateMachine) toCandidate() []SendEvent {
	sm.state=candidate
	sm.leaderId=-1
	sm.voteCount=0
	sm.negVoteCount=0
	sm.nextIndex=nil
	sm.matchIndex=nil
	return nil
}
  

//_____________________________________________________________________________________________________________________


// convert to leader state by changing parameters of state machine
func (sm *StateMachine) toLeader() []SendEvent {
	var sendEvents []SendEvent
	var sendEvent SendEvent
	sm.state=leader
	sm.leaderId=sm.id
	sm.voteCount=0
	sm.negVoteCount=0
	sm.nextIndex=make(map[int]int)
	sm.matchIndex=make(map[int]int)
	sm.Alarm()
	sm.LogStore(sm.lastLogIndex+1,sm.term, nil)
	// Send AppendEntries Request containg leader term entry to all followers
	// request contains all logs beginning with next index for individual peers
	for _,id := range sm.serverIds {
		if id != sm.id {
			sm.nextIndex[id]=sm.lastLogIndex
			sm.matchIndex[id]=0
			sendEvent = sm.Send(id, AppendEntriesReq{sm.term, sm.id, sm.nextIndex[id]-1, sm.log[sm.nextIndex[id]-1].term,sm.log[sm.nextIndex[id]:], sm.commitIndex})
			sendEvents = append(sendEvents,sendEvent)
		}
	}
	return sendEvents
}
  

//_____________________________________________________________________________________________________________________


// handles all requests from layer above to append the data to the replicated log
func (sm *StateMachine) handleAppendEvent(ev Append) {
	var sendEvents []SendEvent
	var sendEvent SendEvent
	switch sm.state {
		case follower:
				;// Ignore
		case candidate:
				;// Ignore
		case leader:
				// store data in log
				sm.LogStore(sm.lastLogIndex+1,sm.term, ev.data)
				// Send AppendEntries Request to all followers
				// containing all logs beginning with next index for individual peers
				for _,id := range sm.serverIds {
					if id != sm.id {
						sendEvent = sm.Send(id, AppendEntriesReq{sm.term, sm.id, sm.nextIndex[id]-1, sm.log[sm.nextIndex[id]-1].term,sm.log[sm.nextIndex[id]:], sm.commitIndex})
						sendEvents = append(sendEvents,sendEvent)
					}
				}
	}
	sm.sendEvents=sendEvents
}
  

//_____________________________________________________________________________________________________________________


// handles election timeout for candidate/follower
// handles heartbeat timeout for leader
func (sm *StateMachine) handleTimeoutEvent(ev Timeout) {
	var sendEvents []SendEvent
	var sendEvent SendEvent
	switch sm.state {
		case follower:
				// election timeout, change to candidate, increment term, vote for self
				sm.sendEvents = sm.toCandidate()
				sm.term++
				sm.votedFor=sm.id
				sm.StateStore()
				// reset election timeout
				sm.Alarm()
				// send vote request to all peers
				for _,id := range sm.serverIds {
					if id != sm.id {
						sendEvent = sm.Send(id, VoteReq{sm.term, sm.id, sm.lastLogIndex, sm.log[sm.lastLogIndex].term})
						sendEvents = append(sendEvents,sendEvent)
					}
				}
		case candidate:
				// election timeout, change to candidate(restart as candidate), increment term, vote for self
				sm.sendEvents = sm.toCandidate()
				sm.term++
				sm.votedFor=sm.id
				sm.StateStore()
				// reset election timeout
				sm.Alarm()
				// send vote request to all peers
				for _,id := range sm.serverIds {
					if id != sm.id {
						sendEvent = sm.Send(id, VoteReq{sm.term, sm.id, sm.lastLogIndex, sm.log[sm.lastLogIndex].term})
						sendEvents = append(sendEvents,sendEvent)
					}
				}
		case leader:
				// hearbeat timeout, reset alarm, add empty log
				// send heart beats(append entry request) to all peers
				// containing all logs beginning with next index for individual peers
				sm.Alarm()
				sm.LogStore(sm.lastLogIndex+1,sm.term, nil)
				for _,id := range sm.serverIds {
					if id != sm.id {
						sendEvent = sm.Send(id, AppendEntriesReq{sm.term, sm.id, sm.nextIndex[id]-1, sm.log[sm.nextIndex[id]-1].term,sm.log[sm.nextIndex[id]:], sm.commitIndex})
						sendEvents = append(sendEvents,sendEvent)
					}
				}
	}
	sm.sendEvents = sendEvents
}
  

//_____________________________________________________________________________________________________________________



// handle all vote request events
func (sm *StateMachine) handleVoteReqEvent(ev VoteReq) {
	var sendEvents []SendEvent
	var sendEvent SendEvent
	switch sm.state {
		case follower:
				// check if not voted in case of same term
				// or event term is greater than term of state machine
				if (sm.votedFor == -1 && sm.term == ev.term)|| sm.term < ev.term {
					sm.term = ev.term
                        		if sm.lastLogIndex==0 || sm.log[sm.lastLogIndex].term<ev.lastLogTerm || (sm.log[sm.lastLogIndex].term==ev.lastLogTerm && sm.lastLogIndex<=ev.lastLogIndex) {
        					// Grant vote
						sm.votedFor = ev.candidateId
						sendEvent = sm.Send(ev.candidateId, VoteResp{sm.id,sm.term,true})
						sendEvents = append(sendEvents,sendEvent)
					} else {
						// Reject vote
						sm.votedFor = -1
						sendEvent = sm.Send(ev.candidateId, VoteResp{sm.id,sm.term,false})
						sendEvents = append(sendEvents,sendEvent)
					}
					sm.StateStore()
					// set election timeout
					sm.Alarm()
				} else {
					// Reject vote
					sendEvent = sm.Send(ev.candidateId, VoteResp{sm.id,sm.term,false})
					sendEvents = append(sendEvents,sendEvent)
				}
		case candidate:
				// check if event term is greater than term of state machine
				if sm.term < ev.term {
					// step down to follower
                        		sm.sendEvents = sm.toFollower()
					sm.term = ev.term
                        		if sm.lastLogIndex==0 || sm.log[sm.lastLogIndex].term<ev.lastLogTerm || (sm.log[sm.lastLogIndex].term==ev.lastLogTerm && sm.lastLogIndex<=ev.lastLogIndex) {
        					// Grant vote
						sm.votedFor = ev.candidateId
						sendEvent = sm.Send(ev.candidateId, VoteResp{sm.id,sm.term,true})
						sendEvents = append(sendEvents,sendEvent)
					} else {
						// Reject vote
						sm.votedFor = -1
						sendEvent = sm.Send(ev.candidateId, VoteResp{sm.id,sm.term,false})
						sendEvents = append(sendEvents,sendEvent)
					}
					sm.StateStore()
					// set election timeout
					sm.Alarm()
				} else {
					// Reject vote
					sendEvent = sm.Send(ev.candidateId, VoteResp{sm.id,sm.term,false})
					sendEvents = append(sendEvents,sendEvent)
				}
		case leader:
				// check if event term is greater than term of state machine
				if sm.term < ev.term {
					// step down to follower
                        		sm.sendEvents = sm.toFollower()
					sm.term = ev.term
                        		if sm.lastLogIndex==0 || sm.log[sm.lastLogIndex].term<ev.lastLogTerm || (sm.log[sm.lastLogIndex].term==ev.lastLogTerm && sm.lastLogIndex<=ev.lastLogIndex) {
        					// Grant vote
						sm.votedFor = ev.candidateId
						sendEvent = sm.Send(ev.candidateId, VoteResp{sm.id,sm.term,true})
						sendEvents = append(sendEvents,sendEvent)
					} else {
						// Reject vote
						sm.votedFor = -1
						sendEvent = sm.Send(ev.candidateId, VoteResp{sm.id,sm.term,false})
						sendEvents = append(sendEvents,sendEvent)
					}
					sm.StateStore()
					// set election timeout
					sm.Alarm()
				} else {
					// Reject vote
					sendEvent = sm.Send(ev.candidateId, VoteResp{sm.id,sm.term,false})
					sendEvents = append(sendEvents,sendEvent)
				}	
	}
	sm.sendEvents = sendEvents
}
  

//_____________________________________________________________________________________________________________________



// handle all vote responses
func (sm *StateMachine) handleVoteRespEvent(ev VoteResp) {
	var sendEvents []SendEvent
	switch sm.state {
		case follower:
				// delayed response
				// update term if event term greater, else reject
				if ev.term > sm.term {
					sm.term = ev.term
					sm.votedFor = -1
					sm.StateStore()
					sm.Alarm()
				}
		case candidate:
				// ignore if event term less than term of state machine
				if ev.term < sm.term {
					;// Ignore delayed response
				} else if ev.term > sm.term {
					// if event term greater, step down to follower
					sm.term = ev.term
					sm.sendEvents = sm.toFollower()
					sm.votedFor = -1
					sm.StateStore()
					// set election timeout
					sm.Alarm()
				} else if ev.voteGranted == true {
					// response of same term, and vote request granted
					sm.voteCount++
					// check for majority acks
					if sm.voteCount>len(sm.serverIds)/2 {
						// move to leader state and send heart beats to all
						sendEvents = sm.toLeader()
					}
				} else {
					// nack
					sm.negVoteCount++
					// if nack count in majority, step back to follower state
					if sm.negVoteCount>=(len(sm.serverIds)+1)/2 {
						sm.sendEvents = sm.toFollower()
						sm.StateStore()
						// set election timeout
						sm.Alarm()
					}
					
				}
		case leader:
				// delayed response
				// update term if event term greater, else reject
				if ev.term > sm.term {
					sm.term = ev.term
					// step down to follower state
					sm.sendEvents = sm.toFollower()
					sm.votedFor = -1
					sm.StateStore()
					sm.Alarm()
				}
	}
	sm.sendEvents = sendEvents
}

  

//_____________________________________________________________________________________________________________________


// handle all append entries request events
func (sm *StateMachine) handleAppendEntriesReq(ev AppendEntriesReq) {
	var sendEvents []SendEvent
	var sendEvent SendEvent
	switch sm.state {
		case follower:
				// handle request if event term equal or greater, else reject
				if ev.term >= sm.term {
					// reset election timeout
                        		sm.Alarm()
					if sm.term < ev.term {
						// update term if event term greater
						sm.term = ev.term
						sm.votedFor = -1
						sm.StateStore()
						if sm.leaderId != ev.leaderId {
							sm.leaderId = ev.leaderId
						}
						
					}
					// if term at previous log index match
					if sm.lastLogIndex >= ev.prevLogIndex && sm.log[ev.prevLogIndex].term==ev.prevLogTerm {
						// if follower is out of sync with leader
						// delete all uncommitted entries after prevLogIndex sent by leader
						// and append entries sent by leader
						if sm.lastLogIndex > ev.prevLogIndex {
							sm.LogDelete(ev.prevLogIndex+1)
						}
						for _,v := range ev.entries {
							sm.LogStore(sm.lastLogIndex+1,v.term,v.data)
						}
						// advance commit index, and store it on persistent storage
						// can be used by file system server to check and apply committed logs
						if sm.commitIndex < ev.leaderCommit {
							sm.commitIndex = ev.leaderCommit
							sm.StateStore()
						}
						// send positive response
						sendEvent = sm.Send(ev.leaderId, AppendEntriesResp{ev,sm.id,sm.term,true})
						sendEvents = append(sendEvents,sendEvent)
					} else {
					// reject request
						sendEvent = sm.Send(ev.leaderId, AppendEntriesResp{ev,sm.id,sm.term,false})
						sendEvents = append(sendEvents,sendEvent)
					}
				} else {
					// reject request
					sendEvent = sm.Send(ev.leaderId, AppendEntriesResp{ev,sm.id,sm.term,false})
					sendEvents = append(sendEvents,sendEvent)
				}
		case candidate:
				// handle request if event term equal or greater, else reject
				if ev.term >= sm.term {
					// step down to follower state
					sm.sendEvents = sm.toFollower()
					// set election timeout
                        		sm.Alarm()
					if sm.term < ev.term {
						sm.term = ev.term
						sm.votedFor = -1
						sm.StateStore()
						if sm.leaderId != ev.leaderId {
							sm.leaderId = ev.leaderId
						}
						
					}
					// if term at previous log index match
					if sm.lastLogIndex >= ev.prevLogIndex && sm.log[ev.prevLogIndex].term==ev.prevLogTerm {
						// when follower is out of sync with leader
						// delete all uncommitted entries after prevLogIndex sent by leader
						// and append entries sent by leader
						if sm.lastLogIndex > ev.prevLogIndex {
							sm.LogDelete(ev.prevLogIndex+1)
						}
						for _,v := range ev.entries {
							sm.LogStore(sm.lastLogIndex+1,v.term,v.data)
						}
						// advance commit index, and store it on persistent storage
						// can be used by file system server to check and apply committed logs
						if sm.commitIndex < ev.leaderCommit {
							sm.commitIndex = ev.leaderCommit
							sm.StateStore()
						}
						// send positive response
						sendEvent = sm.Send(ev.leaderId, AppendEntriesResp{ev,sm.id,sm.term,true})
						sendEvents = append(sendEvents,sendEvent)
					} else {
						// reject request
						sendEvent = sm.Send(ev.leaderId, AppendEntriesResp{ev,sm.id,sm.term,false})
						sendEvents = append(sendEvents,sendEvent)
					}
				} else {
					// reject request
					sendEvent = sm.Send(ev.leaderId, AppendEntriesResp{ev,sm.id,sm.term,false})
					sendEvents = append(sendEvents,sendEvent)
				}
		case leader:
				// handle request if event term greater, else reject
				if ev.term > sm.term {
					// step down to follower
					sm.sendEvents = sm.toFollower()
					// set election timeout
                        		sm.Alarm()
					sm.term = ev.term
					sm.votedFor = -1
					sm.StateStore()
					if sm.leaderId != ev.leaderId {
						sm.leaderId = ev.leaderId
					}
					// if term at previous log index match
					if sm.lastLogIndex >= ev.prevLogIndex && sm.log[ev.prevLogIndex].term==ev.prevLogTerm {
						// when follower is out of sync with leader
						// delete all uncommitted entries after prevLogIndex sent by leader
						// and append entries sent by leader
						if sm.lastLogIndex > ev.prevLogIndex {
							sm.LogDelete(ev.prevLogIndex+1)
						}
						for _,v := range ev.entries {
							sm.LogStore(sm.lastLogIndex+1,v.term,v.data)
						}
						// advance commit index, and store it on persistent storage
						// can be used by file system server to check and apply committed logs
						if sm.commitIndex < ev.leaderCommit {
							sm.commitIndex = ev.leaderCommit
							sm.StateStore()
						}
						// send positive response
						sendEvent = sm.Send(ev.leaderId, AppendEntriesResp{ev,sm.id,sm.term,true})
						sendEvents = append(sendEvents,sendEvent)
					} else {
						// reject request
						sendEvent = sm.Send(ev.leaderId, AppendEntriesResp{ev,sm.id,sm.term,false})
						sendEvents = append(sendEvents,sendEvent)
					}
				} else {
					// reject request
					sendEvent = sm.Send(ev.leaderId, AppendEntriesResp{ev,sm.id,sm.term,false})
					sendEvents = append(sendEvents,sendEvent)
				}
	}
	sm.sendEvents = sendEvents
	
}
  

//_____________________________________________________________________________________________________________________



// handle all append entries response events
func (sm *StateMachine) handleAppendEntriesResp(ev AppendEntriesResp) {
	var sendEvents []SendEvent
	var sendEvent SendEvent
	switch sm.state {
		case follower:
				// delayed response
				// update term if event term greater, else reject
				if ev.term > sm.term {
					sm.term = ev.term
					sm.votedFor = -1
					sm.StateStore()
					// set election timeout
					sm.Alarm()
				}
		case candidate:
				// delayed response
				// update term if event term greater, else reject
				if ev.term > sm.term {
					// step down to follower
					sm.sendEvents = sm.toFollower()
					sm.term = ev.term
					sm.votedFor = -1
					sm.StateStore()
					// reset election timeout
					sm.Alarm()
				}
		case leader:
				// ignore if event term less than term of state machine
				if ev.term < sm.term {
					;//Ignore delayed response
				} else if ev.term > sm.term {
					// if event term greater, step down to follower
					sm.sendEvents = sm.toFollower()
					sm.term = ev.term
					sm.votedFor = -1
					sm.StateStore()
					// set election timeout
					sm.Alarm()
				} else if ev.success == true {
					// response of same term, and append entries request success
					// update next index and match index
					sm.nextIndex[ev.fromId] = sm.nextIndex[ev.fromId] + len(ev.msg.entries)
					sm.matchIndex[ev.fromId] = sm.nextIndex[ev.fromId]-1
					// sort match indices
					var temparr []int
					for _,v:=range sm.matchIndex {
						temparr=append(temparr,v)
					}
					sort.Ints(temparr)
					tempindex := temparr[(len(temparr)-1)/2]
					// update commit index based on one less than median value of sorted match indices
					// represent the index replicated on majority of servers
					if tempindex > sm.commitIndex && sm.log[tempindex].term==sm.term {
						sm.commitIndex = tempindex
						sm.StateStore()
					}
				} else {
					// Resend append entry request with decremented nextIndex of follwer
					sm.nextIndex[ev.fromId] = sm.nextIndex[ev.fromId] - 1
					sendEvent = sm.Send(ev.fromId, AppendEntriesReq{sm.term, sm.id, sm.nextIndex[ev.fromId]-1, sm.log[sm.nextIndex[ev.fromId]-1].term,sm.log[sm.nextIndex[ev.fromId]:], sm.commitIndex})
					sendEvents = append(sendEvents,sendEvent)
				}
	}
	sm.sendEvents = sendEvents
}
  

//_____________________________________________________________________________________________________________________



func (sm *StateMachine) ProcessEvent (ev interface{}) {
	switch ev.(type) {
		case Append:
				sm.handleAppendEvent(ev.(Append))
		case Timeout:
				sm.handleTimeoutEvent(ev.(Timeout))
		case VoteReq:
				sm.handleVoteReqEvent(ev.(VoteReq))
		case VoteResp:
				sm.handleVoteRespEvent(ev.(VoteResp))
		case AppendEntriesReq:
				sm.handleAppendEntriesReq(ev.(AppendEntriesReq))
		case AppendEntriesResp:
				sm.handleAppendEntriesResp(ev.(AppendEntriesResp))
		default: 
				//Ignore
				sm.sendEvents=nil;
	}
}

