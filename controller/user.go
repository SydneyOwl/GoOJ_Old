package controller

import (
	"Gooj/config"
	"Gooj/logger"
	"Gooj/middleware"
	"Gooj/model"
	"Gooj/util"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func Register(c *gin.Context) {
	var registerUserInfo *model.GoojUser = new(model.GoojUser)
	if err := c.ShouldBind(registerUserInfo); err != nil {
		logger.Warn("绑定失败")
		c.JSON(http.StatusOK, gin.H{
			"Status": config.ResolveInfoError,
			"Info":   "ResolveInfoError",
		})
		return
	}
	if registerUserInfo.CaptchaID == "" || registerUserInfo.CaptchaRes == "" || registerUserInfo.Username == "" || registerUserInfo.Password == "" {
		c.JSON(http.StatusOK, gin.H{
			"Status": config.ResolveInfoError,
			"msg":    "CHP or userinfo can't be empty",
		})
		logger.Debug("参数不足！")
		return
	}
	if !config.GetExpSettings().ExperimentalSettings.DisableCaptcha && !util.VerifyCaptcha(registerUserInfo.CaptchaID, registerUserInfo.CaptchaRes) {
		c.JSON(http.StatusOK, gin.H{
			"Status": config.CaptchaError,
			"msg":    "Captcha error",
		})
		logger.Debug("验证码错误！")
		return
	}
	goReg := func(isDupl interface{}) {
		var username string
		var errdb error
		var dupl bool
		var typeUser int
		switch userType := isDupl.(type) {
		case *model.GoojStuUser:
			typeUser = config.TypeStudent
			username = userType.GoojUser.Username
			dupl, errdb = model.IsDuplUser(username, config.TypeStudent)
		case *model.GoojTeaUser:
			typeUser = config.TypeTeacher
			username = userType.GoojUser.Username
			dupl, errdb = model.IsDuplUser(username, config.TypeTeacher)
		default:
			c.JSON(http.StatusOK, gin.H{
				"Status": config.InternalServerError,
				"Info":   "Unknown Error",
			})
			logger.Debug("未识别的类型！")
			return
		}
		if errdb == nil {
			if !dupl {
				if typeUser == config.TypeStudent {
					var newStu *model.GoojStuUser = new(model.GoojStuUser)
					newStu.GoojUser = *registerUserInfo
					newStu.GoojUser.LastLoginAt = time.Now()
					errdb = model.CreateUser(newStu)
				} else {
					var newTea *model.GoojTeaUser = new(model.GoojTeaUser)
					newTea.GoojUser = *registerUserInfo
					newTea.GoojUser.LastLoginAt = time.Now()
					errdb = model.CreateUser(newTea)
				}
				if errdb != nil {
					logger.Warn("数据库错误-无法创建用户！")
					c.JSON(http.StatusOK, gin.H{
						"Status": config.InternalServerError, "Info": "InternalServerError",
					})
					return
				} else {
					queryID := new(model.GoojUser)
					if typeUser == config.TypeStudent {
						model.GetUserByUsername(queryID, username, config.TypeStudent)
					} else {
						model.GetUserByUsername(queryID, username, config.TypeTeacher)
					}
					middleware.GenerateToken(c, queryID.ID, registerUserInfo.Username, registerUserInfo.Identity)
				}
			} else {
				logger.Debug("用户重复！")
				c.JSON(http.StatusOK, gin.H{
					"Status": config.DuplicatedUsername,
					"Info":   "DuplicatedUsername",
				})
				return
			}
		} else {
			logger.Warn("数据库出错！")
			c.JSON(http.StatusOK, gin.H{
				"Status": config.InternalServerError, "Info": "InternalServerError",
			})
			return
		}
	}
	// if

	if registerUserInfo.Identity == config.TypeTeacher {
		isDupl := new(model.GoojTeaUser)
		isDupl.GoojUser.Username = registerUserInfo.Username
		goReg(isDupl)
	} else if registerUserInfo.Identity == config.TypeStudent {
		isDupl := new(model.GoojStuUser)
		isDupl.GoojUser.Username = registerUserInfo.Username
		goReg(isDupl)
	} else {
		c.JSON(http.StatusOK, gin.H{
			"Status": config.InternalServerError,
			"Info":   "Unknown Error",
		})
		fmt.Println(registerUserInfo.Identity)
		logger.Debug("未识别的类型！")
		return
	}
}

// Login 登录
func Login(c *gin.Context) {
	var loginReq *model.GoojUser = new(model.GoojUser)
	if c.ShouldBind(loginReq) == nil {
		if loginReq.CaptchaID == "" || loginReq.CaptchaRes == "" || loginReq.Username == "" || loginReq.Password == "" {
			c.JSON(http.StatusOK, gin.H{
				"Status": config.ResolveInfoError,
				"Info":   "CHP or userinfo can't be empty",
			})
			logger.Debug("参数不足！")
			return
		}
		if !config.GetExpSettings().ExperimentalSettings.DisableCaptcha && !util.VerifyCaptcha(loginReq.CaptchaID, loginReq.CaptchaRes) {
			c.JSON(http.StatusOK, gin.H{
				"Status": config.CaptchaError,
				"Info":   "Captcha error",
			})
			logger.Debug("验证码错误！")
			return
		}
		if !(loginReq.Identity == config.TypeTeacher || loginReq.Identity == config.TypeStudent) {
			c.JSON(http.StatusOK, gin.H{
				"Status": config.InternalServerError,
				"Info":   "Unknown Error",
			})
			logger.Debug("未识别的类型！")
			return
		}
		var isMatch bool
		var err error
		if loginReq.Identity == config.TypeStudent {
			isMatch, err = model.IsPasswordMatch(loginReq, config.TypeStudent)
		} else {
			isMatch, err = model.IsPasswordMatch(loginReq, config.TypeTeacher)
		}
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"Status": config.InternalServerError,
				"Info":   "InternalServerHasError",
			})
			logger.Warn("匹配账号密码失败")
		}
		if isMatch {
			middleware.GenerateToken(c, loginReq.ID, loginReq.Username, loginReq.Identity)
			if loginReq.Identity == config.TypeStudent {
				model.UpdateLastLoginTime(loginReq, config.TypeStudent)
			} else {
				model.UpdateLastLoginTime(loginReq, config.TypeTeacher)
			}
			logger.Debug("已更新用户！")
			return
		} else {
			c.JSON(http.StatusOK, gin.H{
				"Status": config.WrongUsernameOrPassword,
				"Info":   "WrongUsernameOrPassword",
			})
			logger.Debug("帐密错误！")
			return
		}
	} else {
		c.JSON(http.StatusOK, gin.H{
			"Status": config.InternalServerError,
			"Info":   "Unknown Error",
		})
		logger.Info("绑定错误！")
		return
	}
}
func CheckIsExpired(c *gin.Context) {
	token := c.Request.Header.Get("token")
	// token,_ := c.GetPostForm("token")
	// token := c.Request.Header.Get("token")
	// nzx := make(map[string]string, 1) //注意该结构接受的内容
	// c.BindJSON(&nzx)
	// token := nzx["token"]
	if token == "" {
		c.JSON(http.StatusOK, gin.H{
			"Status": 700,
			"Info":   "No token found",
		})
		logger.Debug("Token未找到！")
		return
	}
	j := middleware.NewJWT()
	// parseToken 解析token包含的信息
	_, err := j.ParseToken(token)
	if err != nil {
		if err == middleware.ErrTokenExpired {
			c.JSON(http.StatusOK, gin.H{
				"Status": 701,
				"Info":   "Token Expired",
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"Status": 702,
			"Info":   "InvaildToken",
		})
		return
	} else {
		c.JSON(http.StatusOK, gin.H{
			"Status": 200,
			"Info":   "TokenUseOK",
		})
	}
}
func UpdateMyInfo(c *gin.Context) {
	//
	userData, err1 := c.Get("claims")
	if userData.(*middleware.CustomClaims).Identity == config.TypeTeacher {
		var updateUser *model.GoojTeaUser = new(model.GoojTeaUser)
		err := c.ShouldBind(updateUser)
		if err != nil {
			logger.Info("Bind user err!")
			c.JSON(http.StatusOK, gin.H{
				"Info":   "InternalServerError",
				"Status": config.ResolveInfoError,
			})
			return
		}
		if !err1 {
			logger.Error("JWT信息获取失败！")
			c.JSON(http.StatusOK, gin.H{
				"Info":   "InternalServerError",
				"Status": config.ResolveInfoError,
			})
			return
		}
		// userID := userData.(*middleware.CustomClaims).ID
		updateUser.GoojUser.ID = userData.(*middleware.CustomClaims).ID
		updateUser.GoojUser.UpdatedInfoAt = time.Now()
		if model.UpdateInfo(updateUser, config.TypeTeacher) != nil {
			c.JSON(http.StatusOK, map[string]interface{}{
				"Status": config.InternalServerError,
				"Info":   "InternalServerErrorPin1",
			})
			logger.Error("无法更新用户")
			return
		} else {
			c.JSON(http.StatusOK, gin.H{
				"Status": config.Success,
				"Info":   "Update Doneee",
			})
		}
	} else {
		var updateUser *model.GoojStuUser = new(model.GoojStuUser)
		err := c.ShouldBind(updateUser)
		if err != nil {
			logger.Info("Bind user err!")
			c.JSON(http.StatusOK, gin.H{
				"Info":   "InternalServerError",
				"Status": config.ResolveInfoError,
			})
			return
		}
		if !err1 {
			logger.Error("JWT信息获取失败！")
			c.JSON(http.StatusOK, gin.H{
				"Info":   "InternalServerError",
				"Status": config.ResolveInfoError,
			})
			return
		}
		// userID := userData.(*middleware.CustomClaims).ID
		updateUser.GoojUser.ID = userData.(*middleware.CustomClaims).ID
		updateUser.GoojUser.UpdatedInfoAt = time.Now()
		if model.UpdateInfo(updateUser, config.TypeStudent) != nil {
			c.JSON(http.StatusOK, map[string]interface{}{
				"Status": config.InternalServerError,
				"Info":   "InternalServerErrorPin1",
			})
			logger.Error("无法更新用户")
			return
		} else {
			c.JSON(http.StatusOK, gin.H{
				"Status": config.Success,
				"Info":   "Update Doneee",
			})
		}
	}
}

//首页
func GetBasicInfo(c *gin.Context) {
	userData, err1 := c.Get("claims")
	if !err1 {
		logger.Error("JWT信息获取失败！")
		c.JSON(http.StatusOK, gin.H{
			"Info":   "InternalServerError",
			"Status": config.ResolveInfoError,
		})
		return
	}
	userID := userData.(*middleware.CustomClaims).ID
	if userData.(*middleware.CustomClaims).Identity == config.TypeStudent {
		// userID := userData.(*middleware.CustomClaims).ID
		query, err := model.GetUserById(userID, config.TypeStudent)
		queryID := query.(*model.GoojStuUser)
		var courseInfo []model.BasicCourseInfo
		var timeline []model.TimeLine
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"Status": config.InternalServerError,
				"Info":   "InternalServerErrorPin1",
			})
			logger.Warn("用户信息绑定失败！")
			return
		}
		model.GetBasicInfo(userID, &courseInfo, &timeline)
		isNotFilled := queryID.Realname == "" || queryID.StudentNumber == "" || queryID.Institute == "" || queryID.Profession == "" || queryID.SelfIntroduction == ""
		c.JSON(http.StatusOK, gin.H{
			"Status": config.Success,
			"Info": gin.H{
				"CourseInfo": courseInfo,
				"Timeline":   timeline,
				"PersonalInfo": gin.H{
					"IsFilled":         !isNotFilled,
					"Institute":        queryID.Institute,
					"Profession":       queryID.Profession,
					"Realname":         queryID.Realname,
					"SelfIntroduction": queryID.SelfIntroduction,
					"StudentNumber":    queryID.StudentNumber,
				},
			},
		})
	} else {
		// userID := userData.(*middleware.CustomClaims).ID
		query, err := model.GetUserById(userID, config.TypeTeacher)
		queryID := query.(*model.GoojTeaUser)
		var courseInfo []model.BasicCourseInfo
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"Status": config.InternalServerError,
				"Info":   "InternalServerErrorPin1",
			})
			logger.Warn("用户信息绑定失败！")
			return
		}
		err = model.GetTeachingCourses(userID, &courseInfo)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"Status": config.InternalServerError,
				"Info":   "InternalServerErrorPin1",
			})
			logger.Warn("dberror")
			return
		}
		isNotFilled := queryID.Realname == "" || queryID.TeacherNumber == "" || queryID.Institute == "" || queryID.Profession == "" || queryID.SelfIntroduction == ""
		c.JSON(http.StatusOK, gin.H{
			"Status": config.Success,
			"Info": gin.H{
				"PersonalInfo": gin.H{
					"IsFilled":         !isNotFilled,
					"Institute":        queryID.Institute,
					"Profession":       queryID.Profession,
					"Realname":         queryID.Realname,
					"SelfIntroduction": queryID.SelfIntroduction,
					"StudentNumber":    queryID.TeacherNumber,
				},
				"TeachingCourse": courseInfo,
			},
		})
	}
}
func ChooseCourse(c *gin.Context) {
	userData, _ := c.Get("claims")
	user := userData.(*middleware.CustomClaims)
	if user.Identity == config.TypeTeacher {
		c.JSON(http.StatusOK, gin.H{
			"Status": config.PermissionDenied,
			"Info":   "Teacher is not allowed",
		})
		logger.Debug("Teacher denied")
		return
	}
	idmap := make(map[string]interface{})
	var err error
	c.ShouldBindJSON(&idmap)
	if idmap["type"] == "add" {
		dupl, err1 := model.IsCourseDupl(int(idmap["ID"].(float64)), user.ID)
		if dupl {
			c.JSON(http.StatusOK, gin.H{
				"Status": config.DuplicatedCourse,
				"Info":   "course duplll",
			})
			return
		}
		if err1 != nil {
			c.JSON(http.StatusOK, gin.H{
				"Status": config.InternalServerError,
				"Info":   "",
			})
			return
		}
		err = model.ChooseCourse(int(idmap["ID"].(float64)), user.ID)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"Status": config.InternalServerError,
				"Info":   "",
			})
			return
		} else {
			c.JSON(http.StatusOK, gin.H{
				"Status": config.Success,
				"Info":   "",
			})
		}
	} else {
		err = model.DeleteChoosedCourse(int(idmap["ID"].(float64)), user.ID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusOK, gin.H{
					"Status": config.DuplicatedCourse,
					"Info":   "No such course",
				})
			} else {
				c.JSON(http.StatusOK, gin.H{
					"Status": config.InternalServerError,
					"Info":   "",
				})
			}
		} else {
			c.JSON(http.StatusOK, gin.H{
				"Status": config.Success,
				"Info":   "",
			})
		}
	}
}
