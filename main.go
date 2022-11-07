package main

import (
	"Gooj/config"
	"Gooj/controller"
	"Gooj/logger"
	"Gooj/middleware"
	"Gooj/util"
	"fmt"
	"os/exec"
	"time"

	"github.com/gin-gonic/gin"
)

func init() {
	logger.InitLog()
	util.ConnectDB()
	rm5050:=exec.Command("kill  $(netstat -nlp | grep :5050| awk '{print $7}' | awk -F\"/\" '{ print $1 }')")
	rm5050.Run()
	command := fmt.Sprintf("sudo %s -dir=%s -http-addr=:%d -file-timeout=%s -release",
	config.GetEnvSettings().EnvironmentSettings.Sandbox.BinaryPath,
	config.GetEnvSettings().EnvironmentSettings.TempCodeStoragePath+"sandbox",
	config.GetEnvSettings().EnvironmentSettings.Sandbox.Port,
	config.GetEnvSettings().EnvironmentSettings.Sandbox.StorageTimeout,
)
	logger.Debug(command)
	cmd:=exec.Command("/bin/bash","-c",command)
	err:=cmd.Start()
	if err!=nil{
		logger.Panic("Sandbox error:"+err.Error())
	}
}

func main() {
	go func() {
		db, _ := util.GetConn().DB()
		for {
			db.Ping()
			time.Sleep(time.Hour * 2)
		}
	}()
	engine := gin.New()
	loggerEngine:=logger.GetLogger()
	engine.Use(middleware.Ginzap(loggerEngine,time.RFC3339, true))
	engine.Use(middleware.RecoveryWithZap(loggerEngine,true))//Self-def panic hand doo
	engine.Use(middleware.Cors())

	engine.NoMethod(util.HandleNoMethod)
	engine.NoRoute(util.HandleNoRoute)

	engine.POST("/register", controller.Register)
	engine.POST("/login", controller.Login)
	engine.POST("/checkExpired", controller.CheckIsExpired)
	engine.GET("/getCaptcha", util.NewCaptcha)
	//用户需要验证登录
	codeContext := engine.Group("/code")
	codeContext.Use(middleware.JWTAuth())
	{
		codeContext.POST("/onlineFmtCode", controller.OnlineFmtCode)
		codeContext.POST("/codeSubmit", controller.FetchFile)
		codeContext.POST("/runCode",controller.RunCode)
	}
	userContent := engine.Group("/user")
	userContent.Use(middleware.JWTAuth())
	{
		userContent.GET("/getBasicInfo", controller.GetBasicInfo)
		userContent.POST("/updateInfo", controller.UpdateMyInfo)
	}
	courseContent := engine.Group("/course")
	courseContent.Use(middleware.JWTAuth())
	{
		courseContent.POST("/addCourse",controller.AddCourse)//TeacherOnly
		courseContent.GET("/getCourses",controller.GetCourses)//学生获取课程/模糊查询
		courseContent.POST("/chooseCourse",controller.ChooseCourse)//选课/取消
		courseContent.GET("/getCourseDetail",controller.GetCourseDetail)//教师学生公用，包括任务信息也在
		courseContent.POST("/removeCourse",controller.DeleteCourse)//删课，教师
	}
	taskContent := engine.Group("/task")
	taskContent.Use(middleware.JWTAuth())
	{
		taskContent.POST("/removeTask",controller.DeleteTask)//TeacherOnly;删除课程
		taskContent.POST("/addTask",controller.AddTask)//TeacherOnly增加任务
		taskContent.GET("/getTaskDetail",controller.GetTaskInfo)//教师获取名任务信息；注意，学生任务信息在课程信息中给出
		taskContent.GET("/getStuTaskAns",controller.GetStudentTaskAns)
		taskContent.POST("/commentCode",controller.CommentCode)//评论
	}
	engine.Run()
}
