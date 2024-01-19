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
	GuacdHost string `gorm:"column:guacd_host"  json:"guacd_host"`
	GuacdPort int    `gorm:"column:guacd_port"  json:"guacd_port"`
	AuthURL   string `gorm:"column:auth_url"  json:"auth_url"`
	Secret    string `gorm:"column:secret"  json:"secret"`

	AYFFToken       string `gorm:"column:ayff_token"  json:"ayff_token"`
	WXPusherToken   string `gorm:"column:wxpusher_token"  json:"wxpusher_token"`
	WXPusherTopicId int    `gorm:"column:wxpusher_topicid"  json:"wxpusher_topicid"`
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

	guacdInfo := ConfigInfo{}
	guacdInfo.GuacdHost = info.GuacdHost
	guacdInfo.GuacdPort = info.GuacdPort
	guacdInfo.AuthURL = info.AuthURL
	guacdInfo.Secret = info.Secret
	guacdInfo.AYFFToken = info.AYFFToken
	guacdInfo.WXPusherToken = info.WXPusherToken
	guacdInfo.WXPusherTopicId = info.WXPusherTopicId

	c.JSON(200, gin.H{
		"err":   "",
		"infos": guacdInfo,
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
	cfg.GuacdHost = cfgInfo.GuacdHost
	cfg.GuacdPort = cfgInfo.GuacdPort
	cfg.AuthURL = cfgInfo.AuthURL
	cfg.Secret = cfgInfo.Secret
	cfg.AYFFToken = cfgInfo.AYFFToken
	cfg.WXPusherToken = cfgInfo.WXPusherToken
	cfg.WXPusherTopicId = cfgInfo.WXPusherTopicId

	dbObj.Select("guacd_host", "guacd_port", "auth_url", "secret",
		"ayff_token", "wxpusher_token", "wxpusher_topicid").Save(cfg)

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
