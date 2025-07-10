package rpcPorts

const postStart = 20050

var mapPorts = map[string]int{
	"httpProxy": 20000,
	"conn":      postStart + 0, "apiserver": postStart + 1, "loginV2": postStart + 2,
	"online": postStart + 3, "robot": postStart + 4, "monitor": postStart + 5,
	"smsproxy": postStart + 6, "payAgent": postStart + 7, "money": postStart + 8,

	"risk": postStart + 9, "lockerSvr": postStart + 10, "idfactory": postStart + 11,
	"boot": postStart + 12, "opBackstage": postStart + 13, "csvExport": postStart + 14,
	"topOpAdmin": postStart + 15, "topUrlServer": postStart + 16, "provider": postStart + 17,
	"appserver": postStart + 18, "userV2": postStart + 19, "promoItem": postStart + 20, "record": postStart + 21, "jobSched": postStart + 22,
	"reward": postStart + 23, "notice": postStart + 24, "filterWord": postStart + 25, "extReport": postStart + 26, "dataWatch": postStart + 27,

	"activity": postStart + 28, "task": postStart + 29, "userIDRisk": postStart + 30, "userwealth": postStart + 31, "vip": postStart + 32,
	"gchat": postStart + 33, "extGame": postStart + 34, "eGame": postStart + 35, "flyGame": postStart + 36, "quickG": postStart + 37,
	"fishG": postStart + 38, "fruitG": postStart + 39, "minesG": postStart + 40, "singleG": postStart + 41,
	"userAgent": postStart + 42, "chDBProxy": postStart + 43, "offlineAgentBackstage": postStart + 44, "dataCollect": postStart + 45,
	"payV2": postStart + 46, "syncImageDb": postStart + 47, "laohuG": postStart + 48, "bjG": postStart + 49, "bullG": postStart + 50,
	"holdemG": postStart + 51, "jhG": postStart + 52, "ermjG": postStart + 53, "xlmjG": postStart + 54,
	"transfer": postStart + 55, "chatG": postStart + 56, "lotteryIssue": postStart + 57, "lottery": postStart + 58, "platformCoin": postStart + 59,
	"crm": postStart + 60, "front": postStart + 61, "opBackstageV2": postStart + 62, "tmpUser": postStart + 63, "rewardTask": postStart + 64,
	"licenseReport": postStart + 65, "common": postStart + 66, "responsible": postStart + 67, "uniVendor": postStart + 68,
}

func GetRpcPort(sname string) int {
	port, ok := mapPorts[sname]
	if !ok {
		return 0
	}
	//检查下有没有端口重复的
	for k, v := range mapPorts {
		if k != sname && v == port {
			return -1
		}
	}
	return port
}

func CheckRpcPort(sname string) int {
	port, ok := mapPorts[sname]
	if !ok {
		return 0
	}
	return port
}
