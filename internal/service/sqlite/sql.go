package sqlite

import "fmt"

func getUserSql() string {
	// 定义创建users表的原生SQL
	return `
CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,          -- 主键自增
    username TEXT NOT NULL UNIQUE,                 -- 用户名：非空、唯一
    password TEXT NOT NULL,                        -- 密码：非空
    nickname TEXT DEFAULT '',                      -- 昵称：默认空字符串
    email TEXT DEFAULT '',                         -- 邮箱：默认空字符串
    is_admin INTEGER NOT NULL DEFAULT 0,           -- 是否管理员，默认非管理员
    totp_secret TEXT DEFAULT '',                     -- TOTP密钥：默认空字符串
    two_factor_enabled INTEGER NOT NULL DEFAULT 0, -- 是否开启双因素认证，默认关闭
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP, -- 创建时间：默认当前时间
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP  -- 更新时间：默认当前时间
);
`
}

func getCreateDefaultUserSql(username, password string, is_admin bool) string {
	adminValue := 0
	if is_admin {
		adminValue = 1
	}
	sql := fmt.Sprintf("INSERT INTO users (username, password, is_admin) VALUES ('%s', '%s', %d);", username, password, adminValue)
	return sql
}

func GetWolSql() string {
	// 定义创建wol表的原生SQL
	return `
CREATE TABLE IF NOT EXISTS wol (
    id INTEGER PRIMARY KEY AUTOINCREMENT,          -- 主键自增
    device_name TEXT NOT NULL,                     -- 设备名称：非空
    mac_addr TEXT NOT NULL UNIQUE,                 -- MAC地址
    broadcast_ip TEXT NOT NULL DEFAULT '192.168.1.255', -- 广播IP
    port INTEGER NOT NULL DEFAULT 9,               -- 唤醒端口
    remark TEXT DEFAULT '',                        -- 备注：默认空字符串
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP, -- 创建时间
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP  -- 更新时间
);
`
}
