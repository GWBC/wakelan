package main

import (
	"os"
	"wakelan/backend/api"
	"wakelan/backend/network"
)

func main() {
	network.NetProtoObj().Init()
	network.PushipOBJ().Start(60)

	web := api.Web{}

	port := ":8081"
	if len(os.Args) >= 2 {
		port = ":" + os.Args[1]
	}

	web.Init(port)
}
