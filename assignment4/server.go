package main

import (
	"bufio"
	"fmt"
	"github.com/cs733-iitb/cs733/assignment1/fs"
	"net"
	"os"
	"strconv"
)

type FileSystem struct {
	port		int
	rf 			RaftNode
	clientCh	map[int] chan string
}

var FSArr =make(map[int]FileSystem)

var crlf = []byte{'\r', '\n'}

func check(obj interface{}) {
	if obj != nil {
		fmt.Println(obj)
		os.Exit(1)
	}
}

func reply(conn *net.TCPConn, msg *fs.Msg) bool {
	var err error
	write := func(data []byte) {
		if err != nil {
			return
		}
		_, err = conn.Write(data)
	}
	var resp string
	switch msg.Kind {
	case 'C': // read response
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
	default:
		fmt.Printf("Unknown response kind '%c'", msg.Kind)
		return false
	}
	resp += "\r\n"
	write([]byte(resp))
	if msg.Kind == 'C' {
		write(msg.Contents)
		write(crlf)
	}
	return err == nil
}

func serve(conn *net.TCPConn) {
	reader := bufio.NewReader(conn)
	for {
		msg, msgerr, fatalerr := fs.GetMsg(reader)
		if fatalerr != nil || msgerr != nil {
			reply(conn, &fs.Msg{Kind: 'M'})
			conn.Close()
			break
		}

		if msgerr != nil {
			if (!reply(conn, &fs.Msg{Kind: 'M'})) {
				conn.Close()
				break
			}
		}
		
		/*response := fs.ProcessMsg(msg)
		if !reply(conn, response) {
			conn.Close()
			break
		}*/
	}
}

func serverMain() {
	for i:=1;i<=5;i++ {
		var newFS FileSystem
		newFS.CreateRaft
	}
	//raftaddr:=[]string{"localhost:8080","localhost:8081","localhost:8082","localhost:8083","localhost:8084"}
	for i:=1;i<=5;i++ {
		go func() {
			tcpaddr, err := net.ResolveTCPAddr("tcp", raftaddr[i-1])
			check(err)
			tcp_acceptor, err := net.ListenTCP("tcp", tcpaddr)
			check(err)

			for {
				tcp_conn, err := tcp_acceptor.AcceptTCP()
				check(err)
				go serve(tcp_conn)
			}
		}
	}
}

func main() {
	serverMain()
}
