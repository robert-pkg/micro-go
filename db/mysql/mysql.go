package mysql

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"github.com/robert-pkg/micro-go/log"
)

var dbPingOnce sync.Once

// InitDb .
func InitDb(c *Config) (*gorm.DB, error) {

	ggorm, err := gorm.Open("mysql", c.DSN)
	if err != nil {
		log.Error("InitDb", "gorm.Open", err)
		return nil, errors.New("gorm.Open")
	}
	err = ggorm.DB().Ping()
	if err != nil {
		log.Error("gORM.DB().Ping", "gorm.Open", err)
		return nil, errors.New("gORM.DB().Ping")
	}

	ggorm.DB().SetMaxIdleConns(c.Idle)
	ggorm.DB().SetMaxOpenConns(c.Active)
	ggorm.DB().SetConnMaxLifetime(c.MaxLiefTime)

	dbPingOnce.Do(func() {
		go dbTimerPing(ggorm)
	})

	// 全局禁用表名复数
	// 如果设置为true,`User`的默认表名为`user`,使用`TableName`设置的表名不受影响
	// ggorm.SingularTable(true)

	// 打开调试SQL模式
	ggorm.LogMode(true)
	ggorm.SetLogger(&logImpl{})

	return ggorm, nil
}

func dbTimerPing(ggorm *gorm.DB) {
	tick := time.NewTicker(time.Second * 300)

	for {
		select {
		case <-tick.C:
			ggorm.DB().Ping()
		}
	}
}

type logImpl struct{}

// Print .
func (l *logImpl) Print(args ...interface{}) {
	var b strings.Builder
	for i, v := range args {
		if i != 0 {
			b.WriteString(" ")
		}

		b.WriteString(fmt.Sprint(v))

	}
	log.Info("gorm", "data", b.String())
}
