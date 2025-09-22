package finder

import (
	"os"
	"path/filepath"

	"github.com/h2non/filetype"
)

/*
使用golang实现find <root> type f 命令
*/
func FindAllImages(root string) []string {
	var files []string
	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // 忽略错误，继续遍历
		}
		if !info.IsDir() {
			absPath, _ := filepath.Abs(path)
			if isImage(absPath) {
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
func FindAllImagesInRoot(root string) []string {
	var files []string
	entries, err := os.ReadDir(root)
	if err != nil {
		return files
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			absPath, _ := filepath.Abs(filepath.Join(root, entry.Name()))
			if isImage(absPath) {
				files = append(files, absPath)
			}
		}
	}
	return files
}

func isImage(fp string) bool {
	file, _ := os.Open(fp)
	defer file.Close()
	head := make([]byte, 261)
	file.Read(head)
	if filetype.IsImage(head) {
		return true
	} else {
		return false
	}
}
