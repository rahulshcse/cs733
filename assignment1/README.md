# Assignment 1- File Server
===========================

## Description

This assignment is to build a simple file server, with a simple read/write interface (there’s no open/delete/rename). The file contents are in memory. There are two features that this file system has that traditional file systems don’t.<br/>
1. Each file has a version, and the API supports a `compare and swap` operation based on the version.<br/>
2. Files can optionally expire after some time.

# Installation Instructions

<code>go get github.com/rahulshcse/cs733/assignment1</code>

Two files are supposed to be there <br/>
1. `fileserver.go` contains the code where all the commands are implemented and where server listens to the request <br/>
2. `fileserver_test.go` contains all the test cases including commands which are fired concurrently evaluating all the necessary scenarios

To run the program only below command is needed (assuming the current directory is set to the assignment1 which has the go files) 
<br/><code>go test</code>


### Protocol Specification

* Write: create a file, or update the file’s contents if it already exists.

  `write <filename> <numbytes> [<exptime>]\r\n`
  
  `<content bytes>\r\n`

  The server responds with:

  `OK <version>\r\n`

  where version is a unique 64‐bit number (in decimal format) assosciated with the filename.

* Read: Given a filename, retrieve the corresponding file.

  `read <filename>\r\n`

  The server responds with the following format (or one of the errors described later)

  `CONTENTS <version> <numbytes> <exptime> \r\n`
  
  `<content bytes>\r\n`

* Compare and swap. This replaces the old file contents with the new content provided the version is still the same.

  `cas <filename> <version> <numbytes> [<exptime>]\r\n`
  
  `<content bytes>\r\n`

  The server responds with the new version if successful (or one of the errors described later)

  `OK <version>\r\n`

* Delete file.

  `delete <filename>\r\n`

  Server response (if successful)

  `OK\r\n`
  
### Included Test Cases

* Write to a file with expire time.
* Read from a file.
* Write to a file without expire time.
* Compare and Swap a file content.
* Delete file which exists.
* Unsuccessfully read from file which does not exist.
* Unsuccessfully try deleting file which does not exist.
* Test response to bad command.
* Concurrency Test(tests 1 to 4) for 100 clients

## Errors Returned

* `ERR_CMD_ERR` - Command line Formatting Error
* `ERR_FILE_NOT_FOUND` - File Not Found Error
* `ERR_VERSION` - Version Mismatch Error
* `ERR_INTERNAL` - Internal Error
