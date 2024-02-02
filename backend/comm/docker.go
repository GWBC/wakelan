package comm

import (
	"bufio"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"os"
	"path"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/registry"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

type DockerClient struct {
	host string
	auth string
}

func (d *DockerClient) conn() (*client.Client, error) {
	if len(d.host) == 0 {
		return client.NewClientWithOpts(client.FromEnv,
			client.WithAPIVersionNegotiation())
	}

	return client.NewClientWithOpts(client.FromEnv,
		client.WithAPIVersionNegotiation(),
		client.WithHost(d.host))
}

// 设置host
// 网络：tcp://ip:2375
func (d *DockerClient) SetHost(host string) {
	d.host = host
}

// 设置用户信息
func (d *DockerClient) SetUserInfo(user string, pwd string) error {
	d.auth = ""

	authCfg := registry.AuthConfig{
		Username: user,
		Password: pwd,
	}

	auth, err := registry.EncodeAuthConfig(authCfg)
	if err != nil {
		return err
	}

	d.auth = auth

	return nil
}

// //////////////////////////////////////////////////////////////////
// docker系统清理
func (d *DockerClient) SystemClean(isDeepClean bool) error {
	cli, err := d.conn()
	if err != nil {
		return err
	}

	defer cli.Close()

	_, err = cli.BuildCachePrune(context.Background(), types.BuildCachePruneOptions{
		All: true,
	})
	if err != nil {
		return err
	}

	_, err = cli.ContainersPrune(context.Background(), filters.NewArgs())
	if err != nil {
		return err
	}

	_, err = cli.NetworksPrune(context.Background(), filters.NewArgs())
	if err != nil {
		return err
	}

	_, err = cli.VolumesPrune(context.Background(), filters.NewArgs())
	if err != nil {
		return err
	}

	filterArgs := filters.NewArgs(filters.KeyValuePair{Key: "dangling", Value: fmt.Sprintf("%v", !isDeepClean)})

	_, err = cli.ImagesPrune(context.Background(), filterArgs)
	if err != nil {
		return err
	}

	return nil
}

// 获取版本
func (d *DockerClient) GetVersion() (types.Version, error) {
	cli, err := d.conn()
	if err != nil {
		return types.Version{}, nil
	}

	defer cli.Close()

	return cli.ServerVersion(context.Background())
}

// //////////////////////////////////////////////////////////////////
// 获取镜像
func (d *DockerClient) GetImages(imageName string) ([]image.Summary, error) {
	cli, err := d.conn()
	if err != nil {
		return []image.Summary{}, nil
	}

	defer cli.Close()

	filterArgs := filters.NewArgs()
	if len(imageName) != 0 {
		filterArgs = filters.NewArgs(filters.KeyValuePair{Key: "reference", Value: imageName})
	}

	return cli.ImageList(context.Background(), types.ImageListOptions{
		All:     true,
		Filters: filterArgs,
	})
}

// 拉取镜像
func (d *DockerClient) PullImage(imageName string) error {
	cli, err := d.conn()
	if err != nil {
		return nil
	}

	defer cli.Close()

	r, err := cli.ImagePull(context.Background(), imageName, types.ImagePullOptions{
		All:          false,
		RegistryAuth: d.auth,
	})

	if err != nil {
		return err
	}

	defer r.Close()

	reader := bufio.NewReader(r)

	buf := make([]byte, 1024)

	for {
		n, err := reader.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}

			return err
		}

		fmt.Println(string(buf[:n]))
	}

	return nil
}

// 删除镜像
func (d *DockerClient) DelImage(imageName string, force bool) ([]image.DeleteResponse, error) {
	cli, err := d.conn()
	if err != nil {
		return []image.DeleteResponse{}, nil
	}

	defer cli.Close()

	return cli.ImageRemove(context.Background(), imageName, types.ImageRemoveOptions{
		Force:         force,
		PruneChildren: true,
	})
}

// 获取镜像详情
func (d *DockerClient) InspectImage(imageName string) (types.ImageInspect, error) {
	cli, err := d.conn()
	if err != nil {
		return types.ImageInspect{}, err
	}

	defer cli.Close()

	info, _, err := cli.ImageInspectWithRaw(context.Background(), imageName)

	return info, err
}

// 镜像查找
func (d *DockerClient) SearchImage(imageName string) ([]registry.SearchResult, error) {
	cli, err := d.conn()
	if err != nil {
		return []registry.SearchResult{}, err
	}

	defer cli.Close()

	return cli.ImageSearch(context.Background(), imageName, types.ImageSearchOptions{})
}

// 导出镜像
func (d *DockerClient) ExportImage(imageName string, fileName string) (string, error) {
	cli, err := d.conn()
	if err != nil {
		return "", err
	}

	defer cli.Close()

	r, err := cli.ImageSave(context.Background(), []string{imageName})
	if err != nil {
		return "", err
	}
	defer r.Close()

	os.MkdirAll(path.Dir(fileName), 0755)

	fileName += ".gz"

	gzFile, err := os.Create(fileName)
	if err != nil {
		return "", err
	}
	defer gzFile.Close()

	gzWriter := gzip.NewWriter(gzFile)
	if err != nil {
		return "", err
	}

	defer gzWriter.Close()

	_, err = io.Copy(gzWriter, r)
	if err != nil {
		return "", err
	}

	return fileName, nil
}

// 导入镜像
func (d *DockerClient) ImportImage(fileName string) error {
	cli, err := d.conn()
	if err != nil {
		return err
	}

	defer cli.Close()

	gzFile, err := os.Open(fileName)
	if err != nil {
		return err
	}

	gzReader, err := gzip.NewReader(gzFile)
	if err != nil {
		return err
	}

	r, err := cli.ImageLoad(context.Background(), gzReader, true)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	return nil
}

// //////////////////////////////////////////////////////////////////
// 查询网络
func (d *DockerClient) GetNetworks() ([]types.NetworkResource, error) {
	cli, err := d.conn()
	if err != nil {
		return []types.NetworkResource{}, err
	}

	defer cli.Close()

	return cli.NetworkList(context.Background(), types.NetworkListOptions{})
}

type DockerNetCreate struct {
	Name    string //网络名称
	Driver  string //驱动：none,macvlan,bridge,host
	Subnet  string //子网
	Gateway string //网关
	Parent  string //父网卡
}

// 创建网络
func (d *DockerClient) AddNetwork(cfg *DockerNetCreate) error {
	cli, err := d.conn()
	if err != nil {
		return err
	}

	defer cli.Close()

	netCfg := types.NetworkCreate{}
	netCfg.Driver = cfg.Driver
	netCfg.IPAM = &network.IPAM{
		Config: []network.IPAMConfig{{
			Subnet:  cfg.Subnet,
			Gateway: cfg.Gateway,
		}},
	}

	netCfg.Options = map[string]string{"parent": cfg.Parent}

	_, err = cli.NetworkCreate(context.Background(), cfg.Name, netCfg)

	return err
}

// 删除网络
func (d *DockerClient) DelNetwork(name string) error {
	cli, err := d.conn()
	if err != nil {
		return err
	}

	defer cli.Close()

	return cli.NetworkRemove(context.Background(), name)
}

// //////////////////////////////////////////////////////////////////
// 获取容器
func (d *DockerClient) GetContainer() ([]types.Container, error) {
	cli, err := d.conn()
	if err != nil {
		return []types.Container{}, err
	}

	defer cli.Close()

	return cli.ContainerList(context.Background(), container.ListOptions{
		All: true,
	})
}

type DockerContainerCreate struct {
	Name       string   //容器名称
	Image      string   //镜像名称
	Cmd        []string //执行命令
	Ports      []string //端口映射  public:private/proto，public-public:private-private/proto
	Privileged bool     //是否开启特权
	NetName    string   //网络名称
}

// 运行容器
func (d *DockerClient) RunContainer(cfg *DockerContainerCreate) error {
	cli, err := d.conn()
	if err != nil {
		return err
	}

	defer cli.Close()

	imgs, err := d.GetImages(cfg.Image)
	if err != nil {
		return err
	}

	if len(imgs) == 0 {
		err = d.PullImage(cfg.Image)
		if err != nil {
			return err
		}
	}

	//容器配置
	containerCfg := container.Config{}
	containerCfg.Image = cfg.Image
	containerCfg.Cmd = append(containerCfg.Cmd, cfg.Cmd...)

	//端口映射配置
	_, portMap, err := nat.ParsePortSpecs(cfg.Ports)
	if err != nil {
		return err
	}

	containerHost := container.HostConfig{}
	containerHost.PortBindings = portMap
	containerHost.Privileged = cfg.Privileged

	//网络配置
	containerNet := network.NetworkingConfig{}
	if len(cfg.NetName) != 0 {
		containerNet.EndpointsConfig = map[string]*network.EndpointSettings{
			cfg.NetName: {},
		}
	}

	_, err = cli.ContainerCreate(context.Background(), &containerCfg, &containerHost, &containerNet, nil, cfg.Name)
	if err != nil {
		return err
	}

	err = cli.ContainerStart(context.Background(), cfg.Name, container.StartOptions{})
	if err != nil {
		return err
	}

	return nil
}
