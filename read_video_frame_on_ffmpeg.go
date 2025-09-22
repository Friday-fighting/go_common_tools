package go_common_tools

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

func ReadVideoFramePath(videoPath string, outputPath string, timePosition float64) error {
	// 检查视频文件是否存在
	if _, err := os.Stat(videoPath); os.IsNotExist(err) {
		return fmt.Errorf("错误：视频文件不存在: %s", videoPath)
	}
	// 获取视频时长
	duration, err := getVideoDuration(videoPath)
	if err != nil {
		return fmt.Errorf("错误：无法获取视频时长: %w", err)
	}
	// 计算指定位置的时间点
	position := duration * timePosition
	fmt.Printf("视频总时长: %.2f秒, 指定位置: %.2f秒\n", duration, position)
	// 使用ffmpeg提取帧
	err = extractFrame(videoPath, position, outputPath)
	if err != nil {
		return fmt.Errorf("错误：抽帧失败: %w", err)
	}
	// 检查输出文件
	fileInfo, err := os.Stat(outputPath)
	if err != nil {
		return fmt.Errorf("错误：无法获取输出文件信息: %w", err)
	}
	fmt.Printf("成功：已提取帧并保存到 %s, 文件大小: %d 字节\n", outputPath, fileInfo.Size())
	return nil
}
func ReadVideoFrameMemory(videoData []byte, timePosition float64) ([]byte, error) {
	// 获取视频时长
	duration, err := getVideoDurationFromMem(videoData)
	if err != nil {
		return nil, fmt.Errorf("错误：无法获取视频时长: %w", err)
	}

	// 计算指定位置的时间点
	position := duration * timePosition
	fmt.Printf("视频总时长: %.2f秒, 指定位置: %.2f秒\n", duration, position)

	// 使用ffmpeg从内存中提取帧
	frameData, err := extractFrameFromMem(videoData, position)
	if err != nil {
		return nil, fmt.Errorf("错误：抽帧失败: %w", err)
	}
	return frameData, nil
}

// 从内存中获取视频时长（秒）
func getVideoDurationFromMem(videoData []byte) (float64, error) {
	cmd := exec.Command("ffprobe",
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		"-i", "pipe:0") // 从标准输入读取

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return 0, err
	}

	var outBuf bytes.Buffer
	cmd.Stdout = &outBuf

	var errBuf bytes.Buffer
	cmd.Stderr = &errBuf

	if err := cmd.Start(); err != nil {
		return 0, err
	}

	// 把视频数据写入stdin
	if _, err := io.Copy(stdin, bytes.NewReader(videoData)); err != nil {
		_ = stdin.Close()
		return 0, fmt.Errorf("写入视频数据失败: %w", err)
	}
	_ = stdin.Close()

	if err := cmd.Wait(); err != nil {
		errMsg := errBuf.String()
		return 0, fmt.Errorf("ffprobe执行失败: %w, 错误信息: %s", err, errMsg)
	}

	// 去除输出中的换行符
	durationStr := strings.TrimSpace(outBuf.String())
	return strconv.ParseFloat(durationStr, 64)
}

// 从内存中提取指定时间点的帧
func extractFrameFromMem(videoData []byte, timePosition float64) ([]byte, error) {
	// 创建临时视频文件
	tmpVideoFile, err := os.CreateTemp("", "video_*.mp4")
	if err != nil {
		return nil, fmt.Errorf("创建临时视频文件失败: %w", err)
	}
	tmpVideoPath := tmpVideoFile.Name()
	defer os.Remove(tmpVideoPath) // 确保临时文件被删除

	// 写入视频数据到临时文件
	if _, err := tmpVideoFile.Write(videoData); err != nil {
		tmpVideoFile.Close()
		return nil, fmt.Errorf("写入临时视频文件失败: %w", err)
	}
	tmpVideoFile.Close()

	// 创建临时输出文件
	tmpOutputFile, err := os.CreateTemp("", "frame_*.jpg")
	if err != nil {
		return nil, fmt.Errorf("创建临时输出文件失败: %w", err)
	}
	tmpOutputPath := tmpOutputFile.Name()
	tmpOutputFile.Close()
	defer os.Remove(tmpOutputPath) // 确保临时文件被删除

	// 使用ffmpeg从临时文件中提取帧
	cmd := exec.Command("ffmpeg",
		"-y", // 自动覆盖输出文件
		"-ss", fmt.Sprintf("%.6f", timePosition),
		"-i", tmpVideoPath,
		"-vframes", "1",
		"-q:v", "2",
		tmpOutputPath)

	var errBuf bytes.Buffer
	cmd.Stderr = &errBuf

	if err := cmd.Run(); err != nil {
		errMsg := errBuf.String()
		return nil, fmt.Errorf("ffmpeg执行失败: %w, 错误信息: %s", err, errMsg)
	}

	// 读取生成的图片
	frameData, err := os.ReadFile(tmpOutputPath)
	if err != nil {
		return nil, fmt.Errorf("读取输出文件失败: %w", err)
	}

	if len(frameData) == 0 {
		return nil, fmt.Errorf("生成的图片为空")
	}

	return frameData, nil
}

// 获取视频时长（秒）
func getVideoDuration(videoPath string) (float64, error) {
	cmd := exec.Command("ffprobe",
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		videoPath)

	out, err := cmd.Output()
	if err != nil {
		return 0, err
	}

	// 去除输出中的换行符
	durationStr := strings.TrimSpace(string(out))
	return strconv.ParseFloat(durationStr, 64)
}

// 提取指定时间点的帧
func extractFrame(videoPath string, timePosition float64, outputPath string) error {
	cmd := exec.Command("ffmpeg",
		"-y", // 自动覆盖输出文件
		"-ss", fmt.Sprintf("%.6f", timePosition),
		"-i", videoPath,
		"-vframes", "1",
		"-q:v", "2",
		outputPath)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
