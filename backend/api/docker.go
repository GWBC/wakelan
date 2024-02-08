package api

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
	"wakelan/backend/comm"
	"wakelan/backend/db"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type ImageInfo struct {
	ID          string `json:"id"`
	Repostitory string `json:"repostitory"`
	Tag         string `json:"tag"`
	Size        int64  `json:"size"`
	CreateTime  string `json:"create_time"`
}

type PortInfo struct {
	IP          string `json:"ip"`
	PrivatePort uint16 `json:"private_port"`
	PublicPort  uint16 `json:"public_port"`
	Type        string `json:"type"`
}

type NetInfo struct {
	Name       string   `json:"name"`
	MacAddress string   `json:"mac"`
	Gateway    string   `json:"gateway"`
	IPAddress  string   `json:"ip"`
	DNSNames   []string `json:"dns"`
}

type VolumeInfo struct {
	Type        string `json:"type"`
	Name        string `json:"name"`
	Source      string `json:"source"`
	Destination string `json:"destination"`
}

type ContainerInfo struct {
	ID         string       `json:"id"`
	Name       string       `json:"name"`
	Image      string       `json:"image"`
	Cmd        string       `json:"cmd"`
	V4Ports    []PortInfo   `json:"v4ports"`
	V6Ports    []PortInfo   `json:"v6ports"`
	Networks   []NetInfo    `json:"networks"`
	VolumeInfo []VolumeInfo `json:"volume_info"`
	RunTime    string       `json:"run_time"`
	State      string       `json:"state"`
	CreateTime string       `json:"create_time"`
}

type DockerClient struct {
	cli *comm.DockerClient
}

func (d *DockerClient) Init() error {
	d.cli = &comm.DockerClient{}

	return nil
}

func (d *DockerClient) loadConfig() {
	cfg := db.DBOperObj().GetConfig()
	if cfg.DockerEnableTCP {
		d.cli.SetHost(fmt.Sprintf("tcp://%s:%d", cfg.DockerSvrIP, cfg.DockerSvrPort))
	}
}

// 获取镜像信息
func (d *DockerClient) GetImages(c *gin.Context) {
	d.loadConfig()

	imgs, err := d.cli.GetImages("")
	if err != nil {
		c.JSON(200, gin.H{
			"err": err.Error(),
		})

		return
	}

	infos := []ImageInfo{}

	for _, img := range imgs {
		if len(img.RepoTags) == 0 {
			continue
		}

		repoAndTag := strings.Split(img.RepoTags[0], ":")
		if len(repoAndTag) < 2 {
			continue
		}

		id := img.ID
		idv := strings.Split(img.ID, ":")
		if len(idv) >= 2 {
			id = idv[1]
		}

		if len(id) > 12 {
			id = id[:12]
		}

		info := ImageInfo{}
		info.Repostitory = repoAndTag[0]
		info.ID = id
		info.Tag = repoAndTag[1]
		info.Size = img.Size
		info.CreateTime = time.Unix(img.Created, 0).Format(comm.TimeFormat)

		infos = append(infos, info)
	}

	c.JSON(200, gin.H{
		"err":   "",
		"infos": infos,
	})
}

// 获取容器信息
func (d *DockerClient) GetContainers(c *gin.Context) {
	d.loadConfig()

	containers, err := d.cli.GetContainers("")
	if err != nil {
		c.JSON(200, gin.H{
			"err": err.Error(),
		})

		return
	}

	infos := []ContainerInfo{}

	for _, container := range containers {
		id := container.ID
		if len(id) > 12 {
			id = id[:12]
		}

		imageName := container.Image
		if imageName == container.ImageID {
			v := strings.Split(container.Image, ":")
			if len(v) == 2 {
				imageName = v[1]

				if len(imageName) > 12 {
					imageName = imageName[:12]
				}
			}
		}

		containerName := ""
		if len(container.Names) != 0 {
			containerName = container.Names[0]
		}

		if strings.Index(containerName, "/") == 0 {
			containerName = containerName[1:]
		}

		info := ContainerInfo{}
		info.ID = id
		info.Name = containerName
		info.Cmd = container.Command
		info.Image = imageName
		info.CreateTime = time.Unix(container.Created, 0).Format(comm.TimeFormat)
		info.RunTime = container.Status
		info.State = container.State
		info.V4Ports = []PortInfo{}
		info.V6Ports = []PortInfo{}
		info.Networks = []NetInfo{}
		info.VolumeInfo = []VolumeInfo{}

		sort.Slice(container.Ports, func(i, j int) bool {
			return container.Ports[i].PublicPort < container.Ports[j].PublicPort
		})

		for _, v := range container.Ports {
			pInfo := PortInfo{}
			pInfo.IP = v.IP
			pInfo.PublicPort = v.PublicPort
			pInfo.PrivatePort = v.PrivatePort
			pInfo.Type = v.Type

			ip := net.ParseIP(pInfo.IP)
			if ip.To4() != nil {
				info.V4Ports = append(info.V4Ports, pInfo)
			} else if ip.To16() != nil {
				info.V6Ports = append(info.V6Ports, pInfo)
			}
		}

		for k, v := range container.NetworkSettings.Networks {
			nInfo := NetInfo{}
			nInfo.Name = k
			nInfo.MacAddress = v.MacAddress
			nInfo.IPAddress = v.IPAddress
			nInfo.Gateway = v.Gateway
			nInfo.DNSNames = v.DNSNames

			info.Networks = append(info.Networks, nInfo)
		}

		for _, v := range container.Mounts {
			vInfo := VolumeInfo{}
			vInfo.Type = string(v.Type)
			vInfo.Name = v.Name
			vInfo.Source = v.Source
			vInfo.Destination = v.Destination

			info.VolumeInfo = append(info.VolumeInfo, vInfo)
		}

		infos = append(infos, info)
	}

	c.JSON(200, gin.H{
		"err":   "",
		"infos": infos,
	})
}

// 删除容器
func (d *DockerClient) DelContainer(c *gin.Context) {
	d.loadConfig()

	name := c.Query("name")
	if len(name) == 0 {
		c.JSON(200, gin.H{
			"err": errors.New("name is empty"),
		})

		return
	}

	err := d.cli.DelContainer(name, true)
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

// 修改容器名称
func (d *DockerClient) RenameContainer(c *gin.Context) {
	d.loadConfig()

	old := c.Query("old")
	new := c.Query("new")

	if len(old) == 0 || len(new) == 0 {
		c.JSON(200, gin.H{
			"err": errors.New("old or new name is empty"),
		})

		return
	}

	err := d.cli.RenameContainer(old, new)
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

// 获取容器日志
func (d *DockerClient) getContainerLogs(c *gin.Context) {
	name := c.Query("name")
	if len(name) == 0 {
		c.JSON(200, gin.H{
			"err": errors.New("name is empty"),
		})

		return
	}

	wbsocket := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	conn, err := wbsocket.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(200, gin.H{
			"err": err.Error(),
		})

		return
	}

	defer func() {
		conn.Close()
	}()

	dec, err := d.cli.InspectContainer(name)
	if err != nil {
		if err != nil {
			c.JSON(200, gin.H{
				"err": err.Error(),
			})

			return
		}
	}

	type Head struct {
		Type     byte
		Reserve1 byte
		Reserve2 byte
		Reserve3 byte
		Len      uint32
	}

	head := Head{}
	headBuf := bytes.NewBuffer(nil)
	bodyBuf := bytes.NewBuffer(nil)

	buf := make([]byte, 4096)

	d.cli.LogsContainer(name, func(r *bufio.Reader) error {
		for {
			//使用tty，只有一种格式，直接转发
			if dec.Config.Tty {
				n, err := r.Read(buf)
				if err != nil {
					return err
				}

				err = conn.WriteMessage(websocket.TextMessage, buf[:n])
				if err != nil {
					return err
				}
			} else {
				//处理包头
				_, err := io.CopyN(headBuf, r, int64(binary.Size(head)))
				if err != nil {
					return err
				}

				err = binary.Read(headBuf, binary.BigEndian, &head)
				if err != nil {
					return err
				}

				//处理包体
				_, err = io.CopyN(bodyBuf, r, int64(head.Len))
				if err != nil {
					return err
				}

				bodyBuf.WriteString("\r")

				err = conn.WriteMessage(websocket.TextMessage, bodyBuf.Bytes())
				if err != nil {
					return err
				}

				headBuf.Reset()
				bodyBuf.Reset()
			}
		}
	})
}

// 终端探测
func (d *DockerClient) probeContainerTerm(name string) string {
	terms := []string{"/bin/bash", "/bin/sh"}

	for _, v := range terms {
		cfg := &comm.ContainerExecConfig{
			Name: name,
			Cmd:  []string{"ls", v},
		}

		err := d.cli.ExecContainer(cfg)
		if err == nil {
			return v
		}
	}

	return ""
}

// 进入容器终端
func (d *DockerClient) enterContainer(c *gin.Context) {
	name := c.Query("name")
	if len(name) == 0 {
		c.JSON(200, gin.H{
			"err": errors.New("name is empty"),
		})

		return
	}

	rows, _ := strconv.Atoi(c.Query("rows"))
	cols, _ := strconv.Atoi(c.Query("cols"))

	term := d.probeContainerTerm(name)
	if len(term) == 0 {
		c.JSON(200, gin.H{
			"err": errors.New("term is empty"),
		})

		return
	}

	wbsocket := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	conn, err := wbsocket.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(200, gin.H{
			"err": err.Error(),
		})

		return
	}

	defer conn.Close()

	buf := make([]byte, 4096)
	ttyCfg := &comm.ContainerExecTTYConfig{
		TermType: "xterm",
		Rows:     uint(rows),
		Columns:  uint(cols),
		ReadFun: func(r io.Reader) error {
			n, err := r.Read(buf)
			if err != nil {
				return err
			}

			err = conn.WriteMessage(websocket.TextMessage, buf[:n])
			if err != nil {
				conn.Close()
				return err
			}

			return nil
		},
		WriteFun: func(w io.Writer) (uint, uint, error) {
			msgType, data, err := conn.ReadMessage()
			if err != nil {
				conn.Close()
				return 0, 0, err
			}

			if msgType != websocket.TextMessage {
				return 0, 0, nil
			}

			cmd := map[string]interface{}{}
			err = json.Unmarshal(data, &cmd)
			if err != nil {
				conn.Close()
				return 0, 0, err
			}

			if cmd["cmd"] == "resize" {
				rows, _ := cmd["rows"].(float64)
				cols, _ := cmd["cols"].(float64)
				return uint(rows), uint(cols), nil
			} else if cmd["cmd"] == "data" {
				d, ok := cmd["data"].(string)
				if ok {
					_, err = w.Write([]byte(d))
				}
			}

			if err != nil {
				conn.Close()
				return 0, 0, err
			}

			return 0, 0, nil
		},
	}

	cfg := &comm.ContainerExecConfig{
		Name:   name,
		Cmd:    []string{term},
		TTYCfg: ttyCfg,
	}

	d.cli.ExecContainer(cfg)
}
