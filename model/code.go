package model

import (
	"gorm.io/gorm"
)

type CodeFile struct {
	//保存的文件
	UploaderId   int
	LanType      string
	FileId       string
	OriginalName string
	TaskID       int
	gorm.Model
	RunStatRaw     string
	MemCost        int    `json:"memory"`
	TimeCost       int    `json:"time"`
	Status         int    `json:"exitStatus"`
	StuCode        string `gorm:"-"`
	TeacherComment string
}

func (CodeFile) TableName() string {
	return "code_files"
}

type PostInfo struct {
	Code     string `json:"code"`
	Language string `json:"language"`
	TaskID   int    `json:"task_id"`
}

/*{
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
