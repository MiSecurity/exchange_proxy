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

package active_sync

import (
	"exchange_proxy/logger"
	"exchange_proxy/models"
	"exchange_proxy/vars"

	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func SendCode(user, deviceId, deviceType, phone, code string) (status bool) {
	status = true
	key := fmt.Sprintf("sms_%v_%v_%v", user, deviceId, phone)
	models.InitRedis()
	ret, err := vars.RedisInstance.HMGet(key, "times", "send_time").Result()
	// logger.Log.Infof("key: %v, ret: %v, len(ret): %v, err: %v", key, ret, len(ret), err)
	now := time.Now().Unix()
	if err == nil && len(ret) == 2 {
		t, _ := ret[0].(string)
		times, _ := strconv.Atoi(t)
		st, _ := ret[1].(string)
		iSendTime, _ := strconv.Atoi(st)
		sendTime := int64(iSendTime)

		if times <= 5 && now-sendTime > 60*60*10 {
			content := generateSmsContent(user, deviceId, code, "", deviceType)
			_, ok := SendSmsAPI(user, phone, deviceId, code, content, "")
			status = ok
			times = times + 1
			sendTime = now
			v := make(map[string]interface{})
			v["times"] = times
			v["send_time"] = sendTime
			_, _ = vars.RedisInstance.HMSet(key, v).Result()
		} else {
			status = false
		}
	}

	return status
}

// 重置短信状态，保证激活之后能再次收到短信，否则要等8小时之后了
func ResetSmsStatus(user, deviceId string) (err error) {
	phone, err := GetUserPhone(vars.UserPhoneUrl, user)
	if err == nil {
		key := fmt.Sprintf("sms_%v_%v_%v", user, deviceId, phone)
		models.InitRedis()
		vars.RedisInstance.Del(key)
	}
	return err
}

// 发送短信
func SendSmsAPI(username, phone, deviceId, code, content, srcIp string) (err error, result bool) {
	client := &http.Client{}
	postData := strings.NewReader(fmt.Sprintf("recipients=%s&content=%s", phone, content))
	req, err := http.NewRequest("POST", vars.SmsApiUrl, postData)
	if err == nil {
		req.Header.Add(vars.SmsApiHeader, vars.SmsApiKey)
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

		resp, err := client.Do(req)
		if err == nil {
			result = true
			defer resp.Body.Close()
			ret, err := ioutil.ReadAll(resp.Body)
			logger.Log.Infof("user:%v, deviceId: %v, code: %v, sms send result: %v, err: %v",
				username, deviceId, code, strings.TrimSpace(string(ret)), err)
		}
	}
	return err, result
}

// 生成短信内容
func generateSmsContent(username, deviceId, code, srcIp, deviceType string) (smsContent string) {
	// 获取用户手机设备信息
	deviceInfo, err := models.GetDeviceInfoByDeviceId(deviceId)
	if err == nil {
		phone := deviceInfo.PhoneNumber
		if deviceInfo.Model != "" {
			deviceType = deviceInfo.Model
		}
		if len(phone) > 5 {
			smsContent = fmt.Sprintf("您的邮箱 %v 正在一台新的设备（%v，手机号码为：%v）上登录，请点击链接查看详情并允许或拒绝，%v/a/?c=%v （此链接8小时后失效）",
				username, deviceType, phone, vars.ActiveUrl, code)
		}
	} else {
		smsContent = fmt.Sprintf("您的邮箱 %v 正在一台新的设备（%v）上登录，请点击链接查看详情并允许或拒绝，%v/a/?c=%v （此链接8小时后失效）",
			username, deviceType, vars.ActiveUrl, code)
	}

	return smsContent
}
