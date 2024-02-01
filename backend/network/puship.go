package network

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
	"wakelan/backend/comm"
	"wakelan/backend/db"
)

type PushIP struct {
	ip string
}

func (p *PushIP) Start(second int) error {
	isPrintLog := false

	if second < 60 {
		second = 60 //最低60秒
	}

	dbObj := db.DBOperObj().GetDB()

	go func() {
		waitTime := 20 //初次等待时间为20秒，不能太短，崩溃拉起后过于频繁

		for {
			time.Sleep(time.Duration(waitTime) * time.Second)
			waitTime = second //调整之后的等待时间

			info := db.DBOperObj().GetConfig()
			p.ip = comm.GetGlobalIP(info.CheckIPAddr)

			if len(p.ip) != 0 {
				if !strings.EqualFold(p.ip, info.IP) {
					msg := fmt.Sprintf("当前地址：%s", p.ip)

					if len(info.AYFFToken) == 0 && len(info.WXPusherToken) == 0 {
						continue
					}

					if !isPrintLog {
						db.DBLog("消息推送", "公网IP %s", p.ip)
					}

					err := errors.New("no match")

					if len(info.AYFFToken) != 0 {
						err = comm.AYFFPushMsg(msg, info.AYFFToken)
					}

					if len(info.WXPusherToken) != 0 && info.WXPusherTopicId != 0 {
						err = comm.WXPusherMsg(msg, info.WXPusherToken, info.WXPusherTopicId)
					}

					if err != nil {
						isPrintLog = true
						db.DBLog("消息推送", "推送失败 %s", err.Error())
					}

					info.IP = p.ip
					dbObj.Select("ip").Save(info)
					isPrintLog = false
				}
			} else {
				if !isPrintLog {
					isPrintLog = true
					db.DBLog("消息推送", "公网IP获取失败")
				}
			}
		}
	}()

	return nil
}

func (p *PushIP) GetIP() string {
	return p.ip
}

var pushipOnce sync.Once
var pushipObj *PushIP

func PushipOBJ() *PushIP {
	pushipOnce.Do(func() {
		pushipObj = &PushIP{}
	})

	return pushipObj
}
