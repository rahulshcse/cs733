package main

import (
	"bufio"
	"fmt"
	raft "github.com/rahulshcse/cs733/assignment4/raft"
	fs "github.com/rahulshcse/cs733/assignment4/fs"
	cluster "github.com/cs733-iitb/cluster/mock"
	log "github.com/cs733-iitb/log"
	"net"
	"os"
	"strconv"
	"encoding/json"
	"time"
)

var clientCount = 0
var offset int = 0
type FsLog struct {
	Index int64
}
type ModifiedMsg struct {
	ClientId	int
	Message		*fs.Msg
}

type ClientHandler struct {
	Response		chan *fs.Msg
	CloseHandler	chan int
}

type FileServer struct {
	Id					int
	Port				int
	Rf 					*raft.RaftNode
	ClientCh			map[int] ClientHandler
	FileStore			fs.FS
	Crlf				[]byte
	NeedClose			chan int
	CloseAll			chan int
	LastProcessedIndex	int64
}

var FileServers map[int] *FileServer

func makeFS(id int, port int) *FileServer {
	fileServer := new(FileServer)
	fileServer.Id = id
	fileServer.Port = port
	fileServer.ClientCh = make(map[int] ClientHandler)
	fileServer.NeedClose = make(chan int, 1)
	fileServer.CloseAll = make(chan int, 1)
	//fileServer.CloseRaft = make(chan int, 1)
	//fileServer.ClosedAll = make(chan int, 1)
	fileServer.FileStore = fs.FS{Dir: make(map[string]*fs.FileInfo, 1000)}
	fileServer.Crlf = []byte{'\r', '\n'}
	//fileServer.Rf = initRaft(id)
	fileServer.LastProcessedIndex = fileServer.GetLastProcessed()
	return fileServer
}

func initRaft(id int) *raft.RaftNode {
	cluster, _ := cluster.NewCluster("raft/cluster_test_config.json")
	raftNode := raft.NewNode(id)
	raftNode.(*raft.RaftNode).Server = cluster.Servers[id]
	peerIds:=raftNode.(*raft.RaftNode).Server.Peers()
	raftNode.(*raft.RaftNode).Sm=raftNode.(*raft.RaftNode).NewSm(id, peerIds)
	return raftNode.(*raft.RaftNode)
}



func getLeader() int {
	raftNodes := make(map[int]raft.Node)
	raftNodes[1] = FileServers[1].Rf
	for _,raftNodeId := range raftNodes[1].(*raft.RaftNode).Server.Peers() {
		raftNodes[raftNodeId] = FileServers[raftNodeId].Rf
	}
	//fmt.Printf("\n\n\n***Leader Validation Test started***\n")
	var ldr raft.Node
	for ldr == nil {
		ldr = raft.GetLeader(raftNodes)
	}
	return ldr.Id()
}


func (fileServer *FileServer) GetLastProcessed() int64 {
	logname:="fslog"
	logname+=strconv.Itoa(fileServer.Id)
	lg,_ := log.Open(logname)
	lg.SetCacheSize(5000) 
	lg.RegisterSampleEntry(FsLog{})
	defer lg.Close()
	if lg.GetLastIndex() == -1 {
		lg.Append(FsLog{0})
		lg.Append(FsLog{0})
		lg.Close()
		return 0
	} else {
		lastProcessed,_ := lg.Get(1)
		lg.Close()
		return lastProcessed.(FsLog).Index
	}
}

func (fileServer *FileServer) SetLastProcessed(lastProcessed int64) {
	fileServer.LastProcessedIndex = lastProcessed
	logname:="fslog"
	logname+=strconv.Itoa(fileServer.Id)
	lg,_ := log.Open(logname)
	lg.SetCacheSize(5000) 
	lg.RegisterSampleEntry(FsLog{})
	defer lg.Close()
	lg.TruncateToEnd(1)
 	lg.Append(FsLog{lastProcessed})
	lg.Close()
}
//var FSArr =make(map[int]FileServer)

//var crlf = []byte{'\r', '\n'}

func (fileServer *FileServer) check(obj interface{}) {
	if obj != nil {
		fmt.Println(obj)
		os.Exit(1)
	}
}

func (fileServer *FileServer) reply(conn *net.TCPConn, msg *fs.Msg, host string, port int) bool {
	var err error
	write := func(data []byte) {
		if err != nil {
			return
		}
		_, err = conn.Write(data)
	}
	var resp string
	switch msg.Kind {
	case 'C': // read Response
		resp = fmt.Sprintf("CONTENTS %d %d %d", msg.Version, msg.Numbytes, msg.Exptime)
	case 'O':
		resp = "OK "
		if msg.Version > 0 {
			resp += strconv.Itoa(msg.Version)
		}
	case 'F':
		resp = "ERR_FILE_NOT_FOUND"
	case 'V':
		resp = "ERR_VERSION " + strconv.Itoa(msg.Version)
	case 'M':
		resp = "ERR_CMD_ERR"
	case 'I':
		resp = "ERR_INTERNAL"
	case 'N':
		resp = "ERR_REDIRECT " + host + ":" + strconv.Itoa(port)
	default:
		fmt.Printf("Unknown Response kind '%c'", msg.Kind)
		return false
	}
	resp += "\r\n"
	write([]byte(resp))
	if msg.Kind == 'C' {
		write(msg.Contents)
		write(fileServer.Crlf)
	}
	return err == nil
}

func (fileServer *FileServer) serve(conn *net.TCPConn) {
	reader := bufio.NewReader(conn)
	currClient := clientCount + 1
	clientCount++
	fileServer.ClientCh[currClient] = ClientHandler{Response: make(chan *fs.Msg, 10), CloseHandler: make(chan int, 1)}
	//fileServer.ClientCh[currClient].Response = make(chan *fs.Msg, 10)
	//fileServer.ClientCh[currClient].CloseHandler = make(chan int, 1)
	for {
		if len(fileServer.CloseAll) != 0 {
			fileServer.clientClose(currClient)
			conn.Close()
			return
		} else {
			msg, msgerr, fatalerr := fileServer.FileStore.GetMsg(reader)
			if fatalerr != nil || msgerr != nil {
				fileServer.reply(conn, &fs.Msg{Kind: 'M'}, "", 0)
				fileServer.clientClose(currClient)
				conn.Close()
				return
			}

			if fileServer.Rf.LeaderId() != fileServer.Id && fileServer.Rf.LeaderId() != -1 {
				fileServer.reply(conn, &fs.Msg{Kind: 'N'}, "localhost", FileServers[fileServer.Rf.LeaderId()].Port)
				fileServer.clientClose(currClient)
				conn.Close()
				return
			} else if fileServer.Rf.LeaderId() == -1 {
				fileServer.reply(conn, &fs.Msg{Kind: 'N'}, "localhost", FileServers[fileServer.Id].Port)
				fileServer.clientClose(currClient)
				conn.Close()
				return
			}

			var modifiedmsg ModifiedMsg
			modifiedmsg.ClientId = currClient
			modifiedmsg.Message = msg
			b, err := json.Marshal(modifiedmsg)

			if err != nil {
				fileServer.reply(conn, &fs.Msg{Kind: 'I'}, "", 0)
				fileServer.clientClose(currClient)
				conn.Close()
				return
			}

			fileServer.Rf.Append(b)
			response := <- fileServer.ClientCh[currClient].Response
			if response.Kind == 'N' {
				if fileServer.Rf.LeaderId() != -1 {
					fileServer.reply(conn, &fs.Msg{Kind: 'N'}, "localhost", FileServers[fileServer.Rf.LeaderId()].Port)
				} else {
					fileServer.reply(conn, &fs.Msg{Kind: 'N'}, "localhost", FileServers[fileServer.Id].Port)
				}
				fileServer.clientClose(currClient)
				conn.Close()
				return
			}
			if !fileServer.reply(conn, response, "", 0) {
				fileServer.clientClose(currClient)
				conn.Close()
				return
			}
		}
	}
}

func (fileServer *FileServer) clientClose (currClient int) {
	close(fileServer.ClientCh[currClient].Response)
	close(fileServer.ClientCh[currClient].CloseHandler)
	delete(fileServer.ClientCh, currClient)
}

func (fileServer *FileServer) backEnd() {
	for {
		select {
			case <- fileServer.NeedClose: 
				fileServer.CloseAll <- 1
				//fmt.Println("BlahBlah ", fileServer.Id)
				time.Sleep(10*time.Second)
				return
			case data := <- fileServer.Rf.CommitCh():
			    if data.Err != nil {
			    	res := ModifiedMsg{}
			    	json.Unmarshal([]byte(data.Data), &res)
		    		if _, ok := fileServer.ClientCh[res.ClientId]; ok {
		    			fileServer.ClientCh[res.ClientId].Response <- &fs.Msg{Kind: 'N'}
		    		}
			    } else {
			    			//fmt.Println("FoundFound1")
			    	for i := fileServer.LastProcessedIndex + 1; i <= data.Index; i++ {
			    			//fmt.Println("FoundFound2")
			    		res := ModifiedMsg{}
			    		//dat := fileServer.Rf.Sm.Log[i]
			    		if len(fileServer.Rf.Sm.Log)>int(i) {
			    			if fileServer.Rf.Sm.Log[i].Data != nil {
				    			json.Unmarshal([]byte(fileServer.Rf.Sm.Log[i].Data), &res)
				    			//fmt.Println("FoundFound3")
					    		if _, ok := fileServer.ClientCh[res.ClientId]; ok {
									fileServer.ClientCh[res.ClientId].Response <- fileServer.FileStore.ProcessMsg(res.Message)
								} else {
									fileServer.FileStore.ProcessMsg(res.Message)
								}
								fileServer.SetLastProcessed(i)
							}
						}
					}
					fileServer.SetLastProcessed(data.Index)
			    }
			default:
				;
		}
	}
}

func (fileServer *FileServer) startServer() {
	tcpaddr, err := net.ResolveTCPAddr("tcp", "localhost:" + strconv.Itoa(fileServer.Port))
	fileServer.check(err)
	if fileServer.Rf.Isdown == 1 {
		fileServer.Rf.Isdown = 0
		fmt.Printf("Node %d up again\n", fileServer.Rf.Id());
		fileServer.Rf.Up()
		if len(fileServer.NeedClose) != 0 {
			<-fileServer.NeedClose
		}
		if len(fileServer.CloseAll) != 0 {
			<-fileServer.CloseAll
		}
		time.Sleep(10*time.Second)
	}
	tcp_acceptor, err := net.ListenTCP("tcp", tcpaddr)
	fileServer.check(err)
	defer tcp_acceptor.Close()
	go fileServer.backEnd()
	go fileServer.Rf.ProcessEvents()
	for {
		if len(fileServer.CloseAll) != 0 {
			fileServer.Rf.Shutdown()
			break
		} else if fileServer.Rf.IsClosedRaft() {
			break
		} else {
			for len(fileServer.CloseAll) == 0 && fileServer.Rf.LeaderId() != fileServer.Id {
				if tcp_acceptor != nil {
					tcp_acceptor.Close()
					tcp_acceptor = nil
				}
				time.Sleep(1*time.Second)
			}
			if len(fileServer.CloseAll) != 0 {
				fileServer.Rf.Shutdown()
				break
			} else if fileServer.Rf.IsClosedRaft() {
				break
			}
			if tcp_acceptor == nil {
				tcp_acceptor, err = net.ListenTCP("tcp", tcpaddr)
				fileServer.check(err)
			}
			tcp_conn, err := tcp_acceptor.AcceptTCP()
			fileServer.check(err)
			go fileServer.serve(tcp_conn)
		}
	}
	tcp_acceptor.Close()
}

func serverMain() {
	FileServers = make(map[int]*FileServer)
	for i:=1;i<=5;i++ {
		FileServers[i] = makeFS(i, 8000 + i + offset)
	}
	rafts := raft.MakeRafts()
	for i:=1;i<=5;i++ {
		FileServers[i].Rf = rafts[i].(*raft.RaftNode)
	}
	for i:=1;i<=5;i++ {
		go FileServers[i].startServer()
	}
}

func (fileServer *FileServer) ServerNodeClose() {
	fileServer.Rf.Shutdown()
	fileServer.Rf.Isdown = 1
	time.Sleep(10*time.Second)
}

func serverNodeClose(i int) {
	FileServers[i].ServerNodeClose()
	time.Sleep(10*time.Second)
	fmt.Printf("Server node %d closed\n", i)
}

func serverClose(i int) {
	FileServers[i].NeedClose <- 1
	FileServers[i].ServerNodeClose()
	time.Sleep(10*time.Second)
	fmt.Printf("Server %d closed\n", i)
}

func serversClose() {
	for i:=1;i<=5;i++ {
		FileServers[i].NeedClose <- 1
	}
	time.Sleep(10*time.Second)
	for i:=1;i<=5;i++ {
		if !FileServers[i].Rf.IsClosedRaft() {
			FileServers[i].Rf.Shutdown()
		}
	}
	time.Sleep(10*time.Second)
}

func main() {
	serverMain()
}
