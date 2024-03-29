package db

import (
	"context"
	"encoding/json"
	"errors"
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
	Mac        string     `gorm:"column:mac;primary_key" json:"mac"`
	IP         string     `gorm:"column:ip" json:"ip"`
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

	RandKey string `gorm:"column:rand_key" json:"rand_key"`

	AYFFToken       string `gorm:"column:ayff_token" json:"ayff_token"`
	WXPusherToken   string `gorm:"column:wxpusher_token" json:"wxpusher_token"`
	WXPusherTopicId int    `gorm:"column:wxpusher_topicid" json:"wxpusher_topicid"`

	Debug       bool `gorm:"column:debug" json:"debug"`
	SharedLimit int  `gorm:"column:shared_limit;default:7" json:"shared_limit"`

	DockerEnableTCP   bool   `gorm:"docker_enable_tcp;default:false" json:"docker_enable_tcp"`
	DockerSvrIP       string `gorm:"docker_svr_ip;default:127.0.0.1" json:"docker_svr_ip"`
	DockerSvrPort     int    `gorm:"docker_svr_port;default:2375" json:"docker_svr_port"`
	ContainerRootPath string `gorm:"container_root_path;default:/opt/container-root" json:"container_root_path"`
	DockerUser        string `gorm:"docker_user" json:"docker_user"`
	DockerPasswd      string `gorm:"docker_passwd" json:"docker_passwd"`

	CheckIPAddr string `gorm:"column:check_ip_addr;default:http://ddns.oray.com/checkip;https://ipinfo.io/ip;" json:"check_ip_addr"`
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
		l.CreatedAt.Format(comm.TimeFormat),
	}

	return json.Marshal(datas)
}

type FileMeta struct {
	MD5       string `gorm:"column:md5;primary_key" json:"md5"`
	Name      string `gorm:"column:name" json:"name"`
	Size      int    `gorm:"column:size" json:"size"`
	Index     int    `gorm:"column:index" json:"index"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

// 处理json编码
func (m *FileMeta) MarshalJSON() ([]byte, error) {
	datas := struct {
		FileMeta
		Time string `json:"time"`
	}{
		*m,
		m.CreatedAt.Format(comm.TimeFormat),
	}

	return json.Marshal(datas)
}

type Message struct {
	gorm.Model
	Msg string `gorm:"column:msg" json:"msg"`
}

// 处理json编码
func (m *Message) MarshalJSON() ([]byte, error) {
	datas := struct {
		Message
		Time string `json:"time"`
	}{
		*m,
		m.CreatedAt.Format(comm.TimeFormat),
	}

	return json.Marshal(datas)
}

type DBOper struct {
	db    *gorm.DB
	level logger.LogLevel
}

func (d *DBOper) initData(db *gorm.DB) {
	var count int64
	db.Model(&GlobalInfo{}).Count(&count)

	if count == 0 {
		// 插入初始数据
		initData := GlobalInfo{}
		initData.GuacdHost = "127.0.0.1"
		initData.GuacdPort = 4822
		initData.RandKey = comm.GenRandKey()
		db.Create(&initData)
	} else {
		cfg := d.GetConfig()
		if len(cfg.RandKey) == 0 {
			cfg.RandKey = comm.GenRandKey()
			ret := db.Save(cfg)
			if ret.Error != nil {
				panic(ret.Error.Error())
			}
		}
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
	d.db.Config.Logger = d
	d.db.Config.Logger.LogMode(logger.Silent)

	db.AutoMigrate(&MacInfo{}, &GlobalInfo{}, &AttachInfo{}, &Log{}, &FileMeta{}, &Message{})

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
		d.db.Config.Logger.LogMode(logger.Info)
	} else {
		d.db.Config.Logger.LogMode(logger.Silent)
	}
}

func (d *DBOper) LogMode(level logger.LogLevel) logger.Interface {
	d.level = level
	return d
}

func (d *DBOper) Info(ctx context.Context, msg string, data ...interface{}) {
}

func (d *DBOper) Warn(ctx context.Context, msg string, data ...interface{}) {
}

func (d *DBOper) Error(ctx context.Context, msg string, data ...interface{}) {
}

func (d *DBOper) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if d.level == logger.Silent {
		return
	}

	sql, rowsAffected := fc()
	if strings.Contains(sql, "logs") {
		return
	}

	info := &Log{}
	info.Cmd = "SQL语句"
	info.Msg = fmt.Sprintf("语句：%s 行数：%d", sql, rowsAffected)

	if err != nil && !errors.Is(err, logger.ErrRecordNotFound) {
		info.Msg += " 错误：" + err.Error()
	}

	d.db.Save(info)
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
