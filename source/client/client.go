package main

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"localhost/utils"
	"net"
	"os"
)

const (
	trans = "tcp"
	bufferSize = 1024 * 1024
	arguments = 3
)
// sendInt encodes the provided integer using big endian and sends it to the provided writer
// It returns an error if the writer cannot be written to
// error will be nil if there's no error
func sendInt(writer *bufio.Writer, num int) (error) {
	intSend := int32(num)
	sendBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(sendBytes, uint32(intSend))
	_, err := writer.Write(sendBytes)
	if err != nil {
		return err
	}
	return writer.Flush()
}

// sendBytes sends the provided byte array to the provided writer
// It returns an int of the number of data it send, and error if the writer cannot be written to
// error will be nil if there's no error
func sendBytes(writer *bufio.Writer, data []byte) (int, error) {
	err := sendInt(writer, len(data))
	if err != nil {
		return -1, err
	}

	for start := 0; start < len(data); start += bufferSize {
		end := start + bufferSize
		if end > len(data) {
			end = len(data)
		}

		chunk := data[start:end]
		_, err := writer.Write(chunk)
		if err != nil {
			return -1, err
		}
		err = writer.Flush()
		if err != nil {
			return -1, err
		}
	}
	return len(data), nil
}

// processFiles sends the provided files to the provided writer
// It first sends the number of files as an integer using sendInt
// It then sends each file using sendBytes
// It prints the name of each file that is sent
func processFiles(files []string, writer *bufio.Writer){
	err := sendInt(writer, len(files))
	utils.HandleError(err)

	for _, fileName := range files {
		file, err := os.Open(fileName)
		if err != nil {
			utils.HandleError(err)
			continue
		}

		defer file.Close()

		fileInfo, err := file.Stat()
		if err != nil {
			utils.HandleError(err)
			continue
		}

		fname := []byte(fileInfo.Name())
		_, err = sendBytes(writer, fname)
		if err != nil {
			utils.HandleError(err)
			continue
		}

		fileData, err := os.ReadFile(fileName)
		if err != nil {
			utils.HandleError(err)
			continue
		}
		sendBytes(writer, fileData)

		fmt.Println("Sent file " + fileInfo.Name())

	}
}

//validates the provided arguments
//returns the ip, port, filenames and error
func validateArgs(args []string) (ip string, port string, filenames []string, err error){
	if len(args) < arguments {
		return "", "", nil, errors.New("invalid number of arguments, <ip> <port> <filename1>...<filenameN>")
	}

	ip = args[0]
	port = args[1]
	filenames = args[2:]

	return ip, port, filenames, nil
}



func main(){

	ip, port, filenames, err := validateArgs(os.Args[1:])
	utils.HandleFatalError(err)
	connetion := utils.ParseIP(ip, port)
	conn, err := net.Dial(trans, connetion)
	utils.HandleFatalError(err)
	defer conn.Close()

	writer := bufio.NewWriter(conn)


	processFiles(filenames, writer)

}
