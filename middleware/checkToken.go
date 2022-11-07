package middleware

import (
	"Gooj/config"
	"Gooj/logger"
	"errors"
	"net/http"
	"time"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"gorm.io/gorm"
)

type GoojUser struct {
	Username string
	Password string
	Identity int
	ID       int `gorm:"primarykey"`
	gorm.Model
}
type JWT struct {
	SigningKey []byte
}

// 载荷，可以加一些自己需要的信息
type CustomClaims struct {
	ID       int
	Username string
	Identity int
	jwt.StandardClaims
}

// 一些常量
var (
	ErrTokenExpired     error = errors.New("token is expired")
	ErrTokenNotValidYet error = errors.New("token not active yet")
	ErrTokenMalformed   error = errors.New("that's not even a token")
	ErrTokenInvalid     error = errors.New("couldn't handle this token")
	SignKey          string
)

func init() {
	SignKey := config.GetJwtSettings().JwtSettings.PrivateKey
	if SignKey == "" {
		SignKey = "Nzxyyds"
	}
}

// CreateToken 生成一个token
func (j *JWT) CreateToken(claims CustomClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.SigningKey)
}

// JWTAuth 中间件，检查token
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Request.Header.Get("token")
		/*nzx := make(map[string]string,1) //注意该结构接受的内容
		  c.ShouldBind(&nzx)
		  token := nzx["token"]*/
		if token == "" {
			c.JSON(http.StatusOK, gin.H{
				"Status": config.NoTokenFound,
				"Info":   "No token found",
			})
			c.Abort()
			return
		}

		j := NewJWT()
		// parseToken 解析token包含的信息
		claims, err := j.ParseToken(token)
		if claims==nil{
			c.JSON(http.StatusOK, gin.H{
				"Status": config.TokenExpired,
				"Info":   "Token Expired",
			})
			logger.Debug("Token已过期")
			c.Abort()
			return
		}
		if !(claims.Identity==config.TypeStudent||claims.Identity==config.TypeTeacher){
			c.JSON(http.StatusOK, gin.H{
				"Status": config.InternalServerError,
				"Info":   "Identity not assessed",
			})
			logger.Warn("身份被篡改！")
			return
		}
		if err != nil {
			if err == ErrTokenExpired{
				c.JSON(http.StatusOK, gin.H{
					"Status": config.TokenExpired,
					"Info":   "Token Expired",
				})
				logger.Debug("Token已过期")
				c.Abort()
				return
			}
            logger.Info("未知错误-token无效!")
			c.JSON(http.StatusOK, gin.H{
				"Status": config.TokenInvaild,
				"Info":   "TokenInvaild",
			})
			c.Abort()
			return
		}
		c.Set("claims", claims)
	}
}

// JWT 签名结构

// 新建一个jwt实例
func NewJWT() *JWT {
	return &JWT{
		[]byte(GetSignKey()),
	}
}

// 获取signKey
func GetSignKey() string {
	return SignKey
}

// 这是SignKey
func SetSignKey(key string) string {
	SignKey = key
	return SignKey
}

func CheckLoginStat() gin.HandlerFunc {
	return func(ctx *gin.Context) {
	}
}

// 解析Token
func (j *JWT) ParseToken(tokenString string) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return j.SigningKey, nil
	})
	if err != nil {
		if ve, ok := err.(*jwt.ValidationError); ok {
			if ve.Errors&jwt.ValidationErrorMalformed != 0 {
				return nil, ErrTokenMalformed
			} else if ve.Errors&jwt.ValidationErrorExpired != 0 {
				return nil, ErrTokenExpired
			} else if ve.Errors&jwt.ValidationErrorNotValidYet != 0 {
				return nil, ErrTokenNotValidYet
			} else {
				return nil, ErrTokenInvalid
			}
		}
	}
	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, ErrTokenInvalid
}

// 生成令牌
func GenerateToken(c *gin.Context, id int, username string,identity int) {
	j := &JWT{
		[]byte(SignKey),
	}
	nissuer := config.GetJwtSettings().JwtSettings.Issuer
	timeout := config.GetJwtSettings().JwtSettings.ExpireTimeout
	if nissuer == "" {
		nissuer = "Nzxyyds"
	}
	if timeout == 0 {
		timeout = 3600
	}
	logger.Debug("已生成token")
	claims := CustomClaims{
		id,
		username,
		identity,
		jwt.StandardClaims{
			NotBefore: int64(time.Now().Unix() - 1000),           
			ExpiresAt: int64(time.Now().Unix() + int64(timeout)), 
			Issuer:    nissuer,                                   
		},
	}

	token, err := j.CreateToken(claims)

	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"Status": config.InternalServerError,
			"Info":    "UnknownErr",
		})
		logger.Warn("无法产生token!")
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"Status": config.Success,
		"token":  token,
	})
}
