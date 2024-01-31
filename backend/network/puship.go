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
	dbObj := db.DBOperObj().GetDB()

	go func() {
		if second < 60 {
			second = 60
		}

		for {
			time.Sleep(time.Duration(second) * time.Second)

			info := db.DBOperObj().GetConfig()
			p.ip = comm.GetGlobalIP(info.CheckIPAddr)

			fmt.Printf("当前地址：%s，老地址：%s", p.ip, info.IP)

			if len(p.ip) != 0 {
				if !strings.EqualFold(p.ip, info.IP) {
					msg := fmt.Sprintf("当前地址：%s", p.ip)

					if len(info.AYFFToken) == 0 && len(info.WXPusherToken) == 0 {
						continue
					}

					err := errors.New("no match")

					if len(info.AYFFToken) != 0 {
						err = comm.AYFFPushMsg(msg, info.AYFFToken)
					}

					if len(info.WXPusherToken) != 0 && info.WXPusherTopicId != 0 {
						err = comm.WXPusherMsg(msg, info.WXPusherToken, info.WXPusherTopicId)
					}

					if err == nil {
						info.IP = p.ip
						dbObj.Select("ip").Save(info)
					}
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
