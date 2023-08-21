package logging

import (
	"log"
	"os"
	"sync"
)

var m sync.Mutex

var logFile string

// LogToFile Log to file
func LogToFile(line []byte) {
	if logFile == "" {
		return
	}

	m.Lock()
	defer m.Unlock()

	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		log.Printf("Failed to open log file: %s", err)
		return
	}

	defer f.Close()
	if _, err = f.WriteString(string(line) + "\n"); err != nil {
		log.Printf("Failed to write log file: %s", err)
	}
}

func SetLogFile(path string) {
	logFile = path
}
