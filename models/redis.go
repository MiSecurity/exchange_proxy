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

package models

import (
	"fmt"
	"time"

	"exchange_proxy/logger"
	"exchange_proxy/vars"

	"github.com/go-redis/redis"
)

func init() {
	InitRedis()
}

func InitRedis() () {
	var err error
	vars.RedisInstance, err = NewRedisClient(vars.RedisConf.Host, vars.RedisConf.Port, vars.RedisConf.Db, vars.RedisConf.Password)
	if err != nil {
		logger.Log.Errorf("connect redis failed, err: %v", err)
	}
}

func NewRedisClient(host string, port int, db int, password string) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:        fmt.Sprintf("%v:%v", host, port),
		Password:    password,    // no password set
		DB:          db,          // use default DB
		ReadTimeout: time.Minute, // set timeout value = 60
	})

	_, err := client.Ping().Result()
	return client, err
}
