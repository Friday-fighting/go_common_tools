package go_common_tools

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// ----------- 简单测试 -----------
func TestExt(t *testing.T) {
	// 把任意视频读进来
	cacheFilePath := filepath.Join("tmp", "cacheDownload")
	os.MkdirAll(cacheFilePath, os.ModePerm)
	localPath, err := DownloadFromUrl("http://qiniutest.adesk.com/common_ai_gen/user_ai_result/20_96_1758537470953_bsshabxz.mp4?version=1758537472&sign=6c17aefb188e60cbe0c0e928025d3fc2&t=68d146f0", cacheFilePath)
	if err != nil {
		panic(err)
	}
	defer os.Remove(localPath)
	video, _ := os.ReadFile(localPath)
	img, err := ExtractFrame(video, 0.5) // 取 50% 处
	if err != nil {
		panic(err)
	}
	// 保存提取的 jpg
	if err := os.WriteFile(filepath.Join("tmp", "extracted.jpg"), img, os.ModePerm); err != nil {
		panic(err)
	}
	fmt.Printf("extracted jpg size=%d\n", len(img))
}
