# Assignment 3-  RAFT Node
==========================

# Description

This assignment is about building a “Raft node”, a wrapper over the state machine. It provides system services for the state machine.<br/>


## Raft Node Description

Raft Node performs following actions<br/>
1. `Implements the interface`<br/>
2. `Create Event objects from Append(), incoming network messages, and timeouts. and feed them to the state machine`<br/>
3. `Interpret Actions for Alarm, Network and Log events`<br/>


## Installation Instructions

<code>go get github.com/rahulshcse/cs733/assignment3</code>

Two files are supposed to be there <br/>
1. `raft.go` contains the code which implements Raft Node.<br/>
1. `raft_sm.go` contains the code where algorithm and event handling code for RAFT state machine is implemented.<br/>
2. `raft_test.go` contains all the test cases.<br/>

To run the program below command is needed (assuming the current directory is set to the assignment2 which has the go files) 
<br/><code>go test</code>
  
  
### Included Test Cases

* Leader Validation Test.
* Append Request Replication Test.
* Network Partition Test.
* Network Partitions Heal Test.
* Nodes Shutdown Test.
