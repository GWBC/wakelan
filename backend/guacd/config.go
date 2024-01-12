package guacd

type ConnInfo struct {
	Host string
	Port int16
}

type DstInfo struct {
	GuacdSvr struct {
		Host string `json:"host"`
		Port int    `json:"port"`
	}
	Remote struct {
		Host       string `json:"host"`
		Port       int    `json:"port"`
		User       string `json:"user"`
		Pwd        string `json:"pwd"`
		Type       int    `json:"type"` //0:rdp 1:vnc 2:ssh 3:telnel
		IgnoreCert bool
		Width      string `json:"width"`
		Height     string `json:"height"`
		DPI        string
		AudioInfo  []string
		ImageInfo  []string
		VideoInfo  []string
		TimeZone   string
	} `json:"remote"`
	Sftp struct {
		Enable    bool   `json:"enable"`
		Up        bool   `json:"up"`
		Down      bool   `json:"down"`
		RootPath  string `json:"rootPath"`
		Keepalive int    `json:"keepalive"`
		Host      string `json:"host"`
		Port      int    `json:"port"`
		User      string `json:"user"`
		Pwd       string `json:"pwd"`
	} `json:"sftp"`
}

func (dst *DstInfo) SetGuacdServer(t int, host string, port int16) {
	dst.Remote.Type = t
	dst.GuacdSvr.Host = host
	dst.GuacdSvr.Port = int(port)
	if dst.GuacdSvr.Port == 0 {
		dst.GuacdSvr.Port = 4822
	}

	dst.Remote.IgnoreCert = true

	if len(dst.Remote.Width) == 0 {
		dst.Remote.Width = "1920"
	}

	if len(dst.Remote.Height) == 0 {
		dst.Remote.Height = "1080"
	}

	dst.Remote.DPI = "300"

	dst.Remote.AudioInfo = []string{"audio/L16", "rate=44100", "channels=2"}
	dst.Remote.ImageInfo = []string{"image/png", "image/jpeg"}
	dst.Remote.TimeZone = "Asia/Shanghai"
}
