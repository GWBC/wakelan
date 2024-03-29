package api

import (
	"encoding/json"
	"fmt"
	"strconv"
	"wakelan/backend/db"

	"github.com/gin-gonic/gin"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

type DynamicPassword struct {
	AuthURL string `gorm:"column:auth_url"  json:"auth_url"`
	Secret  string `gorm:"column:secret"  json:"secret"`
}

type ConfigInfo struct {
	IP        string `gorm:"column:ip" json:"ip"`
	GuacdHost string `gorm:"column:guacd_host"  json:"guacd_host"`
	GuacdPort int    `gorm:"column:guacd_port"  json:"guacd_port"`
	AuthURL   string `gorm:"column:auth_url"  json:"auth_url"`
	Secret    string `gorm:"column:secret"  json:"secret"`

	AYFFToken       string `gorm:"column:ayff_token"  json:"ayff_token"`
	WXPusherToken   string `gorm:"column:wxpusher_token"  json:"wxpusher_token"`
	WXPusherTopicId int    `gorm:"column:wxpusher_topicid"  json:"wxpusher_topicid"`

	Debug       bool `gorm:"column:debug" json:"debug"`
	SharedLimit int  `gorm:"column:shared_limit" json:"shared_limit"`

	DockerEnableTCP   bool   `gorm:"docker_enable_tcp" json:"docker_enable_tcp"`
	DockerSvrIP       string `gorm:"docker_svr_ip" json:"docker_svr_ip"`
	DockerSvrPort     int    `gorm:"docker_svr_port" json:"docker_svr_port"`
	ContainerRootPath string `gorm:"container_root_path" json:"container_root_path"`
	DockerUser        string `gorm:"docker_user" json:"docker_user"`
	DockerPasswd      string `gorm:"docker_passwd" json:"docker_passwd"`

	CheckIPAddr string `gorm:"column:check_ip_addr;" json:"check_ip_addr"`
}

type System struct {
}

func (r *System) Init() {

}

// 获取日志总数
func (r *System) GetLogSize(c *gin.Context) {
	var totalRows int64
	dbObj := db.DBOperObj().GetDB()
	dbObj.Model(&db.Log{}).Count(&totalRows)

	infos := map[string]int64{}
	infos["total"] = totalRows

	c.JSON(200, gin.H{
		"err":   "",
		"infos": infos,
	})
}

// 获取日志
func (r *System) GetLog(c *gin.Context) {
	infos := []db.Log{}
	dbObj := db.DBOperObj().GetDB()

	strPageSize := c.Query("pageSize")
	pageSize, err := strconv.Atoi(strPageSize)
	if err != nil {
		c.JSON(200, gin.H{
			"err":   "参数错误",
			"infos": "",
		})
		return
	}

	strPage := c.Query("page")
	page, err := strconv.Atoi(strPage)
	if err != nil {
		c.JSON(200, gin.H{
			"err":   "参数错误",
			"infos": "",
		})
		return
	}

	dbObj.Order("updated_at desc").Limit(pageSize).Offset((page - 1) * pageSize).Find(&infos)

	c.JSON(200, gin.H{
		"err":   "",
		"infos": infos,
	})
}

// 获取配置信息
func (r *System) GetConfigInfo(c *gin.Context) {
	info := &db.GlobalInfo{}
	dbObj := db.DBOperObj().GetDB()
	dbObj.Find(info)

	cfg := ConfigInfo{}
	cfg.IP = info.IP
	cfg.GuacdHost = info.GuacdHost
	cfg.GuacdPort = info.GuacdPort
	cfg.AuthURL = info.AuthURL
	cfg.Secret = info.Secret
	cfg.AYFFToken = info.AYFFToken
	cfg.WXPusherToken = info.WXPusherToken
	cfg.WXPusherTopicId = info.WXPusherTopicId

	cfg.Debug = info.Debug
	cfg.SharedLimit = info.SharedLimit

	cfg.DockerEnableTCP = info.DockerEnableTCP
	cfg.DockerSvrIP = info.DockerSvrIP
	cfg.DockerSvrPort = info.DockerSvrPort
	cfg.ContainerRootPath = info.ContainerRootPath
	cfg.DockerUser = info.DockerUser
	cfg.DockerPasswd = info.DockerPasswd
	if len(info.DockerPasswd) != 0 {
		cfg.DockerPasswd = "******"
	}

	cfg.CheckIPAddr = info.CheckIPAddr

	c.JSON(200, gin.H{
		"err":   "",
		"infos": cfg,
	})
}

// 设置配置信息
func (r *System) SetConfig(c *gin.Context) {
	info := c.Query("info")
	cfgInfo := ConfigInfo{}

	err := json.Unmarshal([]byte(info), &cfgInfo)
	if err != nil {
		c.JSON(200, gin.H{
			"err":   "参数错误",
			"infos": "",
		})
		return
	}

	cfg := &db.GlobalInfo{}
	dbObj := db.DBOperObj().GetDB()
	dbObj.Find(cfg)
	cfg.Debug = cfgInfo.Debug
	cfg.SharedLimit = cfgInfo.SharedLimit
	cfg.GuacdHost = cfgInfo.GuacdHost
	cfg.GuacdPort = cfgInfo.GuacdPort
	cfg.AuthURL = cfgInfo.AuthURL
	cfg.Secret = cfgInfo.Secret
	cfg.AYFFToken = cfgInfo.AYFFToken
	cfg.WXPusherToken = cfgInfo.WXPusherToken
	cfg.WXPusherTopicId = cfgInfo.WXPusherTopicId
	cfg.DockerEnableTCP = cfgInfo.DockerEnableTCP
	cfg.DockerSvrIP = cfgInfo.DockerSvrIP
	cfg.DockerSvrPort = cfgInfo.DockerSvrPort
	cfg.DockerUser = cfgInfo.DockerUser
	cfg.DockerPasswd = cfgInfo.DockerPasswd
	cfg.CheckIPAddr = cfgInfo.CheckIPAddr

	dbObj.Select("guacd_host", "guacd_port", "auth_url", "secret",
		"ayff_token", "wxpusher_token", "wxpusher_topicid",
		"debug", "shared_limit", "check_ip_addr", "docker_enable_tcp",
		"docker_svr_ip", "docker_svr_port", "docker_user", "docker_passwd").Save(cfg)

	db.DBOperObj().SwitchLogger()

	c.JSON(200, gin.H{
		"err":   "",
		"infos": "",
	})
}

// 生成动态密码
func (r *System) GenDynamicPassword(c *gin.Context) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "网络唤醒",
		AccountName: "gwbc",
		Algorithm:   otp.AlgorithmSHA512,
	})

	if err != nil {
		c.JSON(200, gin.H{
			"err":   err.Error(),
			"infos": "",
		})

		return
	}

	pwd := DynamicPassword{}

	//不能使用key.URL，手机动态识别的用户信息错误
	pwd.AuthURL = fmt.Sprintf("otpauth://totp/%s?secret=%s&issuer=%s", key.AccountName(), key.Secret(), key.Issuer())
	pwd.Secret = key.Secret()

	c.JSON(200, gin.H{
		"err":   "",
		"infos": pwd,
	})
}
