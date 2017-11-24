package main

import (
	"encoding/json"
	"testing"
)

func TestYcmdRequest_MarshalJSON(t *testing.T) {
	ycmdRequest := &YcmdRequest{
		LineNum:          5,
		ColumnNum:        17,
		Filepath:         "/tmp/wtf.py",
		FileContents:     SomeRandomPython,
		Filetypes:        []string{"python"},
		CommandArguments: []string{"GoTo"},
		CompleterTarget:  "python",
	}
	blob, err := json.Marshal(ycmdRequest)
	if err != nil {
		t.Log(err)
		t.FailNow()
	}
	t.Log(string(blob))
	if string(blob) != ExpectedJson {
		t.FailNow()
	}
}

const SomeRandomPython = `
from pygments.style import Style
from pygments.token import Keyword, Name, Comment, String


class IgorStyle(Style):
    """
    Pygments version of the official colors for Igor Pro procedures.
    """
    default_style = ""

    styles = {
        Comment:                'italic #FF0000',
        Keyword:                '#0000FF',
        Name.Function:          '#C34E00',
        Name.Decorator:         '#CC00A3',
        Name.Class:             '#007575',
        String:                 '#009C00'
    }
`
const ExpectedJson = `{"column_num":17,"command_arguments":["GoTo"],"completer_target":"python","file_data":{"/tmp/wtf.py":{"contents":"\nfrom pygments.style import Style\nfrom pygments.token import Keyword, Name, Comment, String\n\n\nclass IgorStyle(Style):\n    \"\"\"\n    Pygments version of the official colors for Igor Pro procedures.\n    \"\"\"\n    default_style = \"\"\n\n    styles = {\n        Comment:                'italic #FF0000',\n        Keyword:                '#0000FF',\n        Name.Function:          '#C34E00',\n        Name.Decorator:         '#CC00A3',\n        Name.Class:             '#007575',\n        String:                 '#009C00'\n    }\n","filetypes":["python"]}},"filepath":"/tmp/wtf.py","line_num":5}`
