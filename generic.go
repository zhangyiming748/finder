package finder

import (
	"os"
	"path/filepath"
)

/*
使用golang实现find <root> type d 命令
*/
func FindAllFolders(root string) []string {
	var folders []string
	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // 忽略错误，继续遍历
		}
		if info.IsDir() {
			absPath, _ := filepath.Abs(path)
			folders = append(folders, absPath)
		}
		return nil
	})
	return folders
}

/*
使用golang实现find <root> type f 命令,包含root文件夹本身
*/
func FindAllFiles(root string) []string {
	var files []string
	files = append(files, root)
	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // 忽略错误，继续遍历
		}
		if !info.IsDir() {
			absPath, _ := filepath.Abs(path)
			files = append(files, absPath)
		}
		return nil
	})
	return files
}
