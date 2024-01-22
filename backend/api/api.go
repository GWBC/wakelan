package api

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
	"wakelan/backend/db"

	"github.com/gin-gonic/gin"
	"github.com/pquerna/otp/totp"
)

type Web struct {
}

func (a *Web) SetWakeAPI(r *gin.Engine) {
	api := WakeApi{}
	api.Init()

	group := r.Group("/api/wake")
	group.GET("/getip", api.getGlobalIP)
	group.GET("/getinterfaces", api.getInterfaces)
	group.GET("/probenetwork", api.probeNetwork)
	group.GET("/cleannetworklist", api.cleanNetworklist)
	group.GET("/getnetworklist", api.getNetworklist)
	group.GET("/wakeLan", api.wakeLan)
	group.GET("/operstar", api.operStar)
	group.GET("/opencard", api.openCard)
	group.GET("/getselectnetcard", api.getSelectNetCard)
	group.GET("/pingpc", api.pingPC)
	group.GET("/editpcinfo", api.editPCInfo)
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
}

// ///////////////////////////////////////////////////////////////////
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
		token, err := TokenTMObj().GenToken(minute)
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

func (a *Web) CheckToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !strings.Contains(c.FullPath(), "/api/") {
			c.Next()
			return
		}

		token, err := c.Cookie("token")
		if err != nil || token == "" {
			c.JSON(200, gin.H{
				"err":   "token 无效",
				"infos": "",
			})
			c.Abort()
			return
		}

		if TokenTMObj().VerifyToken(token) {
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

	//设置唤醒接口
	a.SetWakeAPI(r)

	//设置远程接口
	a.SetRemoteAPI(r)

	//设置系统接口
	a.SetSystemAPI(r)

	//设置文件传输接口
	a.SetFileAPI(r)

	// 启动服务
	r.Run(port)
}
