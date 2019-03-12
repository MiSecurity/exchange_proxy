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
	"exchange_proxy/models"
	"exchange_proxy/util"
	"exchange_proxy/vars"
	"strings"

	"encoding/json"
	"fmt"
	"time"
)

func DoActiveDevice(user, deviceId, deviceType string, device models.Device) {
	_, isActive := GetDeviceActiveFlag(user, deviceId)
	if !isActive {
		// 设置进入设备激活流程的FLAG
		SetDeviceActiveFlag(user, deviceId)
		// 获取用户手机号
		phone, err := GetUserPhone(vars.UserPhoneUrl, user)
		if err == nil {
			// 生成激活码
			_, code := SetActiveCode(user, deviceId, device)
			// 发送激活码
			smsStatus := SendCode(user, deviceId, deviceType, phone, code)
			// 发送成功的话删除已发送过的激活码
			if smsStatus {
				_ = DelActiveCodeFlag(user, deviceId)
			}
		}
	} else {
		// logger.Log.Infof("%v, 已经激活过，忽略请求", device)
		// 已经激活过，忽略请求
	}
}

// 获取设备激活标识
func GetDeviceActiveFlag(username, deviceId string) (err error, result bool) {
	models.InitRedis()
	key := fmt.Sprintf("active_%v_%v", username, deviceId)
	v, err := vars.RedisInstance.Exists(key).Result()
	if v == 1 {
		result = true
	}
	// logger.Log.Infof("key:%v, v: %v, err: %v, result: %v", key, v, err, result)
	return err, result
}

// 设置设备激活标识
func SetDeviceActiveFlag(username, deviceId string) {
	models.InitRedis()
	key := fmt.Sprintf("active_%v_%v", username, deviceId)
	_, _ = vars.RedisInstance.Set(key, deviceId, 60*time.Second).Result()
}

// 激活码是否生成过，生成过的话返回老激活码
func GetActiveCodeFlag(username, deviceId string) (has bool, code string) {
	key := fmt.Sprintf("icodeF_%v_%v", username, deviceId)
	models.InitRedis()
	ret, err := vars.RedisInstance.Get(key).Result()

	if err == nil && ret != "" {
		has = true
		code = ret
	}
	return has, code
}

// 设置验证码的标识，判断是否生成过验证码，12小时后失效
func SetActiveCodeFlag(username, deviceId, code string) {
	key := fmt.Sprintf("icodeF_%v_%v", username, deviceId)
	models.InitRedis()
	vars.RedisInstance.Set(key, code, 60*60*12*time.Second)
}

// 删除生成激活码的标识
func DelActiveCodeFlag(username, deviceId string) (err error) {
	key := fmt.Sprintf("icodeF_%v_%v", username, deviceId)
	models.InitRedis()
	_, err = vars.RedisInstance.Del(key).Result()
	return err
}

// 设置激活码的值及超时时间，该函数不对外公开
func setActiveCode(device models.Device, code string) (err error) {
	key := fmt.Sprintf("icode_%v", code)
	deviceStr, err := json.Marshal(device)
	v := fmt.Sprintf("%v-_-%v", device.User, string(deviceStr))
	models.InitRedis()
	_, err = vars.RedisInstance.Set(key, v, 60*60*12*time.Second).Result()
	return err
}

// 激活码保存到redis中，有效期为8小时，不会多次生成
func SetActiveCode(username, deviceId string, device models.Device) (err error, code string) {
	has, oldCode := GetActiveCodeFlag(username, deviceId)
	if has {
		code = oldCode
		err = setActiveCode(device, oldCode)
	} else {
		code = util.GenerateAtiveCode(username, deviceId)
		SetActiveCodeFlag(username, deviceId, code)
		err = setActiveCode(device, code)
	}
	return err, code
}

// 删除激活码
func RemoveActiveCode(code string) (err error) {
	key := fmt.Sprintf("icode_%v", code)
	models.InitRedis()
	_, err = vars.RedisInstance.Del(key).Result()
	return err
}

// 获取激活码的值
func getActiveCodeValue(code string) (err error, has bool, user string, deviceInfo models.Device) {
	key := fmt.Sprintf("icode_%v", code)
	models.InitRedis()
	c, err := vars.RedisInstance.Get(key).Result()
	if err == nil {
		t := strings.Split(c, "-_-")
		if len(t) == 2 {
			user = t[0]
			deviceStr := t[1]
			err = json.Unmarshal([]byte(deviceStr), &deviceInfo)
			if err == nil {
				has = true
			}
		}
	}
	return err, has, user, deviceInfo
}

// 检测激活码
func CheckActiveCode(code string) (err error, has bool, user string, deviceInfo models.Device) {
	err, has, user, deviceInfo = getActiveCodeValue(code)
	return err, has, user, deviceInfo
}

// 激活时，验证激活码
func VerifyActiveCode(code string) (result bool, user string, device models.Device) {
	err, has, user, deviceInfo := CheckActiveCode(code)
	if err == nil && has {
		status := deviceInfo.State
		if status == 1 {
			result = true
		}
	}
	return result, user, deviceInfo
}

// 在激活流程中使用过激活码后重置激活码的状态，设为已激活并在2小时后自动删除
func ResetActiveCodeStatus(code, username string, device models.Device, state int) (result bool, err error) {
	models.InitRedis()
	key := fmt.Sprintf("icode_%v", code)
	device.State = state
	err = setActiveCode(device, code)
	if err == nil {
		result, err = vars.RedisInstance.Expire(key, 60*60*2*time.Second).Result()
		err = ResetSmsStatus(username, device.DeviceId)
	}
	return result, err
}
