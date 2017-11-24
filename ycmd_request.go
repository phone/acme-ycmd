package main

import "encoding/json"

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

type YcmdGoToSingleResponse struct {
	LineNum     int
	ColumnNum   int
	Filepath    string
	Description string
}

type YcmdGoToMultiResponse []YcmdGoToSingleResponse
