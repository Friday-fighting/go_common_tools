package go_common_tools

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// ----------- 压缩包类型探测 -----------
func DetectArchiveType(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch {
	case ext == ".zip":
		return "zip"
	case ext == ".tar":
		return "tar"
	case ext == ".gz" && strings.HasSuffix(strings.ToLower(path), ".tar.gz"):
		return "tgz"
	}
	return ""
}

// ----------- 单文件读为 FileItem -----------
func ReadSingleFile(path string) ([]*FileItem, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return []*FileItem{{
		Data:     data,
		MimeType: http.DetectContentType(data),
		Ext:      strings.ToLower(filepath.Ext(path)),
	}}, nil
}

// ----------- ZIP 解压 -----------
func ExtractZip(zipPath string) (res []*FileItem, err error) {
	res = []*FileItem{}
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	for _, f := range r.File {
		// 目录跳过
		if f.FileInfo().IsDir() {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return nil, err
		}
		data, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			return nil, err
		}
		res = append(res, &FileItem{
			Data:     data,
			MimeType: http.DetectContentType(data),
			Ext:      strings.ToLower(filepath.Ext(f.Name)),
		})
	}
	return res, nil
}

// ----------- TAR / TAR.GZ 解压 -----------
func ExtractTar(tarPath string, gzipped bool) (res []*FileItem, err error) {
	res = []*FileItem{}
	f, err := os.Open(tarPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var r io.Reader = f
	if gzipped {
		gz, err := gzip.NewReader(f)
		if err != nil {
			return nil, err
		}
		defer gz.Close()
		r = gz
	}
	tr := tar.NewReader(r)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		// 只处理普通文件
		if hdr.Typeflag != tar.TypeReg {
			continue
		}
		data, err := io.ReadAll(tr)
		if err != nil {
			return nil, err
		}
		res = append(res, &FileItem{
			Data:     data,
			MimeType: http.DetectContentType(data),
			Ext:      strings.ToLower(filepath.Ext(hdr.Name)),
		})
	}
	return res, nil
}
