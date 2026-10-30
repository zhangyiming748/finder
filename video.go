package finder

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/h2non/filetype"
	"github.com/h2non/filetype/matchers/isobmff"
)

// 常见视频扩展名，第一步仅凭扩展名即可认定是视频文件
var videoExtensions = map[string]struct{}{
	".3g2": {}, ".3gp": {}, ".amv": {}, ".asf": {}, ".avi": {}, ".divx": {},
	".dv": {}, ".f4v": {}, ".flv": {}, ".gvi": {}, ".gxf": {}, ".hevc": {},
	".ismv": {}, ".m1v": {}, ".m2ts": {}, ".m2v": {}, ".m4v": {}, ".mkv": {},
	".mov": {}, ".mp2v": {}, ".mp4": {}, ".mp4v": {}, ".mpg": {}, ".mpe": {},
	".mpeg": {}, ".mpv": {}, ".mts": {}, ".mxf": {}, ".mxg": {}, ".nsv": {},
	".nut": {}, ".nuv": {}, ".ogm": {}, ".ogv": {}, ".ogx": {}, ".qt": {},
	".rm": {}, ".rmvb": {}, ".swf": {}, ".tod": {}, ".tp": {}, ".trp": {},
	".ts": {}, ".tts": {}, ".vob": {}, ".vro": {}, ".webm": {}, ".wm": {},
	".wmv": {}, ".wtv": {}, ".xesc": {}, ".y4m": {},
}

// 本质是图片、但由视频编码衍生或被 ffprobe/ffmpeg 解析成视频流的格式，兜底前必须排除
var videoBasedImageExtensions = map[string]struct{}{
	".avif": {}, ".avifs": {}, // AV1 静帧
	".heic": {}, ".heif": {}, ".heics": {}, ".heifs": {}, // HEVC 静帧
	".webp": {},             // VP8/VP9
	".jxl":  {},             // JPEG XL
	".jp2":  {}, ".jpx": {}, // JPEG 2000
	".mjpg": {}, ".mjpeg": {}, ".mpo": {}, // 单帧 Motion JPEG
	".apng": {}, // PNG 动画
	".gif":  {}, // GIF 动画
}

// ISO-BMFF（ftyp）容器中的静态图片 brand。AVIF / HEIF 即使扩展名被改成 .jpg，
// filetype 也识别不出是图片，最终会被 ffprobe 当成视频流，因此需要自行解析 ftyp
var stillImageBrands = map[string]struct{}{
	"avif": {}, "avis": {}, "heic": {}, "heix": {}, "heim": {}, "heis": {},
	"hevm": {}, "hevs": {}, "mif1": {}, "msf1": {},
}

func isStillImageContainer(head []byte) bool {
	if !isobmff.IsISOBMFF(head) {
		return false
	}
	majorBrand, _, compatibleBrands := isobmff.GetFtyp(head)
	if _, ok := stillImageBrands[majorBrand]; ok {
		return true
	}
	for _, brand := range compatibleBrands {
		if _, ok := stillImageBrands[brand]; ok {
			return true
		}
	}
	return false
}

// 文件头是纯文本（TXT、CSV、Markdown、字幕等）时不可能是视频，
// 直接排除，避免为每个文本文件启动一次 ffprobe / ffmpeg
func isTextContent(head []byte) bool {
	if len(head) == 0 {
		return false
	}
	for _, b := range head {
		switch {
		case b == '\t' || b == '\n' || b == '\r' || b == '\v' || b == '\f':
		case b >= 0x20 && b <= 0x7e:
		case b >= 0x80: // UTF-8 的多字节字符
		default:
			return false
		}
	}
	return true
}

/*
使用golang实现find <root> -type f 命令
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
使用golang实现find <root> -type f -name "*.mkv" 命令
*/
func FindAllMkvVideos(root string) []string {
	var files []string
	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // 忽略错误，继续遍历
		}
		if !info.IsDir() {
			absPath, _ := filepath.Abs(path)
			if isVideo(absPath) {
				if strings.ToLower(filepath.Ext(absPath)) == ".mkv" {
					files = append(files, absPath)
				}
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
	// 第一步：忽略大小写按扩展名判断，开销最小，也能兼容文件头不规范的视频
	ext := strings.ToLower(filepath.Ext(fp))
	if _, ok := videoExtensions[ext]; ok {
		return true
	}
	// 视频编码衍生的图片（AVIF、HEIC 等）会被 ffprobe 当作视频流，这里直接排除
	if _, ok := videoBasedImageExtensions[ext]; ok {
		return false
	}
	// 第二步：读取文件头，用 magic number 判断
	file, err := os.Open(fp)
	if err == nil {
		defer file.Close()
		head := make([]byte, 261)
		n, _ := file.Read(head)
		head = head[:n]
		if n == 0 {
			return false
		}
		if filetype.IsVideo(head) {
			return true
		}
		// 图片（JPG/PNG 等）和音频（带封面图的 MP3 等）会被 ffprobe 识别出视频流，
		// 文件头已明确判定类型的直接排除，避免兜底步骤误报
		if filetype.IsImage(head) || filetype.IsAudio(head) ||
			filetype.IsDocument(head) || filetype.IsArchive(head) || filetype.IsFont(head) ||
			isStillImageContainer(head) || isTextContent(head) {
			return false
		}
	}
	// 第三步：ffprobe 兜底，直接检测文件里是否包含视频流，开销最大所以放最后
	return hasVideoStream(fp)
}

// hasVideoStream 使用 ffprobe 检测文件中是否存在视频流；ffprobe 不存在时告警并退回 ffmpeg
func hasVideoStream(fp string) bool {
	if ffprobe, err := exec.LookPath("ffprobe"); err == nil {
		out, _ := exec.Command(ffprobe, "-v", "error", "-select_streams", "v:0", "-show_entries", "stream=codec_type", "-of", "csv=p=0", fp).Output()
		return strings.TrimSpace(string(out)) == "video"
	}
	warnFfprobeMissingOnce.Do(func() {
		log.Printf("警告：未找到 ffprobe，已退回使用 ffmpeg 检测视频流")
	})
	ffmpeg, err := exec.LookPath("ffmpeg")
	if err != nil {
		warnFfmpegMissingOnce.Do(func() {
			log.Printf("警告：未找到 ffmpeg，无法检测视频流，相关文件将被跳过")
		})
		return false
	}
	// ffmpeg 没有输出文件时必定以非 0 退出且信息都写在 stderr，所以只能解析输出内容
	out, _ := exec.Command(ffmpeg, "-hide_banner", "-i", fp).CombinedOutput()
	return strings.Contains(string(out), "Video:")
}

var (
	warnFfprobeMissingOnce sync.Once
	warnFfmpegMissingOnce  sync.Once
)
