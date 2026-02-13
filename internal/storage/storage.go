package storage

import (
	"bufio"
	"os"
	"sync"
)

type FileManager struct {
	mu sync.Mutex
	baseDir string
}
func NewFileManager(baseDir string) *FileManager {
	if _, err := os.Stat(baseDir); os.IsNotExist(err) {
		_ = os.Mkdir(baseDir, 0755)
	}
	return &FileManager{
		baseDir: baseDir,
	}
}

func (fm *FileManager) SaveMessage(roomName string, message []byte) error {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	filePath := fm.baseDir + "/" + roomName + ".log"
	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := f.Write(message); err != nil {
		return err
	}
	if _, err := f.WriteString("\n"); err != nil {
		return err
	}

	return nil
}

func (fm *FileManager) LoadHistory(roomName string) ([][]byte, error) {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	filePath := fm.baseDir + "/" + roomName + ".log"
	f, err := os.Open(filePath)
	if os.IsNotExist(err) {
		return nil, nil 
	}
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var history [][]byte
	scanner := bufio.NewScanner(f)
	

	for scanner.Scan() {
		line := scanner.Bytes()
		msgCopy := make([]byte, len(line))
		copy(msgCopy, line)
		history = append(history, msgCopy)
	}

	return history, scanner.Err()
}