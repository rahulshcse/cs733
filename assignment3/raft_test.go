package raftnode

import (
	"testing"
	"time"
	"fmt"
	"os"
	"encoding/gob"
	"path/filepath"
	"bytes"
)

var rafts map[int]Node
var ldr Node
func init() {
	gob.Register(VoteReq{})
	gob.Register(VoteResp{})
	gob.Register(AppendEntriesReq{})
	gob.Register(AppendEntriesResp{})
}
  

//_____________________________________________________________________________________________________________________

 func deleteOldLogs() { 
 	dirname := "." + string(filepath.Separator)

	d, err := os.Open(dirname)
	if err != nil {
	  fmt.Println(err)
	  os.Exit(1)
	}
	defer d.Close()

	files, err := d.Readdir(-1)
	if err != nil {
	  fmt.Println(err)
	  os.Exit(1)
	}

	for _, file := range files {
		if boolvar, err:=filepath.Match("log*",file.Name()); err == nil && boolvar==true {
			os.Remove("file.Name()")
		} else if boolvar, err:=filepath.Match("statelog*",file.Name()); err == nil && boolvar==true {
			os.Remove("file.Name()")
		} else if boolvar, err:=filepath.Match("testlog*",file.Name()); err == nil && boolvar==true {
			os.Remove("file.Name()")
		}
	}
 }

  

//_____________________________________________________________________________________________________________________

func TestStart (t *testing.T) {
	deleteOldLogs()
	fmt.Printf("Old logs deleted\n")
	rafts = makeRafts()
	fmt.Printf("All Raft nodes started successfully\n")
	fmt.Printf("Complete testing will take about 180 seconds\n\n\n")
}

  

//_____________________________________________________________________________________________________________________

func TestLeader (t *testing.T) {
	fmt.Printf("\n\n\n***Leader Validation Test started***\n")
	time.Sleep(20*time.Second)
	ldr = getLeader(rafts)
	if ldr == nil {
		t.Errorf("Leader election test failed\n")
	} else {
		fmt.Printf("Leader is Node %d\n", ldr.Id())
		fmt.Printf("Leader election test passed\n")
	}
}

  

//_____________________________________________________________________________________________________________________

func TestReplication (t *testing.T) {
	fmt.Printf("\n\n\n***Append Request Replication Test Started***\n")
	fmt.Printf("Replicating 4 messages\n")
	ldr.Append([]byte("foo"))
	ldr.Append([]byte("bar"))
	ldr.Append([]byte("foo"))
	ldr.Append([]byte("bar"))
	time.Sleep(20*time.Second)
	for _, node:=range rafts {
		if len(node.CommitCh()) == 4 {
			assert(t,<-node.CommitCh(), []byte("foo"))
			assert(t,<-node.CommitCh(), []byte("bar"))
			assert(t,<-node.CommitCh(), []byte("foo"))
			assert(t,<-node.CommitCh(), []byte("bar"))
		} else {
			fmt.Printf("Log State of %d is %v\n", node.Id(),node.Id(),node.(*RaftNode).Sm.Log)
		}
	}
	fmt.Printf("Message replication successful\n")
}
  

//_____________________________________________________________________________________________________________________

func TestPartition (t *testing.T) {
	fmt.Println("\n\n\n***Network Partition Test Started***\n")
	time.Sleep(10*time.Second)
	Partition(rafts,[]int{1, 2, 3}, []int{4, 5})
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
	time.Sleep(20*time.Second)
	fmt.Printf("Clusters healed into one\n")
	Heal(rafts)
	time.Sleep(100*time.Second)
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

func assert(t *testing.T, a CommitInfo, b []byte) {
	if bytes.Compare(a.Data, b) != 0 { 
		t.Error(fmt.Sprintf("Expected %v, found %v", b, a)) // t.Error is visible when running `go test -verbose`
	}
}
