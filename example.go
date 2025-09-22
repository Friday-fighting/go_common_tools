package go_common_tools

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
)

// ExtractFrame 从视频二进制流中提取指定帧并返回 JPG 二进制流。
// 当 position < 1 时，表示取总帧数的百分比（0.0~1.0）。
func ExtractFrame(videoData []byte, position float64) ([]byte, error) {
	// 1. 先把视频数据喂给 ffprobe，拿到总帧数
	totalFrames, err := getTotalFrames(videoData)
	if err != nil {
		return nil, fmt.Errorf("ffprobe failed: %w", err)
	}

	// 2. 计算要提取的帧号
	var frameNum int
	if position < 1 {
		frameNum = int(float64(totalFrames) * position)
	} else {
		frameNum = int(position)
	}
	if frameNum < 0 {
		frameNum = 0
	}
	if frameNum >= totalFrames {
		frameNum = totalFrames - 1
	}

	// 3. 用 ffmpeg 提取该帧为 jpg
	return extractFrameJpeg(videoData, frameNum)
}

// getTotalFrames 通过 ffprobe 获取总帧数
func getTotalFrames(videoData []byte) (int, error) {
	cmd := exec.Command("ffprobe",
		"-v", "error",
		"-select_streams", "v:0",
		"-count_frames",
		"-show_entries", "stream=nb_read_frames",
		"-of", "csv=p=0",
		"-i", "-", // 从 stdin 读
	)
	cmd.Stdin = bytes.NewReader(videoData)

	out, err := cmd.Output()
	if err != nil {
		return 0, err
	}
	var total int
	if _, err := fmt.Sscanf(string(out), "%d", &total); err != nil {
		return 0, errors.New("cannot parse ffprobe output")
	}
	return total, nil
}

// extractFrameJpeg 提取指定帧并返回 JPG 二进制
func extractFrameJpeg(videoData []byte, frameNum int) ([]byte, error) {
	cmd := exec.Command("ffmpeg",
		"-i", "-", // 从 stdin 读视频
		"-vf", fmt.Sprintf("select=eq(n\\,%d)", frameNum),
		"-vsync", "vfr", // 只输出选中的帧
		"-q:v", "2", // jpg 质量
		"-f", "image2pipe",
		"-vcodec", "mjpeg",
		"-", // 输出到 stdout
	)
	cmd.Stdin = bytes.NewReader(videoData)

	var buf bytes.Buffer
	cmd.Stdout = &buf
	stderr := &bytes.Buffer{}
	cmd.Stderr = stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("ffmpeg error: %w, stderr: %s", err, stderr.String())
	}
	if buf.Len() == 0 {
		return nil, errors.New("ffmpeg produced empty image")
	}
	return buf.Bytes(), nil
}
