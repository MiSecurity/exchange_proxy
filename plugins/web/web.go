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

package web

import (
	"encoding/json"
	"exchange_proxy/logger"
	"exchange_proxy/models"
	"exchange_proxy/plugins/active_sync"
	"html/template"
	"net/http"
)

type (
	ApiData struct {
		Name         string
		DeviceModel  string
		DeviceType   string
		DeviceId     string
		Imei         string
		PhoneNumber  string
		ActiveStatus string
		Code         string
	}

	RespData struct {
		Code    int
		Data    string
		Message string
	}
)

func Activation(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	c := r.FormValue("c")

	tSync, _ := template.ParseFiles("plugins/web/templates/activeSync.html")
	tState, _ := template.ParseFiles("plugins/web/templates/deviceState.html")

	var (
		deviceName string
		deviceMode string
		imei       string
		deviceId   string
		deviceNum  int

		apiDate ApiData
	)

	if c != "" {
		// 检查激活码
		err, has, user, deviceInfo := active_sync.CheckActiveCode(c)
		// 激活码是存在的
		if err == nil && has {
			// 获取当前用户的设备数
			deviceNum = models.GetDeviceNum(user)
			deviceId = deviceInfo.DeviceId

			apiDate.Name = user
			apiDate.DeviceId = deviceId
			apiDate.DeviceModel = deviceInfo.DeviceType
			apiDate.DeviceType = deviceInfo.DeviceType
			apiDate.Code = c

			// 取出手机端通过WBXML协议获取的数据
			wbxmlInfo, err := models.GetDeviceInfoByDeviceId(deviceId)
			if err == nil {
				deviceName = wbxmlInfo.FriendlyName
				deviceName = wbxmlInfo.Model
				imei = wbxmlInfo.IMEI
				phone := wbxmlInfo.PhoneNumber

				apiDate.DeviceModel = deviceMode
				apiDate.Imei = imei
				apiDate.DeviceType = deviceName
				apiDate.PhoneNumber = phone
			}
			CodeStatus := deviceInfo.State
			deviceStatue := models.GetDeviceState(user, deviceId)
			logger.Log.Infof("user: %v, deviceId: %v, CodeStatus: %v, deviceStatue: %v", user, deviceId, CodeStatus, deviceStatue)
			switch CodeStatus {
			case 0:
				// 显示已激活的状态页面
				if deviceStatue == 0 {
					apiDate.ActiveStatus = "STATE_ACTIVED"
					_ = tState.Execute(w, apiDate)
					// 显示已拒绝状态的页面
				} else if deviceStatue == 3 {
					apiDate.ActiveStatus = "STATE_REJECTED"
					_ = tState.Execute(w, apiDate)
				}
			default:
				// 激活的设备数已经超过10个
				if deviceNum >= 10 {
					apiDate.ActiveStatus = "STATE_EXCEED"
					_ = tState.Execute(w, apiDate)
				} else {
					// 显示激活页面
					_ = tSync.Execute(w, apiDate)
				}
			}
		} else {
			// 激活码不存在
			apiDate.ActiveStatus = "STATE_INVALID"
			_ = tState.Execute(w, apiDate)
		}
	} else {
		//	激活码为空
		apiDate.ActiveStatus = "STATE_INVALID"
		_ = tState.Execute(w, apiDate)
	}
}

func ActiveDevice(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		_ = r.ParseForm()
		c := r.FormValue("c")
		respData := RespData{Code: 500, Data: "", Message: "设备不存在"}
		if c != "" {
			result, user, device := active_sync.VerifyActiveCode(c)
			if result {
				_ = models.ActiveDevice(user, device.DeviceId)
				_, _ = active_sync.ResetActiveCodeStatus(c, user, device, 0)
				respData.Code = 200
				respData.Message = "已经允许设备访问"
			}
		}
		data, _ := json.Marshal(respData)
		_, _ = w.Write(data)
	}
}

func IgnoreDevice(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		_ = r.ParseForm()
		c := r.FormValue("c")
		respData := RespData{Code: 500, Data: "", Message: "设备不存在"}
		if c != "" {
			result, user, device := active_sync.VerifyActiveCode(c)
			if result {
				_ = models.IgnoreDevice(user, device.DeviceId)
				_, _ = active_sync.ResetActiveCodeStatus(c, user, device, 3)
				respData.Code = 200
				respData.Message = "已忽略该设备"
			}
		}
		data, _ := json.Marshal(respData)
		_, _ = w.Write(data)
	}
}

func NotFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(404)
}
