package config

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)
type LogConf struct{
	OutputLevel string `yaml:"OutputLevel"`
	LogFile LogFile `yaml:"LogFile"`
}
type LogSettings struct{
	LogData LogConf `yaml:"LogSettings"`
}
type LogFile struct {
	LogPath string `yaml:"LogPath"`
	MaxSize int `yaml:"MaxSize"`
	MaxAge int `yaml:"MaxAge"`
	MaxBackups int `yaml:"MaxBackups"`
	Compress bool `yaml:"Compress"`
}

type EnvConf struct{
	TempCodeStoragePath string `yaml:"TempCodeStoragePath"`
	Sandbox Sandbox `yaml:"Sandbox"`
	Golang Golang `yaml:"Golang"`
}
type Sandbox struct{
	BinaryPath string `yaml:"BinaryPath"`
	StorageTimeout string `yaml:"StorageTimeout"`
	Address string `yaml:"Address"`
	Port int `yaml:"Port"`
	TimeLimit uint64 `yaml:"TimeLimit"`
	MemoryLimit uint64 `yaml:"MemoryLimit"`
	StackLimit uint64 `yaml:"StackLimit"`
	ProcLimit uint64 `yaml:"ProcLimit"`
	CpuRateLimit uint64 `yaml:"CpuRateLimit"`
}
type Golang struct{
	Gopath string `yaml:"GoPath"`
	GocodePath string `yaml:"GocodePath"`
	GofmtPath string `yaml:"GofmtPath"`
}
type EnvironmentSettings struct{
	EnvironmentSettings EnvConf `yaml:"EnvironmentSettings"`
}
type ExperimentalConf struct{
	EnableAutoFmt bool `yaml:"EnableAutoFmt"`
	DisableCaptcha bool `yaml:"DisableCaptcha"`
}
type ExperimentalSettings struct{
	ExperimentalSettings ExperimentalConf `yaml:"ExperimentalSettings"`
}

type DatabaseConf struct{
	Address string `yaml:"Address"`
	Port int `yaml:"Port"`
	User string `yaml:"User"`
	Password string `yaml:"Password"`
	DBName string `yaml:"DBName"`
}
type DatabaseSettings struct{
	DatabaseSettings DatabaseConf `yaml:"DatabaseSettings"`
}
type JwtConf struct{
	PrivateKey string `yaml:"PrivateKey"`
	Issuer string `yaml:"Issuer"`
	ExpireTimeout int `yaml:"ExpireTimeout"`
}
type JwtSettings struct{
	JwtSettings JwtConf `yaml:"JwtSettings"`
}
var(
	logSettings LogSettings
	envSettings EnvironmentSettings
	expSettings ExperimentalSettings
	dbSettings DatabaseSettings
	jwtSettings JwtSettings
)

func init(){
	yamlFile, err := ioutil.ReadFile("./config.yaml")
	if err != nil {
		panic("配置文件读取失败！")
	}
	err = yaml.Unmarshal(yamlFile, &logSettings)
	if err != nil {
		panic("配置文件存在错误！")
	}
	if logSettings.LogData.OutputLevel==""{
		logSettings.LogData.OutputLevel="Debug"
	}
	if logSettings.LogData.LogFile.MaxSize==0{
		logSettings.LogData.LogFile.MaxAge=128
	}
	if logSettings.LogData.LogFile.MaxAge==0{
		logSettings.LogData.LogFile.MaxAge=7
	}
	if logSettings.LogData.LogFile.MaxBackups==0{
		logSettings.LogData.LogFile.MaxBackups=30
	}
	err = yaml.Unmarshal(yamlFile, &envSettings)
	if err != nil {
		panic("配置文件存在错误！")
	}
	err = yaml.Unmarshal(yamlFile, &expSettings)
	if err != nil {
		panic("配置文件存在错误！")
	}
	err = yaml.Unmarshal(yamlFile, &dbSettings)
	if err != nil {
		panic("配置文件存在错误！")
	}
	err = yaml.Unmarshal(yamlFile, &jwtSettings)
	if err != nil {
		panic("配置文件存在错误！")
	}
}
func GetLogSettings() LogSettings{
	return logSettings
}
func GetEnvSettings() EnvironmentSettings{
	return envSettings
}
func GetExpSettings() ExperimentalSettings{
	return expSettings
}
func GetDBSettings() DatabaseSettings{
	return dbSettings
}
func GetJwtSettings() JwtSettings{
	return jwtSettings
}