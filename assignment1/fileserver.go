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

func read(conn net.Conn,commands []string,noReply *bool) {

}

func write(conn net.Conn,commands []string,noReply *bool) {

}

func cas(conn net.Conn,commands []string,noReply *bool) {

}

func deleteEntry(conn net.Conn,commands []string) {

}

func checkTimeStamp() {
    
    for key, content := range filestore {
        
        now:=time.Now()
        
        if(now.After(content.lastLived) && content.exptime!=0) {
            mutex.Lock()  
              delete(filestore,key)
            mutex.Unlock()  
        }
    }
}

func handleRequest(conn net.Conn) {

  defer conn.Close()
  
  for {
      buffer := make([]byte, 1024)
      bufsize, err := conn.Read(buffer)
      
      if err != nil {
        log.Print("Error reading:", err.Error())
      }

      buffer= buffer[:bufsize]

      commands := string(buffer)
      commands = strings.TrimSpace(commands)
      lineSeparator := strings.Split(commands,"\r\n")

      arrayOfCommands:= strings.Fields(lineSeparator[0])
      var newArrayOfCommands[] string
      if len(lineSeparator) >1 {
          newArrayOfCommands = make([] string,len(arrayOfCommands),len(arrayOfCommands)+1)
          copy(newArrayOfCommands,arrayOfCommands)
          newArrayOfCommands=append(newArrayOfCommands,lineSeparator[1])
      } else {
          newArrayOfCommands= make([] string,len(arrayOfCommands))
          copy(newArrayOfCommands,arrayOfCommands)
      }   

      checkTimeStamp()

      var noReply bool= false
      
      if(arrayOfCommands[0]=="read") {
            read(conn,newArrayOfCommands[1:],&noReply)

        } else if(arrayOfCommands[0]=="write") {
            write(conn,newArrayOfCommands[1:])

        } else if(arrayOfCommands[0]=="cas") {
            cas(conn,newArrayOfCommands[1:],&noReply)

        } else if(arrayOfCommands[0]=="delete") {
            deleteEntry(conn,newArrayOfCommands[1:]) 

        } else {
            conn.Write([]byte("ERRCMDERR\r\n"))
        }
    }
}
