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
使用golang实现find <root> type d -maxdepth 1 命令 但不包含root文件夹本身
*/
func FindAllFoldersInRoot(root string) []string {
	var folders []string

	// 获取根目录的绝对路径
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return folders
	}

	// 添加根目录本身
	//folders = append(folders, absRoot)

	// 读取根目录下的所有文件和文件夹
	entries, err := os.ReadDir(absRoot)
	if err != nil {
		return folders
	}

	// 遍历第一层目录中的所有项
	for _, entry := range entries {
		if entry.IsDir() {
			// 构造完整路径
			fullPath := filepath.Join(absRoot, entry.Name())
			folders = append(folders, fullPath)
		}
	}

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
