package sqlite

import (
	"context"
	"fmt"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

type Sqliter interface {
	InitSqlite()
}

type Sqlite struct {
	RootPath string
}

func NewSqlite(RootPath string) Sqliter {
	Sqlite := &Sqlite{
		RootPath: RootPath,
	}
	Sqlite.InitSqlite()
	Sqlite.IintTable()
	return Sqlite
}

func (s *Sqlite) InitSqlite() {
	dbConfig := gdb.ConfigNode{
		Link:  fmt.Sprintf("sqlite::@file(%sdb.sqlite3)", s.RootPath),
		Debug: true,
	}
	// 注册数据库配置（分组名 default）
	gdb.AddConfigNode("default", dbConfig)
	// 获取数据库实例
}

func (s *Sqlite) IintTable() {
	ctx := context.Background()
	db1 := g.DB("default")

	// 创建用户表
	_, err := db1.Exec(ctx, getUserSql())
	if err != nil {
		g.Log().Error(ctx, err)
	}

	// 创建WOl表
	_, err = db1.Exec(ctx, GetWolSql())
	if err != nil {
		g.Log().Error(ctx, err)
	}

}
