package logger

import (
	"Gooj/config"
	"os"

	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	log      *zap.Logger
	infoLog  = &lumberjack.Logger{}
	errorLog = &lumberjack.Logger{}
)

func InitLog() {
	conf := config.GetLogSettings()
	var coreArr []zapcore.Core
	//获取编码器
	encoderConfig := zap.NewProductionEncoderConfig()     //NewJSONEncoder()输出json格式，NewConsoleEncoder()输出普通文本格式
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder //指定时间格式
	encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder    //按级别显示不同颜色，不需要的话取值zapcore.CapitalLevelEncoder就可以了
	//encoderConfig.EncodeCaller = zapcore.FullCallerEncoder        //显示完整文件路径
	encoder := zapcore.NewConsoleEncoder(encoderConfig)

	//日志级别
	var levelOutput = zap.DebugLevel
	switch conf.LogData.OutputLevel {
	case "Debug":
		levelOutput = zap.DebugLevel
	case "Info":
		levelOutput = zap.InfoLevel
	case "Warn":
		levelOutput = zap.WarnLevel
	case "Error":
		levelOutput = zap.ErrorLevel
	case "DPanic":
		levelOutput = zap.DPanicLevel
	case "Panic":
		levelOutput = zap.PanicLevel
	case "Fatal":
		levelOutput = zap.FatalLevel
	default:
		panic("警告：日志过滤等级错误")
	}
	highPriority := zap.LevelEnablerFunc(func(lev zapcore.Level) bool { //error级别
		return lev >= levelOutput
	})
	lowPriority := zap.LevelEnablerFunc(func(lev zapcore.Level) bool { //info和debug级别,debug级别是最低的
		return lev < levelOutput && lev >= zap.DebugLevel
	})
	if conf.LogData.LogFile.LogPath != "" {
		infoLog = &lumberjack.Logger{
			Filename:   conf.LogData.LogFile.LogPath+"/normal.log",                //日志文件存放目录，如果文件夹不存在会自动创建
			MaxSize:    conf.LogData.LogFile.MaxSize,    //文件大小限制,单位MB
			MaxBackups: conf.LogData.LogFile.MaxBackups, //最大保留日志文件数量
			MaxAge:     conf.LogData.LogFile.MaxAge,     //日志文件保留天数
			Compress:   conf.LogData.LogFile.Compress,   //是否压缩处理
		}
		errorLog = &lumberjack.Logger{
			Filename:   conf.LogData.LogFile.LogPath+"/error.log",               //日志文件存放目录，如果文件夹不存在会自动创建
			MaxSize:    conf.LogData.LogFile.MaxSize,    //文件大小限制,单位MB
			MaxBackups: conf.LogData.LogFile.MaxBackups, //最大保留日志文件数量
			MaxAge:     conf.LogData.LogFile.MaxAge,     //日志文件保留天数
			Compress:   conf.LogData.LogFile.Compress,   //是否压缩处理
		}
	}
	//info文件writeSyncer
	infoFileCore := zapcore.NewCore(encoder, zapcore.NewMultiWriteSyncer(zapcore.AddSync(infoLog)), lowPriority) //第三个及之后的参数为写入文件的日志级别,ErrorLevel模式只记录error级别的日志
	//error文件writeSyncer
	errorFileCore := zapcore.NewCore(encoder, zapcore.NewMultiWriteSyncer(zapcore.AddSync(errorLog), zapcore.AddSync(os.Stdout)), highPriority) //第三个及之后的参数为写入文件的日志级别,ErrorLevel模式只记录error级别的日志

	coreArr = append(coreArr, infoFileCore)
	coreArr = append(coreArr, errorFileCore)
	log = zap.New(zapcore.NewTee(coreArr...)) //zap.AddCaller()为显示文件名和行号，可省略
}
func Debug(err string) {
	log.Debug(err)
}
func Info(err string) {
	log.Info(err)
}
func Warn(err string) {
	log.Warn(err)
}
func Error(err string) {
	log.Error(err)
}
func DPanic(err string) {
	log.DPanic(err)
}
func Panic(err string) {
	log.Panic(err)
}
func Fatal(err string){
	log.Fatal(err)
}
func GetLogger() *zap.Logger{
	return log
}