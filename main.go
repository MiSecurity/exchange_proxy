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

package main

import (
	"exchange_proxy/logger"
	"exchange_proxy/plugins/active_sync"
	"exchange_proxy/plugins/owa"
	"exchange_proxy/plugins/web"
	"exchange_proxy/vars"

	"fmt"
	"net/http"

	"github.com/toolkits/slice"
)

func main() {
	mux := http.NewServeMux()
	staticFiles := http.FileServer(http.Dir("plugins/web/public"))
	mux.Handle("/static/", http.StripPrefix("/static/", staticFiles))
	mux.HandleFunc("/", owa.OwaHandler(owa.OwaRedirect))
	mux.HandleFunc("/owa/", owa.OwaHandler(owa.OwaRedirect))
	mux.HandleFunc("/Microsoft-Server-ActiveSync", active_sync.ActiveSyncHandler(active_sync.SyncRedirect))
	mux.HandleFunc("/a/", web.Activation)
	mux.HandleFunc("/a/activedevice", web.ActiveDevice)
	mux.HandleFunc("/a/ignoredevice", web.IgnoreDevice)

	mux.HandleFunc("/EWS/", web.NotFound)
	mux.HandleFunc("/ecp/", web.NotFound)
	mux.HandleFunc("/rpc/", web.NotFound)
	mux.HandleFunc("/Autodiscover/", web.NotFound)

	s := &http.Server{
		Addr:    fmt.Sprintf(":%v", vars.MailConfig.Port),
		Handler: mux,
	}

	go func() {
		listen80Port()
	}()

	var err error
	if vars.MailConfig.TLS {
		err = s.ListenAndServeTLS(vars.MailConfig.Cert, vars.MailConfig.Key)
	} else {
		err = s.ListenAndServe()
	}
	logger.Log.Println(err)
}

func listen80Port() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if slice.ContainsString(vars.MailConfig.Host, r.Host) {
			logger.Log.Debugf("redirect http://%v -> https://%v", r.Host, r.Host)
			w.Header().Set("Location", fmt.Sprintf("https://%v", r.Host))
			w.WriteHeader(302)
		}
	})

	s := &http.Server{
		Addr:    ":80",
		Handler: mux,
	}

	_ = s.ListenAndServe()
}
