package appbase

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/robert-pkg/micro-go/log"
)

// Application 应用程序
type Application interface {
	Init()
	Run()
	OnQuit()
}

// WaitForQuit .
func WaitForQuit(app Application) {

	c := make(chan os.Signal, 1)
	signal.Notify(c)
	for {
		s := <-c
		log.Infof("catch signal:%d", s)

		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGSTOP, syscall.SIGINT:
			app.OnQuit()
			return
		case syscall.SIGHUP:
			// TODO reload
		}
	}

}
