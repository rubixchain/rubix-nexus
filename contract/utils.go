package contract

import (
	"encoding/json"
	"fmt"
	"os"
)

func parseContractMsgFromJSON(contractMsgFile string) (string, error) {
	// Check if the path is relative or absolute
	if !os.IsPathSeparator(contractMsgFile[0]) && contractMsgFile[0] != '.' {
		// If it's a relative path, convert it to an absolute path
		absPath, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("failed to get current working directory: %w", err)
		}
		contractMsgFile = absPath + string(os.PathSeparator) + contractMsgFile
	}

	file, err := os.Open(contractMsgFile)
	if err != nil {
		return "", fmt.Errorf("failed to open contract message file: %w", err)
	}
	defer file.Close()

	var contractMsgIntf map[string]interface{}
	if err := json.NewDecoder(file).Decode(&contractMsgIntf); err != nil {
		return "", fmt.Errorf("failed to decode contract message: %w", err)
	}

	contractMsgBytes, err := json.Marshal(contractMsgIntf)
	if err != nil {
		return "", fmt.Errorf("failed to marshal contract message: %w", err)
	}


	return string(contractMsgBytes), nil
}
