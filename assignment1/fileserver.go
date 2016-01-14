package main

import (
    "net"
    "strings"
    "strconv"
    "time"
    "sync"
    "log"
)

const (
    HOST = "localhost"
    PORT = "8080"
    TYPE = "tcp"
)

type Fileserve struct {
  content string
  numbytes int64
  version int64
  exptime int64
  lastLived time.Time
}

 var filestore =make(map[string]Fileserve)

 var ver int64 = 512
 var maxfilenamesize int64 = 250

 var mutex = &sync.RWMutex{}

func main() {
	serverMain()
}
func serverMain() {

    l, conn_error := net.Listen(TYPE, HOST+":"+PORT)
    if conn_error != nil {
        log.Print("Error listening:", err.Error())
    }
    defer l.Close()
    

    for {
        conn, conn_error := l.Accept()
        if conn_error != nil {
            log.Print("Error accepting: ", err.Error())
        }

        go handleRequest(conn)
    }
}

func read(conn net.Conn,commands []string) {
	var genError error
  filename:=strings.TrimSpace(commands[0])
  mutex.RLock()  
    m_instance:=filestore[filename]
  mutex.RUnlock()
  if(m_instance.version==0) {
      _, genError=conn.Write([]byte("ERR_FILE_NOT_FOUND\r\n"))
    } else {
      _, genError=conn.Write([]byte("CONTENTS "+strconv.FormatInt(m_instance.version,10) +" "+strconv.FormatInt(m_instance.numbytes,10) +" "+" "+strconv.FormatInt(m_instance.exptime,10) +"\r\n"+m_instance.content+"\r\n"))
    }
	checkError(genError, conn)
}

func write(conn net.Conn,commands []string) {
	var genError error
	filename:= strings.TrimSpace(commands[0])
	numbytes,_:= strconv.ParseInt(commands[1],10,64)
	var content string
	var exptime int64
	var lastLived time.Time
	if(len(commands)!=3){
	 exptime,_:= strconv.ParseInt(commands[2],10,64)
	 lastLived:=time.Now().Add(time.Duration(exptime)*time.Second)
	 content= strings.TrimSpace(commands[4])
	} else {
	 content= strings.TrimSpace(commands[3])
	}
	unique_version+=1

	m_instance:= Filestore{
	content,
	numbytes,
	unique_version,
	exptime,
	lastLived,
	}

	mutex.Lock()
	filestore[filename]=m_instance
	mutex.Unlock()  
	_, genError=conn.Write([]byte("OK "+strconv.FormatInt(unique_version,10)+"\r\n"))
	checkError(genError, conn)

}

func cas(conn net.Conn,commands []string) {
	var genError error
	filename:= strings.TrimSpace(commands[0])
	version,_:= strconv.ParseInt(commands[1],10,64)
	numbytes,_:= strconv.ParseInt(commands[2],10,64)
	var content string
	var exptime int64
	var lastLived time.Time
	if(len(commands)!=4){
	 exptime,_:= strconv.ParseInt(commands[2],10,64)
	 lastLived:=time.Now().Add(time.Duration(exptime)*time.Second)
	 content= strings.TrimSpace(commands[4])
	} else {
	 content= strings.TrimSpace(commands[3])
	}
	if(filestore[filename].version==0) {
          conn.Write([]byte("ERR_FILE_NOT_FOUND\r\n"))
	} else if(filestore[filename].version!=version) {
          conn.Write([]byte("ERR_VERSION\r\n"))
      	}
	else {
		unique_version+=1
		m_instance:= Filestore{
		content,
		numbytes,
		unique_version,
		exptime,
		lastLived,
		}

		mutex.Lock()
		filestore[filename]=m_instance
		mutex.Unlock()  
		_, genError=conn.Write([]byte("OK "+strconv.FormatInt(unique_version,10)+"\r\n"))
		checkError(genError, conn)
	}
}

func deleteEntry(conn net.Conn,commands []string) {
  filename:=strings.TrimSpace(commands[0])
	var genError error
  mutex.Lock() 
    m_instance:=filestore[filename]

  //checking if the version is zero then that means there is no such value in the map
  if(m_instance.version==0) {
      _, genError=conn.Write([]byte("ERR_FILE_NOT_FOUND\r\n"))
    } else {
          delete(filestore,filename)
          _, genError=conn.Write([]byte("OK\r\n"))
    }
  mutex.Unlock()
	checkError(genError, conn)
}

func checkTimeStamp() {
    
    for filename, content := range filestore {
        
        now:=time.Now()
        
        if(now.After(content.lastLived) && content.exptime!=0) {
            mutex.Lock()  
              delete(filestore,filename)
            mutex.Unlock()  
        }
    }
}

func handleRequest(conn net.Conn) {

  defer conn.Close()
  
  for {
	var genError error
      buffer := make([]byte, 1024)
      bufsize, genError := conn.Read(buffer)
	checkError(genError, conn)

      buffer= buffer[:bufsize]

      commands := string(buffer)
      commands = strings.TrimSpace(commands)
      lineSeparator := strings.Split(commands,"\r\n")

      arrayOfCommands:= strings.Fields(lineSeparator[0])
      var newArrayOfCommands[] string
	  newArrayOfCommands = make([] string,len(arrayOfCommands),len(arrayOfCommands)+1)
	  copy(newArrayOfCommands,arrayOfCommands)
	  newArrayOfCommands=append(newArrayOfCommands,lineSeparator[1])

      checkTimeStamp()
      
      if(arrayOfCommands[0]=="read") {
            read(conn,newArrayOfCommands[1:])

        } else if(arrayOfCommands[0]=="write") {
            write(conn,newArrayOfCommands[1:])

        } else if(arrayOfCommands[0]=="cas") {
            cas(conn,newArrayOfCommands[1:])

        } else if(arrayOfCommands[0]=="delete") {
            deleteEntry(conn,newArrayOfCommands[1:]) 

        } else {
            _, genError=conn.Write([]byte("ERR_CMD_ERR\r\n"))
		checkError(genError, conn)
        }
    }
}
func checkError(genError error, conn net.Conn) {
	if genError != nil {
		err := "ERR_INTERNAL\r\n"
		_, genError = conn.Write([]byte(err))
	}
}
