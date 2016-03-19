package raftnode
import (
	"sort"
	log "github.com/cs733-iitb/log"
	"strconv"
	"errors"
	"time"
	"bufio"
	"fmt"
	"os"
)

type Append struct {
	Data			[]byte
}

type LogEntry struct {
	Term			int
	Data			[]byte
}

type VoteReq struct {
	Term			int
	CandidateId		int
	LastLogIndex		int
	LastLogTerm		int
}
type VoteResp struct {
	FromId			int
	Term			int
	VoteGranted		bool
}

type AppendEntriesReq struct {
	Term			int
	LeaderId		int
	PrevLogIndex		int
	PrevLogTerm		int
	Entries			[]LogEntry
	LeaderCommit 		int
}

type AppendEntriesResp struct {
	Msg			AppendEntriesReq
	FromId			int
	Term			int
	Success			bool
}

type Timeout struct {
}

type SendEvent struct {
	DestId			int
	Event			interface{}
}

type StateMachine struct {
	Id 			int
	LeaderId		int
	RaftNode		*RaftNode
	ServerIds		[]int
	State			int
	Term			int
	VoteCount		int
	NegVoteCount		int
	VotedFor		int
	Log			[]LogEntry
	LastLogIndex		int
	CommitIndex		int
	NextIndex		map[int]int
	MatchIndex		map[int]int
	SendEvents		[]SendEvent
	lg			*log.Log
	statelg			*log.Log
}
type SMPersist struct {
	Term			int
	CommitIndex		int
	VotedFor		int
}
//Constants to be used to determine state
const (
	follower  		int = 1
	candidate 		int = 2
	leader    		int = 3
)
  

//_____________________________________________________________________________________________________________________



func (sm *StateMachine) Alarm(timertype int) {
	if timertype == 1 {
		(*sm.RaftNode).Timer.Reset(time.Millisecond*time.Duration((*sm.RaftNode).HeartbeatTimeout))
	} else if timertype == 2 {
		(*sm.RaftNode).Timer.Reset(time.Millisecond*time.Duration((*sm.RaftNode).ElectionTimeout))
	} else if timertype == 3 {
		(*sm.RaftNode).Timer = time.NewTimer(time.Millisecond*time.Duration((*sm.RaftNode).HeartbeatTimeout))
	} else if timertype == 4 {
		(*sm.RaftNode).Timer = time.NewTimer(time.Millisecond*time.Duration((*sm.RaftNode).HeartbeatTimeout))
	} else {
	}
}

  
//_____________________________________________________________________________________________________________________


func (sm *StateMachine) StateStore() {
	logname:="statelog"
	logname+=strconv.Itoa(sm.Id)
	lg,_ := log.Open(logname)
	lg.SetCacheSize(5000)
     	defer lg.Close()
	lg.RegisterSampleEntry(SMPersist{})
     	logentry:=SMPersist{sm.Term,sm.CommitIndex,sm.VotedFor}
     	lg.Append(logentry)
}
  

//_____________________________________________________________________________________________________________________


// Append log, increment last log index, and update log on disk
func (sm *StateMachine) InitLog() {
	logname:="log"
	logname+=strconv.Itoa(sm.Id)
	lg,_ := log.Open(logname)
     	logentry:=LogEntry{0,nil}
     	lg.Append(logentry)
	sm.Log=append(sm.Log,LogEntry{0,nil})
	sm.LastLogIndex++	
     	defer lg.Close()
}
  

//_____________________________________________________________________________________________________________________


// Append log, increment last log index, and update log on disk
func (sm *StateMachine) LogStore(index int, term int, data []byte) {
	logname:="log"
	logname+=strconv.Itoa(sm.Id)
	lg,_ := log.Open(logname)
	lg.SetCacheSize(5000)	
	lg.RegisterSampleEntry(LogEntry{})
     	defer lg.Close()
     	logentry:=LogEntry{term,data}
     	lg.Append(logentry)
	sm.Log=append(sm.Log,LogEntry{term,data})
	sm.LastLogIndex++
}
  

//_____________________________________________________________________________________________________________________


// Delete log entries starting from index, update last log index, update log on disk
func (sm *StateMachine) LogDelete(index int) {
	logname:="log"
	logname+=strconv.Itoa(sm.Id)
	lg,_ := log.Open(logname)
	lg.SetCacheSize(5000)
	lg.RegisterSampleEntry(LogEntry{})
     	defer lg.Close()
     	lg.TruncateToEnd(int64(index))
	sm.Log=sm.Log[:index]
	sm.LastLogIndex=index-1
}
  

//_____________________________________________________________________________________________________________________


// return a send event
func (sm *StateMachine) Send(destId int,event interface{}) SendEvent {
	return SendEvent{destId,event}
}
  

//_____________________________________________________________________________________________________________________


// convert to follower state by changing parameters of state machine
func (sm *StateMachine) toFollower() []SendEvent {
	sm.State=follower
	sm.LeaderId=-1
	sm.VoteCount=0
	sm.NegVoteCount=0
	sm.NextIndex=nil
	sm.MatchIndex=nil
	return nil
}
  

//_____________________________________________________________________________________________________________________


// convert to candidate state by changing parameters of state machine
func (sm *StateMachine) toCandidate() []SendEvent {
	sm.State=candidate
	sm.LeaderId=-1
	sm.VoteCount=0
	sm.NegVoteCount=0
	sm.NextIndex=nil
	sm.MatchIndex=nil
	return nil
}
  

//_____________________________________________________________________________________________________________________


// convert to leader state by changing parameters of state machine
func (sm *StateMachine) toLeader() []SendEvent {
	var sendEvents []SendEvent
	var sendEvent SendEvent
	sm.State=leader
	sm.LeaderId=sm.Id
	sm.VoteCount=0
	sm.NegVoteCount=0
	sm.NextIndex=make(map[int]int)
	sm.MatchIndex=make(map[int]int)
	sm.Alarm(1)
	sm.LogStore(sm.LastLogIndex+1,sm.Term, nil)
	// Send AppendEntries Request containg leader term entry to all followers
	// request contains all logs beginning with next index for individual peers
	for _,id := range sm.ServerIds {
		if id != sm.Id {
			sm.NextIndex[id]=sm.LastLogIndex
			sm.MatchIndex[id]=0
			sendEvent = sm.Send(id, AppendEntriesReq{sm.Term, sm.Id, sm.NextIndex[id]-1, sm.Log[sm.NextIndex[id]-1].Term,sm.Log[sm.NextIndex[id]:], sm.CommitIndex})
			sendEvents = append(sendEvents,sendEvent)
		}
	}
	return sendEvents
}
  

//_____________________________________________________________________________________________________________________


// handles all requests from layer above to append the data to the replicated log
func (sm *StateMachine) handleAppendEvent(ev Append) {
	//var sendEvents []SendEvent
	//var sendEvent SendEvent
	switch sm.State {
		case follower:
				(*sm.RaftNode).CommitChannel <- CommitInfo{Data:ev.Data, Index:-1, Err:errors.New("I am a follower")}
		case candidate:
				(*sm.RaftNode).CommitChannel <- CommitInfo{Data:ev.Data, Index:-1, Err:errors.New("I am a candidate")}
		case leader:
				// store data in log
				sm.LogStore(sm.LastLogIndex+1,sm.Term, ev.Data)
	}
}
  

//_____________________________________________________________________________________________________________________


// handles election timeout for candidate/follower
// handles heartbeat timeout for leader
func (sm *StateMachine) handleTimeoutEvent(ev Timeout) {
	var sendEvents []SendEvent
	var sendEvent SendEvent
	switch sm.State {
		case follower:
				// election timeout, change to candidate, increment term, vote for self
				sm.SendEvents = sm.toCandidate()
				sm.Term++
				sm.VotedFor=sm.Id
				sm.StateStore()
				// reset election timeout
				sm.Alarm(4)
				// send vote request to all peers
				for _,id := range sm.ServerIds {
					if id != sm.Id {
						sendEvent = sm.Send(id, VoteReq{sm.Term, sm.Id, sm.LastLogIndex, sm.Log[sm.LastLogIndex].Term})
						sendEvents = append(sendEvents,sendEvent)
					}
				}
		case candidate:
				// election timeout, change to candidate(restart as candidate), increment term, vote for self
				sm.SendEvents = sm.toCandidate()
				sm.Term++
				sm.VotedFor=sm.Id
				sm.StateStore()
				// reset election timeout
				sm.Alarm(4)
				// send vote request to all peers
				for _,id := range sm.ServerIds {
					if id != sm.Id {
						sendEvent = sm.Send(id, VoteReq{sm.Term, sm.Id, sm.LastLogIndex, sm.Log[sm.LastLogIndex].Term})
						sendEvents = append(sendEvents,sendEvent)
					}
				}
		case leader:
				// hearbeat timeout, reset alarm, add empty log
				// send heart beats(append entry request) to all peers
				// containing all logs beginning with next index for individual peers
				sm.Alarm(3)
				sm.LogStore(sm.LastLogIndex+1,sm.Term, nil)
				for _,id := range sm.ServerIds {
					if id != sm.Id {
						sendEvent = sm.Send(id, AppendEntriesReq{sm.Term, sm.Id, sm.NextIndex[id]-1, sm.Log[sm.NextIndex[id]-1].Term,sm.Log[sm.NextIndex[id]:], sm.CommitIndex})
						sendEvents = append(sendEvents,sendEvent)
					}
				}
	}
	sm.SendEvents = sendEvents
}
  

//_____________________________________________________________________________________________________________________



// handle all vote request events
func (sm *StateMachine) handleVoteReqEvent(ev VoteReq) {
	var sendEvents []SendEvent
	var sendEvent SendEvent
	switch sm.State {
		case follower:
				// check if not voted in case of same term
				// or event term is greater than term of state machine
				if (sm.VotedFor == -1 && sm.Term == ev.Term)|| sm.Term < ev.Term {
					sm.Term = ev.Term
                        		if sm.LastLogIndex==0 || sm.Log[sm.LastLogIndex].Term<ev.LastLogTerm || (sm.Log[sm.LastLogIndex].Term==ev.LastLogTerm && sm.LastLogIndex<=ev.LastLogIndex) {
        					// Grant vote
						sm.VotedFor = ev.CandidateId
						sendEvent = sm.Send(ev.CandidateId, VoteResp{sm.Id,sm.Term,true})
						sendEvents = append(sendEvents,sendEvent)
					} else {
						// Reject vote
						sm.VotedFor = -1
						sendEvent = sm.Send(ev.CandidateId, VoteResp{sm.Id,sm.Term,false})
						sendEvents = append(sendEvents,sendEvent)
					}
					sm.StateStore()
					// set election timeout
					sm.Alarm(2)
				} else {
					// Reject vote
					sendEvent = sm.Send(ev.CandidateId, VoteResp{sm.Id,sm.Term,false})
					sendEvents = append(sendEvents,sendEvent)
				}
		case candidate:
				// check if event term is greater than term of state machine
				if sm.Term < ev.Term {
					// step down to follower
                        		sm.SendEvents = sm.toFollower()
					sm.Term = ev.Term
                        		if sm.LastLogIndex==0 || sm.Log[sm.LastLogIndex].Term<ev.LastLogTerm || (sm.Log[sm.LastLogIndex].Term==ev.LastLogTerm && sm.LastLogIndex<=ev.LastLogIndex) {
        					// Grant vote
						sm.VotedFor = ev.CandidateId
						sendEvent = sm.Send(ev.CandidateId, VoteResp{sm.Id,sm.Term,true})
						sendEvents = append(sendEvents,sendEvent)
					} else {
						// Reject vote
						sm.VotedFor = -1
						sendEvent = sm.Send(ev.CandidateId, VoteResp{sm.Id,sm.Term,false})
						sendEvents = append(sendEvents,sendEvent)
					}
					sm.StateStore()
					// set election timeout
					sm.Alarm(2)
				} else {
					// Reject vote
					sendEvent = sm.Send(ev.CandidateId, VoteResp{sm.Id,sm.Term,false})
					sendEvents = append(sendEvents,sendEvent)
				}
		case leader:
				// check if event term is greater than term of state machine
				if sm.Term < ev.Term {
					// step down to follower
                        		sm.SendEvents = sm.toFollower()
					sm.Term = ev.Term
                        		if sm.LastLogIndex==0 || sm.Log[sm.LastLogIndex].Term<ev.LastLogTerm || (sm.Log[sm.LastLogIndex].Term==ev.LastLogTerm && sm.LastLogIndex<=ev.LastLogIndex) {
        					// Grant vote
						sm.VotedFor = ev.CandidateId
						sendEvent = sm.Send(ev.CandidateId, VoteResp{sm.Id,sm.Term,true})
						sendEvents = append(sendEvents,sendEvent)
					} else {
						// Reject vote
						sm.VotedFor = -1
						sendEvent = sm.Send(ev.CandidateId, VoteResp{sm.Id,sm.Term,false})
						sendEvents = append(sendEvents,sendEvent)
					}
					sm.StateStore()
					// set election timeout
					sm.Alarm(2)
				} else {
					// Reject vote
					sendEvent = sm.Send(ev.CandidateId, VoteResp{sm.Id,sm.Term,false})
					sendEvents = append(sendEvents,sendEvent)
				}	
	}
	sm.SendEvents = sendEvents
}
  

//_____________________________________________________________________________________________________________________



// handle all vote responses
func (sm *StateMachine) handleVoteRespEvent(ev VoteResp) {
	var sendEvents []SendEvent
	switch sm.State {
		case follower:
				// delayed response
				// update term if event term greater, else reject
				if ev.Term > sm.Term {
					sm.Term = ev.Term
					sm.VotedFor = -1
					sm.StateStore()
				}
		case candidate:
				// ignore if event term less than term of state machine
				if ev.Term < sm.Term {
					;// Ignore delayed response
				} else if ev.Term > sm.Term {
					// if event term greater, step down to follower
					sm.Term = ev.Term
					sm.SendEvents = sm.toFollower()
					sm.VotedFor = -1
					sm.StateStore()
					// set election timeout
					sm.Alarm(2)
				} else if ev.VoteGranted == true {
					// response of same term, and vote request granted
					sm.VoteCount++
					// check for majority acks
					if sm.VoteCount>len(sm.ServerIds)/2 {
						// move to leader state and send heart beats to all
						sendEvents = sm.toLeader()
					}
				} else {
					// nack
					sm.NegVoteCount++
					// if nack count in majority, step back to follower state
					if sm.NegVoteCount>=(len(sm.ServerIds)+1)/2 {
						sm.SendEvents = sm.toFollower()
						sm.StateStore()
						// set election timeout
						sm.Alarm(2)
					}
					
				}
		case leader:
				// delayed response
				// update term if event term greater, else reject
				if ev.Term > sm.Term {
					sm.Term = ev.Term
					// step down to follower state
					sm.SendEvents = sm.toFollower()
					sm.VotedFor = -1
					sm.StateStore()
				}
	}
	sm.SendEvents = sendEvents
}

  

//_____________________________________________________________________________________________________________________


// handle all append entries request events
func (sm *StateMachine) handleAppendEntriesReq(ev AppendEntriesReq) {
	var sendEvents []SendEvent
	var sendEvent SendEvent
	switch sm.State {
		case follower:
				// handle request if event term equal or greater, else reject
				if ev.Term >= sm.Term {
					// reset election timeout
                        		sm.Alarm(2)
					if sm.Term < ev.Term {
						// update term if event term greater
						sm.Term = ev.Term
						sm.VotedFor = -1
						sm.StateStore()
						
					}
					if sm.LeaderId != ev.LeaderId {
						sm.LeaderId = ev.LeaderId
					}
					// if term at previous log index match
					if sm.LastLogIndex >= ev.PrevLogIndex && sm.Log[ev.PrevLogIndex].Term==ev.PrevLogTerm {
						// if follower is out of sync with leader
						// delete all uncommitted entries after prevLogIndex sent by leader
						// and append entries sent by leader
						if sm.LastLogIndex > ev.PrevLogIndex {
							sm.LogDelete(ev.PrevLogIndex+1)
						}
						for _,v := range ev.Entries {
							sm.LogStore(sm.LastLogIndex+1,v.Term,v.Data)
						}
						// advance commit index, and store it on persistent storage
						// can be used by file system server to check and apply committed logs
						if sm.CommitIndex < ev.LeaderCommit {
							sm.Commit(sm.CommitIndex+1, ev.LeaderCommit)
							sm.CommitIndex = ev.LeaderCommit
							sm.StateStore()
						}
						// send positive response
						sendEvent = sm.Send(ev.LeaderId, AppendEntriesResp{ev,sm.Id,sm.Term,true})
						sendEvents = append(sendEvents,sendEvent)
					} else {
					// reject request
						sendEvent = sm.Send(ev.LeaderId, AppendEntriesResp{ev,sm.Id,sm.Term,false})
						sendEvents = append(sendEvents,sendEvent)
					}
				} else {
					// reject request
					sendEvent = sm.Send(ev.LeaderId, AppendEntriesResp{ev,sm.Id,sm.Term,false})
					sendEvents = append(sendEvents,sendEvent)
				}
		case candidate:
				// handle request if event term equal or greater, else reject
				if ev.Term >= sm.Term {
					// step down to follower state
					sm.SendEvents = sm.toFollower()
					// set election timeout
                        		sm.Alarm(2)
					if sm.Term < ev.Term {
						sm.Term = ev.Term
						sm.VotedFor = -1
						sm.StateStore()
						if sm.LeaderId != ev.LeaderId {
							sm.LeaderId = ev.LeaderId
						}
						
					}
					// if term at previous log index match
					if sm.LastLogIndex >= ev.PrevLogIndex && sm.Log[ev.PrevLogIndex].Term==ev.PrevLogTerm {
						// when follower is out of sync with leader
						// delete all uncommitted entries after prevLogIndex sent by leader
						// and append entries sent by leader
						if sm.LastLogIndex > ev.PrevLogIndex {
							sm.LogDelete(ev.PrevLogIndex+1)
						}
						for _,v := range ev.Entries {
							sm.LogStore(sm.LastLogIndex+1,v.Term,v.Data)
						}
						// advance commit index, and store it on persistent storage
						// can be used by file system server to check and apply committed logs
						if sm.CommitIndex < ev.LeaderCommit {
							sm.Commit(sm.CommitIndex+1, ev.LeaderCommit)
							sm.CommitIndex = ev.LeaderCommit
							sm.StateStore()
						}
						// send positive response
						sendEvent = sm.Send(ev.LeaderId, AppendEntriesResp{ev,sm.Id,sm.Term,true})
						sendEvents = append(sendEvents,sendEvent)
					} else {
						// reject request
						sendEvent = sm.Send(ev.LeaderId, AppendEntriesResp{ev,sm.Id,sm.Term,false})
						sendEvents = append(sendEvents,sendEvent)
					}
				} else {
					// reject request
					sendEvent = sm.Send(ev.LeaderId, AppendEntriesResp{ev,sm.Id,sm.Term,false})
					sendEvents = append(sendEvents,sendEvent)
				}
		case leader:
				// handle request if event term greater, else reject
				if ev.Term > sm.Term {
					// step down to follower
					sm.SendEvents = sm.toFollower()
					// set election timeout
                        		sm.Alarm(2)
					sm.Term = ev.Term
					sm.VotedFor = -1
					sm.StateStore()
					if sm.LeaderId != ev.LeaderId {
						sm.LeaderId = ev.LeaderId
					}
					// if term at previous log index match
					if sm.LastLogIndex >= ev.PrevLogIndex && sm.Log[ev.PrevLogIndex].Term==ev.PrevLogTerm {
						// when follower is out of sync with leader
						// delete all uncommitted entries after prevLogIndex sent by leader
						// and append entries sent by leader
						if sm.LastLogIndex > ev.PrevLogIndex {
							sm.LogDelete(ev.PrevLogIndex+1)
						}
						for _,v := range ev.Entries {
							sm.LogStore(sm.LastLogIndex+1,v.Term,v.Data)
						}
						// advance commit index, and store it on persistent storage
						// can be used by file system server to check and apply committed logs
						if sm.CommitIndex < ev.LeaderCommit {
							sm.Commit(sm.CommitIndex+1, ev.LeaderCommit)
							sm.CommitIndex = ev.LeaderCommit
							sm.StateStore()
						}
						// send positive response
						sendEvent = sm.Send(ev.LeaderId, AppendEntriesResp{ev,sm.Id,sm.Term,true})
						sendEvents = append(sendEvents,sendEvent)
					} else {
						// reject request
						sendEvent = sm.Send(ev.LeaderId, AppendEntriesResp{ev,sm.Id,sm.Term,false})
						sendEvents = append(sendEvents,sendEvent)
					}
				} else {
					// reject request
					sendEvent = sm.Send(ev.LeaderId, AppendEntriesResp{ev,sm.Id,sm.Term,false})
					sendEvents = append(sendEvents,sendEvent)
				}
	}
	sm.SendEvents = sendEvents
	
}
  

//_____________________________________________________________________________________________________________________



// handle all append entries response events
func (sm *StateMachine) handleAppendEntriesResp(ev AppendEntriesResp) {
	var sendEvents []SendEvent
	var sendEvent SendEvent
	switch sm.State {
		case follower:
				// delayed response
				// update term if event term greater, else reject
				if ev.Term > sm.Term {
					sm.Term = ev.Term
					sm.VotedFor = -1
					sm.StateStore()
				}
		case candidate:
				// delayed response
				// update term if event term greater, else reject
				if ev.Term > sm.Term {
					// step down to follower
					sm.SendEvents = sm.toFollower()
					sm.Term = ev.Term
					sm.VotedFor = -1
					sm.StateStore()
				}
		case leader:
				// ignore if event term less than term of state machine
				if ev.Term < sm.Term {
					;//Ignore delayed response
				} else if ev.Term > sm.Term {
					// if event term greater, step down to follower
					sm.SendEvents = sm.toFollower()
					sm.Term = ev.Term
					sm.VotedFor = -1
					sm.StateStore()
					// set election timeout
					sm.Alarm(2)
				} else if ev.Success == true {
					// response of same term, and append entries request success
					// update next index and match index
					sm.NextIndex[ev.FromId] = ev.Msg.PrevLogIndex + len(ev.Msg.Entries)+1
					sm.MatchIndex[ev.FromId] = sm.NextIndex[ev.FromId]-1
					// sort match indices
					var temparr []int
					for _,v:=range sm.MatchIndex {
						temparr=append(temparr,v)
					}
					sort.Ints(temparr)
					tempindex := temparr[(len(temparr)-1)/2]
					// update commit index based on one less than median value of sorted match indices
					// represent the index replicated on majority of servers
					if tempindex > sm.CommitIndex && sm.Log[tempindex].Term==sm.Term {
						sm.Commit(sm.CommitIndex+1, tempindex)
						sm.CommitIndex = tempindex
						sm.StateStore()
					}
				} else {
					// Resend append entry request with decremented nextIndex of follwer
					sm.NextIndex[ev.FromId] = sm.NextIndex[ev.FromId] - 1
					sendEvent = sm.Send(ev.FromId, AppendEntriesReq{sm.Term, sm.Id, sm.NextIndex[ev.FromId]-1, sm.Log[sm.NextIndex[ev.FromId]-1].Term,sm.Log[sm.NextIndex[ev.FromId]:], sm.CommitIndex})
					sendEvents = append(sendEvents,sendEvent)
				}
	}
	sm.SendEvents = sendEvents
}
  

//_____________________________________________________________________________________________________________________



func (sm *StateMachine) ProcessEvent (ev interface{}) {
	filename:="testlog"
	filename+=strconv.Itoa(sm.Id)
	filename+=".dat"
	f, err := os.OpenFile(filename, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		//log.Fatal(err)
	}
	defer f.Close()
	b := bufio.NewWriter(f)
	defer func() {
	if err = b.Flush(); err != nil {
	    //log.Fatal(err)
	}
	}()
	switch ev.(type) {
		case Append:
				sm.handleAppendEvent(ev.(Append))
				evchanged :=ev.(Append)
				fmt.Fprintf(b, "Commit channel size:%d\n",len((*sm.RaftNode).CommitChannel))
				fmt.Fprintf(b, "Got append event\nServer Details\nterm:%d\tstate:%d\tcommitIndex:%d\tvotedFor:%d\tleaderId:%d\tSendevent Size:%d\n", sm.Term, sm.State, sm.CommitIndex, sm.VotedFor, sm.LeaderId,len(sm.SendEvents))
				fmt.Fprintf(b, "data:%q\n\n\n\n", evchanged.Data)
		case Timeout:
				sm.handleTimeoutEvent(ev.(Timeout))
				fmt.Fprintf(b, "Commit channel size:%d\n",len((*sm.RaftNode).CommitChannel))
				fmt.Fprintf(b, "Got timeout event\nServer Details\nterm:%d\tstate:%d\tcommitIndex:%d\tvotedFor:%d\tleaderId:%d\tSendevent Size:%d\n", sm.Term, sm.State, sm.CommitIndex, sm.VotedFor, sm.LeaderId,len(sm.SendEvents))
				fmt.Fprintf(b, "\n\n\n")
		case VoteReq:
				sm.handleVoteReqEvent(ev.(VoteReq))
				evchanged :=ev.(VoteReq)
				fmt.Fprintf(b, "Commit channel size:%d\n",len((*sm.RaftNode).CommitChannel))
				fmt.Fprintf(b, "Got votereq event\nServer Details\nterm:%d\tstate:%d\tcommitIndex:%d\tvotedFor:%d\tleaderId:%d\n", sm.Term, sm.State, sm.CommitIndex, sm.VotedFor, sm.LeaderId)
				fmt.Fprintf(b, "term:%d\ncandidateId:%d\nlastLogIndex:%d\nlastLogTerm:%d\n\n\n\n", evchanged.Term, evchanged.CandidateId, evchanged.LastLogIndex, evchanged.LastLogTerm)
		case VoteResp:
				sm.handleVoteRespEvent(ev.(VoteResp))
				evchanged :=ev.(VoteResp)
				fmt.Fprintf(b, "Commit channel size:%d\n",len((*sm.RaftNode).CommitChannel))
				fmt.Fprintf(b, "Got voteresp event\nServer Details\nterm:%d\tstate:%d\tcommitIndex:%d\tvotedFor:%d\tleaderId:%d\n", sm.Term, sm.State, sm.CommitIndex, sm.VotedFor, sm.LeaderId)
				fmt.Fprintf(b, "fromId:%d\nterm:%d\nvoteGranted:%t\n\n\n\n", evchanged.FromId, evchanged.Term, evchanged.VoteGranted)
		case AppendEntriesReq:
				sm.handleAppendEntriesReq(ev.(AppendEntriesReq))
				evchanged :=ev.(AppendEntriesReq)
				fmt.Fprintf(b, "Commit channel size:%d\n",len((*sm.RaftNode).CommitChannel))
				fmt.Fprintf(b, "Got appendentriesreq event\nServer Details\nterm:%d\tstate:%d\tcommitIndex:%d\tvotedFor:%d\tleaderId:%d\n", sm.Term, sm.State, sm.CommitIndex, sm.VotedFor, sm.LeaderId)
				fmt.Fprintf(b, "term:%d\nleaderId:%d\nprevLogIndex:%d\nprevLogTerm:%d\nleaderCommit:%d\n\n\n\n", evchanged.Term, evchanged.LeaderId, evchanged.PrevLogIndex, evchanged.PrevLogTerm, evchanged.LeaderCommit)
		case AppendEntriesResp:
				sm.handleAppendEntriesResp(ev.(AppendEntriesResp))
				evchanged :=ev.(AppendEntriesResp)
				fmt.Fprintf(b, "Commit channel size:%d\n",len((*sm.RaftNode).CommitChannel))
				fmt.Fprintf(b, "Got appendentriesresp event\nServer Details\nterm:%d\tstate:%d\tcommitIndex:%d\tvotedFor:%d\tleaderId:%d\n", sm.Term, sm.State, sm.CommitIndex, sm.VotedFor, sm.LeaderId)
				fmt.Fprintf(b, "fromId:%d\nterm:%d\nsuccess:%t\n\n\n\n", evchanged.FromId, evchanged.Term, evchanged.Success)
		default: 
				//Ignore
				sm.SendEvents=nil;
	}
}

func (sm *StateMachine) Commit (beginIndex int, endIndex int) {
	for i := beginIndex; i <= endIndex; i++ {
		if sm.Log[i].Data != nil {
			(*sm.RaftNode).CommitChannel <- CommitInfo{Data:sm.Log[i].Data, Index:int64(i), Err:nil}
		}
	}
}
	
