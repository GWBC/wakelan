package db

func GetNetworkCard() string {
	dbObj := DBOperObj().GetDB()

	info := &GlobalInfo{}
	result := dbObj.Select("netcard").Find(info)
	if result.Error == nil {
		if len(info.NetCard) != 0 {
			return info.NetCard
		}
	}

	return ""
}
