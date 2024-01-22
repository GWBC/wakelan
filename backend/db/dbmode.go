package db

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
	"wakelan/backend/comm"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type MacInfo struct {
	IP         string     `gorm:"column:ip;primary_key" json:"ip"`
	Mac        string     `gorm:"column:mac;primary_key" json:"mac"`
	MANUF      string     `gorm:"column:manuf" json:"manuf"`
	AttachInfo AttachInfo `gorm:"foreignkey:mac;references:mac" json:"attach_info"`
}

type AttachInfo struct {
	Mac      string `gorm:"column:mac;primary_key" json:"mac"`
	Star     bool   `gorm:"column:star" json:"star"`
	Describe string `gorm:"column:describe" json:"describe"`
	Remote   string `gorm:"column:remote" json:"remote"`
}

type GlobalInfo struct {
	gorm.Model
	IP        string `gorm:"column:ip"`
	NetCard   string `gorm:"column:netcard"`
	GuacdHost string `gorm:"column:guacd_host" json:"guacd_host"`
	GuacdPort int    `gorm:"column:guacd_port" json:"guacd_port"`
	Secret    string `gorm:"column:secret" json:"secret"`
	AuthURL   string `gorm:"column:auth_url" json:"auth_url"`

	AYFFToken       string `gorm:"column:ayff_token" json:"ayff_token"`
	WXPusherToken   string `gorm:"column:wxpusher_token" json:"wxpusher_token"`
	WXPusherTopicId int    `gorm:"column:wxpusher_topicid" json:"wxpusher_topicid"`

	Debug bool `gorm:"column:debug" json:"debug"`
}

type Log struct {
	gorm.Model
	Cmd string `gorm:"column:cmd" json:"cmd"`
	Msg string `gorm:"column:msg" json:"msg"`
}

// 处理json编码
func (l *Log) MarshalJSON() ([]byte, error) {
	datas := struct {
		Log
		Time string `json:"time"`
	}{
		*l,
		l.CreatedAt.Format("2006-01-02 15:04:05"),
	}

	return json.Marshal(datas)
}

type FileMeta struct {
	MD5       string `gorm:"column:md5;primary_key" json:"md5"`
	Name      string `gorm:"column:name" json:"name"`
	Size      int    `gorm:"column:size" json:"size"`
	Index     int    `gorm:"column:index" json:"index"`
	CreatedAt time.Time
}

// 处理json编码
func (m *FileMeta) MarshalJSON() ([]byte, error) {
	datas := struct {
		FileMeta
		Time string `json:"time"`
	}{
		*m,
		m.CreatedAt.Format("2006-01-02 15:04:05"),
	}

	return json.Marshal(datas)
}

type DBOper struct {
	db *gorm.DB
}

func (d *DBOper) initData(db *gorm.DB) {
	var count int64
	db.Model(&GlobalInfo{}).Count(&count)

	if count == 0 {
		// 插入初始数据
		initData := GlobalInfo{}
		initData.GuacdHost = "127.0.0.1"
		initData.GuacdPort = 4822
		db.Create(&initData)
	}
}

func (d *DBOper) Init() error {
	dbPath := filepath.Join(comm.Pwd(), "data/")
	os.MkdirAll(dbPath, 0755)

	db, err := gorm.Open(sqlite.Open(filepath.Join(dbPath, "data.db")), &gorm.Config{})

	if err != nil {
		return err
	}

	d.db = db

	db.AutoMigrate(&MacInfo{}, &GlobalInfo{}, &AttachInfo{}, &Log{}, &FileMeta{})

	d.SwitchLogger()
	d.initData(db)

	return nil
}

func (d *DBOper) GetDB() *gorm.DB {
	return d.db
}

func (d *DBOper) GetConfig() *GlobalInfo {
	info := &GlobalInfo{}
	d.db.Find(info)

	return info
}

func (d *DBOper) SwitchLogger() {
	cfg := d.GetConfig()

	if cfg.Debug {
		d.db.Config.Logger = logger.New(d, logger.Config{
			SlowThreshold: 200 * time.Millisecond, //慢日志阈值
			LogLevel:      logger.Info,            //日志等级
			Colorful:      false,                  //是否彩色打印
		})
	} else {
		d.db.Config.Logger = d.db.Config.Logger.LogMode(logger.Silent)
	}
}

func (d *DBOper) DBLog(cmd string, format string, a ...any) {
	if len(a) < 4 {
		return
	}

	info := &Log{}
	info.Cmd = cmd
	info.Msg = a[3].(string)

	if strings.Contains(info.Msg, "logs") {
		return
	}

	d.db.Save(info)
}

func (d *DBOper) Printf(format string, a ...any) {
	d.DBLog("SQL语句", format, a...)
}

// /////////////////////////////////////////////////
var dbOperObj *DBOper
var dbOperOnce sync.Once

func DBOperObj() *DBOper {
	dbOperOnce.Do(func() {
		dbOper := &DBOper{}
		err := dbOper.Init()

		if err != nil {
			return
		}

		dbOperObj = dbOper
	})

	return dbOperObj
}

func DBLog(cmd string, format string, a ...any) {
	info := &Log{}
	info.Cmd = cmd
	info.Msg = fmt.Sprintf(format, a...)
	dbObj := DBOperObj().GetDB()
	dbObj.Save(info)
}
