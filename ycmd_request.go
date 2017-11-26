package main

import (
	"encoding/json"
	"fmt"
	"strings"
)

const (
	YcmdEventFileReadyToParse          = 1
	YcmdEventBufferUnload              = 2
	YcmdEventBufferVisit               = 3
	YcmdEventInsertLeave               = 4
	YcmdEventCurrentIdentifierFinished = 5
)

type YcmdRequest struct {
	LineNum          int
	ColumnNum        int
	Filepath         string
	FileContents     string
	Filetypes        []string
	CommandArguments []string
	CompleterTarget  string
}

func (r *YcmdRequest) MarshalJSON() ([]byte, error) {
	blob := map[string]interface{}{
		"line_num":   r.LineNum,
		"column_num": r.ColumnNum,
		"filepath":   r.Filepath,
		"file_data": map[string]interface{}{
			r.Filepath: map[string]interface{}{
				"filetypes": r.Filetypes,
				"contents":  r.FileContents,
			},
		},
	}
	if r.CommandArguments != nil {
		blob["command_arguments"] = r.CommandArguments
	}
	if r.CompleterTarget != "" {
		blob["completer_target"] = r.CompleterTarget
	}
	return json.Marshal(blob)
}

type FileLocation struct {
	LineNum     int    `json:"line_num"`
	ColumnNum   int    `json:"column_num"`
	Filepath    string `json:"filepath"`
	Description string `json:"description"`
}

func (y *FileLocation) String() string {
	return fmt.Sprintf("%s:%d:%d", y.Filepath, y.LineNum, y.ColumnNum)
}

func (y *FileLocation) DescriptionString() string {
	if len(y.Description) > 0 {
		return fmt.Sprintf("%s:%d:%s%s", y.Filepath, y.LineNum, strings.Repeat(" ", y.ColumnNum), y.Description)
	} else {
		return fmt.Sprintf("%s:%s", y.Filepath, y.LineNum)
	}
}

func (y *FileLocation) Addr() string {
	return fmt.Sprintf("%d-+#%d", y.LineNum, y.ColumnNum)
}

type FileLocations []FileLocation

func (ym FileLocations) String() string {
	options := make([]string, 0, len(ym))
	for _, y := range ym {
		options = append(options, y.DescriptionString())
	}
	return strings.Join(options, "\n")
}
