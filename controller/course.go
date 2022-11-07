package controller

import (
	"Gooj/config"
	"Gooj/logger"
	"Gooj/middleware"
	"Gooj/model"
	"Gooj/util"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func GetCourses(c *gin.Context) {

	userData, _ := c.Get("claims")
	user := userData.(*middleware.CustomClaims)
	if user.Identity == config.TypeTeacher {
		c.JSON(http.StatusOK, gin.H{
			"Status": config.PermissionDenied,
			"Info":   "Teacher is not allowed",
		})
		logger.Debug("Tea den")
		return
	}
	courseList := make([]model.BasicCourseInfo, 0)
	myChoice := model.MyChoice(user.ID)
	Db := util.GetConn().Debug().Table("course_info ci").Joins("inner join gooj_tea_users gt").Where("ci.teacher_id=gt.id")
	page, _ := strconv.Atoi(c.Query("page"))
	pageSize, _ := strconv.Atoi(c.Query("pageSize"))

	if name, isExist := c.GetQuery("name"); isExist && name != "" {
		Db = Db.Where("ci.course_name LIKE ? OR gt.realname LIKE ?", "%"+name+"%", "%"+name+"%")
	}
	if queryType, isExist := c.GetQuery("type"); isExist && queryType == "getTotalPages" {
		var total int64
		Db.Count(&total)
		pageNum := total / int64(pageSize)
		if total%int64(pageSize) != 0 {
			pageNum++
		}
		c.JSON(http.StatusOK, gin.H{
			"Status": config.Success,
			"total":  pageNum,
		})
		return
	}
	if page > 0 && pageSize > 0 {
		Db = Db.Limit(pageSize).Offset((page - 1) * pageSize)
	}
	if err := model.GetCourseWithTeacherName(Db, &courseList); err != nil {
		logger.Warn("db err!")
		c.JSON(http.StatusOK, gin.H{
			"Status": config.InternalServerError,
			"Info":   "err",
		})
		return
	}
	for k, v := range courseList {
		for _, v1 := range myChoice {
			if v.ID == v1 {
				courseList[k].Choosed = true
			}
		}
	}

	fmt.Println(courseList)
	c.JSON(http.StatusOK, gin.H{
		"Status": config.Success,
		"Info": gin.H{
			"CourseList": courseList,
		},
	})
}

func AddCourse(c *gin.Context) {
	userData, _ := c.Get("claims")
	user := userData.(*middleware.CustomClaims)
	if user.Identity == config.TypeStudent {
		c.JSON(http.StatusOK, gin.H{
			"Status": config.PermissionDenied,
			"Info":   "Studnet is not allowed",
		})
		logger.Debug("Stu den")
		return
	}
	var course *model.CourseInfo = &model.CourseInfo{}
	err := c.ShouldBind(course)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"Status": config.InternalServerError,
			"Info":   "InternalServerErrorPin1",
		})
		logger.Warn("信息绑定失败！" + err.Error())
		return
	}
	course.StartTime, _ = time.ParseInLocation("2006-01-02T15:04:05Z", course.SeTime[0], time.Local)
	course.EndTime, _ = time.ParseInLocation("2006-01-02T15:04:05Z", course.SeTime[1], time.Local)
	tag := ""
	for _, v := range course.ArrTag {
		tag += "," + v
	}
	tag = strings.Trim(tag, ",")
	course.ClassTags = tag
	course.TeacherId = user.ID
	err = model.AddCourse(course)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"Status": config.InternalServerError,
			"Info":   "InternalServerErrorPin1",
		})
		logger.Warn("无法增加教师课程！")
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"Status": config.Success,
		"Info":   "",
	})
}
func GetCourseDetail(c *gin.Context) {
	userData, _ := c.Get("claims")
	user := userData.(*middleware.CustomClaims)
	courseid,_ := c.GetQuery("course_id")
	ans,_:=strconv.Atoi(courseid)
	if user.Identity == config.TypeStudent {
		stuCourse := new(model.StudentCourseDetail)
		err:=model.GetCourseDetail(ans, user.ID, user.Identity, nil, stuCourse)
		if err!=nil{
			c.JSON(http.StatusOK, gin.H{
				"Status": config.InternalServerError,
				"Info":   "InternalServerErrorPin1",
			})
			logger.Warn("无法学生课程！"+err.Error())
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"Status": config.Success,
			"Info":   stuCourse,
		})
	} else {
		teaCourse:=new(model.TeacherCourseDetail)
		err:=model.GetCourseDetail(ans,user.ID,user.Identity,teaCourse,nil)//选课名单
		if err!=nil{
			c.JSON(http.StatusOK, gin.H{
				"Status": config.InternalServerError,
				"Info":   "InternalServerErrorPin1",
			})
			logger.Warn("无法教师课程！"+err.Error())
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"Status": config.Success,
			"Info":   teaCourse,
		})
	}
}
func DeleteCourse(c *gin.Context){
	userData, _ := c.Get("claims")
	user := userData.(*middleware.CustomClaims)
	if user.Identity == config.TypeStudent {
		c.JSON(http.StatusOK, gin.H{
			"Status": config.PermissionDenied,
			"Info":   "Student not allowed area",
		})
		logger.Debug("denied student")
		return
	}
	deleteCourse:=new(model.TeacherCourseDetail)
	err:=c.ShouldBind(deleteCourse)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"Status": config.InternalServerError,
			"Info":   "InternalServerErrorPin1",
		})
		logger.Warn("信息绑定失败！" + err.Error())
		return
	}
	err=model.DeleteAllRelatedCourses(deleteCourse.ID)
	if err!=nil{
		c.JSON(http.StatusOK, gin.H{
			"Status": config.InternalServerError,
			"Info":   "InternalServerErrorPin1",
		})
		logger.Warn("删除课程失败！")
		return
	}
	c.JSON(http.StatusOK,gin.H{
		"Status":config.Success,
		"Info":"",
	})
}