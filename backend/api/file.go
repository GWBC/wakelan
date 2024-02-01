package api

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path"
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
	fMD5 := c.Query("file")
	if len(fMD5) == 0 {
		err := errors.New("文件名为空")
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	fName := c.Query("name")

	file, err := os.Open(path.Join(f.filecachePath, fMD5))
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	fileSize := fileInfo.Size()

	if len(fName) != 0 {
		//设置文件名称
		c.Header("Content-Disposition",
			fmt.Sprintf("attachment; filename=\"%s\"", fName))
	} else {
		fName = fMD5
	}

	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Length", strconv.FormatInt(fileSize, 10))
	c.Header("Accept-Ranges", "bytes")

	rangeHeader := c.GetHeader("Range")
	if len(rangeHeader) == 0 {
		db.DBLog("文件传输", "下载文件:%s", fName)
		io.Copy(c.Writer, file)
		return
	}

	rangeHeader = strings.Replace(rangeHeader, "bytes=", "", 1)
	rangeHeader = strings.Replace(rangeHeader, " ", "", -1)
	rangeSplits := strings.Split(rangeHeader, ",")

	if len(rangeSplits) == 0 {
		c.String(http.StatusRequestedRangeNotSatisfiable, err.Error())
		return
	}

	//判断是否为断点下载的开始
	if strings.Contains(rangeSplits[0], "0-0") {
		db.DBLog("文件传输", "下载文件:%s", fName)
	}

	for _, rangeSplit := range rangeSplits {
		rangePoints := strings.Split(rangeSplit, "-")
		start, err := strconv.ParseInt(rangePoints[0], 10, 64)
		if err != nil {
			c.String(http.StatusRequestedRangeNotSatisfiable, err.Error())
			return
		}

		var end int64
		if rangePoints[1] == "" {
			// 如果没有指定结束位置，表示请求到文件末尾
			end = fileSize - 1
		} else {
			// 如果指定了结束位置，解析结束位置
			end, err = strconv.ParseInt(rangePoints[1], 10, 64)
			if err != nil {
				c.String(http.StatusRequestedRangeNotSatisfiable, err.Error())
				return
			}
		}

		// 检查请求的范围是否有效
		if start < 0 || end >= fileSize || start > end {
			c.String(http.StatusRequestedRangeNotSatisfiable, "Invalid range")
			return
		}

		contentLen := end - start + 1

		// 设置响应头，告诉客户端文件的范围，发送206状态码
		c.Header("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, fileSize))
		c.Header("Content-length", strconv.FormatInt(contentLen, 10))
		c.Status(http.StatusPartialContent)

		file.Seek(start, 0)
		io.CopyN(c.Writer, file, contentLen)
	}
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
