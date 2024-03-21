package api

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"wakelan/backend/db"

	"github.com/gin-gonic/gin"
)

func Download(filePath string, c *gin.Context) {
	orgName := c.Query("file")
	if len(orgName) == 0 {
		err := errors.New("文件名为空")
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	fName := c.Query("name")

	file, err := os.Open(path.Join(filePath, orgName))
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

	if len(fName) == 0 {
		fName = orgName
	}

	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Length", strconv.FormatInt(fileSize, 10))
	c.Header("Accept-Ranges", "bytes")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", fName)) //设置文件名称

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
		c.String(http.StatusRequestedRangeNotSatisfiable, "Invalid range")
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
