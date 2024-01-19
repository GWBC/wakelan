package db

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
	"wakelan/backend/comm"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type MacInfo struct {
	IP     string     `gorm:"column:ip;primary_key" json:"ip"`
	Mac    string     `gorm:"column:mac" json:"mac"`
	MANUF  string     `gorm:"column:manuf" json:"manuf"`
	Star   StarInfo   `gorm:"foreignkey:ip;references:ip" json:"star_info"`
	Remote RemoteInfo `gorm:"foreignkey:ip;references:ip" json:"remote_info"`
}

type StarInfo struct {
	IP   string `gorm:"column:ip;primary_key" json:"ip"`
	Star bool   `gorm:"column:star" json:"star"`
}

type RemoteInfo struct {
	IP     string `gorm:"column:ip;primary_key" json:"ip"`
	Remote string `gorm:"column:remote" json:"remote"`
}

type GlobalInfo struct {
	gorm.Model
	IP        string `gorm:"column:ip"`
	NetCard   string `gorm:"column:netcard"`
	GuacdHost string `gorm:"column:guacd_host"  json:"guacd_host"`
	GuacdPort int    `gorm:"column:guacd_port"  json:"guacd_port"`
	Secret    string `gorm:"column:secret"  json:"secret"`
	AuthURL   string `gorm:"column:auth_url"  json:"auth_url"`

	AYFFToken       string `gorm:"column:ayff_token"  json:"ayff_token"`
	WXPusherToken   string `gorm:"column:wxpusher_token"  json:"wxpusher_token"`
	WXPusherTopicId int    `gorm:"column:wxpusher_topicid"  json:"wxpusher_topicid"`
}

type Log struct {
	gorm.Model
	Cmd  string `gorm:"column:cmd"  json:"cmd"`
	Msg  string `gorm:"column:msg"  json:"msg"`
	Time string `gorm:"-"  json:"time"`
}

type FileMeta struct {
	MD5       string    `gorm:"column:md5;primary_key" json:"md5"`
	Name      string    `gorm:"column:name" json:"name"`
	Size      int       `gorm:"column:size" json:"size"`
	Index     int       `gorm:"column:index" json:"index"`
	CreatedAt time.Time `json:"time"`
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

	db.AutoMigrate(&MacInfo{}, &GlobalInfo{}, &StarInfo{}, &RemoteInfo{}, &Log{}, &FileMeta{})

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
