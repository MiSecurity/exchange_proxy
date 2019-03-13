package models

import (
	"exchange_proxy/util/wbxml"
	"exchange_proxy/vars"

	"encoding/json"
	"fmt"
	"strconv"
)

type (
	// 移动设备信息
	Device struct {
		DeviceType string `json:"device_type"`
		DeviceId   string `json:"device_id"`
		User       string `json:"user"`
		State      int    `json:"state"`
		Time       int64  `json:"time"`
	}
	UserPhone struct {
		Code int    `json:"code"`
		Data string `json:"data"`
	}
)

// 注册新设备，默认为未激活状态
func NewDevice(user string, device Device) (err error) {
	InitRedis()
	v := make(map[string]interface{})
	v["user_state"] = 0
	info, _ := json.Marshal(device)
	v[device.DeviceId] = string(info)
	key := fmt.Sprintf("iuser_%v", user)
	_, err = vars.RedisInstance.HMSet(key, v).Result()
	return err
}

// 新建用户，默认为激活状态
func NewUser(user string) (err error) {
	InitRedis()
	key := fmt.Sprintf("iuser_%v", user)
	v := make(map[string]interface{})
	v["user_state"] = 0
	_, err = vars.RedisInstance.HMSet(key, v).Result()
	return err
}

// 获取用户状态：激活与锁定
func GetUserState(user string) (userState int, err error) {
	userState = -1
	InitRedis()
	key := fmt.Sprintf("iuser_%v", user)
	ret, err := vars.RedisInstance.HGet(key, "user_state").Result()
	userState, _ = strconv.Atoi(ret)
	return userState, err
}

// 获取设备状态：激活、未激活、锁定等
func GetDeviceState(user, deviceId string) (state int) {
	state = -1
	InitRedis()
	key := fmt.Sprintf("iuser_%v", user)
	ret, err := vars.RedisInstance.HGet(key, deviceId).Result()
	if err == nil {
		var device Device
		err := json.Unmarshal([]byte(ret), &device)
		if err == nil {
			state = device.State
		}
	}
	return state
}

// 获取设备信息
func GetDeviceInfo(user, deviceId string) (err error, deviceInfo Device) {
	InitRedis()
	key := fmt.Sprintf("iuser_%v", user)
	info, err := vars.RedisInstance.HGet(key, deviceId).Result()
	if err == nil {
		err = json.Unmarshal([]byte(info), &deviceInfo)
	}
	return err, deviceInfo
}

// 设置设备的状态：激活，锁定
// ALLOW = 0
// NEW = 1
// LOCKED = 2
// BLOCK = 3
func SetDeviceState(user, deviceId string, state int) (err error) {
	InitRedis()
	key := fmt.Sprintf("iuser_%v", user)
	err, deviceInfo := GetDeviceInfo(user, deviceId)
	if err == nil {
		deviceInfo.State = state
		info, err := json.Marshal(deviceInfo)
		v := make(map[string]interface{})
		v[deviceId] = string(info)
		if err == nil {
			_, err = vars.RedisInstance.HMSet(key, v).Result()
		}
	}
	return err
}

// 激活设备
func ActiveDevice(user, deviceId string) (err error) {
	state := 0
	err = SetDeviceState(user, deviceId, state)
	return err
}

// 恢复设备
func RestoreDevice(user, deviceId string) (err error) {
	state := 1
	err = SetDeviceState(user, deviceId, state)
	return err
}

// 锁定设备
func LockDevice(user, deviceId string) (err error) {
	state := 2
	err = SetDeviceState(user, deviceId, state)
	return err
}

// 忽略设备
func IgnoreDevice(user, deviceId string) (err error) {
	state := 3
	err = SetDeviceState(user, deviceId, state)
	return err
}

// 指令检测
func CheckCmd(cmd string, cmdType map[string]bool) (result bool) {
	if cmdType[cmd] {
		result = true
	}
	return result
}

// 存储手机设备的信息
func SetDeviceInfo(deviceId string, deviceInfo wbxml.DeviceInfo) (err error) {
	m := make(map[string]interface{})
	deviceInfoStr, _ := json.Marshal(deviceInfo)
	m[deviceId] = string(deviceInfoStr)
	InitRedis()
	key := "DEVICE_INFO_XSEC_MAIL"
	_, err = vars.RedisInstance.HMSet(key, m).Result()
	return err
}

// 通过设备ID查询手机设备的信息
func GetDeviceInfoByDeviceId(deviceId string) (deviceInfo wbxml.DeviceInfo, err error) {
	InitRedis()
	key := "DEVICE_INFO_XSEC_MAIL"
	deviceInfoStr, err := vars.RedisInstance.HGet(key, deviceId).Result()
	err = json.Unmarshal([]byte(deviceInfoStr), &deviceInfo)
	return deviceInfo, err
}

// 获取设备列表
func GetDeviceList(username string) (devices []string, err error) {
	key := fmt.Sprintf("iuser_%v", username)
	InitRedis()
	devices, err = vars.RedisInstance.HVals(key).Result()
	return devices, err
}

// 获取设备数
func GetDeviceNum(username string) (n int) {
	devices, err := GetDeviceList(username)
	if err == nil {
		n = len(devices)
	}
	return n
}
