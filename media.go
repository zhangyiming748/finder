package finder

import (
	"path/filepath"
	"strings"

	"github.com/gabriel-vasile/mimetype"
)

// mediaKind 表示文件属于哪一类媒体
type mediaKind int

const (
	mediaUnknown mediaKind = iota
	mediaVideo
	mediaImage
	mediaAudio
)

// 常见媒体扩展名，第一道关卡只按扩展名判断，不读取文件内容
var mediaExtensions = map[string]mediaKind{
	// 视频
	".3g2": mediaVideo, ".3gp": mediaVideo, ".amv": mediaVideo, ".asf": mediaVideo,
	".avi": mediaVideo, ".divx": mediaVideo, ".dv": mediaVideo, ".f4v": mediaVideo,
	".flv": mediaVideo, ".gvi": mediaVideo, ".gxf": mediaVideo, ".hevc": mediaVideo,
	".ismv": mediaVideo, ".m1v": mediaVideo, ".m2t": mediaVideo, ".m2ts": mediaVideo,
	".m2v": mediaVideo, ".m4v": mediaVideo, ".mj2": mediaVideo, ".mkv": mediaVideo,
	".mov": mediaVideo, ".mp2v": mediaVideo, ".mp4": mediaVideo, ".mp4v": mediaVideo,
	".mpe": mediaVideo, ".mpeg": mediaVideo, ".mpg": mediaVideo, ".mpv": mediaVideo,
	".mts": mediaVideo, ".mxf": mediaVideo, ".mxg": mediaVideo, ".nsv": mediaVideo,
	".nut": mediaVideo, ".nuv": mediaVideo, ".ogm": mediaVideo, ".ogv": mediaVideo,
	".ogx": mediaVideo, ".qt": mediaVideo, ".rm": mediaVideo, ".rmvb": mediaVideo,
	".rv": mediaVideo, ".swf": mediaVideo, ".tod": mediaVideo, ".tp": mediaVideo,
	".trp": mediaVideo, ".ts": mediaVideo, ".tts": mediaVideo, ".vob": mediaVideo,
	".vro": mediaVideo, ".webm": mediaVideo, ".wm": mediaVideo, ".wmv": mediaVideo,
	".wtv": mediaVideo, ".xesc": mediaVideo, ".y4m": mediaVideo,

	// 图片
	".apng": mediaImage, ".arw": mediaImage, ".avif": mediaImage, ".avifs": mediaImage,
	".bmp": mediaImage, ".cr2": mediaImage, ".cr3": mediaImage, ".cur": mediaImage,
	".dds": mediaImage, ".dib": mediaImage, ".dng": mediaImage, ".exr": mediaImage,
	".gif": mediaImage, ".hdr": mediaImage, ".heic": mediaImage, ".heics": mediaImage,
	".heif": mediaImage, ".heifs": mediaImage, ".ico": mediaImage, ".j2k": mediaImage,
	".jfif": mediaImage, ".jif": mediaImage, ".jpe": mediaImage, ".jpeg": mediaImage,
	".jpf": mediaImage, ".jpg": mediaImage, ".jp2": mediaImage, ".jpx": mediaImage,
	".jxl": mediaImage, ".nef": mediaImage, ".orf": mediaImage, ".pbm": mediaImage,
	".pef": mediaImage, ".pgm": mediaImage, ".png": mediaImage, ".pnm": mediaImage,
	".ppm": mediaImage, ".psb": mediaImage, ".psd": mediaImage, ".raf": mediaImage,
	".raw": mediaImage, ".rw2": mediaImage, ".srw": mediaImage, ".svg": mediaImage,
	".tga": mediaImage, ".tif": mediaImage, ".tiff": mediaImage, ".webp": mediaImage,
	".xpm": mediaImage,

	// 音频
	".aac": mediaAudio, ".ac3": mediaAudio, ".aif": mediaAudio, ".aifc": mediaAudio,
	".aiff": mediaAudio, ".amr": mediaAudio, ".ape": mediaAudio, ".au": mediaAudio,
	".awb": mediaAudio, ".caf": mediaAudio, ".dff": mediaAudio, ".dsf": mediaAudio,
	".dts": mediaAudio, ".eac3": mediaAudio, ".flac": mediaAudio, ".kar": mediaAudio,
	".m4a": mediaAudio, ".m4b": mediaAudio, ".m4p": mediaAudio, ".mka": mediaAudio,
	".mid": mediaAudio, ".midi": mediaAudio, ".mp3": mediaAudio, ".mpc": mediaAudio,
	".oga": mediaAudio, ".ogg": mediaAudio, ".opus": mediaAudio, ".ra": mediaAudio,
	".ram": mediaAudio, ".snd": mediaAudio, ".spx": mediaAudio, ".tta": mediaAudio,
	".wav": mediaAudio, ".wave": mediaAudio, ".wma": mediaAudio, ".wv": mediaAudio,
}

// detectMediaKind 先按扩展名判断，扩展名不认识时再用文件内容（mimetype）兜底，
// 用于识别扩展名写错的文件
func detectMediaKind(fp string) mediaKind {
	if kind, ok := mediaExtensions[strings.ToLower(filepath.Ext(fp))]; ok {
		return kind
	}
	mtype, err := mimetype.DetectFile(fp)
	if err != nil {
		return mediaUnknown
	}
	return mediaKindByMIME(mtype.String())
}

func mediaKindByMIME(mime string) mediaKind {
	switch {
	case strings.HasPrefix(mime, "video/"):
		return mediaVideo
	case mime == "application/vnd.rn-realmedia", mime == "application/vnd.rn-realmedia-vbr",
		mime == "application/x-shockwave-flash":
		return mediaVideo
	case strings.HasPrefix(mime, "image/"):
		return mediaImage
	case strings.HasPrefix(mime, "audio/"):
		return mediaAudio
	case mime == "application/ogg", mime == "application/x-ogg":
		return mediaAudio
	default:
		return mediaUnknown
	}
}
