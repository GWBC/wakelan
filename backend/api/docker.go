package api

import (
	"wakelan/backend/comm"

	"github.com/gin-gonic/gin"
)

type ImagesInfo struct {
	Repostitory string `json: "repostitory"`
	Tag         string `json: "tag"`
	ID          string `json: "id"`
	CreateTime  string `json: "create_time"`
	Size        int64  `json: "size"`
}

type DockerClientApi struct {
	cli *comm.DockerClient
}

func (d *DockerClientApi) Init() error {
	d.cli = &comm.DockerClient{}

	return nil
}

// 获取配置信息
func (d *DockerClientApi) GetImages(c *gin.Context) {
	// imgs, err := d.cli.GetImages("")
	// if err != nil {
	// 	c.JSON(200, gin.H{
	// 		"err": err.Error(),
	// 	})

	// 	return
	// }

	c.JSON(200, gin.H{
		"err": "",
	})
}
