package controller

import (
	"Gooj/config"
	"Gooj/logger"
	"Gooj/middleware"
	"Gooj/model"
	"Gooj/util"
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	"gorm.io/gorm"
)

func OnlineFmtCode(c *gin.Context) { //改成沙箱
	postInfo := new(model.PostInfo)
	err := c.ShouldBind(&postInfo)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"Status": config.ResolveInfoError, "Info": "Bind not in param",
		})
		logger.Error("参数不足！")
		return
	}
	if postInfo.Code == "" {
		result := gin.H{"Status": config.ResolveInfoError, "Info": "Empty code received"}
		c.JSON(http.StatusOK, result)
	} else {
		lanType := postInfo.Language
		fileUUID := uuid.Must(uuid.NewV4()).String()
		pathC := "/tmp/gooj/fmt/" + fileUUID + "." + lanType
		defer os.Remove(pathC)
		logger.Info(pathC)
		err := ioutil.WriteFile(pathC, []byte(postInfo.Code), 0666)
		if err != nil {
			logger.Warn("创建用户失败！")
			c.JSON(http.StatusOK, gin.H{
				"Status": config.InternalServerError,
				"Info":   "InternalServer",
			})
		}
		var cmd *exec.Cmd
		if lanType == "go" {
			cmd = exec.Command("gofmt", "-w", pathC)
		}
		err1 := cmd.Run()
		if err1 != nil {
			logger.Warn("error exec" + err1.Error())
			if strings.ContainsAny(err1.Error(), "exit status ") {
				c.JSON(http.StatusOK, gin.H{
					"Status": config.FormatError,
					"Info":   "Code syntax has error",
				})
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"Status": config.InternalServerError,
				"Info":   "InternalServer",
			})
			return
		}
		fmted, err := ioutil.ReadFile(pathC)
		if err != nil {
			logger.Warn("读取文件失败！")
			c.JSON(http.StatusOK, gin.H{
				"Status": config.InternalServerError,
				"Info":   "InternalServer",
			})
			return
		}
		result := gin.H{"Status": config.Success, "Fmt_code": string(fmted)}
		c.JSON(http.StatusOK, result)
	}
}
func LocalFmtCode(path string) error {
	cmd := exec.Command("gofmt", "-w", path)
	_, err := cmd.CombinedOutput()
	return err
}

// AutocompleteHandler handles request of code autocompletion.
func Autocomplete(c *gin.Context) {
	code := c.DefaultPostForm("code", "")
	line, err := strconv.Atoi(c.DefaultPostForm("cursorLine", ""))
	if err != nil {
		result := gin.H{"Status": config.InternalServerError, "Info": "InternalServerError"}
		c.JSON(http.StatusOK, result)
		return
	}
	ch, err := strconv.Atoi(c.DefaultPostForm("cursorCh", ""))
	if err != nil {
		result := gin.H{"Status": config.InternalServerError, "Info": "InternalServerError"}
		c.JSON(http.StatusOK, result)
		return
	}

	offset := getCursorOffset(code, line, ch)

	argv := []string{"-f=json", "autocomplete", strconv.Itoa(offset)}
	gocodepath := config.GetEnvSettings().EnvironmentSettings.Golang.Gopath
	if gocodepath == "" {
		gocodepath = "gocode"
	}
	cmd := exec.Command(gocodepath, argv...)

	stdin, _ := cmd.StdinPipe()
	stdin.Write([]byte(code))
	stdin.Close()

	output, err := cmd.CombinedOutput()
	if err != nil {
		result := gin.H{"Status": config.InternalServerError, "Info": "InternalServerError"}
		c.JSON(http.StatusOK, result)
		return
	}
	result := gin.H{"Status": config.Success, "Auto_comp": string(output)}
	c.JSON(http.StatusOK, result)
}

// getCursorOffset calculates the cursor offset.
//
// line is the line number, starts with 0 that means the first line
// ch is the column number, starts with 0 that means the first column
func getCursorOffset(code string, line int, ch int) (offset int) {
	lines := strings.Split(code, "\n")
	// calculate sum length of lines before
	for i := 0; i < line; i++ {
		offset += len(lines[i])
	}

	// calculate length of the current line and column
	curLine := lines[line]
	var buffer bytes.Buffer
	r := []rune(curLine)
	for i := 0; i < ch; i++ {
		buffer.WriteString(string(r[i]))
	}

	offset += len(buffer.String()) // append length of current line
	offset += line                 // append number of '\n'
	return offset
}
func FetchFile(c *gin.Context) {
	db := util.GetConn()
	editor := new(model.PostInfo)
	err := c.ShouldBind(editor) //语言类型
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"Status": config.InternalServerError, "Info": "Internal Server Error",
		})
		logger.Warn("无法绑定！")
		return
	}
	filepath_code := "/dev/null"
	if editor.Language == "" {
		c.JSON(http.StatusOK, gin.H{
			"Status": config.ResolveInfoError, "Info": "Language not in param",
		})
		logger.Debug("参数不足！Language")
		return
	}
	if editor.Language != "go" {
		c.JSON(http.StatusOK, gin.H{
			"Status": config.LanguageNotSupported, "Info": "Language not supported",
		})
		logger.Debug("unsupport Language")
		return
	}
	tempFilePath := config.GetEnvSettings().EnvironmentSettings.TempCodeStoragePath
	fileUUID := uuid.Must(uuid.NewV4()).String()
	//数据库存文件信息
	fileDBInfo := new(model.CodeFile)
	if editor.Code == "" {
		file, err := c.FormFile("file")
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"Status": config.ResolveInfoError, "Info": "file not found in request",
			})
			logger.Warn("参数不足！file isNil")
			return
		}
		fileDBInfo.OriginalName = file.Filename
		filepath_code = tempFilePath + fileUUID + filepath.Base(file.Filename) //文件存储路径
		defer os.Remove(filepath_code)
		err = c.SaveUploadedFile(file, filepath_code)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"Status": config.InternalServerError, "Info": "SaveFileFailed",
			})
			logger.Warn("保存失败！")
			return
		}
		fid, err := util.SendStorageReq(filepath_code)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"Status": config.InternalServerError, "Info": "SaveFileFailed",
			})
			logger.Warn("保存失败！" + err.Error())
			return
		}
		fileDBInfo.FileId = strings.Trim(fid, "\"")
	} else {
		code := editor.Code
		fileDBInfo.OriginalName = fileUUID + "." + editor.Language
		if code == "" {
			c.JSON(http.StatusOK, gin.H{
				"Status": config.ResolveInfoError,
				"Info":   "Code cant be null",
			})
			logger.Debug("code为空")
			return
		}
		filepath_code = tempFilePath + fileUUID + "." + editor.Language
		defer os.Remove(filepath_code)
		// newFile, err := os.Create(filepath_code)
		// if err != nil {
		// 	config.Warn("error create")
		// 	c.JSON(http.StatusOK,gin.H{
		// 		"Status":InternalServerError,
		// 		"Info":"InternalServer",
		// 	})
		// }else{newFile.Close()
		err := ioutil.WriteFile(filepath_code, []byte(code), 0666)
		if err != nil {
			logger.Warn("建立文件失败！")
			c.JSON(http.StatusOK, gin.H{
				"Status": config.InternalServerError,
				"Info":   "InternalServer",
			})
		}
		fid, err := util.SendStorageReq(filepath_code)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"Status": config.InternalServerError, "Info": "SaveFileFailed",
			})
			logger.Warn("保存失败！" + err.Error())
			return
		}
		fileDBInfo.FileId = strings.Trim(fid, "\"")
	}

	userData, err1 := c.Get("claims")
	if !err1 {
		c.JSON(http.StatusOK, gin.H{
			"Status": config.InternalServerError,
			"info":   "InternalServer",
		})
		logger.Warn("无法取得jwt!")
		return
	}
	fileDBInfo.UploaderId = userData.(*middleware.CustomClaims).ID
	fileDBInfo.LanType = editor.Language
	fileDBInfo.TaskID = editor.TaskID
	var err2 error
	codeID := model.CodeFile{}
	if err := db.Where("task_id=? and uploader_id=?", editor.TaskID, fileDBInfo.UploaderId).Find(&codeID).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		err2 = db.Create(&fileDBInfo).Error
	} else if err == nil {//删除重复
		util.DeleteFile(codeID.FileId)
		db.Where("task_id=? and uploader_id=?", editor.TaskID, fileDBInfo.UploaderId).Unscoped().Delete(&fileDBInfo)
		err2=db.Create(&fileDBInfo).Error
	} else {
		c.JSON(http.StatusOK, gin.H{
			"Status": config.InternalServerError, "Info": "SaveFileFailed",
		})
		os.Remove(filepath_code)
		logger.Warn("数据库错误！")
		return
	}
	if err2 != nil {
		c.JSON(http.StatusOK, gin.H{
			"Status": config.InternalServerError, "Info": "SaveFileFailed",
		})
		os.Remove(filepath_code)
		logger.Warn("数据库错误！")
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"Status": config.Success,
		"Info":   fileDBInfo.FileId,
	})
}
func RunCode(c *gin.Context) {
	codeinfo := map[string]string{}
	c.ShouldBindJSON(&codeinfo)
	if codeinfo["language"] != "go" {
		c.JSON(http.StatusOK, gin.H{
			"Status": config.LanguageNotSupported,
			"Info":   codeinfo["language"] + "not supported.Only go is available",
		})
		return
	}
	resp, err := util.SendRunReqWithFileid(codeinfo["fileid"], codeinfo["language"])
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"Status": config.InternalServerError,
			"info":   "InternalServer",
		})
		logger.Warn("Running err!")
		return
	}
	resp = resp[1 : len(resp)-2]
	tmp:=&model.CodeFile{}
	json.Unmarshal([]byte(resp),tmp)
	tmp.RunStatRaw = resp
	util.GetConn().Debug().Model(&model.CodeFile{}).Where("file_id=?",codeinfo["fileid"]).Updates(tmp)
	c.JSON(http.StatusOK, gin.H{
		"Status": config.Success,
		"info":   resp,
	})
}
