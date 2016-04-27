package main


import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
	"os"
)

func DeleteOldLogs() { 
	os.RemoveAll("fslog1")
	os.RemoveAll("fslog2")
	os.RemoveAll("fslog3")
	os.RemoveAll("fslog4")
	os.RemoveAll("fslog5")
	os.RemoveAll("log1")
	os.RemoveAll("log2")
	os.RemoveAll("log3")
	os.RemoveAll("log4")
	os.RemoveAll("log5")
 }

func TestStart (t *testing.T) {
	fmt.Printf("\n\n\n\n\n\n\n\n***File System Test started***\n")
	DeleteOldLogs()
	time.Sleep(5*time.Second)
	fmt.Printf("Old logs deleted\n")
}

func TestRPCMain1(t *testing.T) {
	go serverMain()
	time.Sleep(10 * time.Second)
}

func expect(t *testing.T, response *Msg, expected *Msg, errstr string, err error) int {
	if err != nil {
		t.Fatal("Unexpected error: " + err.Error())
	}
	ok := true
	if response.Kind == 'N' {
		return 1
	}
	if response.Kind != expected.Kind {
		ok = false
		errstr += fmt.Sprintf(" Got kind='%c', expected '%c'", response.Kind, expected.Kind)
	}
	if expected.Version > 0 && expected.Version != response.Version {
		ok = false
		errstr += " Version mismatch"
	}
	if response.Kind == 'C' {
		if expected.Contents != nil &&
			bytes.Compare(response.Contents, expected.Contents) != 0 {
			ok = false
		}
	}
	if !ok {
		t.Fatal("Expected " + errstr)
	}
	return 0
}

func TestRPC_BasicSequential1(t *testing.T) {
	var expint int
	cl := mkClient(t)
	defer cl.close()

	// Read non-existent file cs733net
	m, err := cl.read("cs733net")
	expint = expect(t, m, &Msg{Kind: 'F'}, "file not found", err)
	if expint == 1 {
		cl.close()
		cl = mkClient(t)
		fmt.Println("Client redirected to another leader\n")
		return
	}
}

func TestRPC_BasicSequential2(t *testing.T) {
	var expint int
	serverNodeClose(1)
	time.Sleep(10*time.Second)
	cl := mkClient(t)
	defer cl.close()
	top:
	fmt.Printf("Client connection successful with leader id %d\n", getLeader())
	// Write file cs733net
	data := "Cloud fun"
	m, err := cl.write("cs733net", data, 0)
	expint = expect(t, m, &Msg{Kind: 'O'}, "write success", err)
	if expint == 1 {
		cl.close()
		cl = mkClient(t)
		fmt.Println("Client redirected to another leader\n")
		goto top;
	}

	// Expect to read it back
	m, err = cl.read("cs733net")
	expint = expect(t, m, &Msg{Kind: 'C', Contents: []byte(data)}, "read my write", err)
	if expint == 1 {
		cl.close()
		cl = mkClient(t)
		fmt.Println("Client redirected to another leader\n")
		goto top;
	}

	// CAS in new value
	version1 := m.Version
	data2 := "Cloud fun 2"
	// Cas new value
	m, err = cl.cas("cs733net", version1, data2, 0)
	expint = expect(t, m, &Msg{Kind: 'O'}, "cas success", err)
	if expint == 1 {
		cl.close()
		cl = mkClient(t)
		fmt.Println("Client redirected to another leader\n")
		goto top;
	}

	// Expect Cas to fail with old version
	m, err = cl.cas("cs733net", version1, data, 0)
	if expint == 1 {
		cl.close()
		cl = mkClient(t)
		goto top;
	}
	cl.close()
}

func TestRPC_Binary(t *testing.T) {
	var expint int
	cl := mkClient(t)
	defer cl.close()
	top:
	// Write binary contents
	data := "\x00\x01\r\n\x03" // some non-ascii, some crlf chars
	m, err := cl.write("binfile", data, 0)
	expint = expect(t, m, &Msg{Kind: 'O'}, "write success", err)
	if expint == 1 {
		cl.close()
		cl = mkClient(t)
		fmt.Println("Client redirected to another leader\n")
		goto top;
	}

	go FileServers[1].startServer()

	// Expect to read it back
	m, err = cl.read("binfile")
	expint = expect(t, m, &Msg{Kind: 'C', Contents: []byte(data)}, "read my write", err)
	cl.close()
}

func TestRPC_Chunks(t *testing.T) {
	var expint int
	cl := mkClient(t)
	defer cl.close()
	var err error
	snd := func(chunk string) {
		if err == nil {
			err = cl.send(chunk)
		}
	}
	top:
	// Send the command "write teststream 10\r\nabcdefghij\r\n" in multiple chunks
	// Nagle's algorithm is disabled on a write, so the server should get these in separate TCP packets.
	snd("wr")
	time.Sleep(10 * time.Millisecond)
	snd("ite test")
	time.Sleep(10 * time.Millisecond)
	snd("stream 1")
	time.Sleep(10 * time.Millisecond)
	snd("0\r\nabcdefghij\r")
	time.Sleep(10 * time.Millisecond)
	snd("\n")
	var m *Msg
	m, err = cl.rcv()
	expint = expect(t, m, &Msg{Kind: 'O'}, "writing in chunks should work", err)
	if expint == 1 {
		cl.close()
		cl = mkClient(t)
		fmt.Println("Client redirected to another leader\n")
		goto top;
	}
	cl.close()
}


// nclients write to the same file. At the end the file should be
// any one clients' last write*/

func TestRPC_ConcurrentWrites(t *testing.T) {
	nclients := 5
	niters := 1
	clients := make([]*Client, nclients)
	for i := 0; i < nclients; i++ {
		cl := mkClient(t)
		if cl == nil {
			t.Fatalf("Unable to create client #%d", i)
		}
		defer cl.close()
		clients[i] = cl
	}

	errCh := make(chan error, nclients)
	var sem sync.WaitGroup // Used as a semaphore to coordinate goroutines to begin concurrently
	sem.Add(1)
	ch := make(chan *Msg, nclients*niters) // channel for all replies
	for i := 0; i < nclients; i++ {
		go func(i int, cl *Client) {
			sem.Wait()
			for j := 0; j < niters; j++ {
				str := fmt.Sprintf("cl %d %d", i, j)
				m, err := cl.write("concWrite", str, 0)
				if err != nil {
					errCh <- err
					break
				} else {
					ch <- m
				}
			}
		}(i, clients[i])
	}
	time.Sleep(100 * time.Millisecond) // give goroutines a chance
	sem.Done()                      // Go!

	// There should be no errors
	for i := 0; i < nclients*niters; i++ {
		select {
		case m := <-ch:
			if m.Kind != 'O' && m.Kind != 'N' {
				t.Fatalf("Concurrent write failed with kind=%c", m.Kind)
			} else if m.Kind == 'N' {
				return
			}
		case err := <- errCh:
			t.Fatal(err)
		}
	}
}

//----------------------------------------------------------------------
// Utility functions

type Msg struct {
	// Kind = the first character of the command. For errors, it
	// is the first letter after "ERR_", ('V' for ERR_VERSION, for
	// example), except for "ERR_CMD_ERR", for which the kind is 'M'
	Kind     byte
	Filename string
	Contents []byte
	Numbytes int
	Exptime  int // expiry time in seconds
	Version  int
}

func (cl *Client) read(filename string) (*Msg, error) {
	cmd := "read " + filename + "\r\n"
	return cl.sendRcv(cmd)
}

func (cl *Client) write(filename string, contents string, exptime int) (*Msg, error) {
	var cmd string
	if exptime == 0 {
		cmd = fmt.Sprintf("write %s %d\r\n", filename, len(contents))
	} else {
		cmd = fmt.Sprintf("write %s %d %d\r\n", filename, len(contents), exptime)
	}
	cmd += contents + "\r\n"
	return cl.sendRcv(cmd)
}

func (cl *Client) cas(filename string, version int, contents string, exptime int) (*Msg, error) {
	var cmd string
	if exptime == 0 {
		cmd = fmt.Sprintf("cas %s %d %d\r\n", filename, version, len(contents))
	} else {
		cmd = fmt.Sprintf("cas %s %d %d %d\r\n", filename, version, len(contents), exptime)
	}
	cmd += contents + "\r\n"
	return cl.sendRcv(cmd)
}

func (cl *Client) delete(filename string) (*Msg, error) {
	cmd := "delete " + filename + "\r\n"
	return cl.sendRcv(cmd)
}

var errNoConn = errors.New("Connection is closed")

type Client struct {
	conn   *net.TCPConn
	reader *bufio.Reader // a bufio Reader wrapper over conn
}

func mkClient(t *testing.T) *Client {
	var client *Client
	port := 8001 + offset
	i := 0
	raddr, err := net.ResolveTCPAddr("tcp", "localhost:"+ strconv.Itoa(port+i))
	if err == nil {
		conn, err := net.DialTCP("tcp", nil, raddr)
		for err != nil {
			i = (i+1)%5
			//fmt.Printf("\nTrying with Port no : %d\n", port+i)
			raddr, _ = net.ResolveTCPAddr("tcp", "localhost:"+ strconv.Itoa(port+i))
			conn, err = net.DialTCP("tcp", nil, raddr)
		}
		if err == nil {
			client = &Client{conn: conn, reader: bufio.NewReader(conn)}
		}
	}
	if err != nil {
		t.Fatal(err)
	}
	return client
}

func (cl *Client) send(str string) error {
	if cl.conn == nil {
		return errNoConn
	}
	_, err := cl.conn.Write([]byte(str))
	if err != nil {
		err = fmt.Errorf("Write error in SendRaw: %v", err)
		cl.conn.Close()
		cl.conn = nil
	}
	return err
}

func (cl *Client) sendRcv(str string) (msg *Msg, err error) {
	if cl.conn == nil {
		return nil, errNoConn
	}
	err = cl.send(str)
	if err == nil {
		msg, err = cl.rcv()
	}
	return msg, err
}

func (cl *Client) close() {
	if cl != nil && cl.conn != nil {
		cl.conn.Close()
		cl.conn = nil
	}
}

func (cl *Client) rcv() (msg *Msg, err error) {
	// we will assume no errors in server side formatting
	line, err := cl.reader.ReadString('\n')
	if err == nil {
		msg, err = parseFirst(line)
		if err != nil {
			return nil, err
		}
		if msg.Kind == 'C' {
			contents := make([]byte, msg.Numbytes)
			var c byte
			for i := 0; i < msg.Numbytes; i++ {
				if c, err = cl.reader.ReadByte(); err != nil {
					break
				}
				contents[i] = c
			}
			if err == nil {
				msg.Contents = contents
				cl.reader.ReadByte() // \r
				cl.reader.ReadByte() // \n
			}
		}
	}
	if err != nil {
		cl.close()
	}
	return msg, err
}

func parseFirst(line string) (msg *Msg, err error) {
	fields := strings.Fields(line)
	msg = &Msg{}

	// Utility function fieldNum to int
	toInt := func(fieldNum int) int {
		var i int
		if err == nil {
			if fieldNum >=  len(fields) {
				err = errors.New(fmt.Sprintf("Not enough fields. Expected field #%d in %s\n", fieldNum, line))
				return 0
			}
			i, err = strconv.Atoi(fields[fieldNum])
		}
		return i
	}

	if len(fields) == 0 {
		return nil, errors.New("Empty line. The previous command is likely at fault")
	}
	switch fields[0] {
	case "OK": // OK [version]
		msg.Kind = 'O'
		if len(fields) > 1 {
			msg.Version = toInt(1)
		}
	case "CONTENTS": // CONTENTS <version> <numbytes> <exptime> \r\n
		msg.Kind = 'C'
		msg.Version = toInt(1)
		msg.Numbytes = toInt(2)
		msg.Exptime = toInt(3)
	case "ERR_VERSION":
		msg.Kind = 'V'
		msg.Version = toInt(1)
	case "ERR_FILE_NOT_FOUND":
		msg.Kind = 'F'
	case "ERR_CMD_ERR":
		msg.Kind = 'M'
	case "ERR_INTERNAL":
		msg.Kind = 'I'
	case "ERR_REDIRECT":
		msg.Kind = 'N'
	default:
		err = errors.New("Unknown response " + fields[0])
	}
	if err != nil {
		return nil, err
	} else {
		return msg, nil
	}
}
