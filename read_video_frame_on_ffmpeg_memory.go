package go_common_tools

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"math"
	"os/exec"
	"strconv"
	"strings"
)

// ReadSpecifiedFrameFromMem 从视频二进制流中抽取指定帧
// frameParam: >1 表示第几帧(1-base)，<1 表示百分比(0.1=10%)
func ReadSpecifiedFrameFromMem(videoData []byte, frameParam float64) ([]byte, error) {
	if len(videoData) == 0 {
		return nil, errors.New("empty video data")
	}

	// 1. 先把视频流丢给 ffprobe，拿总帧数
	totalFrames, err := getTotalFramesFromMem(videoData)
	if err != nil {
		return nil, fmt.Errorf("getTotalFrames: %w", err)
	}

	// 2. 计算目标帧号
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

	// 3. 用 ffmpeg 抽帧：stdin 喂视频，stdout 接 rawvideo(RGB24)
	cmd := exec.Command("ffmpeg",
		"-i", "pipe:0", // 从 stdin 读
		"-vf", fmt.Sprintf("select=gte(n\\,%d)", targetFrame-1),
		"-frames:v", "1",
		"-f", "rawvideo",
		"-pix_fmt", "rgb24",
		"pipe:1", // 输出到 stdout
	)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	var outBuf bytes.Buffer
	cmd.Stdout = &outBuf

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	// 把完整视频流写进去
	if _, err := io.Copy(stdin, bytes.NewReader(videoData)); err != nil {
		_ = stdin.Close()
		return nil, err
	}
	_ = stdin.Close()

	if err := cmd.Wait(); err != nil {
		return nil, fmt.Errorf("ffmpeg wait: %w", err)
	}

	return outBuf.Bytes(), nil
}

// getTotalFramesFromMem 用 ffprobe 从内存视频流里读总帧数
func getTotalFramesFromMem(videoData []byte) (int, error) {
	cmd := exec.Command("ffprobe",
		"-v", "error",
		"-select_streams", "v:0",
		"-show_entries", "stream=nb_frames",
		"-of", "csv=p=0",
		"pipe:0", // 从 stdin 读
	)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return 0, err
	}
	out, err := cmd.Output()
	if err != nil {
		_ = stdin.Close()
		return 0, err
	}

	go func() {
		_, _ = io.Copy(stdin, bytes.NewReader(videoData))
		_ = stdin.Close()
	}()

	frames, err := strconv.Atoi(strings.TrimSpace(string(out)))
	if err != nil {
		return 0, fmt.Errorf("parse frames: %w", err)
	}
	return frames, nil
}
