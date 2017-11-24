package main

import (
	"encoding/json"
	"io/ioutil"
)

type YcmdSettings struct {
	FilepathCompletionUseWorkingDir          int               `json:"filepath_completion_use_working_dir"`
	AutoTrigger                              int               `json:"auto_trigger"`
	MinNumOfCharsForCompletion               int               `json:"min_num_of_chars_for_completion"`
	MinNumIdentifierCandidateChars           int               `json:"min_num_identifier_candidate_chars"`
	SemanticTriggers                         map[string]string `json:"semantic_triggers"`
	FiletypeSpecificCompletionToDisable      map[string]int    `json:"filetype_specific_completion_to_disable"`
	SeedIdentifiersWithSyntax                int               `json:"seed_identifiers_with_syntax"`
	CollectIdentifiersFromCommentsAndStrings int               `json:"collect_identifiers_from_comments_and_strings"`
	CollectIdentifiersFromTagsFiles          int               `json:"collect_identifiers_from_tags_files"`
	MaxNumIdentifierCandidates               int               `json:"max_num_identifier_candidates"`
	MaxNumCandidates                         int               `json:"max_num_candidates"`
	ExtraConfGloblist                        []string          `json:"extra_conf_globlist"`
	GlobalYcmExtraConf                       string            `json:"global_ycm_extra_conf"`
	ConfirmExtraConf                         int               `json:"confirm_extra_conf"`
	CompleteInComments                       int               `json:"complete_in_comments"`
	CompleteInStrings                        int               `json:"complete_in_strings"`
	MaxDiagnosticsToDisplay                  int               `json:"max_diagnostics_to_display"`
	FiletypeWhitelist                        map[string]int    `json:"filetype_whitelist"`
	FiletypeBlacklist                        map[string]int    `json:"filetype_blacklist"`
	AutoStartCsharpServer                    int               `json:"auto_start_csharp_server"`
	AutoStopCsharpServer                     int               `json:"auto_stop_csharp_server"`
	UseUltiSnipsCompleter                    int               `json:"use_ultisnips_completer"`
	CsharpServerPort                         int               `json:"csharp_server_port"`
	HmacSecret                               string            `json:"hmac_secret"`
	ServerKeepLogfiles                       int               `json:"server_keep_logfiles"`
	GocodeBinaryPath                         string            `json:"gocode_binary_path"`
	GodefBinaryPath                          string            `json:"godef_binary_path"`
	RustSrcPath                              string            `json:"rust_src_path"`
	RacerdBinaryPath                         string            `json:"racerd_binary_path"`
	PythonBinaryPath                         string            `json:"python_binary_path"`
}

func NewYcmdSettingsFromFile(path string) (*YcmdSettings, error) {
	blob, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	ycmdSettings := new(YcmdSettings)
	err = json.Unmarshal(blob, ycmdSettings)
	if err != nil {
		return nil, err
	}
	return ycmdSettings, nil
}

/*
  "auto_start_csharp_server": 1,
  "auto_stop_csharp_server": 1,
  "use_ultisnips_completer": 1,
  "csharp_server_port": 0,
  "hmac_secret": "",
  "server_keep_logfiles": 0,
  "gocode_binary_path": "",
  "godef_binary_path": "",
  "rust_src_path": "",
  "racerd_binary_path": "",
  "python_binary_path": ""
{
  "filepath_completion_use_working_dir": 0,
  "auto_trigger": 1,
  "min_num_of_chars_for_completion": 2,
  "min_num_identifier_candidate_chars": 0,
  "semantic_triggers": {},
  "filetype_specific_completion_to_disable": {
    "gitcommit": 1
  },
  "seed_identifiers_with_syntax": 0,
  "collect_identifiers_from_comments_and_strings": 0,
  "collect_identifiers_from_tags_files": 0,
  "max_num_identifier_candidates": 10,
  "max_num_candidates": 50,
  "extra_conf_globlist": [],
  "global_ycm_extra_conf": "",
  "confirm_extra_conf": 1,
  "complete_in_comments": 0,
  "complete_in_strings": 1,
  "max_diagnostics_to_display": 30,
  "filetype_whitelist": {
    "*": 1
  },
  "filetype_blacklist": {
    "tagbar": 1,
    "qf": 1,
    "notes": 1,
    "markdown": 1,
    "netrw": 1,
    "unite": 1,
    "text": 1,
    "vimwiki": 1,
    "pandoc": 1,
    "infolog": 1,
    "mail": 1
  },
  "auto_start_csharp_server": 1,
  "auto_stop_csharp_server": 1,
  "use_ultisnips_completer": 1,
  "csharp_server_port": 0,
  "hmac_secret": "",
  "server_keep_logfiles": 0,
  "gocode_binary_path": "",
  "godef_binary_path": "",
  "rust_src_path": "",
  "racerd_binary_path": "",
  "python_binary_path": ""
}
*/
