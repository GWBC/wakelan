package comm

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"time"
)

type TokenInfo struct {
	Deadline int64
}

type TokenManager struct {
	key []byte
	iv  []byte
}

func (tm *TokenManager) Init(key []byte, iv []byte) bool {
	if len(iv) < 16 {
		return false
	}

	tm.ChangeKey(key)
	tm.iv = iv[0:16]

	return false
}

func (tm *TokenManager) ChangeKey(key []byte) {
	if len(key) < 32 {
		key = append(key, []byte(GenUniqueKey())...)
	}

	tm.key = key[0:32]
}

func (tm *TokenManager) GenToken(minute int) (string, error) {
	info := TokenInfo{time.Now().Add(time.Duration(minute * int(time.Minute))).Unix()}

	data, err := json.Marshal(info)
	if err != nil {
		return "", err
	}

	data, err = AES_CBC_Seal(data, tm.key, tm.iv, Zero)
	if err != nil {
		return "", err
	}

	return base64.URLEncoding.EncodeToString(data), nil
}

func (tm *TokenManager) VerifyToken(token string) bool {
	data, err := base64.URLEncoding.DecodeString(token)
	if err != nil {
		return false
	}

	data, err = AES_CBC_Open(data, tm.key, tm.iv)
	if err != nil {
		return false
	}

	strData := strings.TrimRight(string(data), "\x00")

	info := TokenInfo{}
	err = json.Unmarshal([]byte(strData), &info)
	if err != nil {
		return false
	}

	if info.Deadline <= time.Now().Unix() {
		return false
	}

	return true
}
