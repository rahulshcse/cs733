package main

import (
    "net"
    "strings"
    "strconv"
    "time"
    "sync"
    "log"
    "bytes"
)

const (
    HOST = "localhost"
    PORT = "8080"
    TYPE = "tcp"
)

//structure declaration for in-memory file
type Fileserve struct {
  content []byte
  numbytes int64
  version int64
  exptime int64
  lastLived time.Time
}

 //in memory file store using map
 var filestore =make(map[string]Fileserve)

 //version which auto increments every time a write or cas request is being successfully processed
 var ver int64 = 512
 var maxfilenamesize int64 = 250

 //mutex to be used for RW operation
 var mutex = &sync.RWMutex{}

func main() {
	serverMain()
}
func serverMain() {

   // Listen for incoming connections
    l, conn_error := net.Listen(TYPE, HOST+":"+PORT)
    if conn_error != nil { // Error check
        log.Print("Error listening:", conn_error.Error())
    }
    // Close the listener at end
    defer l.Close()
    

    for {
        // Accept incoming connection
        conn, conn_error := l.Accept()
        if conn_error != nil {
            log.Print("Error accepting: ", conn_error.Error())
        }

        // Handle connections in a new goroutine
        go handleRequest(conn)
    }
}

// reads data from in-memory file
func read(conn net.Conn,commands []string) int64 {
	var genError error
	// delete expired in-memory file
	checkTimeStamp()  
  	filename:=strings.TrimSpace(commands[0])
  	mutex.RLock()  
 	// data read
    	m_instance:=filestore[filename]
  	mutex.RUnlock()
	// if in-memory file does not exist
  	if(m_instance.version==0) { 
      		_, genError=conn.Write([]byte("ERR_FILE_NOT_FOUND\r\n"))
		return 0
    	} else { // send data to client
		message:=[]byte("CONTENTS "+strconv.FormatInt(m_instance.version,10) +" "+strconv.FormatInt(m_instance.numbytes,10) +" "+strconv.FormatInt(m_instance.exptime,10) +"\r\n")
		message=append(message, (m_instance.content)...)
		message=append(message, ([]byte("\r\n"))...)
      		_, genError=conn.Write(message)
	  	if(checkError(genError, conn)==-1) { //error check
	  		return -1
	  	}
    	}
	return 0
}

// writes data to in-memory file
func write(conn net.Conn,commands []string,data []byte) int64 {
	var genError error
	var numbytes int64
	var exptime int64
	var lastLived time.Time
	// delete expired in-memory file
	checkTimeStamp() 
	// expire-time included in command
	if(len(commands)==3){ 
		 exptime,genError= strconv.ParseInt(commands[2],10,64)
		if(checkError(genError, conn)==-1) { // error check
			return -1
		}
		 lastLived=time.Now().Add(time.Duration(exptime)*time.Second)
	} else if(len(commands)==2) { // expire-time not included in command
	} else { // less number of arguments then needed
		    _, genError=conn.Write([]byte("ERR_CMD_ERR\r\n"))
			return -1
	}
	filename:= strings.TrimSpace(commands[0])
	numbytes,genError= strconv.ParseInt(commands[1],10,64)
	if(checkError(genError, conn)==-1) { // error check
		return -1
	}
	// version increment
	ver+=1 
	// if datasize<number of bytes in command, or "\r\n" not at end of data
	if(int64(len(data))<(numbytes+2) || (!bytes.Equal(data[numbytes:numbytes+2],[]byte("\r\n")))) {
		    	conn.Write([]byte("ERR_CMD_ERR\r\n"))
			return -1
	}
	m_instance:= Fileserve{ 
	data[:numbytes],
	numbytes,
	ver,
	exptime,
	lastLived,
	}

	mutex.Lock()
	filestore[filename]=m_instance // create in-memory file
	mutex.Unlock()  
	// success
	_, genError=conn.Write([]byte("OK "+strconv.FormatInt(ver,10)+"\r\n"))
	if(checkError(genError, conn)==-1) { // error check
		return -1
	}
	// return offset to process next command 
	return int64(numbytes+2)
}

// compare version and swap data of in-memory file
func cas(conn net.Conn,commands []string,data []byte) int64 {
	var genError error
	var version int64
	var numbytes int64
	var exptime int64
	// delete expired in-memory file
	checkTimeStamp()
	var lastLived time.Time
	// expire-time included in command
	if(len(commands)==4){
		 exptime,genError= strconv.ParseInt(commands[3],10,64)
		if(checkError(genError, conn)==-1) { // error check
			return -1
		}
		 lastLived=time.Now().Add(time.Duration(exptime)*time.Second)
	} else if(len(commands)==3) { // expire-time not included in command
	} else { // less number of arguments then needed
		    _, genError=conn.Write([]byte("ERR_CMD_ERR\r\n"))
			return -1
	}
	filename:= strings.TrimSpace(commands[0])
	version,genError= strconv.ParseInt(commands[1],10,64)
	if(checkError(genError, conn)==-1) { // error check
		return -1
	}
	numbytes,genError= strconv.ParseInt(commands[2],10,64)
	if(checkError(genError, conn)==-1) { // error check
		return -1
	}
	if(filestore[filename].version==0) { // file not found
          	conn.Write([]byte("ERR_FILE_NOT_FOUND\r\n"))
		return -1
	} else if(filestore[filename].version!=version) { // version mismatch
		_, _=conn.Write([]byte("ERR_VERSION "+strconv.FormatInt(filestore[filename].version,10)+"\r\n"))
		return int64(numbytes+2)
      	}
	// version increment
	ver+=1
	// if datasize<number of bytes in command, or "\r\n" not at end of data
	if(int64(len(data))<(numbytes+2) || (!bytes.Equal(data[numbytes:numbytes+2],[]byte("\r\n")))) {
		    	conn.Write([]byte("ERR_CMD_ERR\r\n"))
			return -1
	}
	m_instance:= Fileserve{
	data[:numbytes],
	numbytes,
	ver,
	exptime,
	lastLived,
	}

	mutex.Lock()
	filestore[filename]=m_instance // swap data of in-memory file
	mutex.Unlock() 
	// success   
	_, genError=conn.Write([]byte("OK "+strconv.FormatInt(ver,10)+"\r\n"))
	if(checkError(genError, conn)==-1) { // error check
		return -1
	}
	// return offset to process next command 
	return int64(numbytes+2)
}

// delete in-memory file
func deleteEntry(conn net.Conn,commands []string) int64 {
	// delete expired in-memory file
	checkTimeStamp()
  	filename:=strings.TrimSpace(commands[0])
	var genError error
    	m_instance:=filestore[filename]

  	// checking whether in-memory file exists
  	if(m_instance.version==0) {
      		_, genError=conn.Write([]byte("ERR_FILE_NOT_FOUND\r\n"))
		return 0
    	} else {
          	_, genError=conn.Write([]byte("OK\r\n"))
	  	if(checkError(genError, conn)==-1) { // error check
	  		return -1
	  	}
  		mutex.Lock() 
		//delete file
          	delete(filestore,filename)
  		mutex.Unlock()
    	}
	return 0
}

// delete expired in-memory file
func checkTimeStamp() {
    
    for filename, content := range filestore {
        
        now:=time.Now()
        
	// check whether current time exdeeded expire-time of file
        if(now.After(content.lastLived) && content.exptime!=0) {
            mutex.Lock()  
		// delete file
              delete(filestore,filename)
            mutex.Unlock()  
        }
    }
}

// handle per connection file read-write requests
func handleRequest(conn net.Conn) {
	
    	// Close connection at end
	defer conn.Close()
	for {
		var genError error
		var bufsize int
		buffer := make([]byte, 1024)
		bufsize, genError = conn.Read(buffer) // read into buffer
		if(checkError(genError, conn)==-1){ // error check
			conn.Close()
			return
		}
		buffer= buffer[:bufsize]
		for ;!bytes.Equal(buffer[bufsize-2:bufsize],[]byte("\r\n"));{
			tempbuffer := make([]byte, 1024)
			var tempbufsize int
			tempbufsize, genError = conn.Read(tempbuffer)
			tempbuffer=tempbuffer[:tempbufsize]
			buffer=append(buffer,tempbuffer...)
			bufsize=len(buffer)
		}
		// while buffer not empty
		for bufsize>0{
		      	lineSeparator := bytes.SplitN(buffer,[]byte("\r\n"),2)
			buffer=lineSeparator[1]
			tempstr:=strings.TrimSpace(string(lineSeparator[0]))
		      	arrayOfCommands:= strings.Fields(tempstr)
			// delete expired in-memory file
		      	checkTimeStamp()
		      	var returnvalue int64
		      	var datasize int64
		      	if(arrayOfCommands[0]=="read") { // read request
				if(len(arrayOfCommands)<2) { // less arguments than expected
	    				_, genError=conn.Write([]byte("ERR_CMD_ERR\r\n"))
					returnvalue=-1
				}
			    	returnvalue=read(conn,arrayOfCommands[1:])
	
			} else if(arrayOfCommands[0]=="write") { // write request
				if(len(arrayOfCommands)<3) { // less arguments than expected
	    				_, genError=conn.Write([]byte("ERR_CMD_ERR\r\n"))
					returnvalue=-1
				}
				datasize,genError=strconv.ParseInt(arrayOfCommands[2],10,64)
				// fill buffer till datasize reached or exceeded
				for ;int64(len(buffer))<datasize;{
					tempbuffer := make([]byte, 1024)
					var tempbufsize int
					tempbufsize, genError = conn.Read(tempbuffer)
					tempbuffer=tempbuffer[:tempbufsize]
					buffer=append(buffer,tempbuffer...)
					bufsize=len(buffer)
					
				}
				// create in-memory file
			    	returnvalue=write(conn,arrayOfCommands[1:],buffer)

			} else if(arrayOfCommands[0]=="cas") { // compare & swap request
				if(len(arrayOfCommands)<4) { // less arguments than expected
	    				_, genError=conn.Write([]byte("ERR_CMD_ERR\r\n"))
					returnvalue=-1
				}
				datasize,genError=strconv.ParseInt(arrayOfCommands[3],10,64)
				// fill buffer till datasize reached or exceeded
				for ;int64(len(buffer))<datasize;{
					tempbuffer := make([]byte, 1024)
					var tempbufsize int
					tempbufsize, genError = conn.Read(tempbuffer)
					tempbuffer=tempbuffer[:tempbufsize]
					buffer=append(buffer,tempbuffer...)
					bufsize=len(buffer)
					
				}
				// compare version and swap data of in-memory file
			    	returnvalue=cas(conn,arrayOfCommands[1:],lineSeparator[1])

			} else if(arrayOfCommands[0]=="delete") { // delete request
				if(len(arrayOfCommands)<2) { // less arguments than expected
	    				_, genError=conn.Write([]byte("ERR_CMD_ERR\r\n"))
					returnvalue=-1
				}
				// delete in-memory file
			    	returnvalue=deleteEntry(conn,arrayOfCommands[1:]) 

			} else { // unknown request
	    			_, genError=conn.Write([]byte("ERR_CMD_ERR\r\n"))
				returnvalue=-1
			}
			if(returnvalue==-1){ // error situation
				conn.Close()
				return
			}
			//advance buffer to read next command
			buffer=buffer[returnvalue:]
			bufsize=len(buffer)
		}
	}
}

// error check
func checkError(genError error, conn net.Conn) int64 {
	if genError != nil {
		err := "ERR_INTERNAL\r\n"
		_, genError = conn.Write([]byte(err))
		return -1
	}
	return 0
}
