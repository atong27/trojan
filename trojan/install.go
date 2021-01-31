package trojan

import (
	"fmt"
	"github.com/gobuffalo/packr/v2"
	"net"
	"strconv"
	"strings"
	"time"
	"trojan/core"
	"trojan/util"
)

var (
	dockerInstallUrl1 = "https://get.docker.com"
	dockerInstallUrl2 = "https://git.io/docker-install"
	dbDockerRun       = "docker run --name trojan-mariadb --restart=always -p %d:3306 -v /home/mariadb:/var/lib/mysql -e MYSQL_ROOT_PASSWORD=%s -e MYSQL_ROOT_HOST=%% -e MYSQL_DATABASE=trojan -d mariadb:10.2"
)

// InstallMenu 安装目录
func InstallMenu() {
	fmt.Println()
	menu := []string{"Update Trojan", "Certificate Application", "Install MySQL"}
	switch util.LoopInput("Please choose: ", menu, true) {
	case 1:
		InstallTrojan()
	case 2:
		InstallTls()
	case 3:
		InstallMysql()
	default:
		return
	}
}

// InstallDocker 安装docker
func InstallDocker() {
	if !util.CheckCommandExists("docker") {
		util.RunWebShell(dockerInstallUrl1)
		if !util.CheckCommandExists("docker") {
			util.RunWebShell(dockerInstallUrl2)
		} else {
			util.ExecCommand("systemctl enable docker")
			util.ExecCommand("systemctl start docker")
		}
		fmt.Println()
	}
}

// InstallTrojan 安装trojan
func InstallTrojan() {
	fmt.Println()
	box := packr.New("trojan-install", "../asset")
	data, err := box.FindString("trojan-install.sh")
	if err != nil {
		fmt.Println(err)
	}
	if util.ExecCommandWithResult("systemctl list-unit-files|grep trojan.service") != "" && Type() == "trojan-go" {
		data = strings.ReplaceAll(data, "TYPE=0", "TYPE=1")
	}
	util.ExecCommand(data)
	util.OpenPort(443)
	util.ExecCommand("systemctl restart trojan")
	util.ExecCommand("systemctl enable trojan")
}

// InstallTls 安装证书
func InstallTls() {
	domain := ""
	fmt.Println()
	choice := util.LoopInput("Please choose the certificate method: ", []string{"Let's Encrypt 证书", "Custom certificate path"}, true)
	if choice < 0 {
		return
	} else if choice == 1 {
		localIP := util.GetLocalIP()
		fmt.Printf("Ip: %s\n", localIP)
		for {
			domain = util.Input("Please enter the domain name for the certificate: ", "")
			ipList, err := net.LookupIP(domain)
			fmt.Printf("%s resolved ip: %v\n", domain, ipList)
			if err != nil {
				fmt.Println(err)
				fmt.Println("The domain name is wrong, please re-enter")
				continue
			}
			checkIp := false
			for _, ip := range ipList {
				if localIP == ip.String() {
					checkIp = true
				}
			}
			if checkIp {
				break
			} else {
				fmt.Println("The entered domain name is inconsistent with the local ip, please re-enter!")
			}
		}
		util.InstallPack("socat")
		if !util.IsExists("/root/.acme.sh/acme.sh") {
			util.RunWebShell("https://get.acme.sh")
		}
		util.ExecCommand("systemctl stop trojan-web")
		util.OpenPort(80)
		util.ExecCommand(fmt.Sprintf("bash /root/.acme.sh/acme.sh --issue -d %s --debug --standalone --keylength ec-256", domain))
		crtFile := "/root/.acme.sh/" + domain + "_ecc" + "/fullchain.cer"
		keyFile := "/root/.acme.sh/" + domain + "_ecc" + "/" + domain + ".key"
		core.WriteTls(crtFile, keyFile, domain)
	} else if choice == 2 {
		crtFile := util.Input("Please enter the cert file path of the certificate: ", "")
		keyFile := util.Input("Please enter the key file path of the certificate: ", "")
		if !util.IsExists(crtFile) || !util.IsExists(keyFile) {
			fmt.Println("The entered cert or key file does not exist!!")
		} else {
			domain = util.Input("Please enter the domain name corresponding to this certificate: ", "")
			if domain == "" {
				fmt.Println("The input domain name is empty!")
				return
			}
			core.WriteTls(crtFile, keyFile, domain)
		}
	}
	Restart()
	util.ExecCommand("systemctl restart trojan-web")
	fmt.Println()
}

// InstallMysql 安装mysql
func InstallMysql() {
	var (
		mysql  core.Mysql
		choice int
	)
	fmt.Println()
	if util.IsExists("/.dockerenv") {
		choice = 2
	} else {
		choice = util.LoopInput("Please choose: ", []string{"Install docker version of mysql(mariadb)", "Enter a custom mysql connection"}, true)
	}
	if choice < 0 {
		return
	} else if choice == 1 {
		mysql = core.Mysql{ServerAddr: "127.0.0.1", ServerPort: util.RandomPort(), Password: util.RandString(5), Username: "root", Database: "trojan"}
		InstallDocker()
		fmt.Println(fmt.Sprintf(dbDockerRun, mysql.ServerPort, mysql.Password))
		if util.CheckCommandExists("setenforce") {
			util.ExecCommand("setenforce 0")
		}
		util.OpenPort(mysql.ServerPort)
		util.ExecCommand(fmt.Sprintf(dbDockerRun, mysql.ServerPort, mysql.Password))
		db := mysql.GetDB()
		for {
			fmt.Printf("%s mariadb is starting, please wait...\n", time.Now().Format("2006-01-02 15:04:05"))
			err := db.Ping()
			if err == nil {
				db.Close()
				break
			} else {
				time.Sleep(2 * time.Second)
			}
		}
		fmt.Println("mariadb started successfully!")
	} else if choice == 2 {
		mysql = core.Mysql{}
		for {
			for {
				mysqlUrl := util.Input("Please enter the mysql connection address (format: host:port), the default connection address is 127.0.0.1:3306, press enter, otherwise enter a custom connection address: ",
					"127.0.0.1:3306")
				urlInfo := strings.Split(mysqlUrl, ":")
				if len(urlInfo) != 2 {
					fmt.Printf("The input %s does not match the matching format (host:port)\n", mysqlUrl)
					continue
				}
				port, err := strconv.Atoi(urlInfo[1])
				if err != nil {
					fmt.Printf("%s is not a number\n", urlInfo[1])
					continue
				}
				mysql.ServerAddr, mysql.ServerPort = urlInfo[0], port
				break
			}
			mysql.Username = util.Input("Please enter the username of mysql (Enter to use root): ", "root")
			mysql.Password = util.Input(fmt.Sprintf("Please enter the password of the mysql %s user: ", mysql.Username), "")
			db := mysql.GetDB()
			if db != nil && db.Ping() == nil {
				mysql.Database = util.Input("Please enter the name of the database used (it can be created automatically if it does not exist, press Enter to use trojan): ", "trojan")
				db.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s;", mysql.Database))
				break
			} else {
				fmt.Println("Failed to connect to mysql, please re-enter")
			}
		}
	}
	mysql.CreateTable()
	core.WriteMysql(&mysql)
	if userList, _ := mysql.GetData(); len(userList) == 0 {
		AddUser()
	}
	Restart()
	fmt.Println()
}
