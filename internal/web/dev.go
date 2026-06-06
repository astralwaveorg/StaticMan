//go:build !withweb

package web

import "os"

func init() {
	// 开发模式：从本地文件系统读取前端资源
	assets = os.DirFS("internal/web/dist")
}
