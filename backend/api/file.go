package api

import (
	"errors"
	"io"
	"mime/multipart"
	"os"
	"path"
	"strconv"
	"time"
	"wakelan/backend/comm"
	"wakelan/backend/db"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type FileTransfer struct {
	filecachePath string
}

func (f *FileTransfer) Init() error {
	f.filecachePath = path.Join(comm.Pwd(), "data", "filecache")
	err := os.MkdirAll(f.filecachePath, 0755)
	if err != nil {
		return err
	}

	f.autoClean()

	return nil
}

func (f *FileTransfer) autoClean() {
	go func() {
		for {
			t := time.Now()
			datas := []db.FileMeta{}
			dbObj := db.DBOperObj().GetDB()
			dbObj.Find(&datas)

			for _, data := range datas {
				if t.Sub(data.CreatedAt) > 7*24*time.Hour {
					dbObj.Delete(data)
					os.Remove(path.Join(f.filecachePath, data.MD5))
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
		res = dbObj.Find(&datas)
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
		datas = append(datas, db.FileMeta{
			MD5: md5,
		})

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

func (f *FileTransfer) Upload(c *gin.Context) {
	meta := db.FileMeta{}
	meta.MD5 = c.Request.FormValue("md5")
	meta.Name = c.Request.FormValue("name")
	meta.Size, _ = strconv.Atoi(c.Request.FormValue("size"))
	meta.Index, _ = strconv.Atoi(c.Request.FormValue("index"))
	meta.CreatedAt = time.Now()

	var err error

	var out *os.File
	var src multipart.File

	var file *multipart.FileHeader
	fileName := path.Join(f.filecachePath, meta.MD5)

	dbObj := db.DBOperObj().GetDB()

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
