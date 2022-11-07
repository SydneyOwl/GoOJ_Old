package model

import (
	"Gooj/util"
	"time"

	"gorm.io/gorm"
)

type CourseTask struct {
	gorm.Model
	CourseId int       `json:"courseID"`
	Title    string    `json:"title"`
	Detail   string    `json:"detail"`
	Deadline time.Time `json:"deadline"`
}
type TimeLine struct {
	Title      string
	CourseName string
	Deadline   time.Time
	ID         int
}
type TeacherTaskInfo struct{
	ID        int `gorm:"primarykey" json:"ID"`
	CreatedAt time.Time
	StudentList []BasicStudentInfo `gorm:"-"`//已上交的
	Title    string   
	Detail   string    
	Deadline time.Time
	AvgMem float64`gorm:"column:avgmem"`
	AvgTime float64`gorm:"column:avgtime"`
	LeastMemStuID int `gorm:"column:minmemstu"`
	LeastTimeStuID int`gorm:"column:mintimestu"`
}
func (TeacherTaskInfo) TableName() string {
	return "course_tasks"
}
type StudentTaskInfo struct{
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time
	Title    string   
	Detail   string    
	Deadline time.Time
	Finished bool `gorm:"-"`
	Comment string `gorm:"-"`
}
func (StudentTaskInfo) TableName() string {
	return "course_tasks"
}
func DeleteTask(taskid int) error {
	db := util.GetConn()
	err := db.Debug().Where("task_id=?", taskid).Unscoped().Delete(&CodeFile{}).Error
	if err != nil {
		return err
	}
	err = db.Debug().Where("id=?", taskid).Unscoped().Delete(&CourseTask{}).Error
	if err != nil {
		return err
	}
	return nil
}
func AddTask(task *CourseTask) error {
	db := util.GetConn()
	return db.Debug().Create(task).Error
}
func GetTaskInfo(task *TeacherTaskInfo)error{//教师；获取名单等等
	db := util.GetConn()
	err := db.Debug().Where("id=?", task.ID).Find(task).Error
	if err != nil {
		return err
	}
	err = db.Debug().Raw("SELECT AVG(mem_cost) AS avgmem,AVG(time_cost) AS avgtime,(SELECT uploader_id FROM code_files WHERE STATUS=0 AND task_id=9 ORDER BY  mem_cost ASC ) AS minmemstu,(SELECT uploader_id FROM code_files WHERE STATUS=0 AND task_id=9 ORDER BY  time_cost ASC ) AS mintimestu FROM code_files WHERE STATUS=0 AND task_id=?",task.ID).Find(task).Error
	if err != nil {
		return err
	}
	//获取名单
	var uploaders []BasicStudentInfo = make([]BasicStudentInfo, 0)
	if err := db.Debug().Raw("SELECT gs.id,gs.realname,gs.institute,gs.profession FROM gooj_stu_users gs INNER JOIN code_files cf ON cf.uploader_id=gs.ID WHERE cf.task_id=?",task.ID).Scan(&uploaders).Error; err != nil {
		return err
	}
	task.StudentList=uploaders
	return nil
}
func GetStudentTaskAnswer(stuid int,tskid int,stuCompInfo *CodeFile)error{
	db:=util.GetConn()
	return db.Debug().Find(stuCompInfo,"uploader_id=? AND task_id=?",stuid,tskid).Error
}
func TeacherComment(codeid int,comment string)error{
	db:=util.GetConn()
	return db.Debug().Model(&CodeFile{}).Where("id=?",codeid).Update("teacher_comment",comment).Error
}