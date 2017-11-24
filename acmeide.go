package main

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"

	"bytes"

	"9fans.net/go/acme"
	"github.com/phayes/freeport"
)

const HmacSecretLength = 16
const HmacHeaderName = "X-Ycm-Hmac"

var currentSettings *YcmdSettings
var currentSettingsLock sync.Mutex
var currentPort string
var currentPortLock sync.Mutex

func Python() string {
	pyPath, err := exec.LookPath("python")
	if err != nil {
		log.Fatal(err)
	}
	return pyPath
}

func GenerateHmacSecret() string {
	bs := make([]byte, HmacSecretLength)
	_, err := rand.Read(bs)
	if err != nil {
		log.Fatal(err)
	}
	secret := base64.StdEncoding.EncodeToString(bs)
	return string(secret)
}

func WriteNamedTemporaryFileOf(contents string) string {
	f, err := ioutil.TempFile("", "acmeide")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	if _, err = f.Write([]byte(contents)); err != nil {
		log.Fatal(err)
	}
	log.Println("Wrote temporary settings file: " + f.Name())
	return f.Name()
}

func DefaultSettings() *YcmdSettings {
	ycmdSettings, err := NewYcmdSettingsFromFile("./default_settings.json")
	if err != nil {
		log.Fatal(err)
	}
	return ycmdSettings
}

func UpdateCurrentSettings(settings *YcmdSettings) {
	currentSettingsLock.Lock()
	defer currentSettingsLock.Unlock()
	currentSettings = settings
}

func SetCurrentPort(port string) {
	currentPortLock.Lock()
	defer currentPortLock.Unlock()
	currentPort = port
}

func GetCurrentPort() string {
	currentPortLock.Lock()
	defer currentPortLock.Unlock()
	port := currentPort
	return port
}

func GetSettingsJson() string {
	currentSettingsLock.Lock()
	defer currentSettingsLock.Unlock()
	bytes, err := json.Marshal(currentSettings)
	if err != nil {
		log.Fatal(err)
	}
	return string(bytes)
}

func SetHmacSecret(hmacSecret string) {
	currentSettingsLock.Lock()
	defer currentSettingsLock.Unlock()
	currentSettings.HmacSecret = hmacSecret
}

func GetHmacSecret() string {
	currentSettingsLock.Lock()
	defer currentSettingsLock.Unlock()
	return currentSettings.HmacSecret
}

func StartAndWaitForYcmd(pathToYcmd string) {
	port, err := freeport.GetFreePort()
	if err != nil {
		log.Fatal(err)
	}
	SetCurrentPort(strconv.Itoa(port))
	optionsFile := WriteNamedTemporaryFileOf(GetSettingsJson())
	cmd := exec.Command(
		Python(),
		pathToYcmd,
		fmt.Sprintf("--port=%s", GetCurrentPort()),
		fmt.Sprintf("--options_file=%s", optionsFile),
		fmt.Sprintf("--idle_suicide_seconds=%s", "300"),
		fmt.Sprintf("--log=debug"),
		fmt.Sprint("--keep_logfiles"),
		fmt.Sprintf("--stdout=/tmp/ycmd-out.log"),
		fmt.Sprintf("--stderr=/tmp/ycmd-err.log"),
	)
	cmd.Start()
	log.Printf("Started Ycmd on port %s with options file %s\n", GetCurrentPort(), optionsFile)
	cmd.Wait()
}

func YcmdForever(pathToYcmd string) {
	StartAndWaitForYcmd(pathToYcmd)
	//	for {
	//
	//	}
}

func CreateHmac(content, secret []byte) []byte {
	mac := hmac.New(sha256.New, secret)
	mac.Write(content)
	return mac.Sum(nil)
}

func CreateRequestHmac(method, path, body, hmacSecret string) []byte {
	log.Println("Hmacing " + method + " " + path)
	secret, err := base64.StdEncoding.DecodeString(hmacSecret)
	if err != nil {
		log.Fatal(err)
	}
	methodHmac := CreateHmac([]byte(method), secret)
	pathHmac := CreateHmac([]byte(path), secret)
	bodyHmac := CreateHmac([]byte(body), secret)
	joinedHmac := make([]byte, 0, len(methodHmac)+len(pathHmac)+len(bodyHmac))
	joinedHmac = append(joinedHmac, methodHmac...)
	joinedHmac = append(joinedHmac, pathHmac...)
	joinedHmac = append(joinedHmac, bodyHmac...)
	log.Println("Hmac bytes: " + strconv.Itoa(len(joinedHmac)))
	return CreateHmac(joinedHmac, secret)
}

func ExtraHeaders(req *http.Request, method, path, body string) {
	hmacSecret := GetHmacSecret()
	requestHmacBytes := CreateRequestHmac(method, path, body, hmacSecret)
	requestHmac := base64.StdEncoding.EncodeToString(requestHmacBytes)
	req.Header.Add(HmacHeaderName, requestHmac)
	req.Header.Add("Content-Type", "application/json")
}

func CreateRequestForGetHandler(handler string) (*http.Request, error) {
	rawurl := fmt.Sprintf("http://localhost:%s/%s", GetCurrentPort(), handler)
	uri, err := url.Parse(rawurl)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("GET", rawurl, nil)
	if err != nil {
		return nil, err
	}
	ExtraHeaders(req, "GET", uri.Path, "")
	return req, nil
}

func CreateRequestForPostHandler(handler string, body []byte) (*http.Request, error) {
	rawurl := fmt.Sprintf("http://localhost:%s/%s", GetCurrentPort(), handler)
	uri, err := url.Parse(rawurl)
	if err != nil {
		return nil, err
	}
	body2 := make([]byte, len(body))
	copy(body2, body)
	req, err := http.NewRequest("POST", rawurl, bytes.NewBuffer(body2))
	if err != nil {
		return nil, err
	}
	log.Println("BODY2: " + string(body2))
	ExtraHeaders(req, "POST", uri.Path, string(body))
	return req, nil
}

func GetHandler(handler string) (interface{}, error) {
	req, err := CreateRequestForGetHandler(handler)
	if err != nil {
		return nil, err
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	log.Println(resp.Status)
	defer resp.Body.Close()
	blob, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var i interface{}
	err = json.Unmarshal(blob, &i)
	if err != nil {
		return nil, err
	}
	return i, nil
}

func PostHandler(handler string, request *YcmdRequest) (interface{}, error) {
	body, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}
	log.Println(string(body))
	req, err := CreateRequestForPostHandler(handler, body)
	if err != nil {
		return nil, err
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	log.Println(resp.Status)
	defer resp.Body.Close()
	blob, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	log.Printf("Raw Ycmd Response: %s\n", string(blob))
	var i interface{}
	err = json.Unmarshal(blob, &i)
	if err != nil {
		return nil, err
	}
	return i, nil
}

func IsReady() bool {
	var data interface{}
	var err error
	tick := time.Tick(5 * time.Second)
	for range tick {
		data, err = GetHandler("ready")
		if err != nil {
			log.Println(err)
		} else {
			break
		}
	}
	return data.(bool)
}

const PythonTag = "Def Refs Fwd Bck In Out Up Dwn Find Case"

type WindowType int

const (
	NewAcmeWindow    WindowType = iota
	UnknownWindow    WindowType = iota
	CcWindow         WindowType = iota
	PythonWindow     WindowType = iota
	JavaWindow       WindowType = iota
	GoWindow         WindowType = iota
	JavascriptWindow WindowType = iota
	RustWindow       WindowType = iota
	CSharpWindow     WindowType = iota
	DirectoryWindow  WindowType = iota
	WinWindow        WindowType = iota
)

var windowSuffixesLock sync.Mutex
var windowSuffixes = map[string]WindowType{
	".py":  PythonWindow,
	".c":   CcWindow,
	".cpp": CcWindow,
	".cc":  CcWindow,
	".C":   CcWindow,
	".h":   CcWindow,
	".hh":  CcWindow,
	".H":   CcWindow,
	".hpp": CcWindow,
	"/":    DirectoryWindow,
}

func WindowSuffixes() (map[string]WindowType, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}
	winSuffix := fmt.Sprintf("-%s", hostname)

	windowSuffixesLock.Lock()
	defer windowSuffixesLock.Unlock()
	if _, ok := windowSuffixes[winSuffix]; !ok {
		windowSuffixes[winSuffix] = WinWindow
	}
	return windowSuffixes, nil
}

func DetermineWindowType(winName string) WindowType {
	if winName == "" {
		return NewAcmeWindow
	}
	winSuffixes, err := WindowSuffixes()
	if err != nil {
		log.Printf("DetermineWindowType: %s: %s", winName, err.Error())
		return UnknownWindow
	}
	for suffix, winType := range winSuffixes {
		if strings.HasSuffix(winName, suffix) {
			return winType
		}
	}
	return UnknownWindow
}

type AcmeButton int

const (
	AcmeButtonError AcmeButton = iota
	AcmeButtonTwo   AcmeButton = iota
	AcmeButtonThree AcmeButton = iota
)

type AcmeArea int

const (
	AcmeAreaError AcmeArea = iota
	AcmeAreaTag   AcmeArea = iota
	AcmeAreaBody  AcmeArea = iota
)

func WhichAcmeArea(e *acme.Event) (AcmeArea, error) {
	if unicode.IsUpper(e.C2) {
		return AcmeAreaBody, nil
	} else if unicode.IsLower(e.C2) {
		return AcmeAreaTag, nil
	} else {
		return AcmeAreaError, errors.New("unknown area")
	}
}

func WhichAcmeButton(e *acme.Event) (AcmeButton, error) {
	if e.C1 == 'M' && (e.C2 == 'x' || e.C2 == 'X') {
		return AcmeButtonTwo, nil
	} else if e.C1 == 'M' && (e.C2 == 'l' || e.C2 == 'L') {
		return AcmeButtonThree, nil
	} else {
		return AcmeButtonError, errors.New("unknown acme button")
	}
}

func GetAcmeWindowBody(a *acme.Win) (string, error) {
	body, err := a.ReadAll("data")
	if err != nil {
		return "", err
	}
	return string(body), nil
}

type LineAndColumn struct {
	Line   int
	Column int
}

func GetAcmeWindowLineAndColumn(a *acme.Win, body string) (*LineAndColumn, error) {
	err := a.Ctl("addr=dot")
	if err != nil {
		return nil, err
	}
	q0, q1, err := a.ReadAddr()
	if err != nil {
		return nil, err
	}
	log.Printf("Acme Window Addr: %d, %d", q0, q1)
	lastLinePosition := 0
	lineAndColumn := &LineAndColumn{Line: 1}
	for i := 0; i < q0; i++ {
		if body[i] == '\n' {
			lastLinePosition = i
			lineAndColumn.Line++
		}
	}
	lineAndColumn.Column = q0 - lastLinePosition
	return lineAndColumn, nil
}

type Ide interface {
	// Do all the preparation needed to perform the Watch. Should be called
	// before Watch.
	Setup() error

	// Watch for events continuously until the window is closed.
	Watch()
}

type PythonIde struct {
	Id      int
	Name    string
	isSetup bool
	acmeWin *acme.Win
}

func (p *PythonIde) hasIdeTag() (bool, error) {
	var currentTag []byte
	currentTag, err := p.acmeWin.ReadAll("tag")
	if err != nil {
		return false, err
	}
	return strings.Contains(string(currentTag), PythonTag), nil
}

func (p *PythonIde) setupIdeTag() error {
	_, err := p.acmeWin.Write("tag", []byte(PythonTag))
	if err != nil {
		return err
	}
	return nil
}

func (p *PythonIde) Setup() error {
	var err error
	p.acmeWin, err = acme.Open(p.Id, nil)
	if err != nil {
		return err
	}
	hasIdeTag, err := p.hasIdeTag()
	if err != nil {
		return err
	}
	if !hasIdeTag {
		p.setupIdeTag()
	}
	return nil
}

var PythonIdeCommands = map[string]struct{}{
	"Def":  {},
	"Refs": {},
	"Fwd":  {},
	"Bck":  {},
	"In":   {},
	"Out":  {},
	"Up":   {},
	"Dwn":  {},
	"Find": {},
	"Case": {},
}

type IdeCommand struct {
	Command string
	Area    AcmeArea
	Button  AcmeButton
}

func NewIdeCommand(e *acme.Event) *IdeCommand {
	area, _ := WhichAcmeArea(e)
	button, _ := WhichAcmeButton(e)
	ideCommand := &IdeCommand{Command: string(e.Text), Area: area, Button: button}
	return ideCommand
}

func (p *PythonIde) IsIdeCommand(e *acme.Event) bool {
	// The area has to be the tag. We don't process anything in the body.
	area, err := WhichAcmeArea(e)
	if err != nil {
		return false
	}
	if area != AcmeAreaTag {
		return false
	}

	// Make sure the button is valid.
	_, err = WhichAcmeButton(e)
	if err != nil {
		return false
	}

	// We don't override Acme internals, ever.
	if e.Flag&1 != 0 {
		return false
	}

	// We don't override user inserted strings. Commands have to match the IDE commands.
	_, ok := PythonIdeCommands[string(e.Text)]
	return ok
}

func (p *PythonIde) HandleCommand(i *IdeCommand) error {
	if i.Command == "Def" {
		body, err := GetAcmeWindowBody(p.acmeWin)
		if err != nil {
			return err
		}
		lineAndColumn, err := GetAcmeWindowLineAndColumn(p.acmeWin, body)
		if err != nil {
			return err
		}
		ycmdRequest := &YcmdRequest{
			LineNum:          lineAndColumn.Line,
			ColumnNum:        lineAndColumn.Column,
			Filepath:         p.Name,
			FileContents:     body,
			CommandArguments: []string{"GoTo"},
			Filetypes:        []string{"python"},
		}
		rsp, err := PostHandler("run_completer_command", ycmdRequest)
		log.Println(rsp)
	}
	return nil
}

func (p *PythonIde) Watch() {
	events := p.acmeWin.EventChan()
	for {
		e, ok := <-events
		if !ok {
			break
		}
		if p.IsIdeCommand(e) {
			ideCommand := NewIdeCommand(e)
			p.HandleCommand(ideCommand)
		} else {
			p.acmeWin.WriteEvent(e)
			continue
		}
	}
}

func NewPythonIde(winId int, winName string) *PythonIde {
	return &PythonIde{Id: winId, Name: winName}
}

func NewIde(winId int, winName string) Ide {
	windowType := DetermineWindowType(winName)
	if windowType == PythonWindow {
		return NewPythonIde(winId, winName)
	}
	return nil
}

func WatchWindow(winId int, winName string) {
	log.Printf("Found window: %s\n", winName)

	ide := NewIde(winId, winName)
	if ide == nil {
		return
	}
	ide.Setup()
	ide.Watch()
	log.Printf("Finished watching %s\n", winName)
}

func main() {
	argv := os.Args
	if len(argv) < 1 {
		log.Fatal(errors.New("Path to ycmd required."))
	}
	logReader, err := acme.Log()
	if err != nil {
		log.Fatal(err)
	}
	winInfos, err := acme.Windows()
	if err != nil {
		log.Fatal(err)
	}
	UpdateCurrentSettings(DefaultSettings())
	SetHmacSecret(GenerateHmacSecret())
	go YcmdForever(argv[1])
	if IsReady() {
		log.Println("Ycmd Ready!")
	}
	// Keep Ycmd alive.
	go func() {
		for {
			// IsReady waits a while interally before querying, so it's fine to hot loop here.
			IsReady()
		}
	}()
	for _, winInfo := range winInfos {
		go WatchWindow(winInfo.ID, winInfo.Name)
	}

	for {
		logEvent, err := logReader.Read()
		if err != nil {
			log.Println(err)
		}
		if logEvent.Op == "new" {
			go WatchWindow(logEvent.ID, logEvent.Name)
		}
	}
}
