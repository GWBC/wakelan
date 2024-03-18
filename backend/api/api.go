package api

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
	"wakelan/backend/comm"
	"wakelan/backend/db"

	"github.com/gin-gonic/gin"
	"github.com/pquerna/otp/totp"
)

type Web struct {
}

func (a *Web) SetPublicAPI(r *gin.Engine) {
	group := r.Group("/api/public")
	group.GET("/getRandKey", func(c *gin.Context) {
		cfg := db.DBOperObj().GetConfig()

		c.JSON(200, gin.H{
			"err":   "",
			"infos": cfg.RandKey,
		})
	})
}

func (a *Web) SetWakeAPI(r *gin.Engine) {
	api := WakeApi{}
	api.Init()

	group := r.Group("/api/wake")
	group.GET("/getip", api.getGlobalIP)
	group.GET("/getinterfaces", api.getInterfaces)
	group.GET("/probenetwork", api.probeNetwork)
	group.GET("/delnetworklist", api.delNetworklist)
	group.GET("/getnetworklist", api.getNetworklist)
	group.GET("/wakeLan", api.wakeLan)
	group.GET("/operstar", api.operStar)
	group.GET("/opencard", api.openCard)
	group.GET("/getselectnetcard", api.getSelectNetCard)
	group.GET("/pingpc", api.pingPC)
	group.GET("/editpcinfo", api.editPCInfo)
	group.GET("/addnetworklist", api.addNetworklist)
}

func (a *Web) SetRemoteAPI(r *gin.Engine) {
	api := &Remote{}
	api.Init()

	group := r.Group("/api/remote")
	group.GET("/conn", api.remote)
	group.POST("/setting", api.setting)
}

func (a *Web) SetSystemAPI(r *gin.Engine) {
	api := &System{}
	api.Init()

	group := r.Group("/api/system")
	group.GET("/logsize", api.GetLogSize)
	group.GET("/log", api.GetLog)
	group.GET("/configinfo", api.GetConfigInfo)
	group.GET("/setconfig", api.SetConfig)
	group.GET("/genpwd", api.GenDynamicPassword)
}

func (a *Web) SetFileAPI(r *gin.Engine) {
	api := &FileTransfer{}
	api.Init()

	group := r.Group("/api/file")
	group.GET("/meta", api.GetFileMeta)
	group.POST("/upload", api.Upload)
	group.GET("/download", api.Download)
	group.GET("/genkey", api.GenKey)
	group.GET("/getMsg", api.GetMessage)
	group.POST("/addMsg", api.AddMessage)
}

func (a *Web) SetDockerClientApi(r *gin.Engine) {
	api := &DockerClient{}
	api.Init()

	group := r.Group("/api/docker")
	group.GET("/getImages", api.GetImages)
	group.GET("/getContainers", api.GetContainers)
	group.GET("/delContainer", api.DelContainer)
	group.GET("/renameContainer", api.RenameContainer)
	group.GET("/getLogsContainer", api.GetContainerLogs)
	group.GET("/enterContainer", api.EnterContainer)
	group.GET("/operContainer", api.OperContainer)
	group.GET("/getNetworkCards", api.GetNewtworkCards)
	group.GET("/delNetworkCard", api.DelNetworkCard)
	group.GET("/addNetworkCard", api.AddNetworkCard)
	group.GET("/localNetworkCard", api.LocalNetworkCard)
	group.GET("/getImageDetails", api.GetImageDetails)
	group.GET("/delImage", api.DelImage)
	group.GET("/queryImage", api.SearchImage)
	group.GET("/pullImage", api.PullImage)
	group.GET("/getPullImageLog", api.GetPullImageLog)
	group.POST("/runContainer", api.RunContainer)
	group.GET("/pushImage", api.PushImage)
	group.GET("/modifyImage", api.ModifyImage)
	group.GET("/getPushImageLog", api.GetPushImageLog)
	group.GET("/backupImage", api.BackupImage)
	group.GET("/backupContainer", api.BackupContainer)
}

///////////////////////////////////////////////////////////////////

var tokenManagerObj *comm.TokenManager
var tokenManagerOnce sync.Once

func TokenManager() *comm.TokenManager {
	tokenManagerOnce.Do(func() {
		obj := &comm.TokenManager{}
		cfg := db.DBOperObj().GetConfig()
		obj.Init([]byte(cfg.Secret), []byte("9C1A64F21B7B6A82"))
		tokenManagerObj = obj
	})

	return tokenManagerObj
}

var fileSharedObj *comm.TokenManager
var fileSharedOnce sync.Once

func FileSharedMG() *comm.TokenManager {
	fileSharedOnce.Do(func() {
		obj := &comm.TokenManager{}
		key := comm.GenRandKey()
		obj.Init([]byte(key), []byte(key)[0:16])
		fileSharedObj = obj
	})

	return fileSharedObj
}

///////////////////////////////////////////////////////////////////

func (a *Web) Login(r *gin.Engine) {
	api := r.Group("/api")
	api.GET("/login", func(c *gin.Context) {
		code := c.Query("code")

		info := db.DBOperObj().GetConfig()
		if len(info.Secret) != 0 {
			valid := totp.Validate(code, info.Secret)
			if !valid {
				db.DBLog("登录", "登录失败, key:%s", code)
				c.JSON(200, gin.H{
					"err":   "密钥无效",
					"infos": "",
				})
				return
			}
		}

		minute := 7 * 24 * 60
		token, err := TokenManager().GenToken(minute)
		if err != nil {
			db.DBLog("登录", "登录失败, key:%s, err:%s", code, err.Error())
			c.JSON(200, gin.H{
				"err":   "Token生成失败，err:" + err.Error(),
				"infos": "",
			})
			return
		}

		db.DBLog("登录", "登录成功, key:%s, token:%s", code, token)

		expiration := time.Now().Add(time.Duration(minute) * time.Minute)
		cookie := http.Cookie{
			Name:    "token",
			Value:   token,
			Expires: expiration,
		}

		http.SetCookie(c.Writer, &cookie)

		c.JSON(200, gin.H{
			"err":   "",
			"infos": len(info.Secret),
		})
	})
}

func (a *Web) LoadStatic(r *gin.Engine) {
	webPath := filepath.Join(filepath.Dir(os.Args[0]), "web")

	r.StaticFile("/", filepath.Join(webPath, "index.html"))
	r.StaticFile("/favicon.ico", filepath.Join(webPath, "favicon.ico"))
	r.Static("/assets", filepath.Join(webPath, "assets"))

	//解决其他路径无法访问，报404问题
	r.NoRoute(func(c *gin.Context) {
		c.File(filepath.Join(webPath, "index.html"))
	})
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 允许所有来源访问，也可以设置特定的来源地址
		c.Header("Access-Control-Allow-Origin", "*")

		// 设置其他 CORS 头部，根据需要进行调整
		c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		c.Header("Access-Control-Allow-Headers", "Origin, Authorization, Content-Type, Accept")

		// 允许发送 Cookie，如果有需要的话
		c.Header("Access-Control-Allow-Credentials", "true")

		// 如果是 OPTIONS 请求，直接返回 200 状态码，以处理预检请求
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusOK)
			return
		}

		c.Next()
	}
}

// 处理文件共享，返回是否处理
func (a *Web) ProcFileShared(c *gin.Context) bool {
	if strings.Contains(c.FullPath(), "/api/file") {
		if strings.Contains(c.FullPath(), "/api/file/genkey") {
			return false
		}

		key := c.Query("key")
		if len(key) == 0 {
			return false
		}

		if !FileSharedMG().VerifyToken(key) {
			c.JSON(200, gin.H{
				"err": "无效 key",
			})
			c.Abort()
			return true
		}

		c.Next()
		return true
	}

	return false
}

func (a *Web) CheckToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !strings.Contains(c.FullPath(), "/api/") {
			c.Next()
			return
		}

		if a.ProcFileShared(c) {
			return
		}

		token, err := c.Cookie("token")
		if err != nil || token == "" {
			c.JSON(200, gin.H{
				"err": "token 无效",
			})
			c.Abort()
			return
		}

		if TokenManager().VerifyToken(token) {
			c.Next()
		} else {
			c.JSON(200, gin.H{
				"err":   "token 无效",
				"infos": "",
			})
			c.Abort()
		}
	}
}

func (a *Web) Init(port string) {
	r := gin.Default()

	//允许跨域
	r.Use(CORSMiddleware())

	//加载静态资源
	a.LoadStatic(r)

	//登录
	a.Login(r)

	/////////////////////////////////////////////////
	//开启-后续Token验证
	r.Use(a.CheckToken())

	//设置公共接口
	a.SetPublicAPI(r)

	//设置唤醒接口
	a.SetWakeAPI(r)

	//设置远程接口
	a.SetRemoteAPI(r)

	//设置系统接口
	a.SetSystemAPI(r)

	//设置文件传输接口
	a.SetFileAPI(r)

	//设置容器接口
	a.SetDockerClientApi(r)

	// 启动服务
	r.Run(port)
}
