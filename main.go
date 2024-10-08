/*
- @Time         : 2024/10/8
- @Author       : :)
- @File        	: portRefraction
- @Version      : 1.0
- @Description  : （建议配置:	1.修改ssh端口的源访问白名单为本地环回地址; 2.SshPort不等于ssh的源端口）
*/
package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	DragonBallPort []int `yaml:"DragonBallPort"`
	SshPort        int   `yaml:"SshPort"`
}

type Shiny struct {
	order  []int
	length int
}

var (
	config Config
	shiny  Shiny
	mu     sync.Mutex
	wg     sync.WaitGroup
)

func GetLogFileName() string {
	currentTime := time.Now()
	timestamp := currentTime.Format("20060102-150405")

	//检查是否存在log目录
	currentDirname, err := os.Getwd()
	if err != nil {
		log.Println("[!] 获取当前目录失败")
		return ""
	}

	logDirname := filepath.Join(currentDirname, "log")
	if _, err := os.Stat(logDirname); os.IsNotExist(err) {
		err := os.Mkdir(logDirname, 0755)
		if err != nil {
			log.Println("[!] 创建log目录失败")
			return ""
		}
	}

	return fmt.Sprintf("./log/LOG-%s.log", timestamp)
}

func initLogger() {
	logFileName := GetLogFileName()
	if logFileName == "" {
		log.Println("[!] 初始化日志记录器失败")
		return
	}
	logFile, err := os.OpenFile(logFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Println("[!] 初始化日志记录器失败")
		return
	}
	logOutput := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(logOutput)
	log.Println("[+] 初始化日志记录器成功")
}

func analyzeConfig() {
	data, err := os.ReadFile("config.yml")
	if err != nil {
		log.Fatalf("[!] 读取config.yml文件失败, error :%v", err)
	}

	err = yaml.Unmarshal(data, &config)
	if err != nil {
		log.Fatalf("[!] 解析yaml文件失败, error :%v", err)
	}
}

func trafficForwardingHandler(conn net.Conn) bool {
	defer conn.Close()

	mu.Lock()

	if shiny.length == len(config.DragonBallPort) {
		for i := 0; i < shiny.length; i++ {
			if shiny.order[i] != config.DragonBallPort[i] {
				remoteAddr := conn.RemoteAddr().(*net.TCPAddr)
				log.Println("[!] Connected from:", remoteAddr.IP, " 其未激活ssh端口,连接失败")
				//抹除激活状态
				for i := 0; i < shiny.length; i++ {
					shiny.order[i] = 0
				}
				shiny.length = 0
				mu.Unlock()
				return false
			}
		}
	} else {
		remoteAddr := conn.RemoteAddr().(*net.TCPAddr)
		log.Println("[!] Connected from:", remoteAddr.IP, " 其未激活ssh端口,连接失败")
		//抹除激活状态
		for i := 0; i < shiny.length; i++ {
			shiny.order[i] = 0
		}
		shiny.length = 0
		mu.Unlock()
		return false
	}
	//抹除激活状态
	for i := 0; i < shiny.length; i++ {
		shiny.order[i] = 0
	}
	shiny.length = 0
	mu.Unlock()

	remoteAddr := conn.RemoteAddr().(*net.TCPAddr)
	log.Println("[!] Connected from:", remoteAddr.IP, " 其激活ssh端口,连接成功")
	serverConn, _ := net.Dial("tcp", "localhost:22")
	go func() {
		io.Copy(serverConn, conn)
	}()
	io.Copy(conn, serverConn)
	serverConn.Close()
	return true
}

func sshPortHandler() {
	defer wg.Done()
	address := "0.0.0.0" + ":" + strconv.Itoa(config.SshPort)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("[!] sshPort监听失败, error :%v", err)
	} else {
		log.Printf("[+] %d端口 监听成功", config.SshPort)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("[!] sshPort接收连接失败, error :%v", err)
		}
		go trafficForwardingHandler(conn)
	}

}

func dragonBallPortHandler() {
	for i := 0; i < len(config.DragonBallPort); i++ {
		port := config.DragonBallPort[i]
		wg.Add(1)
		go func(port int) {
			defer wg.Done()
			address := "0.0.0.0" + ":" + strconv.Itoa(port)
			listener, err := net.Listen("tcp", address)
			if err != nil {
				log.Printf("[!] %d端口 监听失败, error :%v", port, err)
				return
			} else {
				log.Printf("[+] %d端口 监听成功", port)
			}

			for {
				_, err := listener.Accept()
				if err != nil {
					log.Printf("[!] %d端口 接收连接失败, error :%v", port, err)
				}
				mu.Lock()
				if shiny.length >= len(config.DragonBallPort) {
					//这里抹不抹除激活状态都一样
					for i := 0; i < shiny.length; i++ {
						shiny.order[i] = 0
					}
					shiny.length = 0
				}
				shiny.order[shiny.length] = port
				shiny.length++

				mu.Unlock()
			}
		}(port)
	}
}

func main() {
	initLogger()
	analyzeConfig()

	shiny.order = make([]int, len(config.DragonBallPort))
	shiny.length = 0

	wg.Add(1)
	go sshPortHandler()
	dragonBallPortHandler()
	wg.Wait()
}
