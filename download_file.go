package go_common_tools

import (
	"crypto/rand"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// DownloadTo 把远端文件下载到指定目录，文件名格式：
// <毫秒时间戳>_<8位随机字母>.<原扩展名>
func DownloadFromUrl(rawURL string, dir string) (string, error) {
	// 1. 解析 URL 并提取扩展名
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("invalid url: %w", err)
	}
	// 取路径最后一截，去掉查询参数
	base := filepath.Base(u.Path)
	if base == "" || base == "/" {
		base = "file"
	}
	ext := strings.ToLower(filepath.Ext(base))
	if ext == "" {
		ext = ".bin" // 默认兜底
	}

	// 2. 生成 8 位随机字母
	letter := make([]byte, 8)
	rand.Read(letter) // err 始终为 nil
	for i := range letter {
		letter[i] = 'a' + (letter[i] % 26)
	}

	// 3. 拼文件名：时间戳_随机串.扩展名
	name := fmt.Sprintf("%d_%s%s",
		time.Now().UnixMilli(), // 毫秒级时间戳
		string(letter),
		ext,
	)
	localPath := filepath.Join(dir, name)

	// 4. 创建文件
	out, err := os.Create(localPath)
	if err != nil {
		return "", err
	}
	defer out.Close()

	// 5. 下载
	resp, err := http.Get(rawURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("bad status: %s", resp.Status)
	}

	// 6. 写盘
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return "", fmt.Errorf("download failed: %w", err)
	}
	return localPath, err
}
