# Assignment 4-  Replicated File System using Raft
===================================================

# Description

This assignment combines file system of the first assignment with the raft node from the third assignment. File system code used for this assignment has been taken from the following link<br/>
https://github.com/cs733-iitb/cs733/tree/master/assignment1<br/>


## File System Description

1. Each file server manages a separate in-memory file system.<br/>
2. Each file server is supported by Raft node. Raft layer underlying servers replicates logs containing client commands processed by the file servers.<br/>
2. Each file server configuration contains configuration of the raft node. It also contains client port so that clients can connect to file server.
3. Error messages used by file system includes `ERR_REDIRECT <new leader URL>`, in addition to other file system errors of Assignment 1. The leader URL has the host and port of the other leader. Client retries connection when ERR_REDIRECT response is received.<br/>
4. Client connections are established, and remain valid, only till underlying Raft node supporting file server is a leader.<br/>
5. The file system is logically in two parts, the front end client handler and the back-end store. The client handler goroutine
receives the command, parses and validates it and Appends write operations to its raft node. A channel is assigned per client handler by frontend, used by it to get reply message from backend. Both reads and writes are replicated.<br/>
6. The backend file store (the map) runs in a separate goroutine, waiting for events from the commit channel. It processes commands
and send replies to the appropriate client channel maintained by frontend. Frontend then forwards the message to appropriate client.<br/>
7. The command is tagged with a client id by frontend so that the backend knows which client handler to send the reply to.<br/>
8. Backend file handler remembers which indexes have been consumed. Backend ensures that the log is consumed in increasing log index numbers, and accounts for missing indexes.


## Installation Instructions

<code>go get github.com/rahulshcse/cs733/assignment4</code>

To run the program below command is needed (assuming the current directory is set to the assignment2 which has the go files) <br/>
<code>go test</code>
  
  
### Included Test Cases

* Sequential Client RPC execution.
* Sequential Client RPC execution after a node failure.
* Sequential Client RPC execution during a node recovery.
* Binary client RPC test.
* Client RPC test in chunks.
* Concurrent writes by multiple clients.
