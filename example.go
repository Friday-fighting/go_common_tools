package go_common_tools

import (
	"log"
	"os"
)

func main() {
	var (
		url = "xxx" //文件链接
		err error
	)
	// 用系统临时目录当工作目录
	tempDir, err := os.MkdirTemp("", "downzip_")
	if err != nil {
		log.Fatalf("创建临时目录失败: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			log.Printf("清理临时目录失败: %v", err)
		}
	}()
	localPath, err := DownloadFromUrl(url, tempDir)
	if err != nil {
		panic(err)
	}
	// 临时下载，无论成败最后删本地文件
	defer os.Remove(localPath)
	fileContents, err := ReadFileContent(localPath)
	if err != nil {
		panic(err)
	}

	for _, item := range fileContents {
		// handle item.item.Data
		switch n := item.Ext; n {
		case "mp4":
			continue
		case "gif":
			continue
		default:
			continue
		}
	}
}
