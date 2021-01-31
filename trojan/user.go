package trojan

import (
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"trojan/core"
	"trojan/util"
)

// UserMenu 用户管理菜单
func UserMenu() {
	fmt.Println()
	menu := []string{"Add User", "Delete User", "Limit Traffic", "Clear Traffic", "Set Expiration", "Cancel Expiration""}
	switch util.LoopInput("Please choose: ", menu, false) {
	case 1:
		AddUser()
	case 2:
		DelUser()
	case 3:
		SetUserQuota()
	case 4:
		CleanData()
	case 5:
		SetupExpire()
	case 6:
		CancelExpire()
	}
}

// AddUser 添加用户
func AddUser() {
	randomUser := util.RandString(4)
	randomPass := util.RandString(8)
	inputUser := util.Input(fmt.Sprintf("Generate a random username: %s, Press enter, otherwise enter a custom username: ", randomUser), randomUser)
	if inputUser == "admin" {
		fmt.Println(util.Yellow("Cannot create a new user with the user name 'admin'!"))
		return
	}
	mysql := core.GetMysql()
	if user := mysql.GetUserByName(inputUser); user != nil {
		fmt.Println(util.Yellow("Existing username: " + inputUser + " User!"))
		return
	}
	inputPass := util.Input(fmt.Sprintf("Generate random password: %s, Press enter, otherwise enter a custom password: ", randomPass), randomPass)
	base64Pass := base64.StdEncoding.EncodeToString([]byte(inputPass))
	if user := mysql.GetUserByPass(base64Pass); user != nil {
		fmt.Println(util.Yellow("Existing password: " + inputPass + " User!"))
		return
	}
	if mysql.CreateUser(inputUser, base64Pass, inputPass) == nil {
		fmt.Println("User added successfully!")
	}
}

// DelUser 删除用户
func DelUser() {
	userList := UserList()
	mysql := core.GetMysql()
	choice := util.LoopInput("Please select the user number to be deleted: ", userList, true)
	if choice == -1 {
		return
	}
	if mysql.DeleteUser(userList[choice-1].ID) == nil {
		fmt.Println("User deleted successfully!")
	}
}

// SetUserQuota 限制用户流量
func SetUserQuota() {
	var (
		limit int
		err   error
	)
	userList := UserList()
	mysql := core.GetMysql()
	choice := util.LoopInput("Please select the number of the user whose traffic is to be restricted: ", userList, true)
	if choice == -1 {
		return
	}
	for {
		quota := util.Input("Please enter user"+userList[choice-1].Username+"Restricted traffic size (unit byte)", "")
		limit, err = strconv.Atoi(quota)
		if err != nil {
			fmt.Printf("%s Not a number, please re-enter!\n", quota)
		} else {
			break
		}
	}
	if mysql.SetQuota(userList[choice-1].ID, limit) == nil {
		fmt.Println("User setup successfully" + userList[choice-1].Username + "Limit traffic" + util.Bytefmt(uint64(limit)))
	}
}

// CleanData 清空用户流量
func CleanData() {
	userList := UserList()
	mysql := core.GetMysql()
	choice := util.LoopInput("Please select the user number to clear traffic: ", userList, true)
	if choice == -1 {
		return
	}
	if mysql.CleanData(userList[choice-1].ID) == nil {
		fmt.Println("Cleared traffic successfully!")
	}
}

// CancelExpire 取消限期
func CancelExpire() {
	userList := UserList()
	mysql := core.GetMysql()
	choice := util.LoopInput("Please select the user number to cancel the expiration: ", userList, true)
	if choice == -1 {
		return
	}
	if userList[choice-1].UseDays == 0 {
		fmt.Println(util.Yellow("The selected user has not set an expiration!"))
		return
	}
	if mysql.CancelExpire(userList[choice-1].ID) == nil {
		fmt.Println("Removed expiration successfully!")
	}
}

// SetupExpire 设置限期
func SetupExpire() {
	userList := UserList()
	mysql := core.GetMysql()
	choice := util.LoopInput("Please select the user number to set a deadline: ", userList, true)
	if choice == -1 {
		return
	}
	useDayStr := util.Input("Please enter the number of days: ", "")
	if useDayStr == "" {
		return
	} else if strings.Contains(useDayStr, "-") {
		fmt.Println(util.Yellow("The number of days. Cannot be negative"))
		return
	} else if !util.IsInteger(useDayStr) {
		fmt.Println(util.Yellow("Input is non-integer!"))
		return
	}
	useDays, _ := strconv.Atoi(useDayStr)
	if mysql.SetExpire(userList[choice-1].ID, uint(useDays)) == nil {
		fmt.Println("Set expiration successfully!")
	}
}

// CleanDataByName 清空指定用户流量
func CleanDataByName(usernames []string) {
	mysql := core.GetMysql()
	if err := mysql.CleanDataByName(usernames); err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println("Cleared traffic successfully!")
	}
}

// UserList 获取用户列表并打印显示
func UserList(ids ...string) []*core.User {
	mysql := core.GetMysql()
	userList, err := mysql.GetData(ids...)
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}
	domain, port := GetDomainAndPort()
	for i, k := range userList {
		pass, err := base64.StdEncoding.DecodeString(k.Password)
		if err != nil {
			pass = []byte("")
		}
		fmt.Printf("%d.\n", i+1)
		fmt.Println("username: " + k.Username)
		fmt.Println("password: " + string(pass))
		fmt.Println("Upload traffic: " + util.Cyan(util.Bytefmt(k.Upload)))
		fmt.Println("Download traffic: " + util.Cyan(util.Bytefmt(k.Download)))
		if k.Quota < 0 {
			fmt.Println("Flow limit: " + util.Cyan("Unlimited"))
		} else {
			fmt.Println("Flow limit: " + util.Cyan(util.Bytefmt(uint64(k.Quota))))
		}
		if k.UseDays == 0 {
			fmt.Println("Date of Expiry: " + util.Cyan("Unlimited"))
		} else {
			fmt.Println("Date of Expiry: " + util.Cyan(k.ExpiryDate))
		}
		fmt.Println("Share link: " + util.Green(fmt.Sprintf("trojan://%s@%s:%d", string(pass), domain, port)))
		fmt.Println()
	}
	return userList
}
