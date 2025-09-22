package go_common_tools

import (
	"errors"
	"fmt"
	"os"
)

func ReadVideoFrame(video any, param float64) ([]byte, error) {
	switch v := video.(type) {
	case string:
		return ReadSpecifiedFrame(v, param) // 路径版
	case []byte:
		if len(v) > 100*1024*1024 { // >100 MB 先落盘
			tmp, err := os.CreateTemp("", "vid*.mp4")
			if err != nil {
				return nil, fmt.Errorf("create temp file: %w", err)
			}
			_, err = tmp.Write(v)
			if err != nil {
				return nil, fmt.Errorf("write temp file: %w", err)
			}
			tmp.Close()
			defer os.Remove(tmp.Name())
			return ReadSpecifiedFrame(tmp.Name(), param)
		}
		return ReadSpecifiedFrameFromMem(v, param) // 小文件走内存
	default:
		return nil, errors.New("unsupported type")
	}
}
