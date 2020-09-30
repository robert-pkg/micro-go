package utils

import (
	"math/rand"
	"net"
	"syscall"
	"time"

	"github.com/robert-pkg/micro-go/log"
)

// LocalIP 返回本机的内网IP地址
func LocalInternalIP() ([]string, error) {

	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}

	var resultList []string
	for _, addr := range addrs {

		////判断是否正确获取到IP
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && !ipnet.IP.IsLinkLocalMulticast() && !ipnet.IP.IsLinkLocalUnicast() {

			if ipnet.IP.To4() != nil {
				//fmt.Println(ipnet.IP.String())
				resultList = append(resultList, ipnet.IP.String())
			}
		}

	}

	return resultList, nil
}

// RandPort 生成随机端口号 (10240, 65535), 保证可用
func RandPort() int {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	for {
		port := r.Intn(65535)
		if port <= 10240 {
			continue
		} else {
			sa := new(syscall.SockaddrInet4)
			sa.Port = port
			sa.Addr = [4]byte{0x7f, 0x00, 0x00, 0x01}
			if s, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, syscall.IPPROTO_TCP); nil != err {
				log.Warn("warn", "warn", err)
			} else {
				if err := syscall.Bind(s, sa); err != nil {
					log.Info("端口被占用", "err", err)
				} else {
					syscall.Close(s)
					//log.Infof("端口%d可用", sa.Port)
					return sa.Port
				}
			}
		}
	}
}
