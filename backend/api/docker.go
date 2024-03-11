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

type IPAMConfig struct {
	Subnet  string `json:"subnet"`
	IPRange string `json:"iprange"`
	Gateway string `json:"gateway"`
}

type NetworkCardInfo struct {
	Name    string            `json:"name"`
	ID      string            `json:"id"`
	Created string            `json:"created"`
	Scope   string            `json:"scope"`
	Driver  string            `json:"driver"`
	Options map[string]string `json:"options"`
	Configs []IPAMConfig      `json:"configs"`
}

type AddrNet struct {
	IP     string `json:"ip"`
	Subnet string `json:"subnet"`
}

type LocalNetworkInfo struct {
	Name  string    `json:"name"`
	Addrs []AddrNet `json:"addrs"`
}

type ExportPortInfo struct {
	Proto string `json:"proto"`
	Port  string `json:"port"`
}

type ImageDetailInfo struct {
	Os           string           `json:"os"`
	OsVersion    string           `json:"os_version"`
	Size         int64            `json:"size"`
	ExposedPorts []ExportPortInfo `json:"exposed_ports"`
	Volumes      []string         `json:"volumes"`
	WorkingDir   string           `json:"working_dir"`
	Env          []string         `json:"env"`
	Cmd          string           `json:"cmd"`
}

type PullLayerInfo struct {
	Id        string `json:"id"`
	Status    string `json:"status"`
	CurSize   int    `json:"cur_size"`
	TotalSize int    `json:"total_size"`
}

type PullLogInfo struct {
	Name    string          `json:"name"`
	Refresh bool            `json:"refresh"`
	Layer   []PullLayerInfo `json:"layer"`
}

type DockerClient struct {
	cli      *comm.DockerClient
	pullChan chan string
	pullLog  PullLogInfo
}

func (d *DockerClient) Init() error {
	d.cli = &comm.DockerClient{}
	d.pullLog = PullLogInfo{}
	d.pullChan = make(chan string, 10)
	d.ASyncPullImage()

	return nil
}

func (d *DockerClient) loadConfig() {
	cfg := db.DBOperObj().GetConfig()
	if cfg.DockerEnableTCP {
		d.cli.SetHost(fmt.Sprintf("tcp://%s:%d", cfg.DockerSvrIP, cfg.DockerSvrPort))
	}
}

// 获取宿主机网卡信息
func (d *DockerClient) LocalNetworkCard(c *gin.Context) {
	faces, err := net.Interfaces()
	if err != nil {
		c.JSON(200, gin.H{
			"err": err.Error(),
		})

		return
	}

	infos := []LocalNetworkInfo{}

	for _, v := range faces {
		flags := v.Flags & net.FlagLoopback
		if flags == net.FlagLoopback {
			continue
		}

		info := LocalNetworkInfo{}
		info.Name = v.Name
		addrs, err := v.Addrs()
		if err != nil {
			continue
		}

		info.Addrs = []AddrNet{}

		for _, addr := range addrs {
			_, subnet, err := net.ParseCIDR(addr.String())
			if err != nil {
				continue
			}

			info.Addrs = append(info.Addrs, AddrNet{IP: strings.Split(addr.String(), "/")[0], Subnet: subnet.String()})
		}

		infos = append(infos, info)
	}

	c.JSON(200, gin.H{
		"err":   "",
		"infos": infos,
	})
}

// 添加网卡
func (d *DockerClient) AddNetworkCard(c *gin.Context) {
	d.loadConfig()

	info := comm.DockerNetCreate{}
	info.Name = c.Query("name")
	info.Driver = c.Query("driver")
	info.Parent = c.Query("parent")
	info.Subnet = c.Query("subnet")
	info.Gateway = c.Query("gateway")

	err := d.cli.AddNetworkCard(&info)
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

// 删除网卡
func (d *DockerClient) DelNetworkCard(c *gin.Context) {
	d.loadConfig()

	name := c.Query("name")

	err := d.cli.DelNetworkCard(name)
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

// 获取网卡信息
func (d *DockerClient) GetNewtworkCards(c *gin.Context) {
	d.loadConfig()

	cards, err := d.cli.GetNetworkCards()
	if err != nil {
		c.JSON(200, gin.H{
			"err": err.Error(),
		})

		return
	}

	infos := []NetworkCardInfo{}

	for _, card := range cards {
		info := NetworkCardInfo{}
		info.Name = card.Name
		info.ID = card.ID[:12]
		info.Created = card.Created.Format(comm.TimeFormat)
		info.Scope = card.Scope
		info.Driver = card.Driver
		info.Options = card.Options
		info.Configs = []IPAMConfig{}
		for _, v := range card.IPAM.Config {
			info.Configs = append(info.Configs, IPAMConfig{
				Subnet:  v.Subnet,
				IPRange: v.IPRange,
				Gateway: v.Gateway,
			})
		}

		infos = append(infos, info)
	}

	sort.Slice(infos, func(i, j int) bool {
		return infos[i].Created > infos[j].Created
	})

	c.JSON(200, gin.H{
		"err":   "",
		"infos": infos,
	})
}

func (d *DockerClient) ASyncPullImage() {
	type Info struct {
		Id     string `json:"id"`
		Status string `json:"status"`

		ProgressDetail struct {
			Current int `json:"current"`
			Total   int `json:"total"`
		} `json:"progressDetail"`
	}

	go func() {
		for v := range d.pullChan {
			d.pullLog.Name = v
			d.pullLog.Layer = []PullLayerInfo{}

			err := d.cli.PullImage(v, func(r *bufio.Reader) error {
				for {
					s, err := r.ReadString('\n')
					if err != nil {
						if err == io.EOF {
							break
						}

						return err
					}

					info := Info{}
					err = json.Unmarshal([]byte(s), &info)
					if err != nil {
						continue
					}

					if info.ProgressDetail.Total == 0 {
						continue
					}

					isFind := false

					for i, p := range d.pullLog.Layer {
						if p.Id == info.Id {
							isFind = true
							d.pullLog.Layer[i].Status = info.Status
							d.pullLog.Layer[i].CurSize = info.ProgressDetail.Current
							d.pullLog.Layer[i].TotalSize = info.ProgressDetail.Total
							break
						}
					}

					if !isFind {
						d.pullLog.Layer = append(d.pullLog.Layer, PullLayerInfo{
							Id:        info.Id,
							Status:    info.Status,
							CurSize:   info.ProgressDetail.Current,
							TotalSize: info.ProgressDetail.Total,
						})
					}
				}

				return nil
			})

			if err != nil {
				d.pullLog.Layer = append(d.pullLog.Layer, PullLayerInfo{
					Id:        "Error",
					Status:    err.Error(),
					CurSize:   0,
					TotalSize: 0,
				})
			} else {
				d.pullLog.Refresh = true
				d.pullLog.Layer = append(d.pullLog.Layer, PullLayerInfo{
					Id:        "Success",
					Status:    "Success",
					CurSize:   0,
					TotalSize: 0,
				})
			}
		}
	}()
}

// 获取拉取日志
func (d *DockerClient) GetPullImageLog(c *gin.Context) {
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

	for {
		logMsg := d.pullLog

		if logMsg.Refresh {
			d.pullLog.Refresh = false
		}

		err := conn.WriteJSON(logMsg)
		if err != nil {
			break
		}

		time.Sleep(2 * time.Second)
	}
}

// 拉取镜像
func (d *DockerClient) PullImage(c *gin.Context) {
	d.loadConfig()

	name := c.Query("name")
	if len(name) == 0 {
		c.JSON(200, gin.H{
			"err": "The name cannot be empty",
		})

		return
	}

	d.pullChan <- name

	c.JSON(200, gin.H{
		"err": "",
	})
}

// 查询镜像
func (d *DockerClient) SearchImage(c *gin.Context) {
	d.loadConfig()

	name := c.Query("name")
	if len(name) == 0 {
		c.JSON(200, gin.H{
			"err": "The name cannot be empty",
		})

		return
	}

	infos, err := d.cli.SearchImage(name)
	if err != nil {
		c.JSON(200, gin.H{
			"err": err.Error(),
		})

		return
	}

	c.JSON(200, gin.H{
		"err":   "",
		"infos": infos,
	})
}

// 删除镜像
func (d *DockerClient) DelImage(c *gin.Context) {
	d.loadConfig()

	id := c.Query("id")
	if len(id) == 0 {
		c.JSON(200, gin.H{
			"err": "The id cannot be empty",
		})

		return
	}

	_, err := d.cli.DelImage(id, true)
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

// 获取镜像详细信息
func (d *DockerClient) GetImageDetails(c *gin.Context) {
	d.loadConfig()

	imgId := c.Query("id")
	if len(imgId) == 0 {
		c.JSON(200, gin.H{
			"err": "The id cannot be empty",
		})

		return
	}

	details, err := d.cli.InspectImage(imgId)
	if err != nil {
		c.JSON(200, gin.H{
			"err": err.Error(),
		})

		return
	}

	infos := ImageDetailInfo{}

	infos.Os = details.Os
	infos.OsVersion = details.OsVersion
	infos.Size = details.Size

	if details.Config != nil {
		for port := range details.Config.ExposedPorts {
			p := ExportPortInfo{}
			p.Proto = port.Proto()
			p.Port = port.Port()
			infos.ExposedPorts = append(infos.ExposedPorts, p)
		}

		for v := range details.Config.Volumes {
			infos.Volumes = append(infos.Volumes, v)
		}

		infos.Env = details.Config.Env
		for i, p := range details.Config.Cmd {
			if i != 0 {
				infos.Cmd += " "
			}

			infos.Cmd += p
		}

		infos.WorkingDir = details.Config.WorkingDir
	}

	c.JSON(200, gin.H{
		"err":   "",
		"infos": infos,
	})
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

// 操作容器
func (d *DockerClient) OperContainer(c *gin.Context) {
	d.loadConfig()

	name := c.Query("name")
	t := c.Query("oper")

	if len(name) == 0 || len(t) == 0 {
		c.JSON(200, gin.H{
			"err": errors.New("name or type is empty"),
		})

		return
	}

	var err error

	if t == "start" {
		err = d.cli.StartlContainer(name)
	} else if t == "stop" {
		err = d.cli.StopContainer(name)
	} else {
		err = d.cli.RestartContainer(name)
	}

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
func (d *DockerClient) GetContainerLogs(c *gin.Context) {
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
		c.JSON(200, gin.H{
			"err": err.Error(),
		})

		return
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
				if err != nil && n == 0 {
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
func (d *DockerClient) EnterContainer(c *gin.Context) {
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
			if err != nil && n == 0 {
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
