package model

import (
	"Gooj/config"
	"Gooj/util"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type CourseInfo struct {
	gorm.Model
	TeacherId          int
	CourseName         string `json:"name"`
	StartTime          time.Time
	EndTime            time.Time
	CoursePic          string
	CourseIntroduction string `json:"intro"`
	IsPublic           bool   `json:"isPublic"`
	Classroom          string `json:"region"`
	ClassTags          string
	ArrTag             []string `json:"tags" gorm:"-"`
	ClassType          string   `json:"type"`
	SeTime             []string `gorm:"-" json:"time"`
}
func (CourseInfo) TableName() string {
	return "course_info"
}
type BasicStudentInfo struct {
	ID         string
	Realname   string
	Institute  string
	Profession string
}

func (BasicStudentInfo) TableName() string {
	return "gooj_stu_users"
}

type TeacherCourseDetail struct {
	ID                 int `json:"course_id"`
	CourseName         string
	StartTime          time.Time
	EndTime            time.Time
	CourseIntroduction string
	IsPublic           bool
	Classroom          string
	ClassTags          string
	ClassType          string
	basicStudentInfo   []BasicStudentInfo
	Tasks              []TeacherTaskInfo `gorm:"-"`
}
func (TeacherCourseDetail) TableName() string {
	return "course_info"
}
type StudentCourseDetail struct {
	ID                 int
	TeacherName        string `gorm:"column:realname"`
	CourseName         string
	StartTime          time.Time
	EndTime            time.Time
	CourseIntroduction string
	IsPublic           bool
	Classroom          string
	ClassTags          string
	ClassType          string
	Tasks              []StudentTaskInfo `gorm:"-"`
}
func (StudentCourseDetail) TableName() string {
	return "course_info"
}
type StudentCourse struct {
	CourseId  int
	StudentId int
	CreatedAt time.Time
	UpdatedAt time.Time
}
func (StudentCourse) TableName() string {
	return "student_course"
}
type BasicCourseInfo struct {
	gorm.Model
	ID          int
	CourseName  string
	StartTime   time.Time
	EndTime     time.Time
	TeacherName string `gorm:"column:realname"`
	Choosed     bool   `gorm:"-"`
}
func (BasicCourseInfo) TableName() string {
	return "course_info"
}

func AddCourse(courseinfo *CourseInfo) error {
	db := util.GetConn()
	err := db.Create(courseinfo).Error
	return err
}
func GetCourseWithTeacherName(Db *gorm.DB, basicCourse *[]BasicCourseInfo) error {
	return Db.Debug().Select("ci.ID,ci.course_name,ci.start_time,ci.end_time,gt.realname").Scan(basicCourse).Error
	// INNER JOIN
}
func IsCourseDupl(courseid int, studentid int) (bool,error){
	db := util.GetConn()
	course := StudentCourse{
		StudentId: studentid,
		CourseId:  courseid,
	}
	err := db.Debug().Where("course_id=?", courseid).First(&course).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false,nil
		}
		return true,err
	} else {
		return true,errors.New("dupl")
	}
}
func ChooseCourse(courseid int, studentid int) error {
	db := util.GetConn()
	course := StudentCourse{
		StudentId: studentid,
		CourseId:  courseid,
	}
	return db.Debug().Create(&course).Error
}
func DeleteChoosedCourse(courseid int, studentid int) error {
	db := util.GetConn()
	course := StudentCourse{
		StudentId: studentid,
		CourseId:  courseid,
	}
	return db.Debug().Where("student_id=? and course_id=?", studentid, courseid).Unscoped().Delete(&course).Error
}
func MyChoice(userid int) []int {
	db := util.GetConn()
	mycourse := make([]int, 0)
	myCS := []StudentCourse{}
	db.Debug().Select("course_id").Where("student_id=?", userid).Find(&myCS)
	for _, v := range myCS {
		mycourse = append(mycourse, v.CourseId)
	}
	fmt.Println(mycourse)
	return mycourse
}
func GetCourseDetail(courseid int, userid int, class int, teacherCourse *TeacherCourseDetail, studentCourse *StudentCourseDetail) error {
	db := util.GetConn()
	if class == config.TypeTeacher {
		if err := db.Debug().Find(teacherCourse,courseid).Error; err != nil {
			return err
		}
		var basicStudentInfo []BasicStudentInfo = make([]BasicStudentInfo, 0)
		if err := db.Debug().Table("student_course sc").Joins("inner join gooj_stu_users gs on sc.student_id=gs.ID ").Where("sc.course_id=?", courseid).Scan(&basicStudentInfo).Error; err != nil {
			return err
		}
		teacherCourse.basicStudentInfo = basicStudentInfo
		tasksInfo:=make([]TeacherTaskInfo,0)
		if err := db.Debug().Find(&tasksInfo,"course_id=?",courseid).Error; err != nil {
			return err
		}
		teacherCourse.Tasks = tasksInfo
		// //获取已提交的在其他api
		// var uploaders []BasicStudentInfo = make([]BasicStudentInfo, 0)
		// if err := db.Debug().Table("gooj_stu_users gs").Joins("inner join student_course sc").Joins("inner join course_tasks ct").Where("sc.course_id=ct.course_id AND gs.ID=sc.student_id AND ct.course_id=?", courseid).Scan(&uploaders).Error; err != nil {
		// 	return err
		// }
	} else {
		if err := db.Debug().Raw("SELECT ci.ID,gt.realname,ci.course_name,ci.start_time,ci.end_time,ci.course_introduction,ci.is_public,ci.classroom,ci.class_tags,ci.class_type  FROM course_info ci INNER JOIN gooj_tea_users gt ON ci.teacher_id=gt.ID INNER JOIN student_course sc ON sc.course_id=ci.ID WHERE sc.student_id=? AND ci.ID=?", userid, courseid).Scan(studentCourse).Error; err != nil {
			return err
		}//课程信息
		tasksInfo:=make([]StudentTaskInfo,0)
		if err := db.Debug().Find(&tasksInfo,"course_id=?",courseid).Error; err != nil {
			return err
		}//任务详情
		tmp:=make([]CodeFile,0)
		if err := db.Debug().Find(&tmp,"uploader_id=?",userid).Error; err != nil {
			return err
		}//任务详情
		for _,v:=range tmp{
			for i1,v1:=range tasksInfo{
				if v.TaskID==int(v1.ID){
					tasksInfo[i1].Comment=v.TeacherComment
					tasksInfo[i1].Finished = true
				}
			}
		}//判断是否完成
		studentCourse.Tasks=tasksInfo
	}//课程详情页有任务
	return nil
}
func DeleteAllRelatedCourses(courseid int) error {
	db := util.GetConn()
	deleteTask := make([]CourseTask, 0)
	err := db.Debug().Where("course_id=?", courseid).Find(&deleteTask).Error //获取所有task
	if err != nil {
		return err
	}
	for _, v := range deleteTask { //删除相关files和tasks
		err := db.Debug().Where("task_id=?", v.ID).Unscoped().Delete(&CodeFile{}).Error
		if err != nil {
			return err
		}
		err = db.Debug().Where("id=?", v.ID).Unscoped().Delete(&CourseTask{}).Error
		if err != nil {
			return err
		}
	}
	//删除课程
	err = db.Debug().Where("ID=?", courseid).Unscoped().Delete(&CourseInfo{}).Error
	if err != nil {
		return err
	}
	return nil
}
