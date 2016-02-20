# Assignment 2-  RAFT state machine
====================================

# Description

This assignment is to build RAFT state machine and test a single instance of it.<br/>


## State Machine Description

State Machine can be in one of the following states<br/>
1. `follower`
2. `candidate`
3. `leader`

State machine handles following input events<br/>
1. `Timeout` : A timeout event is interpreted according to the state. If the state machine is a leader, it is interpreted as a heartbeat
timeout, and if it is a follower or candidate, it is interpreted as an election timeout.
2. `Append` : This is a request from the layer above to append the data to the replicated log. The response is in the form of an eventual Commit action (see next section).
3. `VoteReq` : Message from another Raft state machine to request votes for its candidature.
4. `VoteResp` : Response to a Vote request.
5. `AppendEntriesReq` : Message from another Raft state machine.
6. `AppendEntriesResp` : Response from another Raft state machine in response to a previous AppendEntriesReq.

Output Events Generated<br/>
1. `Send` :  Send this event to a remote node. The event is one of AppendEntriesReq/Resp or VoteReq/Resp.
2. `Commit` : Make the log entries durable.
3. `Alarm` : Set/Reset election/heartbeat timeout.
4. `LogStore` : Store the data at the given index to persistent storage.
5. `StateStore` : Store commitIndex, term and votedFor to persistent storage.


## Installation Instructions

<code>go get github.com/rahulshcse/cs733/assignment2</code>

Two files are supposed to be there <br/>
1. `raft_sm.go` contains the code where algorithm and event handling code for RAFT state machine is implemented.<br/>
2. `raft_sm_test.go` contains all the test cases to test all events for candidate, leader and follower states.<br/>

To run the program below command is needed (assuming the current directory is set to the assignment2 which has the go files) 
<br/><code>go test</code>
  
  
### Included Test Cases

Follower State<br/>
* Timeout Event.
* Append Event.
* Vote Request Event.
* Vote Response Event.
* Append Entries Request Event.
* Append Entries Response Event.

Candidate State<br/>
* Timeout Event.
* Append Event.
* Vote Request Event.
* Vote Response Event.
* Append Entries Request Event.
* Append Entries Response Event.

Leader State<br/>
* Timeout Event.
* Append Event.
* Vote Request Event.
* Vote Response Event.
* Append Entries Request Event.
* Append Entries Response Event.
