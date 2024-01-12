package main

import (
	"wakelan/backend/api"
	"wakelan/backend/network"
)

func main() {
	network.NetProtoObj().Init()
	network.PushipOBJ().Start(60)

	web := api.Web{}
	web.Init()
}
