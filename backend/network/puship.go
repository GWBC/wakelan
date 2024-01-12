package network

import (
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
	dbObj := db.DBOperObj().GetDB()

	go func() {
		for {
			info := db.DBOperObj().GetConfig()
			p.ip = comm.GetGlobalIP()

			if len(p.ip) != 0 {
				if !strings.EqualFold(p.ip, info.IP) {
					info.IP = p.ip
					dbObj.Select("ip").Save(info)

					msg := fmt.Sprintf("当前地址：%s", p.ip)

					if len(info.AYFFToken) != 0 {
						go comm.AYFFPushMsg(msg, info.AYFFToken)
					}

					if len(info.WXPusherToken) != 0 && info.WXPusherTopicId != 0 {
						go comm.WXPusherMsg(msg, info.WXPusherToken, info.WXPusherTopicId)
					}
				}
			}

			time.Sleep(time.Duration(second) * time.Second)
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
