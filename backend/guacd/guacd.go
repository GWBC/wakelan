package guacd

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"net"

	"github.com/gorilla/websocket"
)

var t2s []string

func init() {
	t2s = make([]string, 0)
	t2s = append(t2s, "rdp", "vnc", "ssh", "telnet")
}

type GuacdCtrl struct {
	wsConn *websocket.Conn
	info   *DstInfo
	args   []string
	rwBuf  *bufio.ReadWriter
	id     string
	params map[string]string
}

func (g *GuacdCtrl) Start(wsConn *websocket.Conn, info DstInfo) error {
	dstConn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", info.GuacdSvr.Host, info.GuacdSvr.Port))
	if err != nil {
		return err
	}

	defer dstConn.Close()

	g.wsConn = wsConn
	g.info = &info
	g.rwBuf = bufio.NewReadWriter(bufio.NewReader(dstConn), bufio.NewWriter(dstConn))

	g.params = make(map[string]string)
	g.params["hostname"] = g.info.Remote.Host
	g.params["port"] = strconv.Itoa(g.info.Remote.Port)
	g.params["username"] = g.info.Remote.User
	g.params["password"] = g.info.Remote.Pwd
	g.params["ignore-cert"] = fmt.Sprintf("%t", g.info.Remote.IgnoreCert)

	g.params["enable-sftp"] = fmt.Sprintf("%t", g.info.Sftp.Enable)
	g.params["sftp-hostname"] = g.info.Sftp.Host
	g.params["sftp-port"] = strconv.Itoa(g.info.Sftp.Port)
	g.params["sftp-username"] = g.info.Sftp.User
	g.params["sftp-password"] = g.info.Sftp.Pwd
	g.params["sftp-root-directory"] = g.info.Sftp.RootPath
	g.params["sftp-server-alive-interval"] = strconv.Itoa(g.info.Sftp.Keepalive)
	g.params["sftp-disable-upload"] = fmt.Sprintf("%t", !g.info.Sftp.Up)
	g.params["sftp-disable-download"] = fmt.Sprintf("%t", !g.info.Sftp.Down)

	g.params["font-size"] = "6"
	g.params["color-scheme"] = "white-black"     //白底黑字：black-white 黑底灰字：gray-black 黑底绿字：green-black 黑底白字：white-black
	g.params["terminal-type"] = "xterm-256color" //ansi linux vt100 vt220 xterm xterm-256color

	//telnel删除键为8
	if g.info.Remote.Type == 3 {
		g.params["backspace"] = "8"
	}

	g.login()
	g.tunnel()

	return nil
}

func (g *GuacdCtrl) login() error {
	if g.info.Remote.Type >= len(t2s) {
		return errors.New("proto type error")
	}

	//6.select,3.vnc;
	err := g.send("select", t2s[g.info.Remote.Type])
	if err != nil {
		return err
	}

	//4.args,13.VERSION_1_1_0,8.hostname,4.port,8.password,13.swap-red-blue,9.read-only;
	instruct := ""
	instruct, g.args, err = g.recv()
	if err != nil {
		return err
	}

	fmt.Println(g.args)

	if !strings.EqualFold("args", instruct) {
		return errors.New("proto error")
	}

	args := []string{}
	for _, arg := range g.args {
		v, ok := g.params[arg]
		if ok {
			args = append(args, v)
		} else {
			args = append(args, "")
		}
	}

	//4.size,4.1024,3.768,2.96;
	err = g.send("size", g.info.Remote.Width, g.info.Remote.Height, g.info.Remote.DPI)
	if err != nil {
		return err
	}

	//5.audio,9.audio/ogg;
	err = g.send("audio", g.info.Remote.AudioInfo...)
	if err != nil {
		return err
	}

	//5.video;
	err = g.send("video", g.info.Remote.VideoInfo...)
	if err != nil {
		return err
	}

	//5.image,9.image/png,10.image/jpeg;
	err = g.send("image", g.info.Remote.ImageInfo...)
	if err != nil {
		return err
	}

	//8.timezone,16.America/New_York;
	err = g.send("timezone", g.info.Remote.TimeZone)
	if err != nil {
		return err
	}

	//7.connect,13.VERSION_1_1_0,9.localhost,4.5900,0.,0.,0.;
	err = g.send("connect", args...)
	if err != nil {
		return err
	}

	//5.ready,37.$260d01da-779b-4ee9-afc1-c16bae885cc7;
	ready, res, err := g.recv()
	if err != nil {
		return err
	}

	if !strings.EqualFold(ready, "ready") || len(res) == 0 {
		return errors.New("conn error")
	}

	g.id = res[0]

	return nil
}

func (g *GuacdCtrl) tunnel() {
	go func() {
		cache := bytes.NewBuffer(nil)

		for {
			data, err := g.rwBuf.ReadBytes('.')
			if err != nil {
				break
			}

			n, err := strconv.Atoi(string(data[0 : len(data)-1]))
			if err != nil {
				break
			}

			cache.Write(data)

			_, err = io.CopyN(cache, g.rwBuf, int64(n))
			if err != nil {
				break
			}

			delim, err := g.rwBuf.ReadByte()
			if err != nil {
				break
			}

			cache.WriteByte(delim)

			if delim != ';' {
				continue
			}

			tData := cache.Bytes()
			cache.Reset()

			if bytes.HasPrefix(tData, []byte("0.,")) {
				continue
			}

			err = g.wsConn.WriteMessage(websocket.TextMessage, tData)
			if err != nil {
				break
			}
		}

		g.wsConn.Close()
	}()

	for {
		_, data, err := g.wsConn.ReadMessage()
		if err != nil {
			break
		}

		if bytes.HasPrefix(data, []byte("0.,")) {
			continue
		}

		_, err = g.rwBuf.Write(data)
		if err != nil {
			break
		}

		g.rwBuf.Flush()
	}

	g.wsConn.Close()
}

func (g *GuacdCtrl) send(instruct string, args ...string) error {
	buf := bytes.NewBuffer(nil)
	buf.WriteString(strconv.Itoa(len(instruct)))
	buf.WriteString(".")
	buf.WriteString(instruct)
	buf.WriteString(",")

	for _, arg := range args {
		buf.WriteString(strconv.Itoa(len(arg)))
		buf.WriteString(".")
		buf.WriteString(arg)
		buf.WriteString(",")
	}

	buf.Truncate(buf.Len() - 1)
	buf.WriteString(";")

	_, err := g.rwBuf.Write(buf.Bytes())
	if err != nil {
		return err
	}

	return g.rwBuf.Flush()
}

func (g *GuacdCtrl) recv() (string, []string, error) {
	instruct := ""
	args := []string{}

	cmd, err := g.rwBuf.ReadString(';')
	if err != nil {
		return instruct, args, err
	}

	subCmd := strings.Split(cmd, ",")
	if len(subCmd) == 0 {
		return instruct, args, nil
	}

	instruct = strings.Split(subCmd[0], ".")[1]

	for i := 1; i < len(subCmd); i++ {
		args = append(args, strings.Split(subCmd[i], ".")[1])
	}

	if len(args) != 0 {
		i := len(args) - 1
		args[i] = args[i][:len(args[i])-2]
	}

	return instruct, args, nil
}
