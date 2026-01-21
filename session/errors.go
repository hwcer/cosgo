// Package session 提供会话管理功能，支持内存和Redis存储
package session

import (
	"github.com/hwcer/cosgo/values"
)

// 错误常量定义
// 注意：
// 1. ErrorStorageEmpty: 会话存储未设置
// 2. ErrorSessionEmpty: 会话Token为空
// 3. ErrorSessionNotCreate: 会话未创建
// 4. ErrorSessionNotExist: 会话不存在
// 5. ErrorSessionExpired: 会话已过期
// 6. ErrorSessionIllegal: 会话Token非法
// 7. ErrorSessionUnknown: 会话未知错误
// 8. ErrorSessionReplaced: 会话已被替换
var (
	ErrorStorageEmpty     = values.Errorf(202, "session Storage not set")
	ErrorSessionEmpty     = values.Errorf(203, "session token empty")
	ErrorSessionNotCreate = values.Errorf(204, "session not create")
	ErrorSessionNotExist  = values.Errorf(205, "session not exist")
	ErrorSessionExpired   = values.Errorf(206, "session expire")
	ErrorSessionIllegal   = values.Errorf(207, "session illegal")
	ErrorSessionUnknown   = values.Errorf(208, "session unknown error")
	ErrorSessionReplaced  = values.Errorf(209, "session replaced")
)

// Errorf 创建会话相关错误
func Errorf(format any, args ...any) error {
	return values.Errorf(201, format, args...)
}
