package util

import (
	"Gooj/config"
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

var (
	Address      = "http://" + config.GetEnvSettings().EnvironmentSettings.Sandbox.Address + ":" + strconv.Itoa(config.GetEnvSettings().EnvironmentSettings.Sandbox.Port)
	TimeLimit    = config.GetEnvSettings().EnvironmentSettings.Sandbox.TimeLimit * 1000 * 1000
	MemoryLimit  = config.GetEnvSettings().EnvironmentSettings.Sandbox.MemoryLimit * 1024 * 1024
	StackLimit   = config.GetEnvSettings().EnvironmentSettings.Sandbox.StackLimit * 1024 * 1024
	ProcLimit    = config.GetEnvSettings().EnvironmentSettings.Sandbox.ProcLimit
	CpuRateLimit = config.GetEnvSettings().EnvironmentSettings.Sandbox.CpuRateLimit
)

type WholeFile struct {
	Cmd []Cmd
}
type CmdFile struct {
	Src     string `json:"src,omitempty"`
	Content string `json:"content,omitempty"`
	FileID  string `json:"fileId,omitempty"`
	Name    string `json:"name,omitempty"`
	Max     int64  `json:"max,omitempty"`
	Pipe    bool   `json:"pipe,omitempty"`
}
type Cmd struct {
	Args  []string  `json:"args"`
	Env   []string  `json:"env,omitempty"`
	Files []CmdFile `json:"files,omitempty"`

	CPULimit uint64 `json:"cpuLimit"`
	// RealCPULimit      uint64 `json:"realCpuLimit"`
	// ClockLimit        uint64 `json:"clockLimit"`
	MemoryLimit  uint64 `json:"memoryLimit"`
	StackLimit   uint64 `json:"stackLimit"`
	ProcLimit    uint64 `json:"procLimit"`
	CPURateLimit uint64 `json:"cpuRateLimit"`
	// CPUSetLimit       string `json:"cpuSetLimit"`
	// StrictMemoryLimit bool   `json:"strictMemoryLimit"`
	CopyIn  map[string]CmdFile `json:"copyIn"`
	CopyOut []string           `json:"copyOut"`
}

func SendStorageReq(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// 实例化multipart
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// 创建multipart 文件字段
	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return "", err
	}
	// 写入文件数据到multipart
	_, err = io.Copy(part, file)
	if err != nil {
		return "", err
	}
	err = writer.Close()
	if err != nil {
		return "", err
	}
	//创建请求
	req, err := http.NewRequest("POST", Address+"/file", body)
	if err != nil {
		return "", err
	}
	//不要忘记加上writer.FormDataContentType()，
	//该值等于content-type :multipart/form-data; boundary=xxxxx
	req.Header.Add("Content-Type", writer.FormDataContentType())
	resp, err := http.DefaultClient.Do(req)
	defer func() {
		if resp != nil {
			resp.Body.Close()
		}
	}()
	if err != nil || resp.StatusCode != 200 {
		return "", errors.New("resp status err")
	}
	id, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(id), nil
}

/*
{
    "cmd": [{
        "args": ["/usr/bin/g++", "a.cc", "-o", "a"],
        "env": ["PATH=/usr/bin:/bin"],
        "files": [{
            "content": ""
        }, {
            "name": "stdout",
            "max": 10240
        }, {
            "name": "stderr",
            "max": 10240
        }],
        "cpuLimit": 10000000000,
        "memoryLimit": 104857600,
        "procLimit": 50,
        "copyIn": {
            "a.cc": {
                "content": "#include <iostream>\nusing namespace std;\nint main() {\nint a, b;\ncin >> a >> b;\ncout << a + b << endl;\n}"
            }
        },
        "copyOut": ["stdout", "stderr"],
        "copyOutCached": ["a.cc", "a"],
        "copyOutDir": "1"
    }]
}*/
func SendRunReqWithFileid(fileId string, lanType string) (string,error){
	modelfile := WholeFile{
		Cmd: make([]Cmd, 1),
	}
	modelfile.Cmd[0] = Cmd{}
	modelfile.Cmd[0].Env = []string{"PATH=/usr/bin:/bin"}
	modelfile.Cmd[0].Files = []CmdFile{
		{Name: "stdout",
			Max: 10240},
		{Name: "stderr",
			Max: 10240},
	}
	modelfile.Cmd[0].CPULimit = TimeLimit
	modelfile.Cmd[0].MemoryLimit = MemoryLimit
	modelfile.Cmd[0].ProcLimit = ProcLimit
	modelfile.Cmd[0].CPURateLimit = CpuRateLimit
	modelfile.Cmd[0].StackLimit = StackLimit
	//specified
	modelfile.Cmd[0].Env = []string{"PATH=/usr/bin:/bin","GOPATH=/w","GOCACHE=/w/.cache/go-build"}
	modelfile.Cmd[0].Args = []string{"/usr/local/go/bin/go","run", "main.go"}
	modelfile.Cmd[0].CopyIn = map[string]CmdFile{
		"main.go":{
			FileID: fileId,
		},
	}
	modelfile.Cmd[0].CopyOut = []string{"stdout","stderr"}
	result,_:=json.Marshal(modelfile)
	req, _ := http.NewRequest("POST", Address+"/run", bytes.NewBuffer(result))
    // req.Header.Set("X-Custom-Header", "myvalue")
    req.Header.Set("Content-Type", "application/json")
    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return "",err
    }
    defer resp.Body.Close()
    body, _ := ioutil.ReadAll(resp.Body)
	return string(body),nil
}
func DeleteFile(fileid string){
	req, err := http.NewRequest("DELETE", Address+"/file/"+fileid, strings.NewReader(""))
	if err != nil {
		return
	}
	rsp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	defer rsp.Body.Close()
}
func GetFile(fileid string)string{
	req, err := http.NewRequest("GET", Address+"/file/"+fileid, strings.NewReader(""))
	if err != nil {
		return ""
	}
	rsp, err := http.DefaultClient.Do(req)
	if err != nil {
		return ""
	}
	defer rsp.Body.Close()
	body, _ := ioutil.ReadAll(rsp.Body)
	return string(body)
}