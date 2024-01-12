package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"wakelan/backend/db"
	"wakelan/backend/guacd"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type Remote struct {
	t2s map[int]string
}

func (r *Remote) Init() {
	r.t2s = make(map[int]string)

	//0:rdp 1:vnc 2:ssh 3:telnel
	r.t2s[0] = "RDP"
	r.t2s[1] = "VNC"
	r.t2s[2] = "SSH"
	r.t2s[3] = "TELNEL"
}

func (r *Remote) setting(c *gin.Context) {
	data, _ := c.GetRawData()

	info := &db.RemoteInfo{}
	info.IP = c.Query("ip")
	info.Remote = string(data)
	dbObj := db.DBOperObj().GetDB()
	result := dbObj.Save(info)

	if result.Error != nil {
		c.JSON(200, gin.H{
			"err": result.Error.Error(),
		})

		return
	}

	c.JSON(200, gin.H{
		"err": "",
	})
}

func (r *Remote) remote(c *gin.Context) {
	connInfo, err := url.QueryUnescape(c.Query("info"))
	if err != nil {
		fmt.Println(err.Error())
		c.JSON(200, gin.H{
			"err": err.Error(),
		})

		return
	}

	if len(connInfo) == 0 {
		fmt.Println(err.Error())
		c.JSON(200, gin.H{
			"err": "参数错误",
		})

		return
	}

	info := guacd.DstInfo{}
	err = json.Unmarshal([]byte(connInfo), &info)
	if err != nil {
		fmt.Println(err.Error())
		c.JSON(200, gin.H{
			"err": err.Error(),
		})

		return
	}

	cfg := db.DBOperObj().GetConfig()

	if len(cfg.GuacdHost) == 0 {
		cfg.GuacdHost = "127.0.0.1"
	}

	if cfg.GuacdPort == 0 {
		cfg.GuacdPort = 4822
	}

	info.SetGuacdServer(info.Remote.Type, cfg.GuacdHost, int16(cfg.GuacdPort))

	db.DBLog("远程连接", "主机：%s，类型：%v，Guacd：%s:%d",
		info.Remote.Host,
		r.t2s[info.Remote.Type],
		cfg.GuacdHost, cfg.GuacdPort)

	wbsocket := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	protocol := c.Request.Header.Get("Sec-Websocket-Protocol")
	conn, err := wbsocket.Upgrade(c.Writer, c.Request, http.Header{
		"Sec-Websocket-Protocol": {protocol},
	})
	if err != nil {
		c.JSON(200, gin.H{
			"err": err.Error(),
		})

		return
	}

	defer conn.Close()

	guacd := guacd.GuacdCtrl{}
	err = guacd.Start(conn, info)
	if err != nil {
		c.JSON(200, gin.H{
			"err": err.Error(),
		})
	}

	db.DBLog("远程断开", "主机：%s，类型：%v，Guacd：%s:%d",
		info.Remote.Host, r.t2s[info.Remote.Type],
		cfg.GuacdHost, cfg.GuacdPort)
}
