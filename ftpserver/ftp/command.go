package ftp

import (
	"fmt"
)

type Command struct {
	Command string
	Other   string
}

func parseCommand(data []byte) (*Command, error) {
	cmd := &Command{}

	inCommand := true
	otherStart := 0
	for idx := 0; idx < len(data); idx++ {
		if inCommand {
			if data[idx] == ' ' {
				cmd.Command = string(data[0:idx])
				inCommand = false
				otherStart = idx + 1
				continue
			}
		} else {
			if data[idx] == '\r' && len(data) > idx+1 && data[idx+1] == '\n' {
				cmd.Other = string(data[otherStart:idx])
				return cmd, nil
			}
		}
	}

	return nil, fmt.Errorf("reached end")
}
