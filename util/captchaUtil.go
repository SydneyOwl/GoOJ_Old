package util

import (
	"Gooj/config"
	"Gooj/logger"
	"image/color"
	"math/rand"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mojocn/base64Captcha"
)
var store = base64Captcha.DefaultMemStore

// 图形验证码
func NewCaptcha(c *gin.Context) {
	rand.Seed(time.Now().UnixNano())
	var ch *base64Captcha.Captcha
	switch rand.Intn(2) {
	case 0:
		driver := base64Captcha.DriverMath{
			Height:          60,   // 验证码图片高度
			Width:           120,   // 验证码图片宽度
			NoiseCount:      0,     // 干扰词数量
			ShowLineOptions: 2 | 4, // 线条数量
			BgColor: &color.RGBA{ // 背景颜色
				R: 128,
				G: 98,
				B: 112,
				A: 0,
			},
			Fonts: []string{"ApothecaryFont.ttf"}, // 字体
		}

		ch = base64Captcha.NewCaptcha(&driver, store)
	case 1:
		driver := base64Captcha.DriverString{
			Height:          60,
			Width:           120,
			NoiseCount:      0,
			ShowLineOptions: 2 | 4,
			Length:          5,
			Source:          "1234567890abcdefghijklmnopqrstuvwxyz",
			BgColor: &color.RGBA{ // 背景颜色
				R: 128,
				G: 98,
				B: 112,
				A: 0,
			},
			Fonts: []string{"ApothecaryFont.ttf"},
		}
		ch = base64Captcha.NewCaptcha(&driver, store)
	default:
		logger.Warn("无法产生随机数！")
		c.JSON(http.StatusOK, gin.H{
			"Status": config.InternalServerError,
			"Info":   "InternamServerError",
		})
		return
	}
	id, b64s, err := ch.Generate()
	if err != nil {
		logger.Warn("无法产生验证码！")
		c.JSON(http.StatusOK, gin.H{
			"Status": config.InternalServerError,
			"Info":   "Gen CapErr",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"Status": config.Success,
		"ID":     id,
		"Captcha":    b64s,
		"Info":   "",
	})
}
func VerifyCaptcha(captchaID string,captchaAns string)bool{
		return store.Verify(captchaID,captchaAns,true)
}