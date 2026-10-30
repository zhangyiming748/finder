package finder

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// 最小可识别的文件头，让 filetype 能把测试文件识别为图片 / 音频
var (
	pngHeader  = []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	jpegHeader = []byte{0xFF, 0xD8, 0xFF, 0xE0}
	gifHeader  = []byte("GIF89a")
	mp3Header  = []byte("ID3\x04\x00\x00\x00\x00\x00\x00")
	wavHeader  = []byte("RIFF\x24\x00\x00\x00WAVEfmt ")
	oggHeader  = []byte("OggS\x00\x02\x00\x00")
)

func mediaHeader(ext string) []byte {
	switch strings.ToLower(ext) {
	case ".png":
		return pngHeader
	case ".jpg", ".jpeg":
		return jpegHeader
	case ".gif":
		return gifHeader
	case ".mp3":
		return mp3Header
	case ".wav":
		return wavHeader
	case ".ogg":
		return oggHeader
	default:
		return nil
	}
}

func writeTestFile(t *testing.T, path string, content []byte) {
	t.Helper()
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatalf("Failed to create test file %s: %v", path, err)
	}
}

func TestFindAllFolders(t *testing.T) {
	// Step 1: 创建临时测试目录结构
	tempDir := t.TempDir()

	// Step 2: 创建测试文件夹结构
	folders := []string{"folder1", "folder2", "folder1/subfolder1", "folder2/subfolder2"}
	for _, folder := range folders {
		err := os.MkdirAll(filepath.Join(tempDir, folder), 0755)
		if err != nil {
			t.Fatalf("Failed to create test folder %s: %v", folder, err)
		}
	}

	// Step 3: 执行测试
	result := FindAllFolders(tempDir)

	// Step 4: 验证结果数量（FindAllFolders 会把 root 本身也计入结果）
	expectedCount := len(folders) + 1
	if len(result) != expectedCount {
		t.Errorf("Expected %d folders, got %d", expectedCount, len(result))
	}

	// Step 5: 验证结果内容
	folderMap := make(map[string]bool)
	for _, folder := range result {
		folderMap[folder] = true
	}

	for _, folder := range folders {
		expectedPath := filepath.Join(tempDir, folder)
		if !folderMap[expectedPath] {
			t.Errorf("Expected folder %s not found in result", expectedPath)
		}
	}

	if !folderMap[tempDir] {
		t.Errorf("Expected root folder %s not found in result", tempDir)
	}
}

func TestFindAllFiles(t *testing.T) {
	// Step 1: 创建临时测试目录结构
	tempDir := t.TempDir()

	// Step 2: 创建测试文件夹和文件
	folders := []string{"folder1", "folder2"}
	for _, folder := range folders {
		err := os.MkdirAll(filepath.Join(tempDir, folder), 0755)
		if err != nil {
			t.Fatalf("Failed to create test folder %s: %v", folder, err)
		}
	}

	files := []string{"file1.txt", "file2.log", "folder1/file3.txt", "folder2/file4.log"}
	for _, file := range files {
		fullPath := filepath.Join(tempDir, file)
		f, err := os.Create(fullPath)
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", file, err)
		}
		f.Close()
	}

	// Step 3: 执行测试
	result := FindAllFiles(tempDir)

	// Step 4: 验证结果数量（FindAllFiles 会把 root 本身也计入结果）
	expectedCount := len(files) + 1
	if len(result) != expectedCount {
		t.Errorf("Expected %d files, got %d", expectedCount, len(result))
	}

	// Step 5: 验证结果内容
	fileMap := make(map[string]bool)
	for _, file := range result {
		fileMap[file] = true
	}

	for _, file := range files {
		expectedPath := filepath.Join(tempDir, file)
		if !fileMap[expectedPath] {
			t.Errorf("Expected file %s not found in result", expectedPath)
		}
	}

	if !fileMap[tempDir] {
		t.Errorf("Expected root %s not found in result", tempDir)
	}
}

func TestFindAllVideos(t *testing.T) {
	// Step 1: 创建临时测试目录
	tempDir := t.TempDir()

	// Step 2: 创建一些测试文件（使用真实视频文件的扩展名但内容为空）
	videoFiles := []string{"video1.mp4", "video2.avi", "sub/video3.mov"}
	otherFiles := []string{"text.txt", "image.jpg", "doc.pdf"}

	// Step 3: 创建子目录
	err := os.MkdirAll(filepath.Join(tempDir, "sub"), 0755)
	if err != nil {
		t.Fatalf("Failed to create sub directory: %v", err)
	}

	// Step 4: 创建视频文件
	for _, file := range videoFiles {
		fullPath := filepath.Join(tempDir, file)
		f, err := os.Create(fullPath)
		if err != nil {
			t.Fatalf("Failed to create test video file %s: %v", file, err)
		}
		f.Close()
	}

	// Step 5: 创建其他类型文件
	for _, file := range otherFiles {
		fullPath := filepath.Join(tempDir, file)
		f, err := os.Create(fullPath)
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", file, err)
		}
		f.Close()
	}

	// Step 6: 执行测试
	result := FindAllVideos(tempDir)

	// Step 7: 验证结果数量
	expectedCount := len(videoFiles)
	if len(result) != expectedCount {
		t.Errorf("Expected %d video files, got %d", expectedCount, len(result))
	}

	// Step 8: 验证结果内容
	videoMap := make(map[string]bool)
	for _, file := range result {
		videoMap[file] = true
	}

	for _, file := range videoFiles {
		expectedPath := filepath.Join(tempDir, file)
		if !videoMap[expectedPath] {
			t.Errorf("Expected video file %s not found in result", expectedPath)
		}
	}
}

func TestFindAllVideosInRoot(t *testing.T) {
	// Step 1: 创建临时测试目录
	tempDir := t.TempDir()

	// Step 2: 创建根目录和子目录
	err := os.MkdirAll(filepath.Join(tempDir, "sub"), 0755)
	if err != nil {
		t.Fatalf("Failed to create sub directory: %v", err)
	}

	// Step 3: 创建根目录和子目录视频文件列表
	rootVideoFiles := []string{"video1.mp4", "video2.avi"}
	subVideoFiles := []string{"sub/video3.mov"}

	// Step 4: 创建根目录视频文件
	for _, file := range rootVideoFiles {
		fullPath := filepath.Join(tempDir, file)
		f, err := os.Create(fullPath)
		if err != nil {
			t.Fatalf("Failed to create test video file %s: %v", file, err)
		}
		f.Close()
	}

	// Step 5: 创建子目录视频文件
	for _, file := range subVideoFiles {
		fullPath := filepath.Join(tempDir, file)
		f, err := os.Create(fullPath)
		if err != nil {
			t.Fatalf("Failed to create test video file %s: %v", file, err)
		}
		f.Close()
	}

	// Step 6: 执行测试
	result := FindAllVideosInRoot(tempDir)

	// Step 7: 验证结果数量（应该只包含根目录的视频文件）
	expectedCount := len(rootVideoFiles)
	if len(result) != expectedCount {
		t.Errorf("Expected %d video files in root, got %d", expectedCount, len(result))
	}

	// Step 8: 验证结果内容
	videoMap := make(map[string]bool)
	for _, file := range result {
		videoMap[file] = true
	}

	// Step 9: 检查根目录视频文件是否都包含在结果中
	for _, file := range rootVideoFiles {
		expectedPath := filepath.Join(tempDir, file)
		if !videoMap[expectedPath] {
			t.Errorf("Expected root video file %s not found in result", expectedPath)
		}
	}

	// Step 10: 检查子目录视频文件不应该在结果中
	for _, file := range subVideoFiles {
		expectedPath := filepath.Join(tempDir, file)
		if videoMap[expectedPath] {
			t.Errorf("Subdirectory video file %s should not be in result", expectedPath)
		}
	}
}

func TestFindAllImages(t *testing.T) {
	// Step 1: 创建临时测试目录
	tempDir := t.TempDir()

	// Step 2: 定义图像文件和其他类型文件列表
	imageFiles := []string{"image1.jpg", "image2.png", "sub/image3.gif"}
	otherFiles := []string{"text.txt", "video.mp4", "doc.pdf"}

	// Step 3: 创建子目录
	err := os.MkdirAll(filepath.Join(tempDir, "sub"), 0755)
	if err != nil {
		t.Fatalf("Failed to create sub directory: %v", err)
	}

	// Step 4: 创建图像文件（写入真实文件头，保证能被 filetype 识别）
	for _, file := range imageFiles {
		fullPath := filepath.Join(tempDir, file)
		writeTestFile(t, fullPath, mediaHeader(filepath.Ext(file)))
	}

	// Step 5: 创建其他类型文件
	for _, file := range otherFiles {
		fullPath := filepath.Join(tempDir, file)
		f, err := os.Create(fullPath)
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", file, err)
		}
		f.Close()
	}

	// Step 6: 执行测试
	result := FindAllImages(tempDir)

	// Step 7: 验证结果数量
	expectedCount := len(imageFiles)
	if len(result) != expectedCount {
		t.Errorf("Expected %d image files, got %d", expectedCount, len(result))
	}

	// Step 8: 验证结果内容
	imageMap := make(map[string]bool)
	for _, file := range result {
		imageMap[file] = true
	}

	for _, file := range imageFiles {
		expectedPath := filepath.Join(tempDir, file)
		if !imageMap[expectedPath] {
			t.Errorf("Expected image file %s not found in result", expectedPath)
		}
	}
}

func TestFindAllImagesInRoot(t *testing.T) {
	// Step 1: 创建临时测试目录
	tempDir := t.TempDir()

	// Step 2: 创建根目录和子目录
	err := os.MkdirAll(filepath.Join(tempDir, "sub"), 0755)
	if err != nil {
		t.Fatalf("Failed to create sub directory: %v", err)
	}

	// Step 3: 创建根目录和子目录图像文件列表
	rootImageFiles := []string{"image1.jpg", "image2.png"}
	subImageFiles := []string{"sub/image3.gif"}

	// Step 4: 创建根目录图像文件（写入真实文件头，保证能被 filetype 识别）
	for _, file := range rootImageFiles {
		fullPath := filepath.Join(tempDir, file)
		writeTestFile(t, fullPath, mediaHeader(filepath.Ext(file)))
	}

	// Step 5: 创建子目录图像文件（写入真实文件头，保证能被 filetype 识别）
	for _, file := range subImageFiles {
		fullPath := filepath.Join(tempDir, file)
		writeTestFile(t, fullPath, mediaHeader(filepath.Ext(file)))
	}

	// Step 6: 执行测试
	result := FindAllImagesInRoot(tempDir)

	// Step 7: 验证结果数量（应该只包含根目录的图像文件）
	expectedCount := len(rootImageFiles)
	if len(result) != expectedCount {
		t.Errorf("Expected %d image files in root, got %d", expectedCount, len(result))
	}

	// Step 8: 验证结果内容
	imageMap := make(map[string]bool)
	for _, file := range result {
		imageMap[file] = true
	}

	// Step 9: 检查根目录图像文件是否都包含在结果中
	for _, file := range rootImageFiles {
		expectedPath := filepath.Join(tempDir, file)
		if !imageMap[expectedPath] {
			t.Errorf("Expected root image file %s not found in result", expectedPath)
		}
	}

	// Step 10: 检查子目录图像文件不应该在结果中
	for _, file := range subImageFiles {
		expectedPath := filepath.Join(tempDir, file)
		if imageMap[expectedPath] {
			t.Errorf("Subdirectory image file %s should not be in result", expectedPath)
		}
	}
}

func TestFindAllAudios(t *testing.T) {
	// Step 1: 创建临时测试目录
	tempDir := t.TempDir()

	// Step 2: 定义音频文件和其他类型文件列表
	audioFiles := []string{"audio1.mp3", "audio2.wav", "sub/audio3.ogg"}
	otherFiles := []string{"text.txt", "image.jpg", "doc.pdf"}

	// Step 3: 创建子目录
	err := os.MkdirAll(filepath.Join(tempDir, "sub"), 0755)
	if err != nil {
		t.Fatalf("Failed to create sub directory: %v", err)
	}

	// Step 4: 创建音频文件（写入真实文件头，保证能被 filetype 识别）
	for _, file := range audioFiles {
		fullPath := filepath.Join(tempDir, file)
		writeTestFile(t, fullPath, mediaHeader(filepath.Ext(file)))
	}

	// Step 5: 创建其他类型文件
	for _, file := range otherFiles {
		fullPath := filepath.Join(tempDir, file)
		f, err := os.Create(fullPath)
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", file, err)
		}
		f.Close()
	}

	// Step 6: 执行测试
	result := FindAllAudios(tempDir)

	// Step 7: 验证结果数量
	expectedCount := len(audioFiles)
	if len(result) != expectedCount {
		t.Errorf("Expected %d audio files, got %d", expectedCount, len(result))
	}

	// Step 8: 验证结果内容
	audioMap := make(map[string]bool)
	for _, file := range result {
		audioMap[file] = true
	}

	for _, file := range audioFiles {
		expectedPath := filepath.Join(tempDir, file)
		if !audioMap[expectedPath] {
			t.Errorf("Expected audio file %s not found in result", expectedPath)
		}
	}
}

func TestFindAllAudiosInRoot(t *testing.T) {
	// Step 1: 创建临时测试目录
	tempDir := t.TempDir()

	// Step 2: 创建根目录和子目录
	err := os.MkdirAll(filepath.Join(tempDir, "sub"), 0755)
	if err != nil {
		t.Fatalf("Failed to create sub directory: %v", err)
	}

	// Step 3: 创建根目录和子目录音频文件列表
	rootAudioFiles := []string{"audio1.mp3", "audio2.wav"}
	subAudioFiles := []string{"sub/audio3.ogg"}

	// Step 4: 创建根目录音频文件（写入真实文件头，保证能被 filetype 识别）
	for _, file := range rootAudioFiles {
		fullPath := filepath.Join(tempDir, file)
		writeTestFile(t, fullPath, mediaHeader(filepath.Ext(file)))
	}

	// Step 5: 创建子目录音频文件（写入真实文件头，保证能被 filetype 识别）
	for _, file := range subAudioFiles {
		fullPath := filepath.Join(tempDir, file)
		writeTestFile(t, fullPath, mediaHeader(filepath.Ext(file)))
	}

	// Step 6: 执行测试
	result := FindAllAudiosInRoot(tempDir)

	// Step 7: 验证结果数量（应该只包含根目录的音频文件）
	expectedCount := len(rootAudioFiles)
	if len(result) != expectedCount {
		t.Errorf("Expected %d audio files in root, got %d", expectedCount, len(result))
	}

	// Step 8: 验证结果内容
	audioMap := make(map[string]bool)
	for _, file := range result {
		audioMap[file] = true
	}

	// Step 9: 检查根目录音频文件是否都包含在结果中
	for _, file := range rootAudioFiles {
		expectedPath := filepath.Join(tempDir, file)
		if !audioMap[expectedPath] {
			t.Errorf("Expected root audio file %s not found in result", expectedPath)
		}
	}

	// Step 10: 检查子目录音频文件不应该在结果中
	for _, file := range subAudioFiles {
		expectedPath := filepath.Join(tempDir, file)
		if audioMap[expectedPath] {
			t.Errorf("Subdirectory audio file %s should not be in result", expectedPath)
		}
	}
}

func TestFindAllVideosIgnoresImages(t *testing.T) {
	// 合法的 1x1 JPEG：ffprobe 会把它解析成 mjpeg 视频流，不能因此被判为视频
	const jpegBase64 = "/9j/4AAQSkZJRgABAQEAYABgAAD/2wBDAAgGBgcGBQgHBwcJCQgKDBQNDAsLDBkSEw8UHRofHh0aHBwgJC4nICIsIxwcKDcpLDAxNDQ0Hyc5PTgyPC4zNDL/wAALCAABAAEBAREA/8QAFAABAAAAAAAAAAAAAAAAAAAACf/EABQQAQAAAAAAAAAAAAAAAAAAAAD/2gAIAQEAAD8AKp//2Q=="
	// 合法的 AVIF（AV1 静帧）：ffprobe 会把它解析成 av1 视频流，且 filetype 无法识别它的文件头
	const avifBase64 = "AAAAIGZ0eXBhdmlmAAAAAGF2aWZtaWYxbWlhZk1BMUIAAAD5bWV0YQAAAAAAAAAvaGRscgAAAAAAAAAAcGljdAAAAAAAAAAAAAAAAFBpY3R1cmVIYW5kbGVyAAAAAA5waXRtAAAAAAABAAAAHmlsb2MAAAAARAAAAQABAAAAAQAAASEAAAAkAAAAKGlpbmYAAAAAAAEAAAAaaW5mZQIAAAAAAQAAYXYwMUNvbG9yAAAAAGppcHJwAAAAS2lwY28AAAAUaXNwZQAAAAAAAAAQAAAAEAAAABBwaXhpAAAAAAMICAgAAAAMYXYxQ4EADAAAAAATY29scm5jbHgAAgACAAIAAAAAF2lwbWEAAAAAAAAAAQABBAECgwQAAAAsbWRhdAoKAgAABUz/xq+QBDIWEACXgBBAggAAEABcEytNbG+TY8nR6A=="
	jpegBytes, err := base64.StdEncoding.DecodeString(jpegBase64)
	if err != nil {
		t.Fatalf("Failed to decode jpeg fixture: %v", err)
	}
	avifBytes, err := base64.StdEncoding.DecodeString(avifBase64)
	if err != nil {
		t.Fatalf("Failed to decode avif fixture: %v", err)
	}

	tempDir := t.TempDir()
	writeTestFile(t, filepath.Join(tempDir, "photo1.jpg"), jpegBytes)
	writeTestFile(t, filepath.Join(tempDir, "photo2.JPG"), jpegBytes)
	writeTestFile(t, filepath.Join(tempDir, "photo3.avif"), avifBytes)
	writeTestFile(t, filepath.Join(tempDir, "video.mp4"), nil)

	result := FindAllVideos(tempDir)
	if len(result) != 1 || filepath.Base(result[0]) != "video.mp4" {
		t.Fatalf("Expected only video.mp4, got %v", result)
	}
}
