package trojan

import (
	"encoding/base64"
	"fmt"
	"trojan/core"
	"trojan/util"
)

var clientPath = "/root/config.json"

// GenClientJson 生成客户端json
func GenClientJson() {
	fmt.Println()
	var user core.User
	domain, port := GetDomainAndPort()
	mysql := core.GetMysql()
	userList, err := mysql.GetData()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	if len(userList) == 1 {
		user = *userList[0]
	} else {
		UserList()
		choice := util.LoopInput("Please select the user number to generate the configuration file: ", userList, true)
		if choice < 0 {
			return
		}
		user = *userList[choice-1]
	}
	pass, err := base64.StdEncoding.DecodeString(user.Password)
	if err != nil {
		fmt.Println(util.Red("Base64 decoding failed: " + err.Error()))
		return
	}
	if !core.WriteClient(port, string(pass), domain, clientPath) {
		fmt.Println(util.Red("Failed to generate configuration file!"))
	} else {
		fmt.Println("Successfully generated configuration file: " + util.Green(clientPath))
	}
}
