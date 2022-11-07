package model

import (
	"Gooj/config"
	"Gooj/util"
	"errors"
	"gorm.io/gorm"
	"time"
)

//Teacher->1
type GoojUser struct {
	ID            int    `gorm:"primarykey"`
	Username      string `json:"username"`
	Password      string `json:"password"`
	Identity      int    `gorm:"-" json:"identity"`
	LastLoginAt   time.Time
	CreatedAt     time.Time `gorm:"<-:create"`
	UpdatedInfoAt time.Time
	DeletedAt     gorm.DeletedAt `gorm:"index"`
	CaptchaID     string         `gorm:"-" json:"captchaID"`
	CaptchaRes    string         `gorm:"-" json:"captchaRes"`
}
type GoojStuUser struct {
	Realname         string   `json:"realname"`
	StudentNumber    string   `json:"studentNumber"`
	Institute        string   `json:"institute"`
	Profession       string   `json:"profession"`
	SelfIntroduction string   `json:"selfIntroduction"`
	GoojUser         GoojUser `gorm:"embedded"`
}
type GoojTeaUser struct {
	Realname         string   `form:"realname"`
	TeacherNumber    string   `form:"teacherNumber"`
	Institute        string   `form:"institute"`
	Profession       string   `form:"profession"`
	SelfIntroduction string   `form:"selfIntroduction"`
	GoojUser GoojUser `gorm:"embedded"`
}

func IsDuplUser(username string, class int) (bool, error) {
	db := util.GetConn()
	var err error
	if class == config.TypeStudent {
		err = db.Where("username=?", username).First(new(GoojStuUser)).Error
	} else {
		err = db.Where("username=?", username).First(new(GoojTeaUser)).Error
	}
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
func CreateUser(user interface{}) error {
	db := util.GetConn()
	var errdb error
	switch userKind := user.(type) {
	case *GoojStuUser:
		errdb = db.Create(userKind).Error
	case *GoojTeaUser:
		errdb = db.Create(userKind).Error
	}
	return errdb
}
func GetUserByUsername(queryUser *GoojUser, username string, class int) error {
	db := util.GetConn()
	if class == config.TypeTeacher {
		if err := db.Table("gooj_tea_users").First(queryUser, "username=?", username).Error; err != nil {
			return err
		}
		return nil
	} else {
		if err := db.Table("gooj_stu_users").First(queryUser, "username=?", username).Error; err != nil {
			return err
		}
		return nil
	}
}
func GetUserById(id int, class int) (interface{}, error) {
	db := util.GetConn()
	if class == config.TypeStudent {
		queryID := new(GoojStuUser)
		err := db.First(queryID, "ID=?", id).Error
		return queryID, err
	} else {
		queryID := new(GoojTeaUser)
		err := db.First(queryID, "ID=?", id).Error
		return queryID, err
	}
}
func IsPasswordMatch(loginReq *GoojUser, class int) (bool, error) {
	db := util.GetConn()
	var err error
	if class == config.TypeStudent {
		err = db.Table("gooj_stu_users").Where("Username=? and Password=?", loginReq.Username, loginReq.Password).First(loginReq).Error
	} else {
		err = db.Table("gooj_tea_users").Where("Username=? and Password=?", loginReq.Username, loginReq.Password).First(loginReq).Error
	}
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		} else {
			return false, err
		}
	} else {
		return true, nil
	}
}
func UpdateLastLoginTime(GoojUser *GoojUser, class int) error {
	db := util.GetConn()
	if class == config.TypeStudent {
		return db.Table("gooj_stu_users").Model(GoojUser).Update("last_login_at", time.Now()).Error
	} else {
		return db.Table("gooj_tea_users").Model(GoojUser).Update("last_login_at", time.Now()).Error
	}
}
func UpdateInfo(goojUser interface{}, class int) error {
	db := util.GetConn()
	if class == config.TypeStudent {
		return db.Model(goojUser.(*GoojStuUser)).Updates(goojUser.(*GoojStuUser)).Error
	} else {
		return db.Model(goojUser.(*GoojTeaUser)).Updates(goojUser.(*GoojTeaUser)).Error
	}
}
func GetBasicInfo(userID int, courseInfo *[]BasicCourseInfo, timeline *[]TimeLine)(error) {
	db := util.GetConn()
	//getMoreInfoHere!
	err := db.Table("gooj_stu_users").Select("course_info.id,course_info.course_name,course_info.start_time,course_info.end_time,gooj_tea_users.realname,course_info.deleted_at").
		Joins("INNER JOIN student_course ON gooj_stu_users.ID=student_course.student_id").
		Joins("INNER JOIN course_info ON course_info.ID=student_course.course_id").
		Joins("INNER JOIN gooj_tea_users ON gooj_tea_users.ID=course_info.teacher_id").Where("student_course.student_id=? AND course_info.deleted_at is null",userID).Find(courseInfo).Error
	// if err != nil {
	// 	c.JSON(http.StatusOK, gin.H{
	// 		"Status": config.InternalServerError,
	// 		"Info":   "InternalServerErrorPin1",
	// 	})
	// 	logger.Warn("query user error!")
	// 	return
	// }
	if err!=nil{
		return err
	}
	err = db.Table("gooj_stu_users").Select("gooj_stu_users.ID,course_tasks.title,course_info.course_name,course_tasks.deadline,course_tasks.id,course_tasks.deleted_at").
		Joins("INNER JOIN student_course ON gooj_stu_users.ID=student_course.student_id").
		Joins("INNER JOIN course_info ON course_info.ID=student_course.course_id").
		Joins("INNER JOIN course_tasks ON course_tasks.course_id=student_course.course_id").Where("gooj_stu_users.ID=? AND course_tasks.deleted_at IS NULL", userID).Find(timeline).Error
	// if err != nil {
	// 	c.JSON(http.StatusOK, gin.H{
	// 		"Status": config.InternalServerError,
	// 		"Info":   "InternalServerErrorPin1",
	// 	})
	// 	logger.Warn("query user error!")
	// 	return
	// }
	return err
}
func GetTeachingCourses(id int,courses *[]BasicCourseInfo)error{
	db := util.GetConn()
	return db.Debug().Where("teacher_id=?",id).Find(courses).Error
}