## Exchange_proxy

Exchange_proxy是由go语言开发的Exchange安全代理，可以将内网的Exchange服务器的https服务安全地发布出去，
支持的功能如下：

- WEB端增加OTP二次认证
- 手机端增加设备激活绑定的功能 
- 屏蔽了PC端的EWS协议

在使用该系统前，需要确保有以下基础设施的接口对完成对接：

- OTP动态口令系统
- 短信发送接口
- 通过员工姓名查询员工手机号的接口

以上接口准备并对接完成后，正确配置conf/app.ini即可启动代理服务器了。

```ini
[mail]
hosts = mail.xiaomi.com,mail.sec.lu
backend = https://10.10.10.10
port = 443
ssl = true
cert = certs/ca.crt
key = certs/ca.key
; debug level: Fatal, Error, Warn, Info, Debug
debug_level = info

[redis]
host = 10.10.10.20
port = 6379
db = 0
password = redis_passw0rd

[otp]
url = https://otp_api_url/chk_otp

[sms]
url = http://sms_api_url/api/send_sms
header = X-SMS-Token
key = token

[user_info]
user_phone = http://hr_api_url/findMobile
active_url = https://mail.sec.lu/a/

```

配置文件说明：

- mail节下的配置项是配置邮箱服务器本身的
    - hosts表示邮箱域名，支持配置多个用英文逗号分割的域名
    - backend表示邮箱服务器地址
    - ssl表示是否启用https，必须设为true
    - cert和key分别表示证书的公、私钥，与nginx的证书完全兼容
    - debug_level表示日志级别，默认为info级别

- redis节表示redis服务器的配置
- otp节为动态口令检测API的URL
- sms节表示短信接口的API
- user_info节下的user_phone表示查找手机的接口，active_url表示手机中激活连接的URL

设置好配置文件后，可通过`./main`直接启动代理服务器，如下图所示：
![](http://docs.xsec.io/images/mail_proxy/mail_proxy041.png)

WEB通过外网访问WEB端时，要求必须输入正确的OTP口令才可以登录，如下图所示：
![](http://docs.xsec.io/images/mail_proxy/mail_proxy03.png)

通过手机端访问时，只有通过短信中的提示激活后，方可收发邮件，如下图所示：

- 收到激活短信 
![](http://docs.xsec.io/images/mail_proxy/mail_proxy04.png)

- 激活确认页面

![](http://docs.xsec.io/images/mail_proxy/mail_proxy05.png)

- 激活成功页面
![](http://docs.xsec.io/images/mail_proxy/mail_proxy06.png)

正式上线之前，最好提供相应的管理后台并与内网的管理系统对接，邮件代理管理后台提供以下功能：

- 管理员可查看、修改每个用户的账户与设备状态
- 管理员可查看每个设备的激活进程，方便故障排查
- 用户也可自行管理自己的设备

设备数据保存在redis中，用go/python/php等语言都可以实现，我就不单独提供了。

代理系统的进程可以托管在supervisor或god中，部署了该系统后，可以解决邮件服务器手机端与WEB端的安全，目前的开源版本没有电脑端的安全代理功能，建议在PC端收发邮件时拨入VPN，或者在电脑中用BlueMail客户端收发邮件。