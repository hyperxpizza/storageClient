package storageClient

import (
	"encoding/json"
	"fmt"
	"sync"
)

type BulkFileError struct {
	FailedFiles map[string]error
	mutex       sync.Mutex
	msgBase     string
}

func newBulkFileError(msg string) *BulkFileError {
	return &BulkFileError{
		FailedFiles: map[string]error{},
		mutex:       sync.Mutex{},
		msgBase:     msg,
	}
}

func (err *BulkFileError) addFailedFile(filename string, reason error) {
	err.mutex.Lock()
	defer err.mutex.Unlock()
	err.FailedFiles[filename] = reason
}

func (err *BulkFileError) Error() string {
	errData, _ := json.MarshalIndent(err.FailedFiles, "", " ")
	msg := fmt.Sprintf("%s \n %s \n", err.msgBase, string(errData))
	return msg
}
