package sqlite

import (
	"context"
	"fmt"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"golang.org/x/crypto/bcrypt"
)

type Sqliter interface {
	InitSqlite(debug bool)
}

type Sqlite struct {
	RootPath string
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func CheckPassword(password, hashPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashPassword), []byte(password))
	return err == nil
}

func NewSqlite(RootPath string, debug bool) Sqliter {
	Sqlite := &Sqlite{
		RootPath: RootPath,
	}
	Sqlite.InitSqlite(debug)
	Sqlite.IintTable()
	return Sqlite
}

func (s *Sqlite) InitSqlite(debug bool) {
	dbConfig := gdb.ConfigNode{
		Link:  fmt.Sprintf("sqlite::@file(%sdb.sqlite3)", s.RootPath),
		Debug: debug,
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

	// 创建默认管理员用户
	user_name := g.Cfg().MustGetWithCmd(ctx, `username`, "admin")
	if user_name == nil {
		user_name = g.Cfg().MustGetWithCmd(ctx, `u`, "admin")
	}

	pass_word := g.Cfg().MustGetWithCmd(ctx, `password`, "admin123")
	if pass_word == nil {
		pass_word = g.Cfg().MustGetWithCmd(ctx, `p`, "admin123")
	}
	hashedPassword, err := HashPassword(pass_word.String())
	if err != nil {
		g.Log().Error(ctx, err)
		return
	}
	_, err = db1.Exec(ctx, getCreateDefaultUserSql(user_name.String(), hashedPassword, true))
	if err != nil {
		g.Log().Warning(ctx, err)
	}

	// 创建WOl表
	_, err = db1.Exec(ctx, GetWolSql())
	if err != nil {
		g.Log().Error(ctx, err)
	}

}
