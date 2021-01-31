package trojan

import (
	"crypto/sha256"
	"fmt"
	"trojan/core"
	"trojan/util"
)

// WebMenu web管理菜单
func WebMenu() {
	fmt.Println()
	menu := []string{"Reset the web administrator password", "Modify the displayed domain name (not to apply for a certificate)"}
	switch util.LoopInput("Please choose: ", menu, true) {
	case 1:
		ResetAdminPass()
	case 2:
		SetDomain("")
	}
}

// ResetAdminPass 重置管理员密码
func ResetAdminPass() {
	inputPass := util.Input("Please enter the admin user password: ", "")
	if inputPass == "" {
		fmt.Println("Undo changes!")
	} else {
		encryPass := sha256.Sum224([]byte(inputPass))
		err := core.SetValue("admin_pass", fmt.Sprintf("%x", encryPass))
		if err == nil {
			fmt.Println(util.Green("Successfully reset admin password!"))
		} else {
			fmt.Println(err)
		}
	}
}

// SetDomain 设置显示的域名
func SetDomain(domain string) {
	if domain == "" {
		domain = util.Input("Please enter the domain name address to be displayed: ", "")
	}
	if domain == "" {
		fmt.Println("Undo changes!")
	} else {
		core.WriteDomain(domain)
		Restart()
		fmt.Println("Modified domain successfully!")
	}
}

// GetDomainAndPort 获取域名和端口
func GetDomainAndPort() (string, int) {
	config := core.Load("")
	return config.SSl.Sni, config.LocalPort
}
