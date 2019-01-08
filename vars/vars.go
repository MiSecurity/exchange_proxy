/*

Copyright (c) 2018 sec.lu

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THEq
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.

*/

package vars

import (
	"exchange_proxy/logger"
	"exchange_proxy/settings"

	"path/filepath"
	"strings"

	"github.com/go-redis/redis"
	"github.com/sirupsen/logrus"
	"github.com/vulcand/oxy/forward"
)

type (
	Config struct {
		Host       []string
		Backend    string
		Port       int
		TLS        bool
		Cert       string
		Key        string
		DebugLevel string
	}

	RedisConfig struct {
		Host     string
		Port     int
		Db       int
		Password string
	}
)

var (
	FwdOWA  *forward.Forwarder
	FwdSync *forward.Forwarder

	RedisInstance *redis.Client

	MailConfig Config
	RedisConf  RedisConfig

	RequestCmds  = make(map[string]bool)
	ResponseCmds = make(map[string]bool)
	CurDir       string
)

// 代理用到的API接口的信息
var (
	// 用户激活的URL
	ActiveUrl string

	//OTP接口
	OtpUrl string

	// 短信接口的URL，header头名称与值
	SmsApiUrl    string
	SmsApiHeader string
	SmsApiKey    string

	// 根据用户名查找手机号的接口的URL
	UserPhoneUrl string
)

func init() {

	sec := settings.Cfg.Section("mail")
	hosts := sec.Key("hosts").MustString("mail.xsec.io")
	MailConfig.Host = strings.Split(hosts, ",")
	MailConfig.Port = sec.Key("port").MustInt(443)
	MailConfig.Backend = sec.Key("backend").MustString("https://8.8.8.8")
	MailConfig.TLS = sec.Key("ssl").MustBool(true)
	MailConfig.Cert = sec.Key("cert").MustString("certs/ca.crt")
	MailConfig.Key = sec.Key("key").MustString("certs/ca.key")
	MailConfig.DebugLevel = sec.Key("debug_level").MustString("info")

	secRedis := settings.Cfg.Section("redis")
	RedisConf.Host = secRedis.Key("host").MustString("127.0.0.1")
	RedisConf.Port = secRedis.Key("port").MustInt(6379)
	RedisConf.Db = secRedis.Key("db").MustInt(0)
	RedisConf.Password = secRedis.Key("password").MustString("passw0rd")

	FwdOWA, _ = forward.New()
	FwdSync, _ = forward.New()

	// 获取当前目录
	CurDir, _ = GetCurDir()
	// 初始化过滤指令列表
	initActiveSyncCmds()
	// 初始化代理用的到API接口的值
	initApiInfo()
	// 初始化日志级别
	initDebugLevel()

	logger.Log.Printf("host: %v, port: %v, ssl: %v, path: %v", strings.Join(MailConfig.Host, ","), MailConfig.Port, MailConfig.TLS, CurDir)

}

// 手机端未激活前的过滤指令列表
func initActiveSyncCmds() {
	// 请求的指令
	RequestCmds["SendMail"] = true
	RequestCmds["FolderCreate"] = true
	RequestCmds["FolderDelete"] = true
	RequestCmds["FolderUpdate"] = true
	RequestCmds["MeetingResponse"] = true
	RequestCmds["ItemOperations"] = true
	RequestCmds["SmartForward"] = true
	RequestCmds["SmartReply"] = true
	RequestCmds["MoveItems"] = true

	// 响应的指令
	ResponseCmds["Sync"] = true
	ResponseCmds["Search"] = true
	ResponseCmds["GetAttachment"] = true
	ResponseCmds["GetItemEstimate"] = true
	ResponseCmds["MeetingResponse"] = true
}

// 初始化API接口的值
func initApiInfo() {
	sec := settings.Cfg.Section("otp")
	OtpUrl = sec.Key("url").MustString("")

	sec = settings.Cfg.Section("sms")
	SmsApiUrl = sec.Key("url").MustString("")
	SmsApiHeader = sec.Key("header").MustString("")
	SmsApiKey = sec.Key("key").MustString("")

	secUserInfo := settings.Cfg.Section("user_info")
	UserPhoneUrl = secUserInfo.Key("user_phone").MustString("")
	ActiveUrl = secUserInfo.Key("active_url").MustString("")
}

// 初始化log的级别
func initDebugLevel() {
	level := strings.ToLower(MailConfig.DebugLevel)
	switch level {
	case "info":
		logger.Log.Logger.SetLevel(logrus.InfoLevel)
	case "debug":
		logger.Log.Logger.SetLevel(logrus.DebugLevel)
	case "warn":
		logger.Log.Logger.SetLevel(logrus.WarnLevel)
	case "error":
		logger.Log.Logger.SetLevel(logrus.ErrorLevel)
	case "fatal":
		logger.Log.Logger.SetLevel(logrus.FatalLevel)
	default:
		logger.Log.Logger.SetLevel(logrus.InfoLevel)
	}
}

func GetCurDir() (path string, err error) {
	path, err = filepath.Abs(".")
	return path, err
}
