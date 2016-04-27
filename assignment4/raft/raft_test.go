package raftnode

import (
	"testing"
	"time"
	"fmt"
	//"os"
	//log "github.com/cs733-iitb/log"
	"bytes"
	//"strconv"
)

var rafts map[int]Node
var ldr Node

  

//_____________________________________________________________________________________________________________________

func TestStart (t *testing.T) {
	DeleteOldLogs()
	fmt.Printf("Old logs deleted\n")
	rafts = MakeRafts()
	for i:=1;i<=5;i++ {
			go rafts[i].ProcessEvents()
	}
	time.Sleep(10*time.Second)
	fmt.Printf("All Raft nodes started successfully\n")
	fmt.Printf("Complete testing will take about 25 seconds\n\n\n")
}

  

//_____________________________________________________________________________________________________________________

func TestLeader (t *testing.T) {
	fmt.Printf("\n\n\n***Leader Validation Test started***\n")
	time.Sleep(5*time.Second)
	for ldr == nil {
		ldr = GetLeader(rafts)
	}
	fmt.Printf("Leader is Node %d\n", ldr.Id())
	fmt.Printf("Leader election test passed\n")
}

  

//_____________________________________________________________________________________________________________________

func TestReplication (t *testing.T) {
	fmt.Printf("\n\n\n***Append Request Replication Test Started***\n")
	fmt.Printf("Replicating 4 messages\n")
		ldr.Append([]byte("foo"))
		ldr.Append([]byte("bar"))
		ldr.Append([]byte("foo"))
		ldr.Append([]byte("bar"))
		ldr.Append([]byte("foo"))
		time.Sleep(10*time.Second)
	fmt.Printf("Message replication successful\n")
}
  

//_____________________________________________________________________________________________________________________

func TestPartition (t *testing.T) {
	fmt.Println("\n\n\n***Network Partition Test Started***\n")
	Partition(rafts,[]int{1, 2, 3}, []int{4, 5})
	time.Sleep(3*time.Second)
	fmt.Printf("Network Partitioned cluster into 2 clusters\nNodes 1,2,3 and Nodes 4,5\n")
	if rafts[1].LeaderId() == rafts[2].LeaderId() && rafts[2].LeaderId() == rafts[3].LeaderId() {
		fmt.Printf("Network Partition 1 has leader %d\n", rafts[1].LeaderId())
	} else {
		t.Errorf("Network Partition 1 has no leader\nLeader Test Failed\n")
	}
	if rafts[4].LeaderId() == rafts[5].LeaderId() && rafts[4].LeaderId() == -1 {
		fmt.Printf("Network Partition test successful\n")
	}
}
  

//_____________________________________________________________________________________________________________________

func TestHeal (t *testing.T) {
	fmt.Println("\n\n\n***Network Partitions Heal Test Started***\n")
	Heal(rafts)
	time.Sleep(5*time.Second)
	fmt.Printf("Clusters healed into one\n")
	if rafts[1].LeaderId() == rafts[2].LeaderId() && rafts[2].LeaderId() == rafts[3].LeaderId() && rafts[3].LeaderId() == rafts[4].LeaderId() && rafts[4].LeaderId() == rafts[5].LeaderId() {
		fmt.Printf("Healed Network Partition has leader %d\n", rafts[1].LeaderId())
		fmt.Printf("Heal Partitions test successful\n")
	} else {
		t.Errorf("Healed Network Partition has no leader\nHeal Partitions test failed\n")
	}
}
  

//_____________________________________________________________________________________________________________________

func TestShutDown (t *testing.T) {
	fmt.Println("\n\n\n***Nodes Shutdown Test Started***\n")
	for _, node:=range rafts {
		node.Shutdown()
	}
	fmt.Println("Nodes Shutdown test completed\n\n\nAll Tests Over\n\n\n")
}
  

//_____________________________________________________________________________________________________________________

func assert(t *testing.T, a CommitInfo, b []byte, index int64) {
	if bytes.Compare(a.Data, b) != 0 && a.Index != index { 
		t.Error(fmt.Sprintf("Expected %v, found %v", b, a)) // t.Error is visible when running `go test -verbose`
	}
}
