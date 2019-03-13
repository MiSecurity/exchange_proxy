package util

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
)

type TokenResult struct {
	Code int    `json:"code"`
	Data string `json:"data"`
}

func CheckToken(urlOtp, username, token string) (result bool, err error) {
	users := make([]string, 0)
	users = append(users, username)
	tokens := make([]string, 0)
	tokens = append(tokens, token)
	resp, err := http.PostForm(urlOtp, url.Values{"username": users, "verificationCode": tokens})
	if err == nil {
		defer resp.Body.Close()
		ret, err := ioutil.ReadAll(resp.Body)

		if err == nil {
			tokenResult := TokenResult{}
			err = json.Unmarshal(ret, &tokenResult)
			if err == nil {
				if tokenResult.Code == 200 && tokenResult.Data == "success" {
					result = true
				}
			}
		}
	}
	return result, err
}
