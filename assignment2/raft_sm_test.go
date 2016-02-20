package statemachine

import (
	"testing"
	"fmt"
	"bytes"
)

var sm *StateMachine
var sm_expectedstate *StateMachine


//________________________________________ Follower Tests____________________________________________________________

func TestCaseFollowerAppendEvent(t *testing.T) {
	sm = &StateMachine{id:1, leaderId:2, serverIds:[]int{1,2,3,4,5,6}, state:follower, term:1, voteCount:0, negVoteCount:0, votedFor:2, log:[]LogEntry{{0,nil},{1,nil}}, lastLogIndex:1, commitIndex:1, nextIndex:nil, matchIndex:nil}
	sm_expectedstate =  &StateMachine{id:1, leaderId:2, serverIds:[]int{1,2,3,4,5,6}, state:follower, term:1, voteCount:0, negVoteCount:0, votedFor:2, log:[]LogEntry{{0,nil},{1,nil}}, lastLogIndex:1, commitIndex:1, nextIndex:nil, matchIndex:nil}
	// check whether event ignored
	sm.handleAppendEvent(Append{[]byte("Rahul")})
	expect(t, sm, sm_expectedstate)
}  

//_____________________________________________________________________________________________________________________

func TestCaseFollowerTimeoutEvent(t *testing.T) {
	sm = &StateMachine{id:1, leaderId:2, serverIds:[]int{1,2,3,4,5,6}, state:follower, term:1, voteCount:0, negVoteCount:0, votedFor:2, log:[]LogEntry{{0,nil},{1,nil}}, lastLogIndex:1, commitIndex:1, nextIndex:nil, matchIndex:nil}
	sm_expectedstate =  &StateMachine{id:1, leaderId:-1, serverIds:[]int{1,2,3,4,5,6}, state:candidate, term:2, voteCount:0, negVoteCount:0, votedFor:1, log:[]LogEntry{{0,nil},{1,nil}}, lastLogIndex:1, commitIndex:1, nextIndex:nil, matchIndex:nil}
	// check whether changed to candidate state, and sent vote request to peers 
	sm.handleTimeoutEvent(Timeout{})
	expect(t, sm, sm_expectedstate)
	if len(sm.sendEvents) != 5 {
		t.Error(fmt.Sprintf("Expected Send Events Count: %v, found : %v\n", 5, len(sm.sendEvents))) 
	}
}  

//_____________________________________________________________________________________________________________________

func TestCaseFollowerVoteReqEvent(t *testing.T) {
	//event term greater
	sm = &StateMachine{id:1, leaderId:2, serverIds:[]int{1,2,3,4,5,6}, state:follower, term:1, voteCount:0, negVoteCount:0, votedFor:2, log:[]LogEntry{{0,nil},{1,nil}}, lastLogIndex:1, commitIndex:1, nextIndex:nil, matchIndex:nil}
	//vote granted; term,votedFor changed
	sm_expectedstate =  &StateMachine{id:1, leaderId:2, serverIds:[]int{1,2,3,4,5,6}, state:follower, term:2, voteCount:0, negVoteCount:0, votedFor:-1, log:[]LogEntry{{0,nil},{1,nil}}, lastLogIndex:1, commitIndex:1, nextIndex:nil, matchIndex:nil}
	sm.handleVoteReqEvent(VoteReq{2,3,1,0})
	expect(t, sm, sm_expectedstate)
	//if no vote response
	if len(sm.sendEvents) != 1 {
		t.Error(fmt.Sprintf("Expected Send Events Count: %v, found : %v\n", 1, len(sm.sendEvents))) 
	} else if sm.sendEvents[0].event.(VoteResp).voteGranted!=false || sm.sendEvents[0].event.(VoteResp).fromId!=1 || sm.sendEvents[0].event.(VoteResp).term!=2 {
		// Invalid vote response
		t.Error(fmt.Sprintf("Invalid vote response\n"))
	}
	//leader not as updated, vote not granted
	sm = &StateMachine{id:1, leaderId:2, serverIds:[]int{1,2,3,4,5,6}, state:follower, term:2, voteCount:0, negVoteCount:0, votedFor:2, log:[]LogEntry{{0,nil},{1,nil}}, lastLogIndex:1, commitIndex:1, nextIndex:nil, matchIndex:nil}
	sm_expectedstate =  &StateMachine{id:1, leaderId:2, serverIds:[]int{1,2,3,4,5,6}, state:follower, term:2, voteCount:0, negVoteCount:0, votedFor:2, log:[]LogEntry{{0,nil},{1,nil}}, lastLogIndex:1, commitIndex:1, nextIndex:nil, matchIndex:nil}
	sm.handleVoteReqEvent(VoteReq{2,0,1,1})
	expect(t, sm, sm_expectedstate)
	//if no vote response
	if len(sm.sendEvents) != 1 {
		t.Error(fmt.Sprintf("Expected Send Events Count: %v, found : %v\n", 1, len(sm.sendEvents))) 
	} else if sm.sendEvents[0].event.(VoteResp).voteGranted!=false || sm.sendEvents[0].event.(VoteResp).fromId!=1 || sm.sendEvents[0].event.(VoteResp).term!=2 {
		// Invalid vote response
		t.Error(fmt.Sprintf("Invalid vote response\n"))
	}
	
	//event term equal
	sm = &StateMachine{id:1, leaderId:2, serverIds:[]int{1,2,3,4,5,6}, state:follower, term:1, voteCount:0, negVoteCount:0, votedFor:2, log:[]LogEntry{{0,nil},{1,nil}}, lastLogIndex:1, commitIndex:1, nextIndex:nil, matchIndex:nil}
	//leader not as updated, vote not granted
	sm_expectedstate =  &StateMachine{id:1, leaderId:2, serverIds:[]int{1,2,3,4,5,6}, state:follower, term:2, voteCount:0, negVoteCount:0, votedFor:-1, log:[]LogEntry{{0,nil},{1,nil}}, lastLogIndex:1, commitIndex:1, nextIndex:nil, matchIndex:nil}
	sm.handleVoteReqEvent(VoteReq{2,3,1,0})
	expect(t, sm, sm_expectedstate)
	//if no vote response
	if len(sm.sendEvents) != 1 {
		t.Error(fmt.Sprintf("Expected Send Events Count: %v, found : %v\n", 1, len(sm.sendEvents))) 
	} else if sm.sendEvents[0].event.(VoteResp).voteGranted!=false || sm.sendEvents[0].event.(VoteResp).fromId!=1 || sm.sendEvents[0].event.(VoteResp).term!=2 {
		// Invalid vote response
		t.Error(fmt.Sprintf("Invalid vote response\n"))
	}
	//vote granted; term,votedFor changed
	sm = &StateMachine{id:1, leaderId:2, serverIds:[]int{1,2,3,4,5,6}, state:follower, term:1, voteCount:0, negVoteCount:0, votedFor:2, log:[]LogEntry{{0,nil},{1,nil}}, lastLogIndex:1, commitIndex:1, nextIndex:nil, matchIndex:nil}
	sm_expectedstate =  &StateMachine{id:1, leaderId:2, serverIds:[]int{1,2,3,4,5,6}, state:follower, term:2, voteCount:0, negVoteCount:0, votedFor:3, log:[]LogEntry{{0,nil},{1,nil}}, lastLogIndex:1, commitIndex:1, nextIndex:nil, matchIndex:nil}
	sm.handleVoteReqEvent(VoteReq{2,3,1,1})
	expect(t, sm, sm_expectedstate)
	//if no vote response
	if len(sm.sendEvents) != 1 {
		t.Error(fmt.Sprintf("Expected Send Events Count: %v, found : %v\n", 1, len(sm.sendEvents))) 
	} else if sm.sendEvents[0].event.(VoteResp).voteGranted!=true || sm.sendEvents[0].event.(VoteResp).fromId!=1 || sm.sendEvents[0].event.(VoteResp).term!=2 {
		// Invalid vote response
		t.Error(fmt.Sprintf("Invalid vote response\n"))
	}
	
	//event term equal, candidate more updated
	sm = &StateMachine{id:1, leaderId:2, serverIds:[]int{1,2,3,4,5,6}, state:follower, term:1, voteCount:0, negVoteCount:0, votedFor:2, log:[]LogEntry{{0,nil},{1,nil}}, lastLogIndex:1, commitIndex:1, nextIndex:nil, matchIndex:nil}
	//leader not as updated, vote not granted
	sm_expectedstate =  &StateMachine{id:1, leaderId:2, serverIds:[]int{1,2,3,4,5,6}, state:follower, term:2, voteCount:0, negVoteCount:0, votedFor:-1, log:[]LogEntry{{0,nil},{1,nil}}, lastLogIndex:1, commitIndex:1, nextIndex:nil, matchIndex:nil}
	sm.handleVoteReqEvent(VoteReq{2,3,1,0})
	expect(t, sm, sm_expectedstate)
	//if no vote response
	if len(sm.sendEvents) != 1 {
		t.Error(fmt.Sprintf("Expected Send Events Count: %v, found : %v\n", 1, len(sm.sendEvents))) 
	} else if sm.sendEvents[0].event.(VoteResp).voteGranted!=false || sm.sendEvents[0].event.(VoteResp).fromId!=1 || sm.sendEvents[0].event.(VoteResp).term!=2 {
		// Invalid vote response
		t.Error(fmt.Sprintf("Invalid vote response\n"))
	}
	
	//if candidate term is less
	sm = &StateMachine{id:1, leaderId:2, serverIds:[]int{1,2,3,4,5,6}, state:follower, term:2, voteCount:0, negVoteCount:0, votedFor:2, log:[]LogEntry{{0,nil},{1,nil}}, lastLogIndex:1, commitIndex:1, nextIndex:nil, matchIndex:nil}
	sm_expectedstate =  &StateMachine{id:1, leaderId:2, serverIds:[]int{1,2,3,4,5,6}, state:follower, term:2, voteCount:0, negVoteCount:0, votedFor:2, log:[]LogEntry{{0,nil},{1,nil}}, lastLogIndex:1, commitIndex:1, nextIndex:nil, matchIndex:nil}
	sm.handleVoteReqEvent(VoteReq{1,1,1,1})
	expect(t, sm, sm_expectedstate)
	//if no vote response
	if len(sm.sendEvents) != 1 {
		t.Error(fmt.Sprintf("Expected Send Events Count: %v, found : %v\n", 1, len(sm.sendEvents))) 
	} else if sm.sendEvents[0].event.(VoteResp).voteGranted!=false || sm.sendEvents[0].event.(VoteResp).fromId!=1 || sm.sendEvents[0].event.(VoteResp).term!=2 {
		// Invalid vote response
		t.Error(fmt.Sprintf("Invalid vote response\n"))
	}
	
	//if term same, votedFor not -1
	sm = &StateMachine{id:1, leaderId:2, serverIds:[]int{1,2,3,4,5,6}, state:follower, term:2, voteCount:0, negVoteCount:0, votedFor:2, log:[]LogEntry{{0,nil},{1,nil}}, lastLogIndex:1, commitIndex:1, nextIndex:nil, matchIndex:nil}
	sm_expectedstate =  &StateMachine{id:1, leaderId:2, serverIds:[]int{1,2,3,4,5,6}, state:follower, term:2, voteCount:0, negVoteCount:0, votedFor:2, log:[]LogEntry{{0,nil},{1,nil}}, lastLogIndex:1, commitIndex:1, nextIndex:nil, matchIndex:nil}
	sm.handleVoteReqEvent(VoteReq{2,2,1,1})
	expect(t, sm, sm_expectedstate)
	//if no vote response
	if len(sm.sendEvents) != 1 {
		t.Error(fmt.Sprintf("Expected Send Events Count: %v, found : %v\n", 1, len(sm.sendEvents))) 
	} else if sm.sendEvents[0].event.(VoteResp).voteGranted!=false || sm.sendEvents[0].event.(VoteResp).fromId!=1 || sm.sendEvents[0].event.(VoteResp).term!=2 {
		// Invalid vote response
		t.Error(fmt.Sprintf("Invalid vote response\n"))
	}
	
}  

//_____________________________________________________________________________________________________________________

func TestCaseFollowerVoteRespEvent(t *testing.T) {
	//event term greater
	sm = &StateMachine{id:1, leaderId:2, serverIds:[]int{1,2,3,4,5,6}, state:follower, term:1, voteCount:0, negVoteCount:0, votedFor:2, log:[]LogEntry{{0,nil},{1,nil}}, lastLogIndex:1, commitIndex:1, nextIndex:nil, matchIndex:nil}
	//vote granted; term,votedFor changed
	sm_expectedstate =  &StateMachine{id:1, leaderId:2, serverIds:[]int{1,2,3,4,5,6}, state:follower, term:3, voteCount:0, negVoteCount:0, votedFor:-1, log:[]LogEntry{{0,nil},{1,nil}}, lastLogIndex:1, commitIndex:1, nextIndex:nil, matchIndex:nil}
	sm.handleVoteRespEvent(VoteResp{2,3,false})
	expect(t, sm, sm_expectedstate)
	//if no vote response
	if len(sm.sendEvents) != 0 {
		t.Error(fmt.Sprintf("Expected Send Events Count: %v, found : %v\n", 0, len(sm.sendEvents))) 
	}
	
	//event term not greater
	sm = &StateMachine{id:1, leaderId:2, serverIds:[]int{1,2,3,4,5,6}, state:follower, term:1, voteCount:0, negVoteCount:0, votedFor:2, log:[]LogEntry{{0,nil},{1,nil}}, lastLogIndex:1, commitIndex:1, nextIndex:nil, matchIndex:nil}
	//vote granted; term,votedFor changed
	sm_expectedstate =  &StateMachine{id:1, leaderId:2, serverIds:[]int{1,2,3,4,5,6}, state:follower, term:1, voteCount:0, negVoteCount:0, votedFor:2, log:[]LogEntry{{0,nil},{1,nil}}, lastLogIndex:1, commitIndex:1, nextIndex:nil, matchIndex:nil}
	sm.handleVoteRespEvent(VoteResp{2,1,false})
	expect(t, sm, sm_expectedstate)
	//if no vote response
	if len(sm.sendEvents) != 0 {
		t.Error(fmt.Sprintf("Expected Send Events Count: %v, found : %v\n", 0, len(sm.sendEvents))) 
	}
}  

//_____________________________________________________________________________________________________________________

func TestCaseFollowerAppEntReqEvent(t *testing.T) {
	//event term greater and prev log entries don't match
	sm = &StateMachine{id:1, leaderId:2, serverIds:[]int{1,2,3,4,5,6}, state:follower, term:1, voteCount:0, negVoteCount:0, votedFor:2, log:[]LogEntry{{0,nil}}, lastLogIndex:0, commitIndex:0, nextIndex:nil, matchIndex:nil}
	sm_expectedstate =  &StateMachine{id:1, leaderId:3, serverIds:[]int{1,2,3,4,5,6}, state:follower, term:2, voteCount:0, negVoteCount:0, votedFor:-1, log:[]LogEntry{{0,nil}}, lastLogIndex:0, commitIndex:0, nextIndex:nil, matchIndex:nil}
	sm.handleAppendEntriesReq(AppendEntriesReq{2, 3, 1, 1,[]LogEntry{{2,nil}}, 1})
	expect(t, sm, sm_expectedstate)
	//if no vote response
	if len(sm.sendEvents) != 1 {
		t.Error(fmt.Sprintf("Expected Send Events Count: %v, found : %v\n", 1, len(sm.sendEvents))) 
	} else if sm.sendEvents[0].event.(AppendEntriesResp).success!=false || sm.sendEvents[0].event.(AppendEntriesResp).fromId!=1 || sm.sendEvents[0].event.(AppendEntriesResp).term!=2 {
		// Invalid vote response
		t.Error(fmt.Sprintf("Invalid append entries response\n"))
	}
	
	//prev log entries match
	sm = &StateMachine{id:1, leaderId:2, serverIds:[]int{1,2,3,4,5,6}, state:follower, term:1, voteCount:0, negVoteCount:0, votedFor:2, log:[]LogEntry{{0,nil},{1,nil}}, lastLogIndex:1, commitIndex:0, nextIndex:nil, matchIndex:nil}
	sm_expectedstate =  &StateMachine{id:1, leaderId:2, serverIds:[]int{1,2,3,4,5,6}, state:follower, term:2, voteCount:0, negVoteCount:0, votedFor:-1, log:[]LogEntry{{0,nil},{2,nil}}, lastLogIndex:1, commitIndex:1, nextIndex:nil, matchIndex:nil}
	sm.handleAppendEntriesReq(AppendEntriesReq{2, 2, 0, 0,[]LogEntry{{2,nil}}, 1})
	expect(t, sm, sm_expectedstate)
	//if no vote response
	if len(sm.sendEvents) != 1 {
		t.Error(fmt.Sprintf("Expected Send Events Count: %v, found : %v\n", 1, len(sm.sendEvents))) 
	} else if sm.sendEvents[0].event.(AppendEntriesResp).success!=true || sm.sendEvents[0].event.(AppendEntriesResp).fromId!=1 || sm.sendEvents[0].event.(AppendEntriesResp).term!=2 {
		// Invalid vote response
		t.Error(fmt.Sprintf("Invalid append entries response\n"))
	}
	
	//event term smaller
	sm = &StateMachine{id:1, leaderId:2, serverIds:[]int{1,2,3,4,5,6}, state:follower, term:2, voteCount:0, negVoteCount:0, votedFor:2, log:[]LogEntry{{0,nil}}, lastLogIndex:0, commitIndex:0, nextIndex:nil, matchIndex:nil}
	sm_expectedstate =  &StateMachine{id:1, leaderId:2, serverIds:[]int{1,2,3,4,5,6}, state:follower, term:2, voteCount:0, negVoteCount:0, votedFor:2, log:[]LogEntry{{0,nil}}, lastLogIndex:0, commitIndex:0, nextIndex:nil, matchIndex:nil}
	sm.handleAppendEntriesReq(AppendEntriesReq{1, 3, 0, 0,[]LogEntry{{1,nil}}, 1})
	expect(t, sm, sm_expectedstate)
	//if no vote response
	if len(sm.sendEvents) != 1 {
		t.Error(fmt.Sprintf("Expected Send Events Count: %v, found : %v\n", 1, len(sm.sendEvents))) 
	} else if sm.sendEvents[0].event.(AppendEntriesResp).success!=false || sm.sendEvents[0].event.(AppendEntriesResp).fromId!=1 || sm.sendEvents[0].event.(AppendEntriesResp).term!=2 {
		// Invalid vote response
		t.Error(fmt.Sprintf("Invalid append entries response\n"))
	}
	
}  

//_____________________________________________________________________________________________________________________

func TestCaseFollowerAppEntRespEvent(t *testing.T) {
	//event term greater
	sm = &StateMachine{id:1, leaderId:2, serverIds:[]int{1,2,3,4,5,6}, state:follower, term:1, voteCount:0, negVoteCount:0, votedFor:2, log:[]LogEntry{{0,nil},{1,nil}}, lastLogIndex:1, commitIndex:1, nextIndex:nil, matchIndex:nil}
	//vote granted; term,votedFor changed
	sm_expectedstate =  &StateMachine{id:1, leaderId:2, serverIds:[]int{1,2,3,4,5,6}, state:follower, term:3, voteCount:0, negVoteCount:0, votedFor:-1, log:[]LogEntry{{0,nil},{1,nil}}, lastLogIndex:1, commitIndex:1, nextIndex:nil, matchIndex:nil}
	sm.handleAppendEntriesResp(AppendEntriesResp{AppendEntriesReq{},2,3,false})
	expect(t, sm, sm_expectedstate)
	//if no vote response
	if len(sm.sendEvents) != 0 {
		t.Error(fmt.Sprintf("Expected Send Events Count: %v, found : %v\n", 0, len(sm.sendEvents))) 
	}
	
	//event term not greater
	sm = &StateMachine{id:1, leaderId:2, serverIds:[]int{1,2,3,4,5,6}, state:follower, term:1, voteCount:0, negVoteCount:0, votedFor:2, log:[]LogEntry{{0,nil},{1,nil}}, lastLogIndex:1, commitIndex:1, nextIndex:nil, matchIndex:nil}
	//vote granted; term,votedFor changed
	sm_expectedstate =  &StateMachine{id:1, leaderId:2, serverIds:[]int{1,2,3,4,5,6}, state:follower, term:1, voteCount:0, negVoteCount:0, votedFor:2, log:[]LogEntry{{0,nil},{1,nil}}, lastLogIndex:1, commitIndex:1, nextIndex:nil, matchIndex:nil}
	sm.handleAppendEntriesResp(AppendEntriesResp{AppendEntriesReq{},2,1,false})
	expect(t, sm, sm_expectedstate)
	//if no vote response
	if len(sm.sendEvents) != 0 {
		t.Error(fmt.Sprintf("Expected Send Events Count: %v, found : %v\n", 0, len(sm.sendEvents))) 
	}
}  


//________________________________________ Candidate Tests____________________________________________________________


func TestCaseCandidateAppendEvent(t *testing.T) {
	sm = &StateMachine{id:1, leaderId:-1, serverIds:[]int{1,2,3,4,5,6}, state:candidate, term:1, voteCount:0, negVoteCount:0, votedFor:1, log:[]LogEntry{{0,nil},{1,nil}}, lastLogIndex:1, commitIndex:1, nextIndex:nil, matchIndex:nil}
	sm_expectedstate =  &StateMachine{id:1, leaderId:-1, serverIds:[]int{1,2,3,4,5,6}, state:candidate, term:1, voteCount:0, negVoteCount:0, votedFor:1, log:[]LogEntry{{0,nil},{1,nil}}, lastLogIndex:1, commitIndex:1, nextIndex:nil, matchIndex:nil}
	// check request ignore
	sm.handleAppendEvent(Append{[]byte("Rahul")})
	expect(t, sm, sm_expectedstate)
} 

//_____________________________________________________________________________________________________________________

func TestCaseCandidateTimeoutEvent(t *testing.T) {
	sm = &StateMachine{id:1, leaderId:-1, serverIds:[]int{1,2,3,4,5,6}, state:candidate, term:2, voteCount:0, negVoteCount:0, votedFor:1, log:[]LogEntry{{0,nil},{1,nil}}, lastLogIndex:1, commitIndex:1, nextIndex:nil, matchIndex:nil}
	sm_expectedstate =  &StateMachine{id:1, leaderId:-1, serverIds:[]int{1,2,3,4,5,6}, state:candidate, term:3, voteCount:0, negVoteCount:0, votedFor:1, log:[]LogEntry{{0,nil},{1,nil}}, lastLogIndex:1, commitIndex:1, nextIndex:nil, matchIndex:nil}
	sm.handleTimeoutEvent(Timeout{})
	// election timeout, change to candidate(restart as candidate), increment term, vote for self
	expect(t, sm, sm_expectedstate)
	if len(sm.sendEvents) != 5 {
		t.Error(fmt.Sprintf("Expected Send Events Count: %v, found : %v\n", 5, len(sm.sendEvents))) 
	}
}  

//_____________________________________________________________________________________________________________________

func TestCaseCandidateVoteReqEvent(t *testing.T) {
	//event term greater
	sm = &StateMachine{id:1, leaderId:-1, serverIds:[]int{1,2,3,4,5,6}, state:candidate, term:2, voteCount:0, negVoteCount:0, votedFor:2, log:[]LogEntry{{0,nil},{1,nil}}, lastLogIndex:1, commitIndex:1, nextIndex:nil, matchIndex:nil}
	//vote granted; term,votedFor changed
	sm_expectedstate =  &StateMachine{id:1, leaderId:-1, serverIds:[]int{1,2,3,4,5,6}, state:follower, term:3, voteCount:0, negVoteCount:0, votedFor:3, log:[]LogEntry{{0,nil},{1,nil}}, lastLogIndex:1, commitIndex:1, nextIndex:nil, matchIndex:nil}
	sm.handleVoteReqEvent(VoteReq{3,3,1,1})
	expect(t, sm, sm_expectedstate)
	//if no vote response
	if len(sm.sendEvents) != 1 {
		t.Error(fmt.Sprintf("Expected Send Events Count: %v, found : %v\n", 1, len(sm.sendEvents))) 
	} else if sm.sendEvents[0].event.(VoteResp).voteGranted!=true || sm.sendEvents[0].event.(VoteResp).fromId!=1 || sm.sendEvents[0].event.(VoteResp).term!=3 {
		// Invalid vote response
		t.Error(fmt.Sprintf("Invalid vote response\n"))
	}
	//candidate not as updated, vote not granted
	sm = &StateMachine{id:1, leaderId:-1, serverIds:[]int{1,2,3,4,5,6}, state:candidate, term:2, voteCount:0, negVoteCount:0, votedFor:1, log:[]LogEntry{{0,nil},{1,nil}}, lastLogIndex:1, commitIndex:1, nextIndex:nil, matchIndex:nil}
	sm_expectedstate =  &StateMachine{id:1, leaderId:-1, serverIds:[]int{1,2,3,4,5,6}, state:follower, term:3, voteCount:0, negVoteCount:0, votedFor:-1, log:[]LogEntry{{0,nil},{1,nil}}, lastLogIndex:1, commitIndex:1, nextIndex:nil, matchIndex:nil}
	sm.handleVoteReqEvent(VoteReq{3,3,0,0})
	expect(t, sm, sm_expectedstate)
	//if no vote response
	if len(sm.sendEvents) != 1 {
		t.Error(fmt.Sprintf("Expected Send Events Count: %v, found : %v\n", 1, len(sm.sendEvents))) 
	} else if sm.sendEvents[0].event.(VoteResp).voteGranted!=false || sm.sendEvents[0].event.(VoteResp).fromId!=1 || sm.sendEvents[0].event.(VoteResp).term!=3 {
		// Invalid vote response
		t.Error(fmt.Sprintf("Invalid vote response\n"))
	}
	
	//event term equal
	sm = &StateMachine{id:1, leaderId:-1, serverIds:[]int{1,2,3,4,5,6}, state:candidate, term:2, voteCount:0, negVoteCount:0, votedFor:1, log:[]LogEntry{{0,nil},{1,nil}}, lastLogIndex:1, commitIndex:1, nextIndex:nil, matchIndex:nil}
	sm_expectedstate =  &StateMachine{id:1, leaderId:-1, serverIds:[]int{1,2,3,4,5,6}, state:candidate, term:2, voteCount:0, negVoteCount:0, votedFor:1, log:[]LogEntry{{0,nil},{1,nil}}, lastLogIndex:1, commitIndex:1, nextIndex:nil, matchIndex:nil}
	sm.handleVoteReqEvent(VoteReq{2,3,0,0})
	expect(t, sm, sm_expectedstate)
	//if no vote response
	if len(sm.sendEvents) != 1 {
		t.Error(fmt.Sprintf("Expected Send Events Count: %v, found : %v\n", 1, len(sm.sendEvents))) 
	} else if sm.sendEvents[0].event.(VoteResp).voteGranted!=false || sm.sendEvents[0].event.(VoteResp).fromId!=1 || sm.sendEvents[0].event.(VoteResp).term!=2 {
		// Invalid vote response
		t.Error(fmt.Sprintf("Invalid vote response\n"))
	}
	
	//if candidate term is less
	sm = &StateMachine{id:1, leaderId:-1, serverIds:[]int{1,2,3,4,5,6}, state:candidate, term:2, voteCount:0, negVoteCount:0, votedFor:1, log:[]LogEntry{{0,nil},{1,nil}}, lastLogIndex:1, commitIndex:1, nextIndex:nil, matchIndex:nil}
	sm_expectedstate =  &StateMachine{id:1, leaderId:-1, serverIds:[]int{1,2,3,4,5,6}, state:candidate, term:2, voteCount:0, negVoteCount:0, votedFor:1, log:[]LogEntry{{0,nil},{1,nil}}, lastLogIndex:1, commitIndex:1, nextIndex:nil, matchIndex:nil}
	sm.handleVoteReqEvent(VoteReq{1,1,1,1})
	expect(t, sm, sm_expectedstate)
	//if no vote response
	if len(sm.sendEvents) != 1 {
		t.Error(fmt.Sprintf("Expected Send Events Count: %v, found : %v\n", 1, len(sm.sendEvents))) 
	} else if sm.sendEvents[0].event.(VoteResp).voteGranted!=false || sm.sendEvents[0].event.(VoteResp).fromId!=1 || sm.sendEvents[0].event.(VoteResp).term!=2 {
		// Invalid vote response
		t.Error(fmt.Sprintf("Invalid vote response\n"))
	}
} 

//_____________________________________________________________________________________________________________________

func TestCaseCandidateVoteRespEvent(t *testing.T) {
	
	//if VoteResp term is less, ignore it
	sm = &StateMachine{id:1, leaderId:-1, serverIds:[]int{1,2,3,4,5,6}, state:candidate, term:2, voteCount:0, negVoteCount:0, votedFor:1, log:[]LogEntry{{0,nil},{1,nil}}, lastLogIndex:1, commitIndex:1, nextIndex:nil, matchIndex:nil}
	sm_expectedstate =  &StateMachine{id:1, leaderId:-1, serverIds:[]int{1,2,3,4,5,6}, state:candidate, term:2, voteCount:0, negVoteCount:0, votedFor:1, log:[]LogEntry{{0,nil},{1,nil}}, lastLogIndex:1, commitIndex:1, nextIndex:nil, matchIndex:nil}
	sm.handleVoteRespEvent(VoteResp{2,1,false})
	expect(t, sm, sm_expectedstate)
	//if no vote response
	if len(sm.sendEvents) != 0 {
		t.Error(fmt.Sprintf("Expected Send Events Count: %v, found : %v\n", 0, len(sm.sendEvents))) 
	}
	
	//if candidate term is less
	sm = &StateMachine{id:1, leaderId:-1, serverIds:[]int{1,2,3,4,5,6}, state:candidate, term:2, voteCount:0, negVoteCount:0, votedFor:1, log:[]LogEntry{{0,nil},{1,nil}}, lastLogIndex:1, commitIndex:1, nextIndex:nil, matchIndex:nil}
	sm_expectedstate =  &StateMachine{id:1, leaderId:-1, serverIds:[]int{1,2,3,4,5,6}, state:follower, term:3, voteCount:0, negVoteCount:0, votedFor:-1, log:[]LogEntry{{0,nil},{1,nil}}, lastLogIndex:1, commitIndex:1, nextIndex:nil, matchIndex:nil}
	sm.handleVoteRespEvent(VoteResp{2,3,false})
	expect(t, sm, sm_expectedstate)
	//if no vote response
	if len(sm.sendEvents) != 0 {
		t.Error(fmt.Sprintf("Expected Send Events Count: %v, found : %v\n", 0, len(sm.sendEvents))) 
	}
	
	//if voteGranted true, quorum not reached
	sm = &StateMachine{id:1, leaderId:-1, serverIds:[]int{1,2,3,4,5,6}, state:candidate, term:2, voteCount:0, negVoteCount:0, votedFor:1,log:[]LogEntry{{0,nil},{1,nil}}, lastLogIndex:1, commitIndex:1, nextIndex:nil, matchIndex:nil}
	sm_expectedstate =  &StateMachine{id:1, leaderId:-1, serverIds:[]int{1,2,3,4,5,6}, state:candidate, term:2, voteCount:1, negVoteCount:0, votedFor:1, log:[]LogEntry{{0,nil},{1,nil}}, lastLogIndex:1, commitIndex:1, nextIndex:nil, matchIndex:nil}
	sm.handleVoteRespEvent(VoteResp{2,2,true})
	expect(t, sm, sm_expectedstate)
	//if no vote response
	if len(sm.sendEvents) != 0 {
		t.Error(fmt.Sprintf("Expected Send Events Count: %v, found : %v\n", 1, len(sm.sendEvents))) 
	}
	
	//if voteGranted true, quorum reached, becomes leader
	sm = &StateMachine{id:1, leaderId:-1, serverIds:[]int{1,2,3,4,5,6}, state:candidate, term:2, voteCount:3, negVoteCount:0, votedFor:1, log:[]LogEntry{{0,nil},{1,nil}}, lastLogIndex:1, commitIndex:1, nextIndex:nil, matchIndex:nil}
	sm_expectedstate =  &StateMachine{id:1, leaderId:1, serverIds:[]int{1,2,3,4,5,6}, state:leader, term:2, voteCount:0, negVoteCount:0, votedFor:1, log:[]LogEntry{{0,nil},{1,nil},{2,nil}}, lastLogIndex:2, commitIndex:1, nextIndex:map[int]int{2:2, 3:2, 4:2, 5:2, 6:2}, matchIndex:map[int]int{2:0, 3:0, 4:0, 5:0, 6:0}}
	sm.handleVoteRespEvent(VoteResp{2,2,true})
	expect(t, sm, sm_expectedstate)
	//if no vote response
	if len(sm.sendEvents) != 5 {
		t.Error(fmt.Sprintf("Expected Send Events Count: %v, found : %v\n", 5, len(sm.sendEvents))) 
	}
	
	//if voteGranted false, -ve quorum not reached
	sm = &StateMachine{id:1, leaderId:-1, serverIds:[]int{1,2,3,4,5,6}, state:candidate, term:2, voteCount:0, negVoteCount:0, votedFor:1, log:[]LogEntry{{0,nil},{1,nil}}, lastLogIndex:1, commitIndex:1, nextIndex:nil, matchIndex:nil}
	sm_expectedstate =  &StateMachine{id:1, leaderId:-1, serverIds:[]int{1,2,3,4,5,6}, state:candidate, term:2, voteCount:0, negVoteCount:1, votedFor:1, log:[]LogEntry{{0,nil},{1,nil}}, lastLogIndex:1, commitIndex:1, nextIndex:nil, matchIndex:nil}
	sm.handleVoteRespEvent(VoteResp{2,2,false})
	expect(t, sm, sm_expectedstate)
	//if no vote response
	if len(sm.sendEvents) != 0 {
		t.Error(fmt.Sprintf("Expected Send Events Count: %v, found : %v\n", 0, len(sm.sendEvents))) 
	}
	
	//if voteGranted true, -ve quorum reached, step down to follower
	sm = &StateMachine{id:1, leaderId:-1, serverIds:[]int{1,2,3,4,5,6}, state:candidate, term:2, voteCount:3, negVoteCount:3, votedFor:1, log:[]LogEntry{{0,nil},{1,nil}}, lastLogIndex:1, commitIndex:1, nextIndex:nil, matchIndex:nil}
	sm_expectedstate =  &StateMachine{id:1, leaderId:-1, serverIds:[]int{1,2,3,4,5,6}, state:follower, term:2, voteCount:0, negVoteCount:0, votedFor:1, log:[]LogEntry{{0,nil},{1,nil}}, lastLogIndex:1, commitIndex:1, nextIndex:nil, matchIndex:nil}
	sm.handleVoteRespEvent(VoteResp{2,2,false})
	expect(t, sm, sm_expectedstate)
	//if no vote response
	if len(sm.sendEvents) != 0 {
		t.Error(fmt.Sprintf("Expected Send Events Count: %v, found : %v\n", 0, len(sm.sendEvents))) 
	}
}  

//_____________________________________________________________________________________________________________________

func TestCaseCandidateAppEntReqEvent(t *testing.T) {
	//event term greater and prev log entries don't match
	sm = &StateMachine{id:1, leaderId:-1, serverIds:[]int{1,2,3,4,5,6}, state:candidate, term:1, voteCount:0, negVoteCount:0, votedFor:1, log:[]LogEntry{{0,nil}}, lastLogIndex:0, commitIndex:0, nextIndex:nil, matchIndex:nil}
	sm_expectedstate =  &StateMachine{id:1, leaderId:2, serverIds:[]int{1,2,3,4,5,6}, state:follower, term:2, voteCount:0, negVoteCount:0, votedFor:-1, log:[]LogEntry{{0,nil}}, lastLogIndex:0, commitIndex:0, nextIndex:nil, matchIndex:nil}
	sm.handleAppendEntriesReq(AppendEntriesReq{2, 2, 1, 1,[]LogEntry{{2,nil}}, 1})
	expect(t, sm, sm_expectedstate)
	//if no vote response
	if len(sm.sendEvents) != 1 {
		t.Error(fmt.Sprintf("Expected Send Events Count: %v, found : %v\n", 1, len(sm.sendEvents))) 
	} else if sm.sendEvents[0].event.(AppendEntriesResp).success!=false || sm.sendEvents[0].event.(AppendEntriesResp).fromId!=1 || sm.sendEvents[0].event.(AppendEntriesResp).term!=2 {
		// Invalid vote response
		t.Error(fmt.Sprintf("Invalid append entries response\n"))
	}
	
	//prev log entries match
	sm = &StateMachine{id:1, leaderId:-1, serverIds:[]int{1,2,3,4,5,6}, state:candidate, term:1, voteCount:0, negVoteCount:0, votedFor:1, log:[]LogEntry{{0,nil},{1,nil}}, lastLogIndex:1, commitIndex:0, nextIndex:nil, matchIndex:nil}
	sm_expectedstate =  &StateMachine{id:1, leaderId:2, serverIds:[]int{1,2,3,4,5,6}, state:follower, term:2, voteCount:0, negVoteCount:0, votedFor:-1, log:[]LogEntry{{0,nil},{2,nil}}, lastLogIndex:1, commitIndex:1, nextIndex:nil, matchIndex:nil}
	sm.handleAppendEntriesReq(AppendEntriesReq{2, 2, 0, 0,[]LogEntry{{2,nil}}, 1})
	expect(t, sm, sm_expectedstate)
	//if no vote response
	if len(sm.sendEvents) != 1 {
		t.Error(fmt.Sprintf("Expected Send Events Count: %v, found : %v\n", 1, len(sm.sendEvents))) 
	} else if sm.sendEvents[0].event.(AppendEntriesResp).success!=true || sm.sendEvents[0].event.(AppendEntriesResp).fromId!=1 || sm.sendEvents[0].event.(AppendEntriesResp).term!=2 {
		// Invalid vote response
		t.Error(fmt.Sprintf("Invalid append entries response\n"))
	}
	
	//event term smaller
	sm = &StateMachine{id:1, leaderId:-1, serverIds:[]int{1,2,3,4,5,6}, state:candidate, term:2, voteCount:0, negVoteCount:0, votedFor:1, log:[]LogEntry{{0,nil}}, lastLogIndex:0, commitIndex:0, nextIndex:nil, matchIndex:nil}
	sm_expectedstate =  &StateMachine{id:1, leaderId:-1, serverIds:[]int{1,2,3,4,5,6}, state:candidate, term:2, voteCount:0, negVoteCount:0, votedFor:1, log:[]LogEntry{{0,nil}}, lastLogIndex:0, commitIndex:0, nextIndex:nil, matchIndex:nil}
	sm.handleAppendEntriesReq(AppendEntriesReq{1, 3, 0, 0,[]LogEntry{{1,nil}}, 1})
	expect(t, sm, sm_expectedstate)
	//if no vote response
	if len(sm.sendEvents) != 1 {
		t.Error(fmt.Sprintf("Expected Send Events Count: %v, found : %v\n", 1, len(sm.sendEvents))) 
	} else if sm.sendEvents[0].event.(AppendEntriesResp).success!=false || sm.sendEvents[0].event.(AppendEntriesResp).fromId!=1 || sm.sendEvents[0].event.(AppendEntriesResp).term!=2 {
		// Invalid vote response
		t.Error(fmt.Sprintf("Invalid append entries response\n"))
	}
}  

//_____________________________________________________________________________________________________________________

func TestCaseCandidateAppEntRespEvent(t *testing.T) {
	//event term greater
	sm = &StateMachine{id:1, leaderId:-1, serverIds:[]int{1,2,3,4,5,6}, state:candidate, term:1, voteCount:0, negVoteCount:0, votedFor:1, log:[]LogEntry{{0,nil},{1,nil}}, lastLogIndex:1, commitIndex:1, nextIndex:nil, matchIndex:nil}
	//vote granted; term,votedFor changed
	sm_expectedstate =  &StateMachine{id:1, leaderId:-1, serverIds:[]int{1,2,3,4,5,6}, state:follower, term:3, voteCount:0, negVoteCount:0, votedFor:-1, log:[]LogEntry{{0,nil},{1,nil}}, lastLogIndex:1, commitIndex:1, nextIndex:nil, matchIndex:nil}
	sm.handleAppendEntriesResp(AppendEntriesResp{AppendEntriesReq{},2,3,false})
	expect(t, sm, sm_expectedstate)
	//if no vote response
	if len(sm.sendEvents) != 0 {
		t.Error(fmt.Sprintf("Expected Send Events Count: %v, found : %v\n", 0, len(sm.sendEvents))) 
	}
	
	//event term not greater
	sm = &StateMachine{id:1, leaderId:-1, serverIds:[]int{1,2,3,4,5,6}, state:candidate, term:1, voteCount:0, negVoteCount:0, votedFor:1, log:[]LogEntry{{0,nil},{1,nil}}, lastLogIndex:1, commitIndex:1, nextIndex:nil, matchIndex:nil}
	//vote granted; term,votedFor changed
	sm_expectedstate =  &StateMachine{id:1, leaderId:-1, serverIds:[]int{1,2,3,4,5,6}, state:candidate, term:1, voteCount:0, negVoteCount:0, votedFor:1, log:[]LogEntry{{0,nil},{1,nil}}, lastLogIndex:1, commitIndex:1, nextIndex:nil, matchIndex:nil}
	sm.handleAppendEntriesResp(AppendEntriesResp{AppendEntriesReq{},2,1,false})
	expect(t, sm, sm_expectedstate)
	//if no vote response
	if len(sm.sendEvents) != 0 {
		t.Error(fmt.Sprintf("Expected Send Events Count: %v, found : %v\n", 1, len(sm.sendEvents))) 
	}
}   


//________________________________________ Leader Tests____________________________________________________________


func TestCaseLeaderAppendEvent(t *testing.T) {
	sm = &StateMachine{id:1, leaderId:1, serverIds:[]int{1,2,3,4,5,6}, state:leader, term:1, voteCount:0, negVoteCount:0, votedFor:1, log:[]LogEntry{{0,nil},{1,nil}}, lastLogIndex:1, commitIndex:1, nextIndex:map[int]int{2:2,3:2,4:2,5:2,6:2}, matchIndex:map[int]int{2:1,3:1,4:1,5:1,6:1}}
	sm_expectedstate =  &StateMachine{id:1, leaderId:1, serverIds:[]int{1,2,3,4,5,6}, state:leader, term:1, voteCount:0, negVoteCount:0, votedFor:1, log:[]LogEntry{{0,nil},{1,nil},{1,[]byte("Rahul")}}, lastLogIndex:2, commitIndex:1, nextIndex:map[int]int{2:2,3:2,4:2,5:2,6:2}, matchIndex:map[int]int{2:1,3:1,4:1,5:1,6:1}}
	// append to log and check whether append entries requests sent
	sm.handleAppendEvent(Append{[]byte("Rahul")})
	expect(t, sm, sm_expectedstate)
	if len(sm.sendEvents) != 5 {
		t.Error(fmt.Sprintf("Expected Send Events Count: %v, found : %v\n", 5, len(sm.sendEvents))) 
	}
}  

//_____________________________________________________________________________________________________________________

func TestCaseLeaderTimeoutEvent(t *testing.T) {
	sm = &StateMachine{id:1, leaderId:1, serverIds:[]int{1,2,3,4,5,6}, state:leader, term:1, voteCount:0, negVoteCount:0, votedFor:1, log:[]LogEntry{{0,nil},{1,nil},{1,[]byte("Rahul")}}, lastLogIndex:2, commitIndex:1, nextIndex:map[int]int{2:2,3:2,4:2,5:2,6:2}, matchIndex:map[int]int{2:1,3:1,4:1,5:1,6:1}}
	sm_expectedstate =  &StateMachine{id:1, leaderId:1, serverIds:[]int{1,2,3,4,5,6}, state:leader, term:1, voteCount:0, negVoteCount:0, votedFor:1, log:[]LogEntry{{0,nil},{1,nil},{1,[]byte("Rahul")},{1,nil}}, lastLogIndex:3, commitIndex:1, nextIndex:map[int]int{2:2,3:2,4:2,5:2,6:2}, matchIndex:map[int]int{2:1,3:1,4:1,5:1,6:1}}
	// check whether heart beats sent
	sm.handleTimeoutEvent(Timeout{})
	expect(t, sm, sm_expectedstate)
	if len(sm.sendEvents) != 5 {
		t.Error(fmt.Sprintf("Expected Send Events Count: %v, found : %v\n", 5, len(sm.sendEvents))) 
	}
}  

//_____________________________________________________________________________________________________________________

func TestCaseLeaderVoteReqEvent(t *testing.T) {
	//event term greater
	sm = &StateMachine{id:1, leaderId:1, serverIds:[]int{1,2,3,4,5,6}, state:leader, term:2, voteCount:0, negVoteCount:0, votedFor:1, log:[]LogEntry{{0,nil},{1,nil}}, lastLogIndex:1, commitIndex:1, nextIndex:map[int]int{2:2,3:2,4:2,5:2,6:2}, matchIndex:map[int]int{2:1,3:1,4:1,5:1,6:1}}
	//vote granted; term,votedFor changed
	sm_expectedstate =  &StateMachine{id:1, leaderId:-1, serverIds:[]int{1,2,3,4,5,6}, state:follower, term:3, voteCount:0, negVoteCount:0, votedFor:3, log:[]LogEntry{{0,nil},{1,nil}}, lastLogIndex:1, commitIndex:1, nextIndex:nil, matchIndex:nil}
	sm.handleVoteReqEvent(VoteReq{3,3,1,1})
	expect(t, sm, sm_expectedstate)
	//if no vote response
	if len(sm.sendEvents) != 1 {
		t.Error(fmt.Sprintf("Expected Send Events Count: %v, found : %v\n", 1, len(sm.sendEvents))) 
	} else if sm.sendEvents[0].event.(VoteResp).voteGranted!=true || sm.sendEvents[0].event.(VoteResp).fromId!=1 || sm.sendEvents[0].event.(VoteResp).term!=3 {
		// Invalid vote response
		t.Error(fmt.Sprintf("Invalid vote response\n"))
	}
	//candidate not as updated, vote not granted
	sm = &StateMachine{id:1, leaderId:1, serverIds:[]int{1,2,3,4,5,6}, state:leader, term:2, voteCount:0, negVoteCount:0, votedFor:1, log:[]LogEntry{{0,nil},{1,nil}}, lastLogIndex:1, commitIndex:1, nextIndex:map[int]int{2:2,3:2,4:2,5:2,6:2}, matchIndex:map[int]int{2:1,3:1,4:1,5:1,6:1}}
	sm_expectedstate =  &StateMachine{id:1, leaderId:-1, serverIds:[]int{1,2,3,4,5,6}, state:follower, term:3, voteCount:0, negVoteCount:0, votedFor:-1, log:[]LogEntry{{0,nil},{1,nil}}, lastLogIndex:1, commitIndex:1, nextIndex:nil, matchIndex:nil}
	sm.handleVoteReqEvent(VoteReq{3,3,0,0})
	expect(t, sm, sm_expectedstate)
	//if no vote response
	if len(sm.sendEvents) != 1 {
		t.Error(fmt.Sprintf("Expected Send Events Count: %v, found : %v\n", 1, len(sm.sendEvents))) 
	} else if sm.sendEvents[0].event.(VoteResp).voteGranted!=false || sm.sendEvents[0].event.(VoteResp).fromId!=1 || sm.sendEvents[0].event.(VoteResp).term!=3 {
		// Invalid vote response
		t.Error(fmt.Sprintf("Invalid vote response\n"))
	}
	
	//event term equal
	sm = &StateMachine{id:1, leaderId:1, serverIds:[]int{1,2,3,4,5,6}, state:leader, term:2, voteCount:0, negVoteCount:0, votedFor:1, log:[]LogEntry{{0,nil},{1,nil}}, lastLogIndex:1, commitIndex:1, nextIndex:map[int]int{2:2,3:2,4:2,5:2,6:2}, matchIndex:map[int]int{2:1,3:1,4:1,5:1,6:1}}
	sm_expectedstate =  &StateMachine{id:1, leaderId:1, serverIds:[]int{1,2,3,4,5,6}, state:leader, term:2, voteCount:0, negVoteCount:0, votedFor:1, log:[]LogEntry{{0,nil},{1,nil}}, lastLogIndex:1, commitIndex:1, nextIndex:map[int]int{2:2,3:2,4:2,5:2,6:2}, matchIndex:map[int]int{2:1,3:1,4:1,5:1,6:1}}
	sm.handleVoteReqEvent(VoteReq{2,3,0,0})
	expect(t, sm, sm_expectedstate)
	//if no vote response
	if len(sm.sendEvents) != 1 {
		t.Error(fmt.Sprintf("Expected Send Events Count: %v, found : %v\n", 1, len(sm.sendEvents))) 
	} else if sm.sendEvents[0].event.(VoteResp).voteGranted!=false || sm.sendEvents[0].event.(VoteResp).fromId!=1 || sm.sendEvents[0].event.(VoteResp).term!=2 {
		// Invalid vote response
		t.Error(fmt.Sprintf("Invalid vote response\n"))
	}
	
	//if candidate term is less
	sm = &StateMachine{id:1, leaderId:1, serverIds:[]int{1,2,3,4,5,6}, state:leader, term:2, voteCount:0, negVoteCount:0, votedFor:1, log:[]LogEntry{{0,nil},{1,nil}}, lastLogIndex:1, commitIndex:1, nextIndex:map[int]int{2:2,3:2,4:2,5:2,6:2}, matchIndex:map[int]int{2:1,3:1,4:1,5:1,6:1}}
	sm_expectedstate =  &StateMachine{id:1, leaderId:1, serverIds:[]int{1,2,3,4,5,6}, state:leader, term:2, voteCount:0, negVoteCount:0, votedFor:1, log:[]LogEntry{{0,nil},{1,nil}}, lastLogIndex:1, commitIndex:1, nextIndex:map[int]int{2:2,3:2,4:2,5:2,6:2}, matchIndex:map[int]int{2:1,3:1,4:1,5:1,6:1}}
	sm.handleVoteReqEvent(VoteReq{1,1,1,1})
	expect(t, sm, sm_expectedstate)
	//if no vote response
	if len(sm.sendEvents) != 1 {
		t.Error(fmt.Sprintf("Expected Send Events Count: %v, found : %v\n", 1, len(sm.sendEvents))) 
	} else if sm.sendEvents[0].event.(VoteResp).voteGranted!=false || sm.sendEvents[0].event.(VoteResp).fromId!=1 || sm.sendEvents[0].event.(VoteResp).term!=2 {
		// Invalid vote response
		t.Error(fmt.Sprintf("Invalid vote response\n"))
	}
}   

//_____________________________________________________________________________________________________________________

func TestCaseLeaderVoteRespEvent(t *testing.T) {
	//event term greater
	sm = &StateMachine{id:1, leaderId:2, serverIds:[]int{1,2,3,4,5,6}, state:leader, term:1, voteCount:0, negVoteCount:0, votedFor:2, log:[]LogEntry{{0,nil},{1,nil}}, lastLogIndex:1, commitIndex:1, nextIndex:map[int]int{2:2,3:2,4:2,5:2,6:2}, matchIndex:map[int]int{2:1,3:1,4:1,5:1,6:1}}
	//vote granted; term,votedFor changed
	sm_expectedstate =  &StateMachine{id:1, leaderId:-1, serverIds:[]int{1,2,3,4,5,6}, state:follower, term:3, voteCount:0, negVoteCount:0, votedFor:-1, log:[]LogEntry{{0,nil},{1,nil}}, lastLogIndex:1, commitIndex:1, nextIndex:nil, matchIndex:nil}
	sm.handleVoteRespEvent(VoteResp{2,3,false})
	expect(t, sm, sm_expectedstate)
	//if no vote response
	if len(sm.sendEvents) != 0 {
		t.Error(fmt.Sprintf("Expected Send Events Count: %v, found : %v\n", 0, len(sm.sendEvents))) 
	}
	
	//event term not greater
	sm = &StateMachine{id:1, leaderId:1, serverIds:[]int{1,2,3,4,5,6}, state:leader, term:1, voteCount:0, negVoteCount:0, votedFor:1, log:[]LogEntry{{0,nil},{1,nil}}, lastLogIndex:1, commitIndex:1, nextIndex:map[int]int{2:2,3:2,4:2,5:2,6:2}, matchIndex:map[int]int{2:1,3:1,4:1,5:1,6:1}}
	//vote granted; term,votedFor changed
	sm_expectedstate =  &StateMachine{id:1, leaderId:1, serverIds:[]int{1,2,3,4,5,6}, state:leader, term:1, voteCount:0, negVoteCount:0, votedFor:1, log:[]LogEntry{{0,nil},{1,nil}}, lastLogIndex:1, commitIndex:1, nextIndex:map[int]int{2:2,3:2,4:2,5:2,6:2}, matchIndex:map[int]int{2:1,3:1,4:1,5:1,6:1}}
	sm.handleVoteRespEvent(VoteResp{2,1,false})
	expect(t, sm, sm_expectedstate)
	//if no vote response
	if len(sm.sendEvents) != 0 {
		t.Error(fmt.Sprintf("Expected Send Events Count: %v, found : %v\n", 0, len(sm.sendEvents))) 
	}
}  

//_____________________________________________________________________________________________________________________

func TestCaseLeaderAppEntReqEvent(t *testing.T) {
	//event term greater and prev log entries don't match
	sm = &StateMachine{id:1, leaderId:1, serverIds:[]int{1,2,3,4,5,6}, state:leader, term:1, voteCount:0, negVoteCount:0, votedFor:1, log:[]LogEntry{{0,nil}}, lastLogIndex:0, commitIndex:0, nextIndex:map[int]int{2:2,3:2,4:2,5:2,6:2}, matchIndex:map[int]int{2:1,3:1,4:1,5:1,6:1}}
	sm_expectedstate =  &StateMachine{id:1, leaderId:2, serverIds:[]int{1,2,3,4,5,6}, state:follower, term:2, voteCount:0, negVoteCount:0, votedFor:-1, log:[]LogEntry{{0,nil}}, lastLogIndex:0, commitIndex:0, nextIndex:nil, matchIndex:nil}
	sm.handleAppendEntriesReq(AppendEntriesReq{2, 2, 1, 1,[]LogEntry{{2,nil}}, 1})
	expect(t, sm, sm_expectedstate)
	//if no vote response
	if len(sm.sendEvents) != 1 {
		t.Error(fmt.Sprintf("Expected Send Events Count: %v, found : %v\n", 1, len(sm.sendEvents))) 
	} else if sm.sendEvents[0].event.(AppendEntriesResp).success!=false || sm.sendEvents[0].event.(AppendEntriesResp).fromId!=1 || sm.sendEvents[0].event.(AppendEntriesResp).term!=2 {
		// Invalid vote response
		t.Error(fmt.Sprintf("Invalid append entries response\n"))
	}
	
	//prev log entries match
	sm = &StateMachine{id:1, leaderId:1, serverIds:[]int{1,2,3,4,5,6}, state:leader, term:1, voteCount:0, negVoteCount:0, votedFor:1, log:[]LogEntry{{0,nil},{1,nil}}, lastLogIndex:1, commitIndex:0, nextIndex:map[int]int{2:2,3:2,4:2,5:2,6:2}, matchIndex:map[int]int{2:0,3:0,4:0,5:0,6:0}}
	sm_expectedstate =  &StateMachine{id:1, leaderId:2, serverIds:[]int{1,2,3,4,5,6}, state:follower, term:2, voteCount:0, negVoteCount:0, votedFor:-1, log:[]LogEntry{{0,nil},{2,nil}}, lastLogIndex:1, commitIndex:1, nextIndex:nil, matchIndex:nil}
	sm.handleAppendEntriesReq(AppendEntriesReq{2, 2, 0, 0,[]LogEntry{{2,nil}}, 1})
	expect(t, sm, sm_expectedstate)
	//if no vote response
	if len(sm.sendEvents) != 1 {
		t.Error(fmt.Sprintf("Expected Send Events Count: %v, found : %v\n", 1, len(sm.sendEvents))) 
	} else if sm.sendEvents[0].event.(AppendEntriesResp).success!=true || sm.sendEvents[0].event.(AppendEntriesResp).fromId!=1 || sm.sendEvents[0].event.(AppendEntriesResp).term!=2 {
		// Invalid vote response
		t.Error(fmt.Sprintf("Invalid append entries response\n"))
	}
	
	//event term smaller
	sm = &StateMachine{id:1, leaderId:1, serverIds:[]int{1,2,3,4,5,6}, state:leader, term:2, voteCount:0, negVoteCount:0, votedFor:1, log:[]LogEntry{{0,nil}}, lastLogIndex:0, commitIndex:0, nextIndex:map[int]int{2:2,3:2,4:2,5:2,6:2}, matchIndex:map[int]int{2:1,3:1,4:1,5:1,6:1}}
	sm_expectedstate =  &StateMachine{id:1, leaderId:1, serverIds:[]int{1,2,3,4,5,6}, state:leader, term:2, voteCount:0, negVoteCount:0, votedFor:1, log:[]LogEntry{{0,nil}}, lastLogIndex:0, commitIndex:0, nextIndex:map[int]int{2:2,3:2,4:2,5:2,6:2}, matchIndex:map[int]int{2:1,3:1,4:1,5:1,6:1}}
	sm.handleAppendEntriesReq(AppendEntriesReq{1, 3, 0, 0,[]LogEntry{{1,nil}}, 1})
	expect(t, sm, sm_expectedstate)
	//if no vote response
	if len(sm.sendEvents) != 1 {
		t.Error(fmt.Sprintf("Expected Send Events Count: %v, found : %v\n", 1, len(sm.sendEvents))) 
	} else if sm.sendEvents[0].event.(AppendEntriesResp).success!=false || sm.sendEvents[0].event.(AppendEntriesResp).fromId!=1 || sm.sendEvents[0].event.(AppendEntriesResp).term!=2 {
		// Invalid vote response
		t.Error(fmt.Sprintf("Invalid append entries response\n"))
	}
}   

//_____________________________________________________________________________________________________________________

func TestCaseLeaderAppEntRespEvent(t *testing.T) {
	
	//event term is less
	sm = &StateMachine{id:1, leaderId:1, serverIds:[]int{1,2,3,4,5,6}, state:leader, term:2, voteCount:0, negVoteCount:0, votedFor:1, log:[]LogEntry{{0,nil},{1,nil},{2,nil}}, lastLogIndex:2, commitIndex:1, nextIndex:map[int]int{2:2,3:2,4:2,5:2,6:2}, matchIndex:map[int]int{2:1,3:1,4:1,5:1,6:1}}
	//vote granted; term,votedFor changed
	sm_expectedstate =  &StateMachine{id:1, leaderId:1, serverIds:[]int{1,2,3,4,5,6}, state:leader, term:2, voteCount:0, negVoteCount:0, votedFor:1, log:[]LogEntry{{0,nil},{1,nil},{2,nil}}, lastLogIndex:2, commitIndex:1, nextIndex:map[int]int{2:2,3:2,4:2,5:2,6:2}, matchIndex:map[int]int{2:1,3:1,4:1,5:1,6:1}}
	sm.handleAppendEntriesResp(AppendEntriesResp{AppendEntriesReq{},2,1,false})
	expect(t, sm, sm_expectedstate)
	//if no vote response
	if len(sm.sendEvents) != 0 {
		t.Error(fmt.Sprintf("Expected Send Events Count: %v, found : %v\n", 1, len(sm.sendEvents))) 
	}
	
	//event term is greater
	sm = &StateMachine{id:1, leaderId:1, serverIds:[]int{1,2,3,4,5,6}, state:leader, term:2, voteCount:0, negVoteCount:0, votedFor:1, log:[]LogEntry{{0,nil},{1,nil},{2,nil}}, lastLogIndex:2, commitIndex:1, nextIndex:map[int]int{2:2,3:2,4:2,5:2,6:2}, matchIndex:map[int]int{2:1,3:1,4:1,5:1,6:1}}
	//vote granted; term,votedFor changed
	sm_expectedstate =  &StateMachine{id:1, leaderId:-1, serverIds:[]int{1,2,3,4,5,6}, state:follower, term:3, voteCount:0, negVoteCount:0, votedFor:-1, log:[]LogEntry{{0,nil},{1,nil},{2,nil}}, lastLogIndex:2, commitIndex:1, nextIndex:nil, matchIndex:nil}
	sm.handleAppendEntriesResp(AppendEntriesResp{AppendEntriesReq{},2,3,false})
	expect(t, sm, sm_expectedstate)
	//if no vote response
	if len(sm.sendEvents) != 0 {
		t.Error(fmt.Sprintf("Expected Send Events Count: %v, found : %v\n", 1, len(sm.sendEvents))) 
	}
	
	//event term is equal, failure, resend with decremented nextIndex 
	sm = &StateMachine{id:1, leaderId:1, serverIds:[]int{1,2,3,4,5,6}, state:leader, term:2, voteCount:0, negVoteCount:0, votedFor:1, log:[]LogEntry{{0,nil},{1,nil},{2,nil}}, lastLogIndex:2, commitIndex:1, nextIndex:map[int]int{2:2,3:3,4:3,5:3,6:2}, matchIndex:map[int]int{2:0,3:2,4:2,5:2,6:1}}
	//vote granted; term,votedFor changed
	sm_expectedstate =  &StateMachine{id:1, leaderId:1, serverIds:[]int{1,2,3,4,5,6}, state:leader, term:2, voteCount:0, negVoteCount:0, votedFor:1, log:[]LogEntry{{0,nil},{1,nil},{2,nil}}, lastLogIndex:2, commitIndex:1, nextIndex:map[int]int{2:1,3:3,4:3,5:3,6:2}, matchIndex:map[int]int{2:0,3:2,4:2,5:2,6:1}}
	sm.handleAppendEntriesResp(AppendEntriesResp{AppendEntriesReq{2, 1, 1, 1,[]LogEntry{{2,nil}}, 1},2,2,false})
	expect(t, sm, sm_expectedstate)
	//if no vote response
	if len(sm.sendEvents) != 1 {
		t.Error(fmt.Sprintf("Expected Send Events Count: %v, found : %v\n", 1, len(sm.sendEvents))) 
	}
	
	//event term is equal, success
	sm = &StateMachine{id:1, leaderId:1, serverIds:[]int{1,2,3,4,5,6}, state:leader, term:2, voteCount:0, negVoteCount:0, votedFor:1, log:[]LogEntry{{0,nil},{1,nil},{2,nil}}, lastLogIndex:2, commitIndex:1, nextIndex:map[int]int{2:1,3:3,4:3,5:3,6:2}, matchIndex:map[int]int{2:0,3:2,4:2,5:2,6:1}}
	//vote granted; term,votedFor changed
	sm_expectedstate =  &StateMachine{id:1, leaderId:1, serverIds:[]int{1,2,3,4,5,6}, state:leader, term:2, voteCount:0, negVoteCount:0, votedFor:1, log:[]LogEntry{{0,nil},{1,nil},{2,nil}}, lastLogIndex:2, commitIndex:2, nextIndex:map[int]int{2:3,3:3,4:3,5:3,6:2}, matchIndex:map[int]int{2:2,3:2,4:2,5:2,6:1}}
	sm.handleAppendEntriesResp(AppendEntriesResp{AppendEntriesReq{2, 1, 0, 0,[]LogEntry{{1,nil},{2,nil}}, 1},2,2,true})
	expect(t, sm, sm_expectedstate)
	//if no vote response
	if len(sm.sendEvents) != 0 {
		t.Error(fmt.Sprintf("Expected Send Events Count: %v, found : %v\n", 0, len(sm.sendEvents))) 
	}
}  
  

//_____________________________________________________________________________________________________________________
//_____________________________________________________________________________________________________________________



func testEqMaps(a, b map[int]int) bool {

	if a == nil && b == nil { 
	return true; 
	}

	if a == nil || b == nil { 
	return false; 
	}

	if len(a) != len(b) {
	return false
	}

	for i := range a {
	if a[i] != b[i] {
	    return false
	}
	}

	return true
}  

//_____________________________________________________________________________________________________________________

func testEqArrInt(a, b []int) bool {

	if a == nil && b == nil { 
	return true; 
	}

	if a == nil || b == nil { 
	return false; 
	}

	if len(a) != len(b) {
	return false
	}

	for i := range a {
	if a[i] != b[i] {
	    return false
	}
	}

	return true
}  

//_____________________________________________________________________________________________________________________

func testEqArrLog(a, b []LogEntry) bool {

	if a == nil && b == nil { 
	return true; 
	}

	if a == nil || b == nil { 
	return false; 
	}

	if len(a) != len(b) {
	return false
	}

	for i := range a {
			if a[i].term != b[i].term {
				return false
			}
			if !bytes.Equal(a[i].data, b[i].data)  {
				return false
			}
	}

	return true
}  

//_____________________________________________________________________________________________________________________

func expect(t *testing.T, sm1 *StateMachine, sm2 *StateMachine) {
	if sm1.id != sm2.id {
		t.Error(fmt.Sprintf("Expected Server Id: %v, found Server Id: %v\n", sm2.id, sm1.id)) 
	}
	
	if sm1.leaderId != sm2.leaderId {
		t.Error(fmt.Sprintf("Expected Leader Id: %v, found Leader Id: %v\n", sm2.leaderId, sm1.leaderId)) 
	}
	
	if testEqArrInt(sm1.serverIds,sm2.serverIds)==false {
		t.Error(fmt.Sprintf("Expected %v, found %v\n", sm2.serverIds, sm1.serverIds)) 
	}
	
	if sm1.state != sm2.state {
		t.Error(fmt.Sprintf("Expected Server State: %v, found Server State: %v\n", sm2.state, sm1.state)) 
	}
	
	if sm1.term != sm2.term {
		t.Error(fmt.Sprintf("Expected Server Term: %v, found Server Term: %v\n", sm2.term, sm1.term)) 
	}
	
	if sm1.voteCount != sm2.voteCount {
		t.Error(fmt.Sprintf("Expected Server Vote Count: %v, found Server Vote Count: %v\n", sm2.voteCount, sm1.voteCount)) 
	}
	
	if sm1.negVoteCount != sm2.negVoteCount {
		t.Error(fmt.Sprintf("Expected Server Vote Count: %v, found Server Vote Count: %v\n", sm2.negVoteCount, sm1.negVoteCount)) 
	}
	
	if sm1.votedFor != sm2.votedFor {
		t.Error(fmt.Sprintf("Expected Server Voted For Id: %v, found Server Voted For Id: %v\n", sm2.votedFor, sm1.votedFor)) 
	}
	
	if testEqArrLog(sm1.log,sm2.log)==false {
		t.Error(fmt.Sprintf("Expected %v, found %v\n", sm2.log, sm1.log)) 
	}
	
	if sm1.lastLogIndex != sm2.lastLogIndex {
		t.Error(fmt.Sprintf("Expected Server Last Log Index: %v, found Server Last Log Index: %v\n", sm2.lastLogIndex, sm1.lastLogIndex)) 
	}
	
	if sm1.commitIndex != sm2.commitIndex {
		t.Error(fmt.Sprintf("Expected Server Commit Index: %v, found Server Commit Index: %v\n", sm2.commitIndex, sm1.commitIndex)) 
	}
	
	if testEqMaps(sm1.nextIndex, sm2.nextIndex)==false {
		t.Error(fmt.Sprintf("Expected %v, found %v\n", sm2.nextIndex, sm1.nextIndex)) 
	}
	
	if testEqMaps(sm1.matchIndex, sm2.matchIndex)==false {
		t.Error(fmt.Sprintf("Expected %v, found %v\n", sm2.matchIndex, sm1.matchIndex)) 
	}
	
}

