package utils

import (
	"fmt"
	"log"
	"strings"
)

// HandleFatalError is a utility function to handle fatal errors.
// If the provided error is not nil, it logs the error and exits the program using log.Fatalln.
func HandleFatalError(err error) {
	if err != nil {
		log.Fatalln("Fatal Error:", err)
	}
}

// HandleError is a utility function to handle non-fatal errors.
// If the provided error is not nil, it prints the error message to the console using fmt.Println.
func HandleError(err error) {
	if err != nil {
		fmt.Println("Error:", err)
	}
}

func ParseIP(ip string, port string) string {
	if strings.Contains(ip, ":") {
		return "[" + ip + "]:" + port
	}
	return ip + ":" + port
}
