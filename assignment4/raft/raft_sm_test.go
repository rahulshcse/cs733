package raftnode

import (
	"testing"
	"fmt"
	"bytes"
)

var Sm *StateMachine
var Sm_expectedState *StateMachine


//________________________________________ Follower Tests____________________________________________________________

func TestCaseFollowerAppendEvent(t *testing.T) {
	fmt.Printf("\n\n\n\n\n\n***State Machine Tests Start***\n\n\n")
	Sm = &StateMachine{Id:1, LeaderId:2, ServerIds:[]int{1,2,3,4,5,6}, State:follower, Term:1, VoteCount:0, NegVoteCount:0, VotedFor:2, Log:[]LogEntry{{0,nil,1},{1,nil,1}}, LastLogIndex:1, CommitIndex:1, NextIndex:nil, MatchIndex:nil}
	Sm_expectedState =  &StateMachine{Id:1, LeaderId:2, ServerIds:[]int{1,2,3,4,5,6}, State:follower, Term:1, VoteCount:0, NegVoteCount:0, VotedFor:2, Log:[]LogEntry{{0,nil,1},{1,nil,1}}, LastLogIndex:1, CommitIndex:1, NextIndex:nil, MatchIndex:nil}
	// check whether Event ignored
	Sm.handleAppendEvent(Append{[]byte("Rahul")})
	expect(t, Sm, Sm_expectedState)
}  

//_____________________________________________________________________________________________________________________

func TestCaseFollowerTimeoutEvent(t *testing.T) {
	Sm = &StateMachine{Id:1, LeaderId:2, ServerIds:[]int{1,2,3,4,5,6}, State:follower, Term:1, VoteCount:0, NegVoteCount:0, VotedFor:2, Log:[]LogEntry{{0,nil,1},{1,nil,1}}, LastLogIndex:1, CommitIndex:1, NextIndex:nil, MatchIndex:nil}
	Sm_expectedState =  &StateMachine{Id:1, LeaderId:-1, ServerIds:[]int{1,2,3,4,5,6}, State:candidate, Term:2, VoteCount:0, NegVoteCount:0, VotedFor:1, Log:[]LogEntry{{0,nil,1},{1,nil,1}}, LastLogIndex:1, CommitIndex:1, NextIndex:nil, MatchIndex:nil}
	// check whether changed to candidate State, and Sent vote request to peers 
	Sm.handleTimeoutEvent(Timeout{})
	expect(t, Sm, Sm_expectedState)
	if len(Sm.SendEvents) != 5 {
		t.Error(fmt.Sprintf("Expected Send Events Count: %v, found : %v\n", 5, len(Sm.SendEvents))) 
	}
}  

//_____________________________________________________________________________________________________________________

func TestCaseFollowerVoteReqEvent(t *testing.T) {
	//Event Term greater
	Sm = &StateMachine{Id:1, LeaderId:2, ServerIds:[]int{1,2,3,4,5,6}, State:follower, Term:1, VoteCount:0, NegVoteCount:0, VotedFor:2, Log:[]LogEntry{{0,nil,1},{1,nil,1}}, LastLogIndex:1, CommitIndex:1, NextIndex:nil, MatchIndex:nil}
	//vote granted; Term,VotedFor changed
	Sm_expectedState =  &StateMachine{Id:1, LeaderId:2, ServerIds:[]int{1,2,3,4,5,6}, State:follower, Term:2, VoteCount:0, NegVoteCount:0, VotedFor:-1, Log:[]LogEntry{{0,nil,1},{1,nil,1}}, LastLogIndex:1, CommitIndex:1, NextIndex:nil, MatchIndex:nil}
	Sm.handleVoteReqEvent(VoteReq{2,3,1,0})
	expect(t, Sm, Sm_expectedState)
	//if no vote response
	if len(Sm.SendEvents) != 1 {
		t.Error(fmt.Sprintf("Expected Send Events Count: %v, found : %v\n", 1, len(Sm.SendEvents))) 
	} else if Sm.SendEvents[0].Event.(VoteResp).VoteGranted!=false || Sm.SendEvents[0].Event.(VoteResp).FromId!=1 || Sm.SendEvents[0].Event.(VoteResp).Term!=2 {
		// Invalid vote response
		t.Error(fmt.Sprintf("Invalid vote response\n"))
	}
	//leader not as updated, vote not granted
	Sm = &StateMachine{Id:1, LeaderId:2, ServerIds:[]int{1,2,3,4,5,6}, State:follower, Term:2, VoteCount:0, NegVoteCount:0, VotedFor:2, Log:[]LogEntry{{0,nil,1},{1,nil,1}}, LastLogIndex:1, CommitIndex:1, NextIndex:nil, MatchIndex:nil}
	Sm_expectedState =  &StateMachine{Id:1, LeaderId:2, ServerIds:[]int{1,2,3,4,5,6}, State:follower, Term:2, VoteCount:0, NegVoteCount:0, VotedFor:2, Log:[]LogEntry{{0,nil,1},{1,nil,1}}, LastLogIndex:1, CommitIndex:1, NextIndex:nil, MatchIndex:nil}
	Sm.handleVoteReqEvent(VoteReq{2,0,1,1})
	expect(t, Sm, Sm_expectedState)
	//if no vote response
	if len(Sm.SendEvents) != 1 {
		t.Error(fmt.Sprintf("Expected Send Events Count: %v, found : %v\n", 1, len(Sm.SendEvents))) 
	} else if Sm.SendEvents[0].Event.(VoteResp).VoteGranted!=false || Sm.SendEvents[0].Event.(VoteResp).FromId!=1 || Sm.SendEvents[0].Event.(VoteResp).Term!=2 {
		// Invalid vote response
		t.Error(fmt.Sprintf("Invalid vote response\n"))
	}
	
	//Event Term equal
	Sm = &StateMachine{Id:1, LeaderId:2, ServerIds:[]int{1,2,3,4,5,6}, State:follower, Term:1, VoteCount:0, NegVoteCount:0, VotedFor:2, Log:[]LogEntry{{0,nil,1},{1,nil,1}}, LastLogIndex:1, CommitIndex:1, NextIndex:nil, MatchIndex:nil}
	//leader not as updated, vote not granted
	Sm_expectedState =  &StateMachine{Id:1, LeaderId:2, ServerIds:[]int{1,2,3,4,5,6}, State:follower, Term:2, VoteCount:0, NegVoteCount:0, VotedFor:-1, Log:[]LogEntry{{0,nil,1},{1,nil,1}}, LastLogIndex:1, CommitIndex:1, NextIndex:nil, MatchIndex:nil}
	Sm.handleVoteReqEvent(VoteReq{2,3,1,0})
	expect(t, Sm, Sm_expectedState)
	//if no vote response
	if len(Sm.SendEvents) != 1 {
		t.Error(fmt.Sprintf("Expected Send Events Count: %v, found : %v\n", 1, len(Sm.SendEvents))) 
	} else if Sm.SendEvents[0].Event.(VoteResp).VoteGranted!=false || Sm.SendEvents[0].Event.(VoteResp).FromId!=1 || Sm.SendEvents[0].Event.(VoteResp).Term!=2 {
		// Invalid vote response
		t.Error(fmt.Sprintf("Invalid vote response\n"))
	}
	//vote granted; Term,VotedFor changed
	Sm = &StateMachine{Id:1, LeaderId:2, ServerIds:[]int{1,2,3,4,5,6}, State:follower, Term:1, VoteCount:0, NegVoteCount:0, VotedFor:2, Log:[]LogEntry{{0,nil,1},{1,nil,1}}, LastLogIndex:1, CommitIndex:1, NextIndex:nil, MatchIndex:nil}
	Sm_expectedState =  &StateMachine{Id:1, LeaderId:2, ServerIds:[]int{1,2,3,4,5,6}, State:follower, Term:2, VoteCount:0, NegVoteCount:0, VotedFor:3, Log:[]LogEntry{{0,nil,1},{1,nil,1}}, LastLogIndex:1, CommitIndex:1, NextIndex:nil, MatchIndex:nil}
	Sm.handleVoteReqEvent(VoteReq{2,3,1,1})
	expect(t, Sm, Sm_expectedState)
	//if no vote response
	if len(Sm.SendEvents) != 1 {
		t.Error(fmt.Sprintf("Expected Send Events Count: %v, found : %v\n", 1, len(Sm.SendEvents))) 
	} else if Sm.SendEvents[0].Event.(VoteResp).VoteGranted!=true || Sm.SendEvents[0].Event.(VoteResp).FromId!=1 || Sm.SendEvents[0].Event.(VoteResp).Term!=2 {
		// Invalid vote response
		t.Error(fmt.Sprintf("Invalid vote response\n"))
	}
	
	//Event Term equal, candidate more updated
	Sm = &StateMachine{Id:1, LeaderId:2, ServerIds:[]int{1,2,3,4,5,6}, State:follower, Term:1, VoteCount:0, NegVoteCount:0, VotedFor:2, Log:[]LogEntry{{0,nil,1},{1,nil,1}}, LastLogIndex:1, CommitIndex:1, NextIndex:nil, MatchIndex:nil}
	//leader not as updated, vote not granted
	Sm_expectedState =  &StateMachine{Id:1, LeaderId:2, ServerIds:[]int{1,2,3,4,5,6}, State:follower, Term:2, VoteCount:0, NegVoteCount:0, VotedFor:-1, Log:[]LogEntry{{0,nil,1},{1,nil,1}}, LastLogIndex:1, CommitIndex:1, NextIndex:nil, MatchIndex:nil}
	Sm.handleVoteReqEvent(VoteReq{2,3,1,0})
	expect(t, Sm, Sm_expectedState)
	//if no vote response
	if len(Sm.SendEvents) != 1 {
		t.Error(fmt.Sprintf("Expected Send Events Count: %v, found : %v\n", 1, len(Sm.SendEvents))) 
	} else if Sm.SendEvents[0].Event.(VoteResp).VoteGranted!=false || Sm.SendEvents[0].Event.(VoteResp).FromId!=1 || Sm.SendEvents[0].Event.(VoteResp).Term!=2 {
		// Invalid vote response
		t.Error(fmt.Sprintf("Invalid vote response\n"))
	}
	
	//if candidate Term is less
	Sm = &StateMachine{Id:1, LeaderId:2, ServerIds:[]int{1,2,3,4,5,6}, State:follower, Term:2, VoteCount:0, NegVoteCount:0, VotedFor:2, Log:[]LogEntry{{0,nil,1},{1,nil,1}}, LastLogIndex:1, CommitIndex:1, NextIndex:nil, MatchIndex:nil}
	Sm_expectedState =  &StateMachine{Id:1, LeaderId:2, ServerIds:[]int{1,2,3,4,5,6}, State:follower, Term:2, VoteCount:0, NegVoteCount:0, VotedFor:2, Log:[]LogEntry{{0,nil,1},{1,nil,1}}, LastLogIndex:1, CommitIndex:1, NextIndex:nil, MatchIndex:nil}
	Sm.handleVoteReqEvent(VoteReq{1,1,1,1})
	expect(t, Sm, Sm_expectedState)
	//if no vote response
	if len(Sm.SendEvents) != 1 {
		t.Error(fmt.Sprintf("Expected Send Events Count: %v, found : %v\n", 1, len(Sm.SendEvents))) 
	} else if Sm.SendEvents[0].Event.(VoteResp).VoteGranted!=false || Sm.SendEvents[0].Event.(VoteResp).FromId!=1 || Sm.SendEvents[0].Event.(VoteResp).Term!=2 {
		// Invalid vote response
		t.Error(fmt.Sprintf("Invalid vote response\n"))
	}
	
	//if Term Same, VotedFor not -1
	Sm = &StateMachine{Id:1, LeaderId:2, ServerIds:[]int{1,2,3,4,5,6}, State:follower, Term:2, VoteCount:0, NegVoteCount:0, VotedFor:2, Log:[]LogEntry{{0,nil,1},{1,nil,1}}, LastLogIndex:1, CommitIndex:1, NextIndex:nil, MatchIndex:nil}
	Sm_expectedState =  &StateMachine{Id:1, LeaderId:2, ServerIds:[]int{1,2,3,4,5,6}, State:follower, Term:2, VoteCount:0, NegVoteCount:0, VotedFor:2, Log:[]LogEntry{{0,nil,1},{1,nil,1}}, LastLogIndex:1, CommitIndex:1, NextIndex:nil, MatchIndex:nil}
	Sm.handleVoteReqEvent(VoteReq{2,2,1,1})
	expect(t, Sm, Sm_expectedState)
	//if no vote response
	if len(Sm.SendEvents) != 1 {
		t.Error(fmt.Sprintf("Expected Send Events Count: %v, found : %v\n", 1, len(Sm.SendEvents))) 
	} else if Sm.SendEvents[0].Event.(VoteResp).VoteGranted!=false || Sm.SendEvents[0].Event.(VoteResp).FromId!=1 || Sm.SendEvents[0].Event.(VoteResp).Term!=2 {
		// Invalid vote response
		t.Error(fmt.Sprintf("Invalid vote response\n"))
	}
	
}  

//_____________________________________________________________________________________________________________________

func TestCaseFollowerVoteRespEvent(t *testing.T) {
	//Event Term greater
	Sm = &StateMachine{Id:1, LeaderId:2, ServerIds:[]int{1,2,3,4,5,6}, State:follower, Term:1, VoteCount:0, NegVoteCount:0, VotedFor:2, Log:[]LogEntry{{0,nil,1},{1,nil,1}}, LastLogIndex:1, CommitIndex:1, NextIndex:nil, MatchIndex:nil}
	//vote granted; Term,VotedFor changed
	Sm_expectedState =  &StateMachine{Id:1, LeaderId:2, ServerIds:[]int{1,2,3,4,5,6}, State:follower, Term:3, VoteCount:0, NegVoteCount:0, VotedFor:-1, Log:[]LogEntry{{0,nil,1},{1,nil,1}}, LastLogIndex:1, CommitIndex:1, NextIndex:nil, MatchIndex:nil}
	Sm.handleVoteRespEvent(VoteResp{2,3,false})
	expect(t, Sm, Sm_expectedState)
	//if no vote response
	if len(Sm.SendEvents) != 0 {
		t.Error(fmt.Sprintf("Expected Send Events Count: %v, found : %v\n", 0, len(Sm.SendEvents))) 
	}
	
	//Event Term not greater
	Sm = &StateMachine{Id:1, LeaderId:2, ServerIds:[]int{1,2,3,4,5,6}, State:follower, Term:1, VoteCount:0, NegVoteCount:0, VotedFor:2, Log:[]LogEntry{{0,nil,1},{1,nil,1}}, LastLogIndex:1, CommitIndex:1, NextIndex:nil, MatchIndex:nil}
	//vote granted; Term,VotedFor changed
	Sm_expectedState =  &StateMachine{Id:1, LeaderId:2, ServerIds:[]int{1,2,3,4,5,6}, State:follower, Term:1, VoteCount:0, NegVoteCount:0, VotedFor:2, Log:[]LogEntry{{0,nil,1},{1,nil,1}}, LastLogIndex:1, CommitIndex:1, NextIndex:nil, MatchIndex:nil}
	Sm.handleVoteRespEvent(VoteResp{2,1,false})
	expect(t, Sm, Sm_expectedState)
	//if no vote response
	if len(Sm.SendEvents) != 0 {
		t.Error(fmt.Sprintf("Expected Send Events Count: %v, found : %v\n", 0, len(Sm.SendEvents))) 
	}
}  

//_____________________________________________________________________________________________________________________

func TestCaseFollowerAppEntReqEvent(t *testing.T) {
	//Event Term greater and prev Log entries don't match
	Sm = &StateMachine{Id:1, LeaderId:2, ServerIds:[]int{1,2,3,4,5,6}, State:follower, Term:1, VoteCount:0, NegVoteCount:0, VotedFor:2, Log:[]LogEntry{{0,nil,1}}, LastLogIndex:0, CommitIndex:0, NextIndex:nil, MatchIndex:nil}
	Sm_expectedState =  &StateMachine{Id:1, LeaderId:3, ServerIds:[]int{1,2,3,4,5,6}, State:follower, Term:2, VoteCount:0, NegVoteCount:0, VotedFor:-1, Log:[]LogEntry{{0,nil,1}}, LastLogIndex:0, CommitIndex:0, NextIndex:nil, MatchIndex:nil}
	Sm.handleAppendEntriesReq(AppendEntriesReq{2, 3, 1, 1,[]LogEntry{{2,nil,1}}, 1})
	expect(t, Sm, Sm_expectedState)
	//if no vote response
	if len(Sm.SendEvents) != 1 {
		t.Error(fmt.Sprintf("Expected Send Events Count: %v, found : %v\n", 1, len(Sm.SendEvents))) 
	} else if Sm.SendEvents[0].Event.(AppendEntriesResp).Success!=false || Sm.SendEvents[0].Event.(AppendEntriesResp).FromId!=1 || Sm.SendEvents[0].Event.(AppendEntriesResp).Term!=2 {
		// Invalid vote response
		t.Error(fmt.Sprintf("Invalid append entries response\n"))
	}
	
	//prev Log entries match
	Sm = &StateMachine{Id:1, LeaderId:2, ServerIds:[]int{1,2,3,4,5,6}, State:follower, Term:1, VoteCount:0, NegVoteCount:0, VotedFor:2, Log:[]LogEntry{{0,nil,1},{1,nil,1}}, LastLogIndex:1, CommitIndex:0, NextIndex:nil, MatchIndex:nil}
	Sm_expectedState =  &StateMachine{Id:1, LeaderId:2, ServerIds:[]int{1,2,3,4,5,6}, State:follower, Term:2, VoteCount:0, NegVoteCount:0, VotedFor:-1, Log:[]LogEntry{{0,nil,1},{2,nil,1}}, LastLogIndex:1, CommitIndex:1, NextIndex:nil, MatchIndex:nil}
	Sm.handleAppendEntriesReq(AppendEntriesReq{2, 2, 0, 0,[]LogEntry{{2,nil,1}}, 1})
	expect(t, Sm, Sm_expectedState)
	//if no vote response
	if len(Sm.SendEvents) != 1 {
		t.Error(fmt.Sprintf("Expected Send Events Count: %v, found : %v\n", 1, len(Sm.SendEvents))) 
	} else if Sm.SendEvents[0].Event.(AppendEntriesResp).Success!=true || Sm.SendEvents[0].Event.(AppendEntriesResp).FromId!=1 || Sm.SendEvents[0].Event.(AppendEntriesResp).Term!=2 {
		// Invalid vote response
		t.Error(fmt.Sprintf("Invalid append entries response\n"))
	}
	
	//Event Term Smaller
	Sm = &StateMachine{Id:1, LeaderId:2, ServerIds:[]int{1,2,3,4,5,6}, State:follower, Term:2, VoteCount:0, NegVoteCount:0, VotedFor:2, Log:[]LogEntry{{0,nil,1}}, LastLogIndex:0, CommitIndex:0, NextIndex:nil, MatchIndex:nil}
	Sm_expectedState =  &StateMachine{Id:1, LeaderId:2, ServerIds:[]int{1,2,3,4,5,6}, State:follower, Term:2, VoteCount:0, NegVoteCount:0, VotedFor:2, Log:[]LogEntry{{0,nil,1}}, LastLogIndex:0, CommitIndex:0, NextIndex:nil, MatchIndex:nil}
	Sm.handleAppendEntriesReq(AppendEntriesReq{1, 3, 0, 0,[]LogEntry{{1,nil,1}}, 1})
	expect(t, Sm, Sm_expectedState)
	//if no vote response
	if len(Sm.SendEvents) != 1 {
		t.Error(fmt.Sprintf("Expected Send Events Count: %v, found : %v\n", 1, len(Sm.SendEvents))) 
	} else if Sm.SendEvents[0].Event.(AppendEntriesResp).Success!=false || Sm.SendEvents[0].Event.(AppendEntriesResp).FromId!=1 || Sm.SendEvents[0].Event.(AppendEntriesResp).Term!=2 {
		// Invalid vote response
		t.Error(fmt.Sprintf("Invalid append entries response\n"))
	}
	
}  

//_____________________________________________________________________________________________________________________

func TestCaseFollowerAppEntRespEvent(t *testing.T) {
	//Event Term greater
	Sm = &StateMachine{Id:1, LeaderId:2, ServerIds:[]int{1,2,3,4,5,6}, State:follower, Term:1, VoteCount:0, NegVoteCount:0, VotedFor:2, Log:[]LogEntry{{0,nil,1},{1,nil,1}}, LastLogIndex:1, CommitIndex:1, NextIndex:nil, MatchIndex:nil}
	//vote granted; Term,VotedFor changed
	Sm_expectedState =  &StateMachine{Id:1, LeaderId:2, ServerIds:[]int{1,2,3,4,5,6}, State:follower, Term:3, VoteCount:0, NegVoteCount:0, VotedFor:-1, Log:[]LogEntry{{0,nil,1},{1,nil,1}}, LastLogIndex:1, CommitIndex:1, NextIndex:nil, MatchIndex:nil}
	Sm.handleAppendEntriesResp(AppendEntriesResp{AppendEntriesReq{},2,3,false})
	expect(t, Sm, Sm_expectedState)
	//if no vote response
	if len(Sm.SendEvents) != 0 {
		t.Error(fmt.Sprintf("Expected Send Events Count: %v, found : %v\n", 0, len(Sm.SendEvents))) 
	}
	
	//Event Term not greater
	Sm = &StateMachine{Id:1, LeaderId:2, ServerIds:[]int{1,2,3,4,5,6}, State:follower, Term:1, VoteCount:0, NegVoteCount:0, VotedFor:2, Log:[]LogEntry{{0,nil,1},{1,nil,1}}, LastLogIndex:1, CommitIndex:1, NextIndex:nil, MatchIndex:nil}
	//vote granted; Term,VotedFor changed
	Sm_expectedState =  &StateMachine{Id:1, LeaderId:2, ServerIds:[]int{1,2,3,4,5,6}, State:follower, Term:1, VoteCount:0, NegVoteCount:0, VotedFor:2, Log:[]LogEntry{{0,nil,1},{1,nil,1}}, LastLogIndex:1, CommitIndex:1, NextIndex:nil, MatchIndex:nil}
	Sm.handleAppendEntriesResp(AppendEntriesResp{AppendEntriesReq{},2,1,false})
	expect(t, Sm, Sm_expectedState)
	//if no vote response
	if len(Sm.SendEvents) != 0 {
		t.Error(fmt.Sprintf("Expected Send Events Count: %v, found : %v\n", 0, len(Sm.SendEvents))) 
	}
}  


//________________________________________ Candidate Tests____________________________________________________________


func TestCaseCandidateAppendEvent(t *testing.T) {
	Sm = &StateMachine{Id:1, LeaderId:-1, ServerIds:[]int{1,2,3,4,5,6}, State:candidate, Term:1, VoteCount:0, NegVoteCount:0, VotedFor:1, Log:[]LogEntry{{0,nil,1},{1,nil,1}}, LastLogIndex:1, CommitIndex:1, NextIndex:nil, MatchIndex:nil}
	Sm_expectedState =  &StateMachine{Id:1, LeaderId:-1, ServerIds:[]int{1,2,3,4,5,6}, State:candidate, Term:1, VoteCount:0, NegVoteCount:0, VotedFor:1, Log:[]LogEntry{{0,nil,1},{1,nil,1}}, LastLogIndex:1, CommitIndex:1, NextIndex:nil, MatchIndex:nil}
	// check request ignore
	Sm.handleAppendEvent(Append{[]byte("Rahul")})
	expect(t, Sm, Sm_expectedState)
} 

//_____________________________________________________________________________________________________________________

func TestCaseCandidateTimeoutEvent(t *testing.T) {
	Sm = &StateMachine{Id:1, LeaderId:-1, ServerIds:[]int{1,2,3,4,5,6}, State:candidate, Term:2, VoteCount:0, NegVoteCount:0, VotedFor:1, Log:[]LogEntry{{0,nil,1},{1,nil,1}}, LastLogIndex:1, CommitIndex:1, NextIndex:nil, MatchIndex:nil}
	Sm_expectedState =  &StateMachine{Id:1, LeaderId:-1, ServerIds:[]int{1,2,3,4,5,6}, State:candidate, Term:3, VoteCount:0, NegVoteCount:0, VotedFor:1, Log:[]LogEntry{{0,nil,1},{1,nil,1}}, LastLogIndex:1, CommitIndex:1, NextIndex:nil, MatchIndex:nil}
	Sm.handleTimeoutEvent(Timeout{})
	// election timeout, change to candidate(restart as candidate), increment Term, vote for Self
	expect(t, Sm, Sm_expectedState)
	if len(Sm.SendEvents) != 5 {
		t.Error(fmt.Sprintf("Expected Send Events Count: %v, found : %v\n", 5, len(Sm.SendEvents))) 
	}
}  

//_____________________________________________________________________________________________________________________

func TestCaseCandidateVoteReqEvent(t *testing.T) {
	//Event Term greater
	Sm = &StateMachine{Id:1, LeaderId:-1, ServerIds:[]int{1,2,3,4,5,6}, State:candidate, Term:2, VoteCount:0, NegVoteCount:0, VotedFor:2, Log:[]LogEntry{{0,nil,1},{1,nil,1}}, LastLogIndex:1, CommitIndex:1, NextIndex:nil, MatchIndex:nil}
	//vote granted; Term,VotedFor changed
	Sm_expectedState =  &StateMachine{Id:1, LeaderId:-1, ServerIds:[]int{1,2,3,4,5,6}, State:follower, Term:3, VoteCount:0, NegVoteCount:0, VotedFor:3, Log:[]LogEntry{{0,nil,1},{1,nil,1}}, LastLogIndex:1, CommitIndex:1, NextIndex:nil, MatchIndex:nil}
	Sm.handleVoteReqEvent(VoteReq{3,3,1,1})
	expect(t, Sm, Sm_expectedState)
	//if no vote response
	if len(Sm.SendEvents) != 1 {
		t.Error(fmt.Sprintf("Expected Send Events Count: %v, found : %v\n", 1, len(Sm.SendEvents))) 
	} else if Sm.SendEvents[0].Event.(VoteResp).VoteGranted!=true || Sm.SendEvents[0].Event.(VoteResp).FromId!=1 || Sm.SendEvents[0].Event.(VoteResp).Term!=3 {
		// Invalid vote response
		t.Error(fmt.Sprintf("Invalid vote response\n"))
	}
	//candidate not as updated, vote not granted
	Sm = &StateMachine{Id:1, LeaderId:-1, ServerIds:[]int{1,2,3,4,5,6}, State:candidate, Term:2, VoteCount:0, NegVoteCount:0, VotedFor:1, Log:[]LogEntry{{0,nil,1},{1,nil,1}}, LastLogIndex:1, CommitIndex:1, NextIndex:nil, MatchIndex:nil}
	Sm_expectedState =  &StateMachine{Id:1, LeaderId:-1, ServerIds:[]int{1,2,3,4,5,6}, State:follower, Term:3, VoteCount:0, NegVoteCount:0, VotedFor:-1, Log:[]LogEntry{{0,nil,1},{1,nil,1}}, LastLogIndex:1, CommitIndex:1, NextIndex:nil, MatchIndex:nil}
	Sm.handleVoteReqEvent(VoteReq{3,3,0,0})
	expect(t, Sm, Sm_expectedState)
	//if no vote response
	if len(Sm.SendEvents) != 1 {
		t.Error(fmt.Sprintf("Expected Send Events Count: %v, found : %v\n", 1, len(Sm.SendEvents))) 
	} else if Sm.SendEvents[0].Event.(VoteResp).VoteGranted!=false || Sm.SendEvents[0].Event.(VoteResp).FromId!=1 || Sm.SendEvents[0].Event.(VoteResp).Term!=3 {
		// Invalid vote response
		t.Error(fmt.Sprintf("Invalid vote response\n"))
	}
	
	//Event Term equal
	Sm = &StateMachine{Id:1, LeaderId:-1, ServerIds:[]int{1,2,3,4,5,6}, State:candidate, Term:2, VoteCount:0, NegVoteCount:0, VotedFor:1, Log:[]LogEntry{{0,nil,1},{1,nil,1}}, LastLogIndex:1, CommitIndex:1, NextIndex:nil, MatchIndex:nil}
	Sm_expectedState =  &StateMachine{Id:1, LeaderId:-1, ServerIds:[]int{1,2,3,4,5,6}, State:candidate, Term:2, VoteCount:0, NegVoteCount:0, VotedFor:1, Log:[]LogEntry{{0,nil,1},{1,nil,1}}, LastLogIndex:1, CommitIndex:1, NextIndex:nil, MatchIndex:nil}
	Sm.handleVoteReqEvent(VoteReq{2,3,0,0})
	expect(t, Sm, Sm_expectedState)
	//if no vote response
	if len(Sm.SendEvents) != 1 {
		t.Error(fmt.Sprintf("Expected Send Events Count: %v, found : %v\n", 1, len(Sm.SendEvents))) 
	} else if Sm.SendEvents[0].Event.(VoteResp).VoteGranted!=false || Sm.SendEvents[0].Event.(VoteResp).FromId!=1 || Sm.SendEvents[0].Event.(VoteResp).Term!=2 {
		// Invalid vote response
		t.Error(fmt.Sprintf("Invalid vote response\n"))
	}
	
	//if candidate Term is less
	Sm = &StateMachine{Id:1, LeaderId:-1, ServerIds:[]int{1,2,3,4,5,6}, State:candidate, Term:2, VoteCount:0, NegVoteCount:0, VotedFor:1, Log:[]LogEntry{{0,nil,1},{1,nil,1}}, LastLogIndex:1, CommitIndex:1, NextIndex:nil, MatchIndex:nil}
	Sm_expectedState =  &StateMachine{Id:1, LeaderId:-1, ServerIds:[]int{1,2,3,4,5,6}, State:candidate, Term:2, VoteCount:0, NegVoteCount:0, VotedFor:1, Log:[]LogEntry{{0,nil,1},{1,nil,1}}, LastLogIndex:1, CommitIndex:1, NextIndex:nil, MatchIndex:nil}
	Sm.handleVoteReqEvent(VoteReq{1,1,1,1})
	expect(t, Sm, Sm_expectedState)
	//if no vote response
	if len(Sm.SendEvents) != 1 {
		t.Error(fmt.Sprintf("Expected Send Events Count: %v, found : %v\n", 1, len(Sm.SendEvents))) 
	} else if Sm.SendEvents[0].Event.(VoteResp).VoteGranted!=false || Sm.SendEvents[0].Event.(VoteResp).FromId!=1 || Sm.SendEvents[0].Event.(VoteResp).Term!=2 {
		// Invalid vote response
		t.Error(fmt.Sprintf("Invalid vote response\n"))
	}
} 

//_____________________________________________________________________________________________________________________

func TestCaseCandidateVoteRespEvent(t *testing.T) {
	
	//if VoteResp Term is less, ignore it
	Sm = &StateMachine{Id:1, LeaderId:-1, ServerIds:[]int{1,2,3,4,5,6}, State:candidate, Term:2, VoteCount:0, NegVoteCount:0, VotedFor:1, Log:[]LogEntry{{0,nil,1},{1,nil,1}}, LastLogIndex:1, CommitIndex:1, NextIndex:nil, MatchIndex:nil}
	Sm_expectedState =  &StateMachine{Id:1, LeaderId:-1, ServerIds:[]int{1,2,3,4,5,6}, State:candidate, Term:2, VoteCount:0, NegVoteCount:0, VotedFor:1, Log:[]LogEntry{{0,nil,1},{1,nil,1}}, LastLogIndex:1, CommitIndex:1, NextIndex:nil, MatchIndex:nil}
	Sm.handleVoteRespEvent(VoteResp{2,1,false})
	expect(t, Sm, Sm_expectedState)
	//if no vote response
	if len(Sm.SendEvents) != 0 {
		t.Error(fmt.Sprintf("Expected Send Events Count: %v, found : %v\n", 0, len(Sm.SendEvents))) 
	}
	
	//if candidate Term is less
	Sm = &StateMachine{Id:1, LeaderId:-1, ServerIds:[]int{1,2,3,4,5,6}, State:candidate, Term:2, VoteCount:0, NegVoteCount:0, VotedFor:1, Log:[]LogEntry{{0,nil,1},{1,nil,1}}, LastLogIndex:1, CommitIndex:1, NextIndex:nil, MatchIndex:nil}
	Sm_expectedState =  &StateMachine{Id:1, LeaderId:-1, ServerIds:[]int{1,2,3,4,5,6}, State:follower, Term:3, VoteCount:0, NegVoteCount:0, VotedFor:-1, Log:[]LogEntry{{0,nil,1},{1,nil,1}}, LastLogIndex:1, CommitIndex:1, NextIndex:nil, MatchIndex:nil}
	Sm.handleVoteRespEvent(VoteResp{2,3,false})
	expect(t, Sm, Sm_expectedState)
	//if no vote response
	if len(Sm.SendEvents) != 0 {
		t.Error(fmt.Sprintf("Expected Send Events Count: %v, found : %v\n", 0, len(Sm.SendEvents))) 
	}
	
	//if VoteGranted true, quorum not reached
	Sm = &StateMachine{Id:1, LeaderId:-1, ServerIds:[]int{1,2,3,4,5,6}, State:candidate, Term:2, VoteCount:0, NegVoteCount:0, VotedFor:1,Log:[]LogEntry{{0,nil,1},{1,nil,1}}, LastLogIndex:1, CommitIndex:1, NextIndex:nil, MatchIndex:nil}
	Sm_expectedState =  &StateMachine{Id:1, LeaderId:-1, ServerIds:[]int{1,2,3,4,5,6}, State:candidate, Term:2, VoteCount:1, NegVoteCount:0, VotedFor:1, Log:[]LogEntry{{0,nil,1},{1,nil,1}}, LastLogIndex:1, CommitIndex:1, NextIndex:nil, MatchIndex:nil}
	Sm.handleVoteRespEvent(VoteResp{2,2,true})
	expect(t, Sm, Sm_expectedState)
	//if no vote response
	if len(Sm.SendEvents) != 0 {
		t.Error(fmt.Sprintf("Expected Send Events Count: %v, found : %v\n", 1, len(Sm.SendEvents))) 
	}
	
	//if VoteGranted true, quorum reached, becomes leader
	Sm = &StateMachine{Id:1, LeaderId:-1, ServerIds:[]int{1,2,3,4,5,6}, State:candidate, Term:2, VoteCount:3, NegVoteCount:0, VotedFor:1, Log:[]LogEntry{{0,nil,1},{1,nil,1}}, LastLogIndex:1, CommitIndex:1, NextIndex:nil, MatchIndex:nil}
	Sm_expectedState =  &StateMachine{Id:1, LeaderId:1, ServerIds:[]int{1,2,3,4,5,6}, State:leader, Term:2, VoteCount:0, NegVoteCount:0, VotedFor:1, Log:[]LogEntry{{0,nil,1},{1,nil,1},{2,nil,1}}, LastLogIndex:2, CommitIndex:1, NextIndex:map[int]int{2:2, 3:2, 4:2, 5:2, 6:2}, MatchIndex:map[int]int{2:0, 3:0, 4:0, 5:0, 6:0}}
	Sm.handleVoteRespEvent(VoteResp{2,2,true})
	expect(t, Sm, Sm_expectedState)
	//if no vote response
	if len(Sm.SendEvents) != 5 {
		t.Error(fmt.Sprintf("Expected Send Events Count: %v, found : %v\n", 5, len(Sm.SendEvents))) 
	}
	
	//if VoteGranted false, -ve quorum not reached
	Sm = &StateMachine{Id:1, LeaderId:-1, ServerIds:[]int{1,2,3,4,5,6}, State:candidate, Term:2, VoteCount:0, NegVoteCount:0, VotedFor:1, Log:[]LogEntry{{0,nil,1},{1,nil,1}}, LastLogIndex:1, CommitIndex:1, NextIndex:nil, MatchIndex:nil}
	Sm_expectedState =  &StateMachine{Id:1, LeaderId:-1, ServerIds:[]int{1,2,3,4,5,6}, State:candidate, Term:2, VoteCount:0, NegVoteCount:1, VotedFor:1, Log:[]LogEntry{{0,nil,1},{1,nil,1}}, LastLogIndex:1, CommitIndex:1, NextIndex:nil, MatchIndex:nil}
	Sm.handleVoteRespEvent(VoteResp{2,2,false})
	expect(t, Sm, Sm_expectedState)
	//if no vote response
	if len(Sm.SendEvents) != 0 {
		t.Error(fmt.Sprintf("Expected Send Events Count: %v, found : %v\n", 0, len(Sm.SendEvents))) 
	}
	
	//if VoteGranted true, -ve quorum reached, Step down to follower
	Sm = &StateMachine{Id:1, LeaderId:-1, ServerIds:[]int{1,2,3,4,5,6}, State:candidate, Term:2, VoteCount:3, NegVoteCount:3, VotedFor:1, Log:[]LogEntry{{0,nil,1},{1,nil,1}}, LastLogIndex:1, CommitIndex:1, NextIndex:nil, MatchIndex:nil}
	Sm_expectedState =  &StateMachine{Id:1, LeaderId:-1, ServerIds:[]int{1,2,3,4,5,6}, State:follower, Term:2, VoteCount:0, NegVoteCount:0, VotedFor:1, Log:[]LogEntry{{0,nil,1},{1,nil,1}}, LastLogIndex:1, CommitIndex:1, NextIndex:nil, MatchIndex:nil}
	Sm.handleVoteRespEvent(VoteResp{2,2,false})
	expect(t, Sm, Sm_expectedState)
	//if no vote response
	if len(Sm.SendEvents) != 0 {
		t.Error(fmt.Sprintf("Expected Send Events Count: %v, found : %v\n", 0, len(Sm.SendEvents))) 
	}
}  

//_____________________________________________________________________________________________________________________

func TestCaseCandidateAppEntReqEvent(t *testing.T) {
	//Event Term greater and prev Log entries don't match
	Sm = &StateMachine{Id:1, LeaderId:-1, ServerIds:[]int{1,2,3,4,5,6}, State:candidate, Term:1, VoteCount:0, NegVoteCount:0, VotedFor:1, Log:[]LogEntry{{0,nil,1}}, LastLogIndex:0, CommitIndex:0, NextIndex:nil, MatchIndex:nil}
	Sm_expectedState =  &StateMachine{Id:1, LeaderId:2, ServerIds:[]int{1,2,3,4,5,6}, State:follower, Term:2, VoteCount:0, NegVoteCount:0, VotedFor:-1, Log:[]LogEntry{{0,nil,1}}, LastLogIndex:0, CommitIndex:0, NextIndex:nil, MatchIndex:nil}
	Sm.handleAppendEntriesReq(AppendEntriesReq{2, 2, 1, 1,[]LogEntry{{2,nil,1}}, 1})
	expect(t, Sm, Sm_expectedState)
	//if no vote response
	if len(Sm.SendEvents) != 1 {
		t.Error(fmt.Sprintf("Expected Send Events Count: %v, found : %v\n", 1, len(Sm.SendEvents))) 
	} else if Sm.SendEvents[0].Event.(AppendEntriesResp).Success!=false || Sm.SendEvents[0].Event.(AppendEntriesResp).FromId!=1 || Sm.SendEvents[0].Event.(AppendEntriesResp).Term!=2 {
		// Invalid vote response
		t.Error(fmt.Sprintf("Invalid append entries response\n"))
	}
	
	//prev Log entries match
	Sm = &StateMachine{Id:1, LeaderId:-1, ServerIds:[]int{1,2,3,4,5,6}, State:candidate, Term:1, VoteCount:0, NegVoteCount:0, VotedFor:1, Log:[]LogEntry{{0,nil,1},{1,nil,1}}, LastLogIndex:1, CommitIndex:0, NextIndex:nil, MatchIndex:nil}
	Sm_expectedState =  &StateMachine{Id:1, LeaderId:2, ServerIds:[]int{1,2,3,4,5,6}, State:follower, Term:2, VoteCount:0, NegVoteCount:0, VotedFor:-1, Log:[]LogEntry{{0,nil,1},{2,nil,1}}, LastLogIndex:1, CommitIndex:1, NextIndex:nil, MatchIndex:nil}
	Sm.handleAppendEntriesReq(AppendEntriesReq{2, 2, 0, 0,[]LogEntry{{2,nil,1}}, 1})
	expect(t, Sm, Sm_expectedState)
	//if no vote response
	if len(Sm.SendEvents) != 1 {
		t.Error(fmt.Sprintf("Expected Send Events Count: %v, found : %v\n", 1, len(Sm.SendEvents))) 
	} else if Sm.SendEvents[0].Event.(AppendEntriesResp).Success!=true || Sm.SendEvents[0].Event.(AppendEntriesResp).FromId!=1 || Sm.SendEvents[0].Event.(AppendEntriesResp).Term!=2 {
		// Invalid vote response
		t.Error(fmt.Sprintf("Invalid append entries response\n"))
	}
	
	//Event Term Smaller
	Sm = &StateMachine{Id:1, LeaderId:-1, ServerIds:[]int{1,2,3,4,5,6}, State:candidate, Term:2, VoteCount:0, NegVoteCount:0, VotedFor:1, Log:[]LogEntry{{0,nil,1}}, LastLogIndex:0, CommitIndex:0, NextIndex:nil, MatchIndex:nil}
	Sm_expectedState =  &StateMachine{Id:1, LeaderId:-1, ServerIds:[]int{1,2,3,4,5,6}, State:candidate, Term:2, VoteCount:0, NegVoteCount:0, VotedFor:1, Log:[]LogEntry{{0,nil,1}}, LastLogIndex:0, CommitIndex:0, NextIndex:nil, MatchIndex:nil}
	Sm.handleAppendEntriesReq(AppendEntriesReq{1, 3, 0, 0,[]LogEntry{{1,nil,1}}, 1})
	expect(t, Sm, Sm_expectedState)
	//if no vote response
	if len(Sm.SendEvents) != 1 {
		t.Error(fmt.Sprintf("Expected Send Events Count: %v, found : %v\n", 1, len(Sm.SendEvents))) 
	} else if Sm.SendEvents[0].Event.(AppendEntriesResp).Success!=false || Sm.SendEvents[0].Event.(AppendEntriesResp).FromId!=1 || Sm.SendEvents[0].Event.(AppendEntriesResp).Term!=2 {
		// Invalid vote response
		t.Error(fmt.Sprintf("Invalid append entries response\n"))
	}
}  

//_____________________________________________________________________________________________________________________

func TestCaseCandidateAppEntRespEvent(t *testing.T) {
	//Event Term greater
	Sm = &StateMachine{Id:1, LeaderId:-1, ServerIds:[]int{1,2,3,4,5,6}, State:candidate, Term:1, VoteCount:0, NegVoteCount:0, VotedFor:1, Log:[]LogEntry{{0,nil,1},{1,nil,1}}, LastLogIndex:1, CommitIndex:1, NextIndex:nil, MatchIndex:nil}
	//vote granted; Term,VotedFor changed
	Sm_expectedState =  &StateMachine{Id:1, LeaderId:-1, ServerIds:[]int{1,2,3,4,5,6}, State:follower, Term:3, VoteCount:0, NegVoteCount:0, VotedFor:-1, Log:[]LogEntry{{0,nil,1},{1,nil,1}}, LastLogIndex:1, CommitIndex:1, NextIndex:nil, MatchIndex:nil}
	Sm.handleAppendEntriesResp(AppendEntriesResp{AppendEntriesReq{},2,3,false})
	expect(t, Sm, Sm_expectedState)
	//if no vote response
	if len(Sm.SendEvents) != 0 {
		t.Error(fmt.Sprintf("Expected Send Events Count: %v, found : %v\n", 0, len(Sm.SendEvents))) 
	}
	
	//Event Term not greater
	Sm = &StateMachine{Id:1, LeaderId:-1, ServerIds:[]int{1,2,3,4,5,6}, State:candidate, Term:1, VoteCount:0, NegVoteCount:0, VotedFor:1, Log:[]LogEntry{{0,nil,1},{1,nil,1}}, LastLogIndex:1, CommitIndex:1, NextIndex:nil, MatchIndex:nil}
	//vote granted; Term,VotedFor changed
	Sm_expectedState =  &StateMachine{Id:1, LeaderId:-1, ServerIds:[]int{1,2,3,4,5,6}, State:candidate, Term:1, VoteCount:0, NegVoteCount:0, VotedFor:1, Log:[]LogEntry{{0,nil,1},{1,nil,1}}, LastLogIndex:1, CommitIndex:1, NextIndex:nil, MatchIndex:nil}
	Sm.handleAppendEntriesResp(AppendEntriesResp{AppendEntriesReq{},2,1,false})
	expect(t, Sm, Sm_expectedState)
	//if no vote response
	if len(Sm.SendEvents) != 0 {
		t.Error(fmt.Sprintf("Expected Send Events Count: %v, found : %v\n", 1, len(Sm.SendEvents))) 
	}
}   


//________________________________________ Leader Tests____________________________________________________________


func TestCaseLeaderAppendEvent(t *testing.T) {
	Sm = &StateMachine{Id:1, LeaderId:1, ServerIds:[]int{1,2,3,4,5,6}, State:leader, Term:1, VoteCount:0, NegVoteCount:0, VotedFor:1, Log:[]LogEntry{{0,nil,1},{1,nil,1}}, LastLogIndex:1, CommitIndex:1, NextIndex:map[int]int{2:2,3:2,4:2,5:2,6:2}, MatchIndex:map[int]int{2:1,3:1,4:1,5:1,6:1}}
	Sm_expectedState =  &StateMachine{Id:1, LeaderId:1, ServerIds:[]int{1,2,3,4,5,6}, State:leader, Term:1, VoteCount:0, NegVoteCount:0, VotedFor:1, Log:[]LogEntry{{0,nil,1},{1,nil,1},{1,[]byte("Rahul"),1}}, LastLogIndex:2, CommitIndex:1, NextIndex:map[int]int{2:2,3:2,4:2,5:2,6:2}, MatchIndex:map[int]int{2:1,3:1,4:1,5:1,6:1}}
	// append to Log and check whether append entries requests Sent
	Sm.handleAppendEvent(Append{[]byte("Rahul")})
	expect(t, Sm, Sm_expectedState)
}  

//_____________________________________________________________________________________________________________________

func TestCaseLeaderTimeoutEvent(t *testing.T) {
	Sm = &StateMachine{Id:1, LeaderId:1, ServerIds:[]int{1,2,3,4,5,6}, State:leader, Term:1, VoteCount:0, NegVoteCount:0, VotedFor:1, Log:[]LogEntry{{0,nil,1},{1,nil,1},{1,[]byte("Rahul"),1}}, LastLogIndex:2, CommitIndex:1, NextIndex:map[int]int{2:2,3:2,4:2,5:2,6:2}, MatchIndex:map[int]int{2:1,3:1,4:1,5:1,6:1}}
	Sm_expectedState =  &StateMachine{Id:1, LeaderId:1, ServerIds:[]int{1,2,3,4,5,6}, State:leader, Term:1, VoteCount:0, NegVoteCount:0, VotedFor:1, Log:[]LogEntry{{0,nil,1},{1,nil,1},{1,[]byte("Rahul"),1},{1,nil,1}}, LastLogIndex:3, CommitIndex:1, NextIndex:map[int]int{2:2,3:2,4:2,5:2,6:2}, MatchIndex:map[int]int{2:1,3:1,4:1,5:1,6:1}}
	// check whether heart beats Sent
	Sm.handleTimeoutEvent(Timeout{})
	expect(t, Sm, Sm_expectedState)
	if len(Sm.SendEvents) != 5 {
		t.Error(fmt.Sprintf("Expected Send Events Count: %v, found : %v\n", 5, len(Sm.SendEvents))) 
	}
}  

//_____________________________________________________________________________________________________________________

func TestCaseLeaderVoteReqEvent(t *testing.T) {
	//Event Term greater
	Sm = &StateMachine{Id:1, LeaderId:1, ServerIds:[]int{1,2,3,4,5,6}, State:leader, Term:2, VoteCount:0, NegVoteCount:0, VotedFor:1, Log:[]LogEntry{{0,nil,1},{1,nil,1}}, LastLogIndex:1, CommitIndex:1, NextIndex:map[int]int{2:2,3:2,4:2,5:2,6:2}, MatchIndex:map[int]int{2:1,3:1,4:1,5:1,6:1}}
	//vote granted; Term,VotedFor changed
	Sm_expectedState =  &StateMachine{Id:1, LeaderId:-1, ServerIds:[]int{1,2,3,4,5,6}, State:follower, Term:3, VoteCount:0, NegVoteCount:0, VotedFor:3, Log:[]LogEntry{{0,nil,1},{1,nil,1}}, LastLogIndex:1, CommitIndex:1, NextIndex:nil, MatchIndex:nil}
	Sm.handleVoteReqEvent(VoteReq{3,3,1,1})
	expect(t, Sm, Sm_expectedState)
	//if no vote response
	if len(Sm.SendEvents) != 1 {
		t.Error(fmt.Sprintf("Expected Send Events Count: %v, found : %v\n", 1, len(Sm.SendEvents))) 
	} else if Sm.SendEvents[0].Event.(VoteResp).VoteGranted!=true || Sm.SendEvents[0].Event.(VoteResp).FromId!=1 || Sm.SendEvents[0].Event.(VoteResp).Term!=3 {
		// Invalid vote response
		t.Error(fmt.Sprintf("Invalid vote response\n"))
	}
	//candidate not as updated, vote not granted
	Sm = &StateMachine{Id:1, LeaderId:1, ServerIds:[]int{1,2,3,4,5,6}, State:leader, Term:2, VoteCount:0, NegVoteCount:0, VotedFor:1, Log:[]LogEntry{{0,nil,1},{1,nil,1}}, LastLogIndex:1, CommitIndex:1, NextIndex:map[int]int{2:2,3:2,4:2,5:2,6:2}, MatchIndex:map[int]int{2:1,3:1,4:1,5:1,6:1}}
	Sm_expectedState =  &StateMachine{Id:1, LeaderId:-1, ServerIds:[]int{1,2,3,4,5,6}, State:follower, Term:3, VoteCount:0, NegVoteCount:0, VotedFor:-1, Log:[]LogEntry{{0,nil,1},{1,nil,1}}, LastLogIndex:1, CommitIndex:1, NextIndex:nil, MatchIndex:nil}
	Sm.handleVoteReqEvent(VoteReq{3,3,0,0})
	expect(t, Sm, Sm_expectedState)
	//if no vote response
	if len(Sm.SendEvents) != 1 {
		t.Error(fmt.Sprintf("Expected Send Events Count: %v, found : %v\n", 1, len(Sm.SendEvents))) 
	} else if Sm.SendEvents[0].Event.(VoteResp).VoteGranted!=false || Sm.SendEvents[0].Event.(VoteResp).FromId!=1 || Sm.SendEvents[0].Event.(VoteResp).Term!=3 {
		// Invalid vote response
		t.Error(fmt.Sprintf("Invalid vote response\n"))
	}
	
	//Event Term equal
	Sm = &StateMachine{Id:1, LeaderId:1, ServerIds:[]int{1,2,3,4,5,6}, State:leader, Term:2, VoteCount:0, NegVoteCount:0, VotedFor:1, Log:[]LogEntry{{0,nil,1},{1,nil,1}}, LastLogIndex:1, CommitIndex:1, NextIndex:map[int]int{2:2,3:2,4:2,5:2,6:2}, MatchIndex:map[int]int{2:1,3:1,4:1,5:1,6:1}}
	Sm_expectedState =  &StateMachine{Id:1, LeaderId:1, ServerIds:[]int{1,2,3,4,5,6}, State:leader, Term:2, VoteCount:0, NegVoteCount:0, VotedFor:1, Log:[]LogEntry{{0,nil,1},{1,nil,1}}, LastLogIndex:1, CommitIndex:1, NextIndex:map[int]int{2:2,3:2,4:2,5:2,6:2}, MatchIndex:map[int]int{2:1,3:1,4:1,5:1,6:1}}
	Sm.handleVoteReqEvent(VoteReq{2,3,0,0})
	expect(t, Sm, Sm_expectedState)
	//if no vote response
	if len(Sm.SendEvents) != 1 {
		t.Error(fmt.Sprintf("Expected Send Events Count: %v, found : %v\n", 1, len(Sm.SendEvents))) 
	} else if Sm.SendEvents[0].Event.(VoteResp).VoteGranted!=false || Sm.SendEvents[0].Event.(VoteResp).FromId!=1 || Sm.SendEvents[0].Event.(VoteResp).Term!=2 {
		// Invalid vote response
		t.Error(fmt.Sprintf("Invalid vote response\n"))
	}
	
	//if candidate Term is less
	Sm = &StateMachine{Id:1, LeaderId:1, ServerIds:[]int{1,2,3,4,5,6}, State:leader, Term:2, VoteCount:0, NegVoteCount:0, VotedFor:1, Log:[]LogEntry{{0,nil,1},{1,nil,1}}, LastLogIndex:1, CommitIndex:1, NextIndex:map[int]int{2:2,3:2,4:2,5:2,6:2}, MatchIndex:map[int]int{2:1,3:1,4:1,5:1,6:1}}
	Sm_expectedState =  &StateMachine{Id:1, LeaderId:1, ServerIds:[]int{1,2,3,4,5,6}, State:leader, Term:2, VoteCount:0, NegVoteCount:0, VotedFor:1, Log:[]LogEntry{{0,nil,1},{1,nil,1}}, LastLogIndex:1, CommitIndex:1, NextIndex:map[int]int{2:2,3:2,4:2,5:2,6:2}, MatchIndex:map[int]int{2:1,3:1,4:1,5:1,6:1}}
	Sm.handleVoteReqEvent(VoteReq{1,1,1,1})
	expect(t, Sm, Sm_expectedState)
	//if no vote response
	if len(Sm.SendEvents) != 1 {
		t.Error(fmt.Sprintf("Expected Send Events Count: %v, found : %v\n", 1, len(Sm.SendEvents))) 
	} else if Sm.SendEvents[0].Event.(VoteResp).VoteGranted!=false || Sm.SendEvents[0].Event.(VoteResp).FromId!=1 || Sm.SendEvents[0].Event.(VoteResp).Term!=2 {
		// Invalid vote response
		t.Error(fmt.Sprintf("Invalid vote response\n"))
	}
}   

//_____________________________________________________________________________________________________________________

func TestCaseLeaderVoteRespEvent(t *testing.T) {
	//Event Term greater
	Sm = &StateMachine{Id:1, LeaderId:2, ServerIds:[]int{1,2,3,4,5,6}, State:leader, Term:1, VoteCount:0, NegVoteCount:0, VotedFor:2, Log:[]LogEntry{{0,nil,1},{1,nil,1}}, LastLogIndex:1, CommitIndex:1, NextIndex:map[int]int{2:2,3:2,4:2,5:2,6:2}, MatchIndex:map[int]int{2:1,3:1,4:1,5:1,6:1}}
	//vote granted; Term,VotedFor changed
	Sm_expectedState =  &StateMachine{Id:1, LeaderId:-1, ServerIds:[]int{1,2,3,4,5,6}, State:follower, Term:3, VoteCount:0, NegVoteCount:0, VotedFor:-1, Log:[]LogEntry{{0,nil,1},{1,nil,1}}, LastLogIndex:1, CommitIndex:1, NextIndex:nil, MatchIndex:nil}
	Sm.handleVoteRespEvent(VoteResp{2,3,false})
	expect(t, Sm, Sm_expectedState)
	//if no vote response
	if len(Sm.SendEvents) != 0 {
		t.Error(fmt.Sprintf("Expected Send Events Count: %v, found : %v\n", 0, len(Sm.SendEvents))) 
	}
	
	//Event Term not greater
	Sm = &StateMachine{Id:1, LeaderId:1, ServerIds:[]int{1,2,3,4,5,6}, State:leader, Term:1, VoteCount:0, NegVoteCount:0, VotedFor:1, Log:[]LogEntry{{0,nil,1},{1,nil,1}}, LastLogIndex:1, CommitIndex:1, NextIndex:map[int]int{2:2,3:2,4:2,5:2,6:2}, MatchIndex:map[int]int{2:1,3:1,4:1,5:1,6:1}}
	//vote granted; Term,VotedFor changed
	Sm_expectedState =  &StateMachine{Id:1, LeaderId:1, ServerIds:[]int{1,2,3,4,5,6}, State:leader, Term:1, VoteCount:0, NegVoteCount:0, VotedFor:1, Log:[]LogEntry{{0,nil,1},{1,nil,1}}, LastLogIndex:1, CommitIndex:1, NextIndex:map[int]int{2:2,3:2,4:2,5:2,6:2}, MatchIndex:map[int]int{2:1,3:1,4:1,5:1,6:1}}
	Sm.handleVoteRespEvent(VoteResp{2,1,false})
	expect(t, Sm, Sm_expectedState)
	//if no vote response
	if len(Sm.SendEvents) != 0 {
		t.Error(fmt.Sprintf("Expected Send Events Count: %v, found : %v\n", 0, len(Sm.SendEvents))) 
	}
}  

//_____________________________________________________________________________________________________________________

func TestCaseLeaderAppEntReqEvent(t *testing.T) {
	//Event Term greater and prev Log entries don't match
	Sm = &StateMachine{Id:1, LeaderId:1, ServerIds:[]int{1,2,3,4,5,6}, State:leader, Term:1, VoteCount:0, NegVoteCount:0, VotedFor:1, Log:[]LogEntry{{0,nil,1}}, LastLogIndex:0, CommitIndex:0, NextIndex:map[int]int{2:2,3:2,4:2,5:2,6:2}, MatchIndex:map[int]int{2:1,3:1,4:1,5:1,6:1}}
	Sm_expectedState =  &StateMachine{Id:1, LeaderId:2, ServerIds:[]int{1,2,3,4,5,6}, State:follower, Term:2, VoteCount:0, NegVoteCount:0, VotedFor:-1, Log:[]LogEntry{{0,nil,1}}, LastLogIndex:0, CommitIndex:0, NextIndex:nil, MatchIndex:nil}
	Sm.handleAppendEntriesReq(AppendEntriesReq{2, 2, 1, 1,[]LogEntry{{2,nil,1}}, 1})
	expect(t, Sm, Sm_expectedState)
	//if no vote response
	if len(Sm.SendEvents) != 1 {
		t.Error(fmt.Sprintf("Expected Send Events Count: %v, found : %v\n", 1, len(Sm.SendEvents))) 
	} else if Sm.SendEvents[0].Event.(AppendEntriesResp).Success!=false || Sm.SendEvents[0].Event.(AppendEntriesResp).FromId!=1 || Sm.SendEvents[0].Event.(AppendEntriesResp).Term!=2 {
		// Invalid vote response
		t.Error(fmt.Sprintf("Invalid append entries response\n"))
	}
	
	//prev Log entries match
	Sm = &StateMachine{Id:1, LeaderId:1, ServerIds:[]int{1,2,3,4,5,6}, State:leader, Term:1, VoteCount:0, NegVoteCount:0, VotedFor:1, Log:[]LogEntry{{0,nil,1},{1,nil,1}}, LastLogIndex:1, CommitIndex:0, NextIndex:map[int]int{2:2,3:2,4:2,5:2,6:2}, MatchIndex:map[int]int{2:0,3:0,4:0,5:0,6:0}}
	Sm_expectedState =  &StateMachine{Id:1, LeaderId:2, ServerIds:[]int{1,2,3,4,5,6}, State:follower, Term:2, VoteCount:0, NegVoteCount:0, VotedFor:-1, Log:[]LogEntry{{0,nil,1},{2,nil,1}}, LastLogIndex:1, CommitIndex:1, NextIndex:nil, MatchIndex:nil}
	Sm.handleAppendEntriesReq(AppendEntriesReq{2, 2, 0, 0,[]LogEntry{{2,nil,1}}, 1})
	expect(t, Sm, Sm_expectedState)
	//if no vote response
	if len(Sm.SendEvents) != 1 {
		t.Error(fmt.Sprintf("Expected Send Events Count: %v, found : %v\n", 1, len(Sm.SendEvents))) 
	} else if Sm.SendEvents[0].Event.(AppendEntriesResp).Success!=true || Sm.SendEvents[0].Event.(AppendEntriesResp).FromId!=1 || Sm.SendEvents[0].Event.(AppendEntriesResp).Term!=2 {
		// Invalid vote response
		t.Error(fmt.Sprintf("Invalid append entries response\n"))
	}
	
	//Event Term Smaller
	Sm = &StateMachine{Id:1, LeaderId:1, ServerIds:[]int{1,2,3,4,5,6}, State:leader, Term:2, VoteCount:0, NegVoteCount:0, VotedFor:1, Log:[]LogEntry{{0,nil,1}}, LastLogIndex:0, CommitIndex:0, NextIndex:map[int]int{2:2,3:2,4:2,5:2,6:2}, MatchIndex:map[int]int{2:1,3:1,4:1,5:1,6:1}}
	Sm_expectedState =  &StateMachine{Id:1, LeaderId:1, ServerIds:[]int{1,2,3,4,5,6}, State:leader, Term:2, VoteCount:0, NegVoteCount:0, VotedFor:1, Log:[]LogEntry{{0,nil,1}}, LastLogIndex:0, CommitIndex:0, NextIndex:map[int]int{2:2,3:2,4:2,5:2,6:2}, MatchIndex:map[int]int{2:1,3:1,4:1,5:1,6:1}}
	Sm.handleAppendEntriesReq(AppendEntriesReq{1, 3, 0, 0,[]LogEntry{{1,nil,1}}, 1})
	expect(t, Sm, Sm_expectedState)
	//if no vote response
	if len(Sm.SendEvents) != 1 {
		t.Error(fmt.Sprintf("Expected Send Events Count: %v, found : %v\n", 1, len(Sm.SendEvents))) 
	} else if Sm.SendEvents[0].Event.(AppendEntriesResp).Success!=false || Sm.SendEvents[0].Event.(AppendEntriesResp).FromId!=1 || Sm.SendEvents[0].Event.(AppendEntriesResp).Term!=2 {
		// Invalid vote response
		t.Error(fmt.Sprintf("Invalid append entries response\n"))
	}
}   

//_____________________________________________________________________________________________________________________

func TestCaseLeaderAppEntRespEvent(t *testing.T) {
	
	//Event Term is less
	Sm = &StateMachine{Id:1, LeaderId:1, ServerIds:[]int{1,2,3,4,5,6}, State:leader, Term:2, VoteCount:0, NegVoteCount:0, VotedFor:1, Log:[]LogEntry{{0,nil,1},{1,nil,1},{2,nil,1}}, LastLogIndex:2, CommitIndex:1, NextIndex:map[int]int{2:2,3:2,4:2,5:2,6:2}, MatchIndex:map[int]int{2:1,3:1,4:1,5:1,6:1}}
	//vote granted; Term,VotedFor changed
	Sm_expectedState =  &StateMachine{Id:1, LeaderId:1, ServerIds:[]int{1,2,3,4,5,6}, State:leader, Term:2, VoteCount:0, NegVoteCount:0, VotedFor:1, Log:[]LogEntry{{0,nil,1},{1,nil,1},{2,nil,1}}, LastLogIndex:2, CommitIndex:1, NextIndex:map[int]int{2:2,3:2,4:2,5:2,6:2}, MatchIndex:map[int]int{2:1,3:1,4:1,5:1,6:1}}
	Sm.handleAppendEntriesResp(AppendEntriesResp{AppendEntriesReq{},2,1,false})
	expect(t, Sm, Sm_expectedState)
	//if no vote response
	if len(Sm.SendEvents) != 0 {
		t.Error(fmt.Sprintf("Expected Send Events Count: %v, found : %v\n", 1, len(Sm.SendEvents))) 
	}
	
	//Event Term is greater
	Sm = &StateMachine{Id:1, LeaderId:1, ServerIds:[]int{1,2,3,4,5,6}, State:leader, Term:2, VoteCount:0, NegVoteCount:0, VotedFor:1, Log:[]LogEntry{{0,nil,1},{1,nil,1},{2,nil,1}}, LastLogIndex:2, CommitIndex:1, NextIndex:map[int]int{2:2,3:2,4:2,5:2,6:2}, MatchIndex:map[int]int{2:1,3:1,4:1,5:1,6:1}}
	//vote granted; Term,VotedFor changed
	Sm_expectedState =  &StateMachine{Id:1, LeaderId:-1, ServerIds:[]int{1,2,3,4,5,6}, State:follower, Term:3, VoteCount:0, NegVoteCount:0, VotedFor:-1, Log:[]LogEntry{{0,nil,1},{1,nil,1},{2,nil,1}}, LastLogIndex:2, CommitIndex:1, NextIndex:nil, MatchIndex:nil}
	Sm.handleAppendEntriesResp(AppendEntriesResp{AppendEntriesReq{},2,3,false})
	expect(t, Sm, Sm_expectedState)
	//if no vote response
	if len(Sm.SendEvents) != 0 {
		t.Error(fmt.Sprintf("Expected Send Events Count: %v, found : %v\n", 1, len(Sm.SendEvents))) 
	}
	
	//Event Term is equal, failure, resend with decremented NextIndex 
	Sm = &StateMachine{Id:1, LeaderId:1, ServerIds:[]int{1,2,3,4,5,6}, State:leader, Term:2, VoteCount:0, NegVoteCount:0, VotedFor:1, Log:[]LogEntry{{0,nil,1},{1,nil,1},{2,nil,1}}, LastLogIndex:2, CommitIndex:1, NextIndex:map[int]int{2:2,3:3,4:3,5:3,6:2}, MatchIndex:map[int]int{2:0,3:2,4:2,5:2,6:1}}
	//vote granted; Term,VotedFor changed
	Sm_expectedState =  &StateMachine{Id:1, LeaderId:1, ServerIds:[]int{1,2,3,4,5,6}, State:leader, Term:2, VoteCount:0, NegVoteCount:0, VotedFor:1, Log:[]LogEntry{{0,nil,1},{1,nil,1},{2,nil,1}}, LastLogIndex:2, CommitIndex:1, NextIndex:map[int]int{2:1,3:3,4:3,5:3,6:2}, MatchIndex:map[int]int{2:0,3:2,4:2,5:2,6:1}}
	Sm.handleAppendEntriesResp(AppendEntriesResp{AppendEntriesReq{2, 1, 1, 1,[]LogEntry{{2,nil,1}}, 1},2,2,false})
	expect(t, Sm, Sm_expectedState)
	//if no vote response
	if len(Sm.SendEvents) != 1 {
		t.Error(fmt.Sprintf("Expected Send Events Count: %v, found : %v\n", 1, len(Sm.SendEvents))) 
	}
	
	//Event Term is equal, Success
	Sm = &StateMachine{Id:1, LeaderId:1, ServerIds:[]int{1,2,3,4,5,6}, State:leader, Term:2, VoteCount:0, NegVoteCount:0, VotedFor:1, Log:[]LogEntry{{0,nil,1},{1,nil,1},{2,nil,1}}, LastLogIndex:2, CommitIndex:1, NextIndex:map[int]int{2:1,3:3,4:3,5:3,6:2}, MatchIndex:map[int]int{2:0,3:2,4:2,5:2,6:1}}
	//vote granted; Term,VotedFor changed
	Sm_expectedState =  &StateMachine{Id:1, LeaderId:1, ServerIds:[]int{1,2,3,4,5,6}, State:leader, Term:2, VoteCount:0, NegVoteCount:0, VotedFor:1, Log:[]LogEntry{{0,nil,1},{1,nil,1},{2,nil,1}}, LastLogIndex:2, CommitIndex:2, NextIndex:map[int]int{2:3,3:3,4:3,5:3,6:2}, MatchIndex:map[int]int{2:2,3:2,4:2,5:2,6:1}}
	Sm.handleAppendEntriesResp(AppendEntriesResp{AppendEntriesReq{2, 1, 0, 0,[]LogEntry{{1,nil,1},{2,nil,1}}, 1},2,2,true})
	expect(t, Sm, Sm_expectedState)
	//if no vote response
	if len(Sm.SendEvents) != 0 {
		t.Error(fmt.Sprintf("Expected Send Events Count: %v, found : %v\n", 0, len(Sm.SendEvents))) 
	}
	fmt.Printf("***State Machine Tests End***\n\n\n")
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
			if a[i].Term != b[i].Term {
				return false
			}
			if !bytes.Equal(a[i].Data, b[i].Data)  {
				return false
			}
	}

	return true
}  

//_____________________________________________________________________________________________________________________

func expect(t *testing.T, Sm1 *StateMachine, Sm2 *StateMachine) {
	if Sm1.Id != Sm2.Id {
		t.Error(fmt.Sprintf("Expected Server Id: %v, found Server Id: %v\n", Sm2.Id, Sm1.Id)) 
	}
	
	if Sm1.LeaderId != Sm2.LeaderId {
		t.Error(fmt.Sprintf("Expected Leader Id: %v, found Leader Id: %v\n", Sm2.LeaderId, Sm1.LeaderId)) 
	}
	
	if testEqArrInt(Sm1.ServerIds,Sm2.ServerIds)==false {
		t.Error(fmt.Sprintf("Expected %v, found %v\n", Sm2.ServerIds, Sm1.ServerIds)) 
	}
	
	if Sm1.State != Sm2.State {
		t.Error(fmt.Sprintf("Expected Server State: %v, found Server State: %v\n", Sm2.State, Sm1.State)) 
	}
	
	if Sm1.Term != Sm2.Term {
		t.Error(fmt.Sprintf("Expected Server Term: %v, found Server Term: %v\n", Sm2.Term, Sm1.Term)) 
	}
	
	if Sm1.VoteCount != Sm2.VoteCount {
		t.Error(fmt.Sprintf("Expected Server Vote Count: %v, found Server Vote Count: %v\n", Sm2.VoteCount, Sm1.VoteCount)) 
	}
	
	if Sm1.NegVoteCount != Sm2.NegVoteCount {
		t.Error(fmt.Sprintf("Expected Server Vote Count: %v, found Server Vote Count: %v\n", Sm2.NegVoteCount, Sm1.NegVoteCount)) 
	}
	
	if Sm1.VotedFor != Sm2.VotedFor {
		t.Error(fmt.Sprintf("Expected Server Voted For Id: %v, found Server Voted For Id: %v\n", Sm2.VotedFor, Sm1.VotedFor)) 
	}
	
	/*if testEqArrLog(Sm1.Log,Sm2.Log)==false {
		t.Error(fmt.Sprintf("Expected %v, found %v\n", Sm2.Log, Sm1.Log)) 
	}*/
	
	if Sm1.LastLogIndex != Sm2.LastLogIndex {
		t.Error(fmt.Sprintf("Expected Server Last Log Index: %v, found Server Last Log Index: %v\n", Sm2.LastLogIndex, Sm1.LastLogIndex)) 
	}
	
	if Sm1.CommitIndex != Sm2.CommitIndex {
		t.Error(fmt.Sprintf("Expected Server Commit Index: %v, found Server Commit Index: %v\n", Sm2.CommitIndex, Sm1.CommitIndex)) 
	}
	
	if testEqMaps(Sm1.NextIndex, Sm2.NextIndex)==false {
		t.Error(fmt.Sprintf("Expected %v, found %v\n", Sm2.NextIndex, Sm1.NextIndex)) 
	}
	
	if testEqMaps(Sm1.MatchIndex, Sm2.MatchIndex)==false {
		t.Error(fmt.Sprintf("Expected %v, found %v\n", Sm2.MatchIndex, Sm1.MatchIndex)) 
	}
	
}

