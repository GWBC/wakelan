package api

import (
	"encoding/base64"
	"encoding/json"
	"math/rand"
	"strings"
	"sync"
	"time"
	"wakelan/backend/comm"
	"wakelan/backend/db"
)

type TokenInfo struct {
	Deadline int64
}

type TokenManager struct {
	key []byte
	iv  []byte
}

func (tm *TokenManager) Init() {
	tm.ChangeKey()
	tm.iv = []byte("941CC9A12E47BF17")
}

func (tm *TokenManager) ChangeKey() {
	cfg := db.DBOperObj().GetConfig()
	if len(cfg.Secret) < 32 {
		cfg.Secret += comm.GenUniqueKey()
	}

	tm.key = []byte(cfg.Secret)[0:32]
}

func (tm *TokenManager) GenToken(minute int) (string, error) {
	tokenBytes := make([]byte, 32)
	_, err := rand.Read(tokenBytes)
	if err != nil {
		return "", err
	}

	info := TokenInfo{time.Now().Add(time.Duration(minute * int(time.Minute))).Unix()}

	data, err := json.Marshal(info)
	if err != nil {
		return "", err
	}

	data, err = comm.AES_CBC_Seal(data, tm.key, tm.iv, comm.Zero)
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

	data, err = comm.AES_CBC_Open(data, tm.key, tm.iv)
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

var tokenManagerOnce sync.Once
var tokenManagerObj *TokenManager

func TokenTMObj() *TokenManager {
	tokenManagerOnce.Do(func() {
		tokenManagerObj = &TokenManager{}
		tokenManagerObj.Init()
	})

	return tokenManagerObj
}
