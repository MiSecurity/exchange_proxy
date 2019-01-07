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

package wbxml

import (
	"bytes"
	"encoding/xml"
	"regexp"

	"exchange_proxy/logger"

	. "github.com/magicmonty/activesync-go/activesync"
	. "github.com/magicmonty/activesync-go/activesync/base"
	. "github.com/magicmonty/wbxml-go/wbxml"
)

/*
```
<?xml version="1.0" encoding="utf-8"?>
<O:Provision xmlns:O="Provision" xmlns:S="Settings">
    <S:DeviceInformation>
        <S:Set>
            <S:Model>MIX 2</S:Model>
            <S:IMEI>888833336669999</S:IMEI>
            <S:FriendlyName>MIX 2</S:FriendlyName>
            <S:OS>Android 8.0.0</S:OS>
            <S:PhoneNumber>+8618599999999</S:PhoneNumber>
            <S:UserAgent>Android/8.0.0-EAS-1.3</S:UserAgent>
            <S:MobileOperator>中国联通 (46001)</S:MobileOperator>
        </S:Set>
    </S:DeviceInformation>
    <O:Policies>
        <O:Policy>
            <O:PolicyType>MS-EAS-Provisioning-WBXML</O:PolicyType>
        </O:Policy>
    </O:Policies>
</O:Provision>
```
*/

type (
	Provision struct {
		XMLName           xml.Name `xml:"Provision"`
		Text              string   `xml:",chardata"`
		O                 string   `xml:"O,attr"`
		S                 string   `xml:"S,attr"`
		DeviceInformation struct {
			Text string `xml:",chardata"`
			Set  struct {
				Text           string `xml:",chardata"`
				Model          string `xml:"Model"`
				IMEI           string `xml:"IMEI"`
				FriendlyName   string `xml:"FriendlyName"`
				OS             string `xml:"OS"`
				PhoneNumber    string `xml:"PhoneNumber"`
				UserAgent      string `xml:"UserAgent"`
				MobileOperator string `xml:"MobileOperator"`
			} `xml:"Set"`
		} `xml:"DeviceInformation"`
		Policies struct {
			Text   string `xml:",chardata"`
			Policy struct {
				Text       string `xml:",chardata"`
				PolicyType string `xml:"PolicyType"`
			} `xml:"Policy"`
		} `xml:"Policies"`
	}

	DeviceInfo struct {
		Model          string `json:"model"`
		IMEI           string `json:"imei"`
		FriendlyName   string `json:"friendly_name"`
		PhoneNumber    string `json:"phone_number"`
		MobileOperator string `json:"mobile_operator"`
	}
)

func removeInvalidChars(b []byte) []byte {
	re := regexp.MustCompile("[^\x09\x0A\x0D\x20-\uD7FF\uE000-\uFFFD\u10000-\u10FFFF]")
	return re.ReplaceAll(b, []byte{})
}

func EncodeXML(xmlString []byte) {
	xmlString = removeInvalidChars([]byte(xmlString))
	w := bytes.NewBuffer(make([]byte, 0))
	e := NewEncoder(
		MakeCodeBook(PROTOCOL_VERSION_14_1),
		string(xmlString),
		w)
	err := e.Encode()
	if err != nil {
		logger.Log.Println(err.Error())
	} else {
		logger.Log.Println(w)
	}
}

func getDecodeResult(data ...byte) string {
	var result string
	result, _ = Decode(bytes.NewBuffer(data), MakeCodeBook(PROTOCOL_VERSION_14_1))
	return result
}

func Parse(data []byte) (deviceInfo DeviceInfo, err error) {
	xmldata := getDecodeResult(data...)

	out := Provision{}
	err = xml.Unmarshal([]byte(xmldata), &out)
	if err == nil {
		logger.Log.Debugf("Model: %v", out.DeviceInformation.Set.Model)
		logger.Log.Debugf("Imie: %v", out.DeviceInformation.Set.IMEI)
		logger.Log.Debugf("FriendlyName: %v", out.DeviceInformation.Set.FriendlyName)
		logger.Log.Debugf("PhoneNumber: %v", out.DeviceInformation.Set.PhoneNumber)
		logger.Log.Debugf("MobileOperator: %v", out.DeviceInformation.Set.MobileOperator)

		deviceInfo.Model = out.DeviceInformation.Set.Model
		deviceInfo.IMEI = out.DeviceInformation.Set.IMEI
		deviceInfo.FriendlyName = out.DeviceInformation.Set.FriendlyName
		deviceInfo.PhoneNumber = out.DeviceInformation.Set.PhoneNumber
		deviceInfo.MobileOperator = out.DeviceInformation.Set.MobileOperator
	}

	return deviceInfo, err
}
