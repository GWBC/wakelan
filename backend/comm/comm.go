package comm

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/wxpusher/wxpusher-sdk-go"
	"github.com/wxpusher/wxpusher-sdk-go/model"
)

const TimeFormat = "2006-01-02 15:04:05"

func inc(ip net.IP) net.IP {
	tip := make(net.IP, len(ip))
	copy(tip, ip)

	for i := len(tip) - 1; i >= 0; i-- {
		tip[i]++
		if tip[i] > 0 {
			break
		}
	}

	return tip
}

// 根据mask生成所有ip
func MakeIPs(ipnet net.IPNet) []net.IP {
	ips := []net.IP{}

	for ip := ipnet.IP.Mask(ipnet.Mask); ipnet.Contains(ip); ip = inc(ip) {
		ips = append(ips, ip)
	}

	return ips
}

// 获取当前程序路径
func Pwd() string {
	return filepath.Dir(os.Args[0])
}

// 唤醒机器
func WakeLan(mac string) error {
	targetMac, err := net.ParseMAC(mac)
	if err != nil {
		return err
	}

	buf := bytes.NewBuffer(nil)
	for i := 0; i < 6; i++ {
		buf.WriteByte(0xff)
	}

	for i := 0; i < 16; i++ {
		buf.Write(targetMac)
	}

	sendFun := func(port int) error {
		conn, err := net.Dial("udp", fmt.Sprintf("255.255.255.255:%d", port))
		if err != nil {
			return err
		}

		defer conn.Close()

		_, err = conn.Write(buf.Bytes())

		return err
	}

	ports := []int{7, 9}

	for _, port := range ports {
		for j := 0; j < 6; j++ {
			sendFun(port)
		}
	}

	return nil
}

func GetHttp(url string, headers *map[string]string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return []byte{}, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36 Edg/121.0.0.0")

	if headers != nil {
		for k, v := range *headers {
			req.Header.Set(k, v)
		}
	}

	client := http.Client{
		Timeout: 6 * time.Second,
	}
	rsp, err := client.Do(req)
	if err != nil {
		return []byte{}, err
	}

	defer rsp.Body.Close()

	if rsp.StatusCode != http.StatusOK {
		return []byte{}, fmt.Errorf("http status code:%d", rsp.StatusCode)
	}

	return io.ReadAll(rsp.Body)
}

func GetGlobalIP(url string) string {
	urls := strings.Split(url, ";")

	for _, v := range urls {
		if len(v) == 0 {
			continue
		}
		buf, err := GetHttp(v, nil)
		if err == nil {
			pattern := `(\d+\.\d+\.\d+\.\d+)`
			re := regexp.MustCompile(pattern)
			ip := re.FindString(string(buf))
			if len(ip) == 0 {
				continue
			}

			return ip
		}
	}

	return ""
}

func AYFFPushMsg(msg string, token string) error {
	apiUrl := fmt.Sprintf("https://iyuu.cn/%s.send?text=%s", token, url.QueryEscape(msg))
	_, err := GetHttp(apiUrl, nil)
	if err != nil {
		return err
	}

	return nil
}

func WXPusherMsg(msg string, appToken string, topicId int) error {
	msgModel := model.NewMessage(appToken).SetContent(msg).AddTopicId(topicId)
	res, err := wxpusher.SendMessage(msgModel)

	if err != nil {
		return err
	}

	for _, resp := range res {
		if resp.Code != 1000 {
			err = errors.New(resp.Status)
		}
	}

	return err
}

func IpLess(a, b net.IP) bool {
	if len(b) == 0 {
		return false
	}

	for i := 0; i < len(a); i++ {
		if a[i] < b[i] {
			return true
		} else if a[i] > b[i] {
			return false
		}
	}
	return false
}

type AESFillType int

const (
	PKCS5_PKCS7 AESFillType = iota
	ANSI_X_923
	ISO_10126
	Zero
)

func Fill(data []byte, blockSize int, t AESFillType) []byte {
	l := len(data)
	m := l % blockSize
	if m == 0 {
		return data
	}

	diff := blockSize - m

	switch t {
	case PKCS5_PKCS7:
		data = append(data, bytes.Repeat([]byte{byte(diff)}, diff)...)
	case ANSI_X_923, ISO_10126:
		data = append(data, bytes.Repeat([]byte{0}, diff)...)
		data[len(data)-1] = byte(diff)
	case Zero:
		data = append(data, bytes.Repeat([]byte{0}, diff)...)
	}

	return data
}

// 加密
func AES_CBC_Seal(data []byte, key []byte, iv []byte, t AESFillType) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return []byte{}, err
	}

	ciperObj := cipher.NewCBCEncrypter(block, iv)

	tData := Fill(data, block.BlockSize(), t)
	dst := make([]byte, len(tData))
	ciperObj.CryptBlocks(dst, tData)

	return dst, nil
}

// 解密
func AES_CBC_Open(data []byte, key []byte, iv []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return []byte{}, err
	}

	ciperObj := cipher.NewCBCDecrypter(block, iv)

	dst := make([]byte, len(data))
	ciperObj.CryptBlocks(dst, data)

	return dst, nil
}

type NetInfo struct {
	IP  []string
	MAC string
}

// 获取本机IP和MAC
func GetIPAndMac() ([]NetInfo, error) {
	netInfos := []NetInfo{}

	interfaces, err := net.Interfaces()
	if err != nil {
		return netInfos, err
	}

	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp != 0 && //启动
			iface.Flags&net.FlagLoopback == 0 && //非loopback
			iface.Flags&net.FlagPointToPoint == 0 { //非虚拟
			netInfo := NetInfo{}

			addrs, err := iface.Addrs()
			if err != nil {
				continue
			}

			for _, addr := range addrs {
				netInfo.IP = append(netInfo.IP, addr.String())
			}

			hwAddr := iface.HardwareAddr
			if hwAddr == nil {
				continue
			}

			netInfo.MAC = hwAddr.String()
			netInfos = append(netInfos, netInfo)
		}
	}

	return netInfos, nil
}

// 生成当前机器的唯一key
func GenUniqueKey() string {
	key := "A4540D1E84B1C931013FE8968B61834F"

	netInfo, _ := GetIPAndMac()
	if len(netInfo) != 0 {
		sum := md5.Sum([]byte(netInfo[0].MAC))
		key = hex.EncodeToString(sum[:])
	}

	return key
}

// 生成随机key
func GenRandKey() string {
	bytes := make([]byte, 32)
	_, err := rand.Read(bytes)
	if err != nil {
		return ""
	}

	sum := md5.Sum(bytes)
	return hex.EncodeToString(sum[:])
}
