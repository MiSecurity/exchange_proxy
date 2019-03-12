/*

Copyright (c) 2018 sec.xiaomi.com

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

package active_sync

import (
	"bytes"
	"crypto/tls"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"exchange_proxy/logger"
	"exchange_proxy/models"
	"exchange_proxy/util/wbxml"
	"exchange_proxy/vars"

	"github.com/toolkits/slice"
	"github.com/vulcand/oxy/forward"
	"github.com/vulcand/oxy/testutils"
)

func filterCmd(resp *http.Response) (err error) {
	var bodyBytes []byte
	bodyBytes, err = ioutil.ReadAll(resp.Request.Body)
	// 恢复Req.body的值给ParseForm函数使用
	resp.Request.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
	values := resp.Request.URL.Query()
	if len(values) == 4 {
		cmd := values["Cmd"][0]
		mail := values["User"][0]
		deviceId := values["DeviceId"][0]
		user := ""
		// 取邮箱前缀
		u := strings.Split(mail, "@")
		if len(u) == 2 {
			user = u[0]
		}
		// 取域的\后面部分
		u1 := strings.Split(mail, "\\")
		if len(u1) == 2 {
			user = u1[1]
		}

		deviceState := models.GetDeviceState(user, deviceId)
		logger.Log.Infof("in filterCmd, user: %v, deviceId: %v, deviceState: %v", user, deviceId, deviceState)
		// 设备未激活时，过滤掉特定指令的返回值
		if deviceState != 0 {
			if models.CheckCmd(cmd, vars.ResponseCmds) {
				resp.Body = ioutil.NopCloser(bytes.NewBuffer([]byte("")))
			}
		}
	}

	// 恢复Req.body的值传到下一个处理器中
	resp.Request.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
	return err
}

func SyncRedirect(w http.ResponseWriter, req *http.Request) {
	if slice.ContainsString(vars.MailConfig.Host, req.Host) {
		req.URL = testutils.ParseURI(vars.MailConfig.Backend)
		vars.FwdSync.ServeHTTP(w, req)
	} else {
		w.WriteHeader(444)
	}
}

func ActiveSyncHandler(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {

		if vars.MailConfig.TLS {
			r := forward.RoundTripper(
				&http.Transport{
					TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
				},
			)
			// 对手机设备的处理逻辑
			if strings.HasPrefix(req.RequestURI, "/Microsoft-Server-ActiveSync") {
				values := req.URL.Query()
				if len(values) > 0 {
					mail := values["User"][0]
					deviceId := values["DeviceId"][0]
					deviceType := values["DeviceType"][0]
					cmd := values["Cmd"][0]

					user := ""
					// 取邮箱前缀
					u := strings.Split(mail, "@")
					if len(u) == 2 {
						user = u[0]
					}
					// 取域的\后面部分
					u1 := strings.Split(mail, "\\")
					if len(u1) == 2 {
						user = u1[1]
					}

					device := models.Device{User: user, DeviceId: deviceId, DeviceType: deviceType,
						State: 1, Time: time.Now().Unix()}

					// 如果设备ID为0或为空，直接退出请求
					if len(deviceId) <= 1 {
						w.WriteHeader(444)
					}

					// 获取设备信息的详细信息
					var bodyBytes []byte
					bodyBytes, err := ioutil.ReadAll(req.Body)
					if err == nil {
						// 判断大小，否则可能会把邮件的附件传进来处理
						if len(bodyBytes) < 500 {
							deviceInfo, err := wbxml.Parse(bodyBytes)
							if err == nil && deviceInfo.Model != "" {
								logger.Log.Debugf("deviceInfo: %v, err: %v", deviceInfo, err)
								_ = models.SetDeviceInfo(deviceId, deviceInfo)
							}
						}
					}
					// 用完后恢复req.Body的值，否则之后的处理器不能再用了
					req.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))

					userState, _ := models.GetUserState(user)
					// logger.Log.Infof("user: %v, state: %v", user, userState)
					// 用户状态为1，表示已经锁定的账户，直接退出请求
					if userState == 1 {
						w.WriteHeader(444)
					} else if userState < 0 {
						_ = models.NewUser(user)
					}

					deviceState := models.GetDeviceState(user, deviceId)
					logger.Log.Debugf("user: %v, deviceId: %v, deviceState: %v, deviceType: %v, Cmd: %v",
						user, deviceId, deviceState, deviceType, cmd)
					if deviceState < 0 {
						// 设置不存在时，创建设备
						_ = models.NewDevice(user, device)
						// 未激活前需要过滤掉可以返回给客户端数据的指令及返回的数据
						vars.FwdSync, _ = forward.New(forward.PassHostHeader(true), forward.Logger(logger.Log.Logger),
							r, forward.ResponseModifier(filterCmd))
						h(w, req)
					} else if deviceState == 0 {
						// 设备已激活就直接放行
						vars.FwdSync, _ = forward.New(forward.PassHostHeader(true), forward.Logger(logger.Log.Logger), r, forward.Stream(true))
						h(w, req)
					} else if deviceState == 1 {
						// 设备未激活时，进入激活流程
						DoActiveDevice(user, deviceId, deviceType, device)

						// 未激活前需要过滤掉可以返回给客户端数据的指令及返回的数据
						vars.FwdSync, _ = forward.New(forward.PassHostHeader(true), forward.Logger(logger.Log.Logger),
							r, forward.ResponseModifier(filterCmd))
						h(w, req)
						if models.CheckCmd(cmd, vars.RequestCmds) {
							w.WriteHeader(401)
						}
					} else {
						// 其他状态为锁定或阻止，不代理到后端，在代理层面直接返回
						w.WriteHeader(200)
					}
				} else {
					// 对于OPTIONS指令，暂时直接透传到后端
					vars.FwdSync, _ = forward.New(forward.PassHostHeader(true), forward.Logger(logger.Log.Logger), r)
					h(w, req)
				}

			}
		}
	}
}
