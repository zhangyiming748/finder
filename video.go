package finder

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/h2non/filetype"
)

/*
使用golang实现find <root> type f 命令
*/
func FindAllVideos(root string) []string {
	var files []string
	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // 忽略错误，继续遍历
		}
		if !info.IsDir() {
			absPath, _ := filepath.Abs(path)
			if isVideo(absPath) {
				files = append(files, absPath)
			}
		}
		return nil
	})
	return files
}

/*
使用golang实现查找给定文件路径下的全部文件 不包含子目录
*/
func FindAllVideosInRoot(root string) []string {
	var files []string
	entries, err := os.ReadDir(root)
	if err != nil {
		return files
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			absPath, _ := filepath.Abs(filepath.Join(root, entry.Name()))
			if isVideo(absPath) {
				files = append(files, absPath)
			}
		}
	}
	return files
}

func isVideo(fp string) bool {
	file, _ := os.Open(fp)
	defer file.Close()
	head := make([]byte, 261)
	file.Read(head)
	ext := strings.ToLower(filepath.Ext(fp))
	if filetype.IsVideo(head) {
		return true
	} else if ext == ".rmvb" {
		return true
	} else if ext == ".rm" {
		return true
	} else {
		return false
	}
}
