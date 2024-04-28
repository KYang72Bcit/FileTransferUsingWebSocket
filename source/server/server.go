package main

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"localhost/utils"
	"log"
	"net"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
)

const (
	trans = "tcp"
	bufferSize = 1024 * 1024 // 1MB
	arguments = 3
)
var shouldRun int32 = 1


// receiveInt reads 4 bytes from the provided reader and returns an integer.
// It assumes that the integer was encoded using big endian.
// It returns an error if the reader cannot be read from.
func receiveInt(reader *bufio.Reader) (int, error) {
	receivedByte := make([]byte, 4)
	_, err := reader.Read(receivedByte)
	if err != nil {
		return -1, err
	}
	receiveInt := binary.BigEndian.Uint32(receivedByte)

	return int(receiveInt), nil
}

// receiveBytes reads a byte array from the provided reader.
// It first reads 4 bytes to get the size of the byte array.
// It then reads the byte array in chunks of size bufferSize.
// It returns an error if the reader cannot be read from.
func receiveBytes(reader *bufio.Reader) ([]byte, error) {
	size, err := receiveInt(reader)
	if err != nil {
		return nil, err
	}
	data := make([]byte, size)
	received := 0

	for received < size {
		remaining := size - received
		readSize := bufferSize
		if remaining < readSize {
			readSize = remaining
		}

		n, err := reader.Read(data[received : received+readSize])
		if err != nil {
			return nil, err
		}

		received += n
	}

	return data, nil
}

// makeDir creates a directory with the provided name if it does not exist.
// It returns an error if the directory cannot be created.
func makeDir(storageDir string) error {
	if _, err := os.Stat(storageDir); os.IsNotExist(err) {
		err := os.Mkdir(storageDir, 0755)
		return err
	}
	return nil

}

// handleClient handles a client connection.
// It ignores any errors that occur while reading or writing.
func handleClient(conn net.Conn, storageDir string) {
	reader := bufio.NewReader(conn)
	numFiles, err := receiveInt(reader)
	utils.HandleError(err)
	for i := 0; i < numFiles; i++ {
		fileName, err := receiveBytes(reader)
		if err != nil {
			continue
		}

		fileContent, err := receiveBytes(reader)
		if err != nil {
			continue
		}

		file, err := os.Create(storageDir + "/" + string(fileName))
		if err != nil {
			continue
		}
		defer file.Close()
		file.Write(fileContent)

		fmt.Println("created file " + string(fileName) + " in " + storageDir)

	}
}

// handles the SIGINT signal for graceful showdown the server
func handleSig(sig chan os.Signal, ln net.Listener ){
	<-sig
	fmt.Println("\nServer Exiting...")
	atomic.StoreInt32(&shouldRun, 0)
	ln.Close()
	//os.Remove(socket)
	os.Exit(0)
}

//validateArgs validates the arguments passed to the server and parses them
// It returns the ip, port and storage directory if the arguments are valid
// It returns an error if the arguments are invalid
func validateArgs(args []string) (ip string, port string, storageDir string, err error){
	if len(args) != arguments {
		return "", "", "", errors.New("invalid number of arguments, <ip> <port> <storage Directory>")
	}

	ip = args[0]
	port = args[1]
	storageDir = args[2]

	return ip, port, storageDir, nil
}


func main() {

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT)
	ip, port, storageDir, err := validateArgs(os.Args[1:])
	utils.HandleFatalError(err)

	connetion := utils.ParseIP(ip, port)

	ln, err := net.Listen(trans, connetion)
	utils.HandleFatalError(err)

	defer ln.Close()
	log.Println("Server started tcp connection at  " + connetion)

	err = makeDir(storageDir)
	utils.HandleFatalError(err)

	go handleSig(sig, ln)

	for atomic.LoadInt32(&shouldRun) == 1{
	  	conn, err := ln.Accept()
		  if err != nil {
			// check for the "use of closed network connection" error
			if opErr, ok := err.(*net.OpError); ok && opErr.Op == "accept" {
				log.Println("Server close connection")
				break
			}
			utils.HandleError(err)
		}
		go handleClient(conn, storageDir)
	}

}
