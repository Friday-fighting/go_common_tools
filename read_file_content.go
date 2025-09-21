package go_common_tools

import "fmt"

func ReadFileContent(filePath string) (res []*FileItem, err error) {
	switch DetectArchiveType(filePath) {
	case "zip":
		res, err = ExtractZip(filePath)
	case "tar":
		res, err = ExtractTar(filePath, false)
	case "tgz":
		res, err = ExtractTar(filePath, true)
	default: // 普通文件
		res, err = ReadSingleFile(filePath)
	}
	if err != nil {
		return nil, fmt.Errorf("extract/read failed: %w", err)
	}
	return res, nil
}
