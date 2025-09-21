package go_common_tools

// FileItem 统一描述“一个文件”的二进制流与元数据
type FileItem struct {
	Data     []byte // 文件内容
	MimeType string // http.DetectContentType 结果
	Ext      string // 原始扩展名（含点）
}
