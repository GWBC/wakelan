package api

import (
	"encoding/json"
	"net"
	"net/http"
	"sort"
	"sync"
	"wakelan/backend/comm"
	"wakelan/backend/db"
	"wakelan/backend/network"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"gorm.io/gorm"
)

// 获取网卡信息
type InterfaceInfo struct {
	Name string   `json:"name"`
	Desc string   `json:"desc"`
	IPS  []string `json:"ips"`
}

type WakeApi struct {
}

func (w *WakeApi) Init() {

}

// 获取外网IP
func (w *WakeApi) getGlobalIP(c *gin.Context) {
	c.JSON(200, gin.H{
		"err": "",
		"ip":  network.PushipOBJ().GetIP(),
	})
}

func (w *WakeApi) getInterfaces(c *gin.Context) {
	ifaces, err := network.NetProtoObj().GetInterfaces()
	if err != nil {
		c.JSON(200, gin.H{
			"err": err.Error(),
		})
		return
	}

	ifs := []InterfaceInfo{}
	for _, v := range ifaces {
		i := InterfaceInfo{}
		i.Name = v.Name
		i.Desc = v.Description

		for _, addr := range v.Addresses {
			i.IPS = append(i.IPS, addr.IP.String())
		}

		ifs = append(ifs, i)
	}

	c.JSON(200, gin.H{
		"err":   "",
		"infos": ifs,
	})
}

// 打开网络
func (w *WakeApi) openCard(c *gin.Context) {
	iface, err := network.NetProtoObj().GetInterfaceByName(c.Query("name"))
	if err != nil {
		c.JSON(200, gin.H{
			"err": err.Error(),
		})
		return
	}

	network.NetProtoObj().Close()
	err = network.NetProtoObj().Open(iface, true)
	if err != nil {
		c.JSON(200, gin.H{
			"err": err.Error(),
		})
		return
	}

	info := &db.GlobalInfo{}
	dbObj := db.DBOperObj().GetDB()
	dbObj.Find(info)

	i := InterfaceInfo{}
	i.Name = iface.Name
	i.Desc = iface.Description

	for _, addr := range iface.Addresses {
		i.IPS = append(i.IPS, addr.IP.String())
	}

	data, _ := json.Marshal(i)
	info.NetCard = string(data)

	dbObj.Select("netcard").Save(info)

	c.JSON(200, gin.H{
		"err": "",
	})
}

// 探测网络
func (w *WakeApi) probeNetwork(c *gin.Context) {
	obj := network.NetProtoObj()
	if !obj.IsOpen() {
		c.JSON(200, gin.H{
			"err": "network not open",
		})
		return
	}

	db.DBLog("探测网络", "网卡：%s", obj.GetLocalInfo().Name)

	obj.QueryNet(6)

	c.JSON(200, gin.H{
		"err": "",
	})
}

// 删除网络列表
func (w *WakeApi) delNetworklist(c *gin.Context) {
	ip := c.Query("ip")

	dbObj := db.DBOperObj().GetDB()

	var result *gorm.DB = nil
	if len(ip) == 0 {
		result = dbObj.Delete(&db.MacInfo{}, "1=1")
	} else {
		result = dbObj.Where("ip=?", ip).Delete(&db.MacInfo{})
	}

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

// 添加网络列表
func (w *WakeApi) addNetworklist(c *gin.Context) {
	info := &db.MacInfo{}
	info.IP = c.Query("ip")
	info.Mac = c.Query("mac")
	info.AttachInfo.Mac = info.Mac
	info.AttachInfo.Describe = c.Query("describe")

	if len(info.IP) == 0 || len(info.Mac) == 0 {
		c.JSON(200, gin.H{
			"err": "IP和Mac不能为空",
		})
		return
	}

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

// 获取网络列表
func (w *WakeApi) getNetworklist(c *gin.Context) {
	isAes := c.DefaultQuery("aes", "1")
	infos := []db.MacInfo{}
	datas := network.NetProtoObj().GetResult()

	for _, info := range datas {
		macInfo := db.MacInfo{}
		macInfo.IP = info.IP.String()
		macInfo.Mac = info.Mac.String()
		macInfo.MANUF = info.MANUF

		infos = append(infos, macInfo)
	}

	dbObj := db.DBOperObj().GetDB()

	if len(infos) != 0 {
		dbObj.Save(infos)
	}

	//获取加星
	infos = []db.MacInfo{}

	//注意：AttachInfo是结构字段的名称，不是表名
	result := dbObj.Joins("AttachInfo").Find(&infos, "AttachInfo.star==1")

	sort.Slice(infos, func(i, j int) bool {
		ip1 := net.ParseIP(infos[i].IP)
		ip2 := net.ParseIP(infos[j].IP)
		if isAes == "1" {
			return comm.IpLess(ip1, ip2)
		}

		return !comm.IpLess(ip1, ip2)
	})

	if result.Error != nil {
		c.JSON(200, gin.H{
			"err": result.Error.Error(),
		})
		return
	}

	var tmpInfo = []db.MacInfo{}

	result = dbObj.Joins("AttachInfo").
		Where("AttachInfo.star is null or AttachInfo.star=0").Find(&tmpInfo)

	sort.Slice(tmpInfo, func(i, j int) bool {
		ip1 := net.ParseIP(tmpInfo[i].IP)
		ip2 := net.ParseIP(tmpInfo[j].IP)
		if isAes == "1" {
			return comm.IpLess(ip1, ip2)
		}

		return !comm.IpLess(ip1, ip2)
	})

	if result.Error != nil {
		c.JSON(200, gin.H{
			"err": result.Error.Error(),
		})
		return
	}

	infos = append(infos, tmpInfo...)

	c.JSON(200, gin.H{
		"err":   "",
		"infos": infos,
	})
}

// 唤醒
func (w *WakeApi) wakeLan(c *gin.Context) {
	err := comm.WakeLan(c.Query("mac"))
	if err != nil {
		c.JSON(200, gin.H{
			"err": err.Error(),
		})

		return
	}

	db.DBLog("唤醒", "Mac：%s", c.Query("mac"))

	c.JSON(200, gin.H{
		"err": "",
	})
}

// 操作星
func (w *WakeApi) operStar(c *gin.Context) {
	info := &db.AttachInfo{}
	info.Mac = c.Query("mac")
	info.Star = c.Query("star") == "1"
	dbObj := db.DBOperObj().GetDB()

	if len(info.Mac) == 0 {
		c.JSON(200, gin.H{
			"err": "MAC不能为空",
		})
		return
	}

	result := dbObj.Model(info).Update("star", info.Star)
	if result.Error != nil {
		c.JSON(200, gin.H{
			"err": result.Error.Error(),
		})

		return
	}

	if result.RowsAffected == 0 {
		result = dbObj.Save(info)
	}

	if result.Error != nil {
		c.JSON(200, gin.H{
			"err": result.Error.Error(),
		})

		return
	}

	if info.Star {
		db.DBLog("收藏", "Mac：%s", info.Mac)
	} else {
		db.DBLog("取消收藏", "Mac：%s", info.Mac)
	}

	c.JSON(200, gin.H{
		"err": "",
	})
}

// 编辑机器信息
func (w *WakeApi) editPCInfo(c *gin.Context) {
	info := &db.AttachInfo{}
	info.Mac = c.Query("mac")
	info.Describe = c.Query("describe")
	dbObj := db.DBOperObj().GetDB()

	if len(info.Mac) == 0 {
		c.JSON(200, gin.H{
			"err": "MAC不能为空",
		})
		return
	}

	result := dbObj.Model(info).Update("describe", info.Describe)
	if result.Error != nil {
		c.JSON(200, gin.H{
			"err": result.Error.Error(),
		})

		return
	}

	if result.RowsAffected == 0 {
		result = dbObj.Save(info)
	}

	if result.Error != nil {
		c.JSON(200, gin.H{
			"err": result.Error.Error(),
		})

		return
	}

	db.DBLog("编辑机器信息", "Mac：%s，描述：%s", info.Mac, info.Describe)

	c.JSON(200, gin.H{
		"err": "",
	})
}

// 查询当前选择的网卡
func (w *WakeApi) getSelectNetCard(c *gin.Context) {
	dbObj := db.DBOperObj().GetDB()

	info := &db.GlobalInfo{}
	result := dbObj.Select("netcard").Find(info)
	if result.Error != nil {
		c.JSON(200, gin.H{
			"err": result.Error.Error(),
		})

		return
	}

	if len(info.NetCard) == 0 {
		c.JSON(200, gin.H{
			"err":   "",
			"infos": "",
		})
		return
	}

	datas := map[string]interface{}{}
	err := json.Unmarshal([]byte(info.NetCard), &datas)
	if err != nil {
		c.JSON(200, gin.H{
			"err": result.Error.Error(),
		})

		return
	}

	c.JSON(200, gin.H{
		"err":   "",
		"infos": datas,
	})
}

// ping机器
func (w *WakeApi) pingPC(c *gin.Context) {
	wbsocket := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	lock := sync.Mutex{}
	conn, err := wbsocket.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(200, gin.H{
			"err": err.Error(),
		})

		return
	}

	defer func() {
		lock.Lock()
		conn.Close()
		lock.Unlock()
	}()

	network.NetProtoObj().AddPingRetFun(conn.RemoteAddr().String(), func(ip, mac string) {
		lock.Lock()
		conn.WriteMessage(websocket.TextMessage, []byte(ip+","+mac))
		lock.Unlock()
	})

	defer network.NetProtoObj().DelPingRetFun(conn.RemoteAddr().String())

	for {
		_, data, err := conn.ReadMessage()
		if err != nil {
			break
		}

		type cmd struct {
			Cmd  string `json:"cmd"`
			Data string `json:"data"`
		}

		cmdObj := cmd{}
		err = json.Unmarshal(data, &cmdObj)
		if err != nil {
			continue
		}

		if cmdObj.Cmd == "ping" {
			ips := []string{}

			if len(cmdObj.Data) == 0 {
				dbObj := db.DBOperObj().GetDB()
				macInfos := []db.MacInfo{}
				dbObj.Find(&macInfos)

				for _, v := range macInfos {
					ips = append(ips, v.IP)
				}
			} else {
				ips = append(ips, cmdObj.Data)
			}

			network.NetProtoObj().PingNet(ips)

			iface := network.NetProtoObj().GetLocalInfo()

			addrs, _ := iface.Addrs()
			for _, addr := range addrs {
				data := addr.(*net.IPNet).IP.String() + "," + iface.HardwareAddr.String()

				lock.Lock()
				conn.WriteMessage(websocket.TextMessage, []byte(data))
				lock.Unlock()
			}
		}
	}
}
