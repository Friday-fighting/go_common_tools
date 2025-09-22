package go_common_tools

import (
	"bytes"
	"fmt"
	"math"
	"os/exec"
	"strconv"
	"strings"
)

// 保证 ffmpeg 与 ffprobe 已在 PATH 中
// ReadSpecifiedFrame 读取视频指定帧，返回 RGB24 比特流
// frameParam: >1 表示第几帧(1-base)，<1 表示百分比(0.1=10%)
func ReadSpecifiedFrame(videoPath string, frameParam float64) ([]byte, error) {
	// 1. 获取总帧数
	totalFrames, err := getTotalFramesA(videoPath)
	if err != nil {
		return nil, fmt.Errorf("getTotalFrames: %w", err)
	}

	// 2. 计算目标帧号（1-base）
	var targetFrame int
	if frameParam >= 1 {
		targetFrame = int(math.Round(frameParam))
	} else {
		targetFrame = int(math.Round(frameParam * float64(totalFrames)))
	}
	if targetFrame < 1 {
		targetFrame = 1
	}
	if targetFrame > totalFrames {
		targetFrame = totalFrames
	}

	// 3. 用 ffmpeg 抽帧：跳过(targetFrame-1)帧，读1帧，rawvideo RGB24 输出到 stdout
	cmd := exec.Command("ffmpeg",
		"-i", videoPath,
		"-vf", fmt.Sprintf("select=gte(n\\,%d)", targetFrame-1),
		"-frames:v", "1",
		"-f", "rawvideo",
		"-pix_fmt", "rgb24",
		"-",
	)

	var buf bytes.Buffer
	cmd.Stdout = &buf
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("ffmpeg run: %w", err)
	}

	return buf.Bytes(), nil
}

// getTotalFrames 调用 ffprobe 返回视频总帧数
func getTotalFramesA(videoPath string) (int, error) {
	cmd := exec.Command("ffprobe",
		"-v", "error",
		"-select_streams", "v:0",
		"-show_entries", "stream=nb_frames",
		"-of", "csv=p=0",
		videoPath,
	)
	out, err := cmd.Output()
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(strings.TrimSpace(string(out)))
}
