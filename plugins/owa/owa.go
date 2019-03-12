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

package owa

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/toolkits/slice"
	"github.com/vulcand/oxy/forward"
	"github.com/vulcand/oxy/testutils"

	"exchange_proxy/logger"
	"exchange_proxy/util"
	"exchange_proxy/vars"
)

func addOtp(resp *http.Response) error {
	b, err := ioutil.ReadAll(resp.Body)
	if err == nil {
		err = resp.Body.Close()

		oldHtml := "<div class=\"showPasswordCheck signInCheckBoxText\">"
		newHtml := "<div class=\"signInInputLabel\" id=\"passwordLabel\" aria-hidden=\"true\">动态口令:</div><div><input id=\"customToken\" onfocus=\"g_fFcs=0\" name=\"customToken\" value=\"\" type=\"password\" class=\"signInInputText\" aria-labelledby=\"passwordLabel\"></div>"
		b = bytes.Replace(b, []byte(oldHtml), []byte(newHtml), -1) // replace html

		body := ioutil.NopCloser(bytes.NewReader(b))
		resp.Body = body
		resp.ContentLength = int64(len(b))
		resp.Header.Set("Content-Length", strconv.Itoa(len(b)))
	}
	return err
}

func OwaRedirect(w http.ResponseWriter, req *http.Request) {

	if slice.ContainsString(vars.MailConfig.Host, req.Host) {
		req.URL = testutils.ParseURI(vars.MailConfig.Backend)
		vars.FwdOWA.ServeHTTP(w, req)
	} else {
		u := fmt.Sprintf("https://%v//owa/auth/logon.aspx", vars.MailConfig.Host[0])
		w.Header().Set("Location", u)
		w.WriteHeader(302)
	}
}

func OwaHandler(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {

		// 内网用户不用验证OTP口令
		if strings.HasPrefix(req.RemoteAddr, "10.") || strings.HasPrefix(req.RemoteAddr, "192.") {
			r := forward.RoundTripper(
				&http.Transport{
					TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
				},
			)
			vars.FwdOWA, _ = forward.New(forward.PassHostHeader(true), forward.Logger(logger.Log.Logger), r)
			h(w, req)
		} else {
			if vars.MailConfig.TLS {
				r := forward.RoundTripper(
					&http.Transport{
						TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
					},
				)

				if req.Method == "GET" && strings.HasPrefix(req.RequestURI, "/owa/auth/logon.aspx") {
					vars.FwdOWA, _ = forward.New(forward.PassHostHeader(true), forward.Logger(logger.Log.Logger),
						r, forward.ResponseModifier(addOtp))
					h(w, req)

				} else if req.Method == "POST" && strings.HasPrefix(req.RequestURI, "/owa/auth/logon.aspx") {
					var bodyBytes []byte
					if req.Body != nil {
						bodyBytes, _ = ioutil.ReadAll(req.Body)

						// 恢复Req.body的值给ParseForm函数使用
						req.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
						_ = req.ParseForm()
						username := req.FormValue("username")
						customToken := req.FormValue("customToken")

						// 恢复Req.body的值传到下一个处理器中
						req.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))

						result, err := util.CheckToken(vars.OtpUrl, username, customToken)
						logger.Log.Printf("user: %v, token: %v, result: %v", username, customToken, result)
						if err == nil && result {
							vars.FwdOWA, _ = forward.New(forward.PassHostHeader(true), forward.Logger(logger.Log.Logger), r)
							h(w, req)
						} else {
							w.Header().Set("Location", "/owa/auth/logon.aspx")
							w.WriteHeader(302)
						}
					}
				} else {
					vars.FwdOWA, _ = forward.New(forward.PassHostHeader(true), forward.Logger(logger.Log.Logger), r)
					h(w, req)
				}
			}
		}
	}
}
