package controller

import (
	"Gooj/config"
	"Gooj/logger"
	"Gooj/middleware"
	"Gooj/model"
	"Gooj/util"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func DeleteTask(c *gin.Context) {
	userData, _ := c.Get("claims")
	user := userData.(*middleware.CustomClaims)
	if user.Identity == config.TypeStudent {
		c.JSON(http.StatusOK, gin.H{
			"Status": config.PermissionDenied,
			"Info":   "Studnet is not allowed",
		})
		logger.Debug("Stu deny del task")
		return
	}
	var task *model.CourseTask = &model.CourseTask{}
	err := c.ShouldBind(task)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"Status": config.InternalServerError,
			"Info":   "InternalServerErrorPin1",
		})
		logger.Warn("信息绑定失败！" + err.Error())
		return
	}
	dberr := model.DeleteTask(int(task.ID))
	if dberr != nil {
		c.JSON(http.StatusOK, gin.H{
			"Status": config.InternalServerError,
			"Info":   "InternalServerErrorPin1",
		})
		logger.Warn("RM task err!" + err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"Status": config.Success,
		"Info":   "",
	})
}
func AddTask(c *gin.Context) {
	userData, _ := c.Get("claims")
	user := userData.(*middleware.CustomClaims)
	if user.Identity == config.TypeStudent {
		c.JSON(http.StatusOK, gin.H{
			"Status": config.PermissionDenied,
			"Info":   "Studnet is not allowed",
		})
		logger.Debug("Stu deny del task")
		return
	}
	var task *model.CourseTask = &model.CourseTask{}
	err := c.ShouldBind(task)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"Status": config.InternalServerError,
			"Info":   "InternalServerErrorPin1",
		})
		logger.Warn("信息绑定失败！" + err.Error())
		return
	}
	dberr := model.AddTask(task)
	if dberr != nil {
		c.JSON(http.StatusOK, gin.H{
			"Status": config.InternalServerError,
			"Info":   "InternalServerErrorPin1",
		})
		logger.Warn("add task err!" + err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"Status": config.Success,
		"Info":   "",
	})
}
func GetTaskInfo(c *gin.Context) {
	userData, _ := c.Get("claims")
	user := userData.(*middleware.CustomClaims)
	if user.Identity == config.TypeStudent {
		c.JSON(http.StatusOK, gin.H{
			"Status": config.PermissionDenied,
			"Info":   "Student is not allowed",
		})
		logger.Debug("Stu deny quer task")
		return
	}
	task := &model.TeacherTaskInfo{}
	if tid, isHere := c.GetQuery("task_id"); isHere {
		task.ID, _ = strconv.Atoi(tid)
	}
	dberr := model.GetTaskInfo(task)
	if dberr != nil {
		c.JSON(http.StatusOK, gin.H{
			"Status": config.InternalServerError,
			"Info":   "InternalServerError",
		})
		logger.Warn("quer task err!" + dberr.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"Status": config.Success,
		"Info":   task,
	})
}
func GetStudentTaskAns(c *gin.Context) {
	userData, _ := c.Get("claims")
	user := userData.(*middleware.CustomClaims)
	if user.Identity == config.TypeStudent {
		c.JSON(http.StatusOK, gin.H{
			"Status": config.PermissionDenied,
			"Info":   "Student is not allowed",
		})
		logger.Debug("Stu deny quer task")
		return
	}
	_taskid, _ := c.GetQuery("task_id")
	_stuid, _ := c.GetQuery("stu_id")
	stuid, _ := strconv.Atoi(_stuid)
	taskid, _ := strconv.Atoi(_taskid)
	stuInfo := &model.CodeFile{}
	if err := model.GetStudentTaskAnswer(stuid, taskid, stuInfo); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"Status": config.InternalServerError,
			"Info":   "InternalServerError",
		})
		logger.Warn("无法取得学生答案" + err.Error())
		return
	}
	stuInfo.StuCode = util.GetFile(stuInfo.FileId)
	c.JSON(http.StatusOK, gin.H{
		"Status": config.Success,
		"Info":   stuInfo,
	})
}
func CommentCode(c *gin.Context) {
	userData, _ := c.Get("claims")
	user := userData.(*middleware.CustomClaims)
	if user.Identity == config.TypeStudent {
		c.JSON(http.StatusOK, gin.H{
			"Status": config.PermissionDenied,
			"Info":   "Student is not allowed",
		})
		logger.Debug("Stu deny quer task")
		return
	}
	tmp:=map[string]interface{}{}
	c.ShouldBind(&tmp)
	codeid := tmp["code_id"].(float64)
	comment := tmp["comment"].(string)
	if err := model.TeacherComment(int(codeid), comment); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"Status": config.InternalServerError,
			"Info":   "InternalServerError",
		})
		logger.Warn("无法评论" + err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"Status": config.Success,
		"Info":   "",
	})
}
