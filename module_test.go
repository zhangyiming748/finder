package finder

import (
	"os"
	"path/filepath"
	"testing"
)

// 创建测试文件的辅助函数
func createTestFile(t *testing.T, dir, filename, content string) string {
	t.Helper()
	path := filepath.Join(dir, filename)
	err := os.WriteFile(path, []byte(content), 0644)
	if err != nil {
		t.Fatal(err)
	}
	return path
}

// 创建测试目录的辅助函数
func createTestDir(t *testing.T, dir, name string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	err := os.MkdirAll(path, 0755)
	if err != nil {
		t.Fatal(err)
	}
	return path
}

func TestFindAudio(t *testing.T) {
	// 创建临时目录
	tempDir := t.TempDir()
	
	// 创建一些测试文件
	audioDir := createTestDir(t, tempDir, "audio")
	
	// 创建有效的音频文件 (使用最小的WAV文件头)
	wavContent := []byte("RIFF\x24\x00\x00\x00WAVEfmt \x10\x00\x00\x00\x01\x00\x01\x00D\xac\x00\x00\x88X\x01\x00\x02\x00\x10\x00data\x00\x00\x00\x00")
	createTestFile(t, audioDir, "test1.wav", string(wavContent))
	
	// 创建非音频文件
	createTestFile(t, audioDir, "test.txt", "this is a text file")
	
	// 创建子目录和更多文件
	subDir := createTestDir(t, audioDir, "sub")
	createTestFile(t, subDir, "test2.wav", string(wavContent))
	
	// 调用函数
	result, err := FindAudio(tempDir)
	if err != nil {
		t.Errorf("FindAudio returned error: %v", err)
	}
	
	// 验证结果
	if len(result) != 2 {
		t.Errorf("Expected 2 audio files, got %d", len(result))
	}
	
	// 验证返回的路径是绝对路径
	for _, path := range result {
		if !filepath.IsAbs(path) {
			t.Errorf("Path is not absolute: %s", path)
		}
	}
}

func TestFindImage(t *testing.T) {
	// 创建临时目录
	tempDir := t.TempDir()
	
	// 创建图片目录
	imageDir := createTestDir(t, tempDir, "images")
	
	// 创建有效的图片文件 (使用最小的PNG文件头)
	pngContent := []byte("\x89PNG\r\n\x1a\n\x00\x00\x00\rIHDR\x00\x00\x00\x01\x00\x00\x00\x01\x08\x06\x00\x00\x00\x1f\x15\xc4\x89\x00\x00\x00\nIDATx\x9cc\x00\x01\x00\x00\x05\x00\x01\r\n-\xb4\x00\x00\x00\x00IEND\xaeB`\x82")
	createTestFile(t, imageDir, "test1.png", string(pngContent))
	
	// 创建非图片文件
	createTestFile(t, imageDir, "test.txt", "this is a text file")
	
	// 创建子目录和更多文件
	subDir := createTestDir(t, imageDir, "sub")
	createTestFile(t, subDir, "test2.png", string(pngContent))
	
	// 调用函数
	result, err := FindImage(tempDir)
	if err != nil {
		t.Errorf("FindImage returned error: %v", err)
	}
	
	// 验证结果
	if len(result) != 2 {
		t.Errorf("Expected 2 image files, got %d", len(result))
	}
	
	// 验证返回的路径是绝对路径
	for _, path := range result {
		if !filepath.IsAbs(path) {
			t.Errorf("Path is not absolute: %s", path)
		}
	}
}

func TestFindVideo(t *testing.T) {
	// 创建临时目录
	tempDir := t.TempDir()
	
	// 创建视频目录
	videoDir := createTestDir(t, tempDir, "videos")
	
	// 创建有效的视频文件 (使用最小的MP4文件头)
	mp4Content := []byte("\x00\x00\x00\x18ftypmp42\x00\x00\x00\x00mp42isom<\x00\x00\x00\x01mdat")
	createTestFile(t, videoDir, "test1.mp4", string(mp4Content))
	
	// 创建非视频文件
	createTestFile(t, videoDir, "test.txt", "this is a text file")
	
	// 创建子目录和更多文件
	subDir := createTestDir(t, videoDir, "sub")
	createTestFile(t, subDir, "test2.mp4", string(mp4Content))
	
	// 调用函数
	result, err := FindVideo(tempDir)
	if err != nil {
		t.Errorf("FindVideo returned error: %v", err)
	}
	
	// 验证结果
	if len(result) != 2 {
		t.Errorf("Expected 2 video files, got %d", len(result))
	}
	
	// 验证返回的路径是绝对路径
	for _, path := range result {
		if !filepath.IsAbs(path) {
			t.Errorf("Path is not absolute: %s", path)
		}
	}
}

func TestFindAudioEmptyDirectory(t *testing.T) {
	// 创建空的临时目录
	tempDir := t.TempDir()
	
	// 调用函数
	result, err := FindAudio(tempDir)
	if err != nil {
		t.Errorf("FindAudio returned error: %v", err)
	}
	
	// 验证结果为空
	if len(result) != 0 {
		t.Errorf("Expected 0 audio files, got %d", len(result))
	}
}

func TestFindImageEmptyDirectory(t *testing.T) {
	// 创建空的临时目录
	tempDir := t.TempDir()
	
	// 调用函数
	result, err := FindImage(tempDir)
	if err != nil {
		t.Errorf("FindImage returned error: %v", err)
	}
	
	// 验证结果为空
	if len(result) != 0 {
		t.Errorf("Expected 0 image files, got %d", len(result))
	}
}

func TestFindVideoEmptyDirectory(t *testing.T) {
	// 创建空的临时目录
	tempDir := t.TempDir()
	
	// 调用函数
	result, err := FindVideo(tempDir)
	if err != nil {
		t.Errorf("FindVideo returned error: %v", err)
	}
	
	// 验证结果为空
	if len(result) != 0 {
		t.Errorf("Expected 0 video files, got %d", len(result))
	}
}

func TestFindAudioNonExistentDirectory(t *testing.T) {
	// 使用不存在的目录
	nonExistentDir := filepath.Join(t.TempDir(), "nonexistent")
	
	// 调用函数
	result, err := FindAudio(nonExistentDir)
	if err == nil {
		t.Error("Expected error for non-existent directory, got nil")
	}
	
	// 验证结果为空
	if len(result) != 0 {
		t.Errorf("Expected 0 files, got %d", len(result))
	}
}

func TestFindImageNonExistentDirectory(t *testing.T) {
	// 使用不存在的目录
	nonExistentDir := filepath.Join(t.TempDir(), "nonexistent")
	
	// 调用函数
	result, err := FindImage(nonExistentDir)
	if err == nil {
		t.Error("Expected error for non-existent directory, got nil")
	}
	
	// 验证结果为空
	if len(result) != 0 {
		t.Errorf("Expected 0 files, got %d", len(result))
	}
}

func TestFindVideoNonExistentDirectory(t *testing.T) {
	// 使用不存在的目录
	nonExistentDir := filepath.Join(t.TempDir(), "nonexistent")
	
	// 调用函数
	result, err := FindVideo(nonExistentDir)
	if err == nil {
		t.Error("Expected error for non-existent directory, got nil")
	}
	
	// 验证结果为空
	if len(result) != 0 {
		t.Errorf("Expected 0 files, got %d", len(result))
	}
}