package api

import (
	"errors"
	"io"
	"mime/multipart"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"
	"wakelan/backend/comm"
	"wakelan/backend/db"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type FileTransfer struct {
	filecachePath string
	fileSharedKey string
}

func (f *FileTransfer) Init() error {
	f.filecachePath = path.Join(comm.Pwd(), "data", "filecache")
	err := os.MkdirAll(f.filecachePath, 0755)
	if err != nil {
		return err
	}

	f.autoClean()

	f.fileSharedKey = comm.GenRandKey()

	return nil
}

func (f *FileTransfer) autoClean() {
	go func() {
		for {
			dbObj := db.DBOperObj().GetDB()
			cfg := db.DBOperObj().GetConfig()

			t := time.Now()
			fileMetas := []db.FileMeta{}
			dbObj.Find(&fileMetas)

			if cfg.SharedLimit <= 0 && cfg.SharedLimit >= 30 {
				cfg.SharedLimit = 7
			}

			for _, data := range fileMetas {
				if t.Sub(data.CreatedAt) > time.Duration(cfg.SharedLimit)*24*time.Hour {
					dbObj.Delete(&data)
					os.Remove(path.Join(f.filecachePath, data.MD5))
				}
			}

			messages := []db.Message{}
			dbObj.Find(&messages)

			for _, data := range messages {
				if t.Sub(data.CreatedAt) > time.Duration(cfg.SharedLimit)*24*time.Hour {
					dbObj.Delete(&data)
				}
			}

			time.Sleep(1 * time.Hour)
		}
	}()
}

func (f *FileTransfer) GetFileMeta(c *gin.Context) {
	md5 := c.Query("md5")

	datas := []db.FileMeta{}
	dbObj := db.DBOperObj().GetDB()
	var res *gorm.DB
	if len(md5) == 0 {
		res = dbObj.Order("created_at DESC").Find(&datas)
	} else {
		res = dbObj.Where("md5=?", md5).Find(&datas)
	}

	if res.Error != nil {
		c.JSON(200, gin.H{
			"err": res.Error.Error(),
		})

		return
	}

	if len(datas) == 0 {
		c.JSON(200, gin.H{
			"err":   "",
			"infos": datas,
		})
		return
	}

	for i, data := range datas {
		//判断文件是否存在
		_, err := os.Stat(path.Join(f.filecachePath, data.MD5))
		if os.IsNotExist(err) {
			datas[i].Index = 0
		}
	}

	c.JSON(200, gin.H{
		"err":   "",
		"infos": datas,
	})
}

func (f *FileTransfer) MoveToDockerBackup(name string, fileName string) error {
	re := regexp.MustCompile(`(.*?)_(.*?)_(.*?)_(.*)`)
	e := re.FindAllStringSubmatch(name, -1)
	if len(e) > 0 {
		if strings.EqualFold(e[0][1], "docker") {
			cfg := db.DBOperObj().GetConfig()
			imageBackupPath := path.Join(cfg.ContainerRootPath, "image-backup")
			containerBackupPath := path.Join(cfg.ContainerRootPath, "container-backup")

			in, err := os.OpenFile(fileName, os.O_RDWR, 0666)
			if err != nil {
				return err
			}
			defer in.Close()

			var out *os.File

			if e[0][3] == "image" {
				out, err = os.Create(imageBackupPath + "/" + e[0][4])
			} else {
				out, err = os.Create(containerBackupPath + "/" + e[0][4])
			}

			if err != nil {
				return err
			}
			defer out.Close()

			_, err = io.Copy(out, in)
			if err != nil {
				return err
			}

			db.DBLog("文件转移", "%s 文件转移到容器备份目录", name)
		}
	}

	return nil
}

func (f *FileTransfer) MoveToDockerBackupApi(c *gin.Context) {
	fileName := path.Join(f.filecachePath, c.Query("md5"))
	err := f.MoveToDockerBackup(c.Query("name"), fileName)

	if err != nil {
		c.JSON(200, gin.H{
			"err": err.Error(),
		})

		return
	}

	c.JSON(200, gin.H{
		"err": "",
	})
}

func (f *FileTransfer) Upload(c *gin.Context) {
	meta := db.FileMeta{}
	meta.MD5 = c.PostForm("md5")
	meta.Name = c.PostForm("name")
	meta.Size, _ = strconv.Atoi(c.PostForm("size"))
	meta.Index, _ = strconv.Atoi(c.PostForm("index"))
	meta.CreatedAt = time.Now()

	var err error

	var out *os.File
	var src multipart.File

	var file *multipart.FileHeader
	fileName := path.Join(f.filecachePath, meta.MD5)

	dbObj := db.DBOperObj().GetDB()

	if meta.Index == 0 {
		db.DBLog("文件传输", "上传文件:%s，MD5:%s", meta.Name, meta.MD5)
	}

	if meta.Index != meta.Size {
		if meta.Size == 0 {
			err = errors.New("filesize 0")
			goto UploadErr
		}

		if len(meta.MD5) == 0 {
			err = errors.New("md5 empty")
			goto UploadErr
		}

		file, err = c.FormFile("file")
		if err != nil {
			goto UploadErr
		}

		src, err = file.Open()
		if err != nil {
			goto UploadErr
		}
		defer src.Close()

		if meta.Index == 0 {
			out, err = os.Create(fileName)
		} else {
			out, err = os.OpenFile(fileName, os.O_RDWR|os.O_APPEND, 0666)
		}

		if err != nil {
			goto UploadErr
		}

		defer out.Close()

		_, err = out.Seek(int64(meta.Index), 0)
		if err != nil {
			goto UploadErr
		}

		_, err = io.Copy(out, src)
		if err != nil {
			goto UploadErr
		}

		meta.Index += int(file.Size)
	}

	if meta.Index == meta.Size {
		err = f.MoveToDockerBackup(meta.Name, fileName)
		if err != nil {
			goto UploadErr
		}
	}

	dbObj.Save(&meta)

	c.JSON(200, gin.H{
		"err": "",
	})

	return

UploadErr:
	meta.Index = 0
	if len(meta.MD5) != 0 {
		dbObj.Save(&meta)
	}

	c.JSON(200, gin.H{
		"err": err.Error(),
	})
}

func (f *FileTransfer) Download(c *gin.Context) {
	Download(f.filecachePath, c)
}

func (f *FileTransfer) GenKey(c *gin.Context) {
	key, err := FileSharedMG().GenToken(int(6 * time.Hour))
	if err != nil {
		c.JSON(200, gin.H{
			"err": err.Error(),
		})

		return
	}

	c.JSON(200, gin.H{
		"err":   "",
		"infos": key,
	})
}

func (f *FileTransfer) Verify(c *gin.Context) {
	key := c.Query("key")
	if len(key) == 0 {
		c.JSON(200, gin.H{
			"err":   "无效key",
			"infos": "",
		})
		return
	}

	if !FileSharedMG().VerifyToken(key) {
		c.JSON(200, gin.H{
			"err":   "无效key",
			"infos": "",
		})
		return
	}

	c.JSON(200, gin.H{
		"err": "",
	})
}

//////////////////////////////////////////////////////////

func (f *FileTransfer) GetMessage(c *gin.Context) {
	dbObj := db.DBOperObj().GetDB()
	msgs := []db.Message{}
	res := dbObj.Order("id desc").Find(&msgs)
	if res.Error != nil {
		c.JSON(200, gin.H{
			"err": res.Error.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"err":   "",
		"infos": msgs,
	})
}

func (f *FileTransfer) AddMessage(c *gin.Context) {
	msg := db.Message{}
	c.ShouldBindJSON(&msg)

	if len(msg.Msg) == 0 {
		c.JSON(200, gin.H{
			"err": "",
		})

		return
	}

	db.DBLog("消息", "%s", msg.Msg)

	dbObj := db.DBOperObj().GetDB()
	res := dbObj.Save(&msg)
	if res.Error != nil {
		c.JSON(200, gin.H{
			"err": res.Error.Error(),
		})
		return
	}

	datas := struct {
		db.Message
		Time string `json:"time"`
	}{
		msg,
		msg.CreatedAt.Format(comm.TimeFormat),
	}

	c.JSON(200, gin.H{
		"err":   "",
		"infos": datas,
	})
}
