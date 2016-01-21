package main

import (
	"net"
	"os"
	"strings"
	"testing"
	"time"
	"fmt"
	"bufio"
	"strconv"
)

const servAddr string = "localhost:8080"

//stores versions of file "file.txt" after successful CAS by concurrent clients
var version_test_cas []int64

var conn net.Conn
func TestMain(m *testing.M) {
       	go serverMain()   // launch the server as a goroutine.
        time.Sleep(1 * time.Second) // one second is enough time for the server to start
}   

func TestCase(t *testing.T) {

	//Test server for a single client
	clientTest(t)

	//Test server for 100 clients
	for i:=0; i<100; i++{

		go func(){
			clientTest(t)
		}()

	}

	//Try more validation testcases
	clientTest2(t)
}

	

func clientTest(t *testing.T) {	
	
	//Connect to the Server
	tcpAddr, err := net.ResolveTCPAddr("tcp", servAddr)
	if err != nil {
		os.Exit(1)
	}

	conn, err = net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		os.Exit(1)
	}
	scanner := bufio.NewScanner(conn)




	name := "hi.txt"
	contents := "bye"
	exptime := 300000

	//#1 TestCase SUCCESSFUL WRITE
	fmt.Fprintf(conn, "write %v %v %v\r\n%v\r\n", name, len(contents), exptime, contents)
	scanner.Scan() // read first line
	resp := scanner.Text() // extract the text from the buffer
	arr := strings.Split(resp, " ") // split into OK and <version>
	expect(t, arr[0], "OK")
	version, err := strconv.ParseInt(arr[1], 10, 64) // parse version as number
	if err != nil {
		t.Error("Non-numeric version found")
	}

	//#2 TestCase SUCCESSFUL READ
	fmt.Fprintf(conn, "read %v\r\n", name) // try a read now
	scanner.Scan()

	arr = strings.Split(scanner.Text(), " ")
	expect(t, arr[0], "CONTENTS")
	expect(t, arr[1], fmt.Sprintf("%v", version)) // expect only accepts strings, convert int version to string
	expect(t, arr[2], fmt.Sprintf("%v", len(contents)))	
	scanner.Scan()
	expect(t, contents, scanner.Text())

	//#3 TestCase SUCCESSFUL WRITE without exptime
	fmt.Fprintf(conn, "write %v %v\r\n%v\r\n", name, len(contents), contents)
	scanner.Scan() // read first line
	resp = scanner.Text() // extract the text from the buffer
	arr = strings.Split(resp, " ") // split into OK and <version>
	expect(t, arr[0], "OK")
	version, err = strconv.ParseInt(arr[1], 10, 64) // parse version as number
	if err != nil {
		t.Error("Non-numeric version found")
	}

	//#4 TestCase CAS
	fmt.Fprintf(conn, "cas %v %v %v\r\n%v\r\n", name, version, len(contents), exptime, contents)
	scanner.Scan() // read first line
	resp = scanner.Text() // extract the text from the buffer
	arr = strings.Split(resp, " ") // split into OK and <version>
	for ;arr[0]!="ERR_VERSION"; {
		version, err = strconv.ParseInt(arr[1], 10, 64) // parse version as number
		if err != nil {
			t.Error("Non-numeric version found")
		}
		fmt.Fprintf(conn, "cas %v %v %v\r\n%v\r\n", name, version, len(contents), exptime, contents)
		scanner.Scan() // read first line
		resp = scanner.Text() // extract the text from the buffer
		arr = strings.Split(resp, " ") // split into OK and <version>
	}
	expect(t, arr[0], "OK")
	version, err = strconv.ParseInt(arr[1], 10, 64) // parse version as number
	if err != nil {
		t.Error("Non-numeric version found")
	}
	version_test_cas=append(version_test_cas,version)
}

	

func clientTest2(t *testing.T) {	
	
	//Connect to the Server
	tcpAddr, err := net.ResolveTCPAddr("tcp", servAddr)
	if err != nil {
		os.Exit(1)
	}

	conn, err = net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		os.Exit(1)
	}
	scanner := bufio.NewScanner(conn)

	//#4 TestCase READ version validation after CAS
	fmt.Fprintf(conn, "read %v\r\n", "hi.txt") // try a read now
	scanner.Scan()

	arr := strings.Split(scanner.Text(), " ")
	expect(t, arr[0], "CONTENTS")
	var max_version_cas int64
	for _,version := range version_test_cas {
		if(max_version_cas < version) {
			max_version_cas = version
		}
	}

	// expects file to have most updated version after concurrent CAS by clients
	expect(t, arr[1], fmt.Sprintf("%v", max_version_cas)) 
	scanner.Scan()


	// After execution of 4 testcases by multiple clients individually
	//#5 TestCase SUCCESSFUL DELETE
	fmt.Fprintf(conn, "delete %v\r\n", "hi.txt") // try a delete now on hi.txt
	scanner.Scan()

	str := scanner.Text()
	expect(t, str, "OK")

	//#6 TestCase UNSUCCESSFUL READ after DELETE
	fmt.Fprintf(conn, "read %v\r\n", "hi.txt") // try a read now on hi.txt
	scanner.Scan()

	str = scanner.Text()
	expect(t, str, "ERR_FILE_NOT_FOUND")

	//#7 TestCase UNSUCCESSFUL DELETE after DELETE
	fmt.Fprintf(conn, "delete %v\r\n", "hi.txt") // try a delete now on hi.txt
	scanner.Scan()

	str = scanner.Text()
	expect(t, str, "ERR_FILE_NOT_FOUND")

	//#8 TestCase BAD COMMAND
	fmt.Fprintf(conn, "del %v\r\n", "hi.txt") // try a delete now on hi.txt
	scanner.Scan()

	str = scanner.Text()
	expect(t, str, "ERR_CMD_ERR")
}


// Useful testing function
func expect(t *testing.T, a string, b string) {
	if a != b {
		t.Error(fmt.Sprintf("Expected %v, found %v", b, a)) // t.Error is visible when running `go test -verbose`
	}
}
