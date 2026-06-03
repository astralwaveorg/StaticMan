//go:build withweb

package web

import (
	"embed"
	"io/fs"
)

// 嵌入前端构建产物
//
//go:embed all:dist
var embeddedFS embed.FS

func init() {
	// 生产模式：使用嵌入的文件系统
	sub, _ := fs.Sub(embeddedFS, "dist")
	assets = sub
}