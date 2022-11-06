package main

import (
	"fmt"
	"golang.org/x/text/encoding/simplifiedchinese"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

func main() {
	fmt.Println("获取的用户列表只是登录过的,要自动登录必须在登录时勾选了记住密码")
	// 获取steam路径
	SteamPath, SteamExe := getPath()
	// 获取steam路径下已登录用户的账户
	UserList := getLoginUser(SteamPath)
	// 输出视图,返回用户选择的数据
	choose := view(UserList)
	// win10中注册表获取的路径会有这串东西,去除掉
	SteamExe = strings.Replace(SteamExe, ".FriendlyAppName", "", 1)
	fmt.Println(SteamExe)
	startSteam(UserList[choose], SteamExe)
}

// 获取steam根目录
func getPath() (string, string) {
	// 获取注册表下所有数据这个注册表下存着所有软件的根目录, 键是目录,项是 主程序的描述,比如F:\steam=>Steam
	cmd := exec.Command("cmd", "/c", "reg", "query", "HKEY_CLASSES_ROOT\\Local Settings\\Software\\Microsoft\\Windows\\Shell\\MuiCache", "/s")
	output, _ := cmd.Output()
	// 将获取的所有注册表转换成字符串
	output, _ = simplifiedchinese.GB18030.NewDecoder().Bytes(output)
	// 每一行的数据分隔成数组
	arr := strings.Split(string(output), "\r\n")
	// 遍历数组获取每一行
	for _, v := range arr {
		// 去除行数据两段的空格
		v = strings.TrimSpace(v)
		// 每行中的数据分成三项,项 数据类型 值,将其分隔成数组
		arr2 := strings.Split(v, "    ")
		// 遍历每一行的数据获得 项和值
		for i, v2 := range arr2 {
			// 如果arr[2]的值是Steam,表示找到了Steam的根目录,将其存入遍历
			if i == 2 && v2 == "Steam" {
				// 返回steam所在文件夹 和 返回steam.exe完整路径
				return filepath.Dir(arr2[0]), arr2[0]
			}
		}
	}
	return "", ""
}

// 获取config中已记住密码的用户名
func getLoginUser(path string) map[int]string {
	UserList := make(map[int]string)
	index := 1
	// steam所有记住密码的用户文件
	path = path + "/config/loginusers.vdf"
	// 读取文件的字符串
	buf, _ := os.ReadFile(path)
	// 去除开头的"users"{,否则刚开始的大括号会对正则有影响
	str := strings.Trim(string(buf), "\"users\"\r\n{")
	// 正则获取所有大括号内包含的登录用户
	reg1 := regexp.MustCompile(`"AccountName"\t\t.*`)
	arr := reg1.FindAllStringSubmatch(str, -1)
	for _, v := range arr {
		for _, v2 := range v {
			// 将获取的每个用户名存入map
			username := strings.Split(v2, "\t\t")
			UserList[index] = username[1]
			index++
		}
	}
	return UserList
}

// 输出视图
func view(UserList map[int]string) (choose int) {
	fmt.Println("请选择要登录的用户")
	for i := 1; i <= len(UserList); i++ {
		fmt.Printf("%d.%s\n", i, UserList[i])
	}
	temp := ""
	fmt.Scan(&temp)
	choose, _ = strconv.Atoi(temp)
	return
}

// 将要登录的用户名写入steam的注册表并启动steam
func startSteam(username string, steamExe string) {
	username = strings.Trim(username, `"`)
	fmt.Println(username)
	// 写入注册表到这,/f表示自动同意,如果已经有这个注册表会问是否覆盖,/f自动同意
	cmd := exec.Command("cmd", "/c", "reg", "add", "HKEY_CURRENT_USER\\SOFTWARE\\Valve\\Steam", "/v", "AutoLoginUser", "/t", "REG_SZ", "/d", username, "/f")
	// 要用run等待这次执行完成后再往后执行,否则可能会导致steam还没杀死进程就重启了
	cmd.Run()
	// 杀死steam进程
	killSteam()
	// 启动steam
	cmd1 := exec.Command("cmd", "/c", "start", steamExe)
	cmd1.Run()
}

func killSteam() {
	// 不加/f只是给steam发送终止信号,是steam自己处理,/f表示系统强制杀死进程
	cmd := exec.Command("cmd", "/c", "taskkill", "/pid", "steam.exe", "/f")
	cmd.Run()
}
