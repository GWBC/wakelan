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
