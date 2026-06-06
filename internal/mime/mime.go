package mime

import (
	"os"
)

// IsBinaryByMagic 通过文件魔数检测是否为二进制文件
// 读取文件前 512 字节，检测常见二进制格式
func IsBinaryByMagic(path string) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer f.Close()

	buf := make([]byte, 512)
	n, err := f.Read(buf)
	if err != nil && n == 0 {
		return false, err
	}
	buf = buf[:n]

	// 空文件视为文本
	if n == 0 {
		return false, nil
	}

	// 检测常见二进制魔数
	switch {
	// PNG
	case n >= 4 && buf[0] == 0x89 && buf[1] == 0x50 && buf[2] == 0x4E && buf[3] == 0x47:
		return true, nil
	// JPEG
	case n >= 2 && buf[0] == 0xFF && buf[1] == 0xD8:
		return true, nil
	// GIF
	case n >= 4 && buf[0] == 0x47 && buf[1] == 0x49 && buf[2] == 0x46:
		return true, nil
	// BMP
	case n >= 2 && buf[0] == 0x42 && buf[1] == 0x4D:
		return true, nil
	// WebP
	case n >= 12 && buf[0] == 0x52 && buf[1] == 0x49 && buf[2] == 0x46 && buf[3] == 0x46 &&
		buf[8] == 0x57 && buf[9] == 0x45 && buf[10] == 0x42 && buf[11] == 0x50:
		return true, nil
	// ELF (Linux 可执行文件)
	case n >= 4 && buf[0] == 0x7F && buf[1] == 0x45 && buf[2] == 0x4C && buf[3] == 0x46:
		return true, nil
	// Mach-O (macOS 可执行文件)
	case n >= 4 && (buf[0] == 0xCF && buf[1] == 0xFA && buf[2] == 0xED && buf[3] == 0xFE):
		return true, nil
	case n >= 4 && (buf[0] == 0xFE && buf[1] == 0xED && buf[2] == 0xFA && buf[3] == 0xCF):
		return true, nil
	// PE (Windows 可执行文件)
	case n >= 2 && buf[0] == 0x4D && buf[1] == 0x5A:
		return true, nil
	// ZIP (包括 docx, xlsx, jar 等)
	case n >= 4 && buf[0] == 0x50 && buf[1] == 0x4B && buf[2] == 0x03 && buf[3] == 0x04:
		return true, nil
	// GZIP
	case n >= 2 && buf[0] == 0x1F && buf[1] == 0x8B:
		return true, nil
	// PDF
	case n >= 4 && buf[0] == 0x25 && buf[1] == 0x50 && buf[2] == 0x44 && buf[3] == 0x46:
		return true, nil
	// SQLite
	case n >= 16 && string(buf[:16]) == "SQLite format 3\x00":
		return true, nil
	// MP3 (ID3)
	case n >= 3 && buf[0] == 0x49 && buf[1] == 0x44 && buf[2] == 0x33:
		return true, nil
	// MP4 / MOV (ftyp)
	case n >= 8 && buf[4] == 0x66 && buf[5] == 0x74 && buf[6] == 0x79 && buf[7] == 0x70:
		return true, nil
	// TTF / OTF
	case n >= 4 && string(buf[:4]) == "\x00\x01\x00\x00":
		return true, nil
	case n >= 4 && string(buf[:4]) == "OTTO":
		return true, nil
	// WOFF
	case n >= 4 && string(buf[:4]) == "wOFF":
		return true, nil
	case n >= 4 && string(buf[:4]) == "wOF2":
		return true, nil
	}

	// 检测是否包含大量 NUL 字节（二进制文件的典型特征）
	nulCount := 0
	for _, b := range buf {
		if b == 0 {
			nulCount++
		}
	}
	// 如果 NUL 字节占比超过 5%，认为是二进制文件
	if float64(nulCount)/float64(n) > 0.05 {
		return true, nil
	}

	return false, nil
}
