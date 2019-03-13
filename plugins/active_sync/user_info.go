package active_sync

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"exchange_proxy/models"
)

// 获取手机号
func GetUserPhone(url, user string) (phone string, err error) {
	var userPhone models.UserPhone
	resp, err := http.Get(fmt.Sprintf("%v?username=%v", url, user))
	if err == nil {
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err == nil {
			err = json.Unmarshal(body, &userPhone)
			if err == nil && userPhone.Code == 200 {
				phone = userPhone.Data
			}
		}
	}
	return phone, err
}
