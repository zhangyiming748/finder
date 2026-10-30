# Finder

用 Go 模拟 Unix `find` 命令，递归扫描指定目录并把符合条件的文件 / 文件夹以**绝对路径**字符串切片的形式返回。

适合作为媒体库整理、文件迁移、批量转码等场景的基础组件：调用方只需要拿到路径列表，不需要关心遍历、隐藏目录或格式识别的细节。

## 特性

- 纯 Go 实现，遍历基于 `filepath.Walk` / `os.ReadDir`，单文件错误会被忽略，不中断整体扫描
- 同时提供「递归全部层级」和「仅当前目录」两种粒度的查询
- 图片、音频、视频统一采用「扩展名 → 文件内容」两级识别，扩展名写错的文件也能被正确归类
- 不依赖 ffprobe / ffmpeg 等外部命令，识别结果不随环境安装情况变化
- 返回结果均为绝对路径，便于直接交给后续处理流程

## 环境要求

- Go 1.26.1 及以上
- 依赖 [`github.com/gabriel-vasile/mimetype`](https://github.com/gabriel-vasile/mimetype) v1.4.16（纯 Go、无 cgo，只读取文件头）

```bash
go get github.com/zhangyiming748/finder
```

## 快速开始

```go
package main

import (
	"fmt"

	"github.com/zhangyiming748/finder"
)

func main() {
	videos := finder.FindAllVideos("/path/to/media")
	for _, v := range videos {
		fmt.Println(v)
	}
}
```

返回示例：

```text
/Users/zen/media/movie/01.mp4
/Users/zen/media/movie/02.mkv
/Users/zen/media/series/S01/E01.mov
```

## API

| 函数 | 说明 |
| --- | --- |
| `FindAllFolders(root string) []string` | 递归查找所有层级的文件夹（包含 `root` 本身） |
| `FindAllFoldersInRoot(root string) []string` | 仅查找 `root` 第一层的子文件夹，不包含 `root` 本身 |
| `FindAllFiles(root string) []string` | 递归查找所有层级的文件（结果包含 `root` 本身） |
| `FindAllVideos(root string) []string` | 递归查找所有层级的视频文件 |
| `FindAllMkvVideos(root string) []string` | 递归查找所有层级的 `.mkv` 视频文件 |
| `FindAllVideosInRoot(root string) []string` | 仅查找 `root` 第一层的视频文件，不进入子目录 |
| `FindAllImages(root string) []string` | 递归查找所有层级的图片文件 |
| `FindAllImagesInRoot(root string) []string` | 仅查找 `root` 第一层的图片文件 |
| `FindAllAudios(root string) []string` | 递归查找所有层级的音频文件 |
| `FindAllAudiosInRoot(root string) []string` | 仅查找 `root` 第一层的音频文件 |

所有函数都不会返回 error；遍历过程中无权限的目录、读取失败的单个文件会被跳过，不影响整体扫描。

## 文件识别策略

图片、音频、视频共用同一套两级判断，顺序固定：先看扩展名，再用文件内容兜底。

### 第一级：扩展名

`mediaExtensions` 维护了常见媒体扩展名到类别的映射，匹配时忽略大小写：

- 视频：`.mp4`、`.mkv`、`.avi`、`.mov`、`.wmv`、`.flv`、`.rmvb`、`.ts`、`.m2ts`、`.webm`、`.vob`、`.mxf` 等 50 余种
- 图片：`.jpg`、`.jpeg`、`.png`、`.gif`、`.webp`、`.avif`、`.heic`、`.heif`、`.jxl`、`.jp2`、`.tiff`、`.svg` 等 50 余种
- 音频：`.mp3`、`.flac`、`.wav`、`.aac`、`.m4a`、`.ogg`、`.opus`、`.wma`、`.ape`、`.amr` 等 40 余种

命中即直接归类，不读取文件内容。这一步同时解决三个问题：开销最小；文件头不规范的视频（如某些 AVI）仍能被正确识别；JPG、AVIF 等图片永远不会被 `FindAllVideos` 误判。

### 第二级：文件内容

扩展名不在映射表中时，用 [`github.com/gabriel-vasile/mimetype`](https://github.com/gabriel-vasile/mimetype) 读取文件头识别 MIME，再映射到类别：

| MIME | 归类 |
| --- | --- |
| `video/*`；RealMedia、SWF | 视频 |
| `image/*`（含 AVIF / HEIC / WebP / JPEG XL） | 图片 |
| `audio/*`；`application/ogg` | 音频 |
| 其他（文本、文档、压缩包、未知） | 忽略 |

mimetype 内置了完整的 magic number 规则，能解析 ISO-BMFF（`ftyp`）的 brand、Matroska、AVI 等容器，也能穿透 zip / OLE 识别 Office 文档；同时是纯 Go 实现、无 cgo、不启动子进程，Linux / macOS / Windows 行为一致。因此扩展名写错的文件（例如视频改名成 `.txt`）也能被归到正确类别。

### 修复说明：为什么移除 ffprobe 兜底

**问题现象**：目录里只有 1 个视频和 2 张 JPG，`FindAllVideos` 却返回了 3 个路径；后续还发现纯 TXT 会触发 ffmpeg 处理。

**根因**：旧实现的前两级都没命中时会调用 ffprobe 检测视频流，而 ffprobe 会把静态图片解析成视频流：

```bash
ffprobe -select_streams v:0 -show_entries stream=codec_type -of csv=p=0 photo.jpg
# video   ← image2 解复用器 + mjpeg
ffprobe -select_streams v:0 -show_entries stream=codec_type -of csv=p=0 photo.avif
# video   ← AV1 静帧，ftyp 容器被当成 mov/mp4
```

当时使用的 `filetype` v1.1.3 不认识 AVIF 的 `avif` brand，一张内容为 AVIF、扩展名为 `.jpg` 的文件会在扩展名和文件头两关都漏过，最终被 ffprobe 判成视频。TXT 虽然没有被判成视频，但每个文本文件都会启动一次子进程，扫描大目录时开销明显。

**旧方案的堆叠**：为了堵住这些误报，代码陆续加了视频衍生图片扩展名排除、图片 / 音频 / 文档 / 压缩包 / 字体排除、ftyp brand 解析、纯文本判定，以及 ffprobe 缺失时退回 ffmpeg 的分支。规则越多，维护成本和边界漏洞也越多，而且始终依赖外部可执行文件。

**新方案**：统一为两级判断，代码集中在 [media.go](media.go) 的 `detectMediaKind`：

1. 扩展名命中 `mediaExtensions` 就直接归类；`.jpg`、`.avif`、`.heic` 等一律归为图片，因此图片不会进入视频结果
2. 扩展名未命中时，用 mimetype 读取文件头；文本归为“其他”，视频 / 图片 / 音频各归其类，扩展名写错的文件也能纠正

判定入口保持不变：[video.go](video.go) 的 `isVideo`、[image.go](image.go) 的 `isImage`、[audio.go](audio.go) 的 `isAudio` 都只比较 `detectMediaKind` 的返回值。ffprobe / ffmpeg 调用、告警用的 `sync.Once`、ftyp brand 解析、纯文本判定和文档类排除代码已全部删除，依赖也只剩 mimetype。

### 边界与取舍

- **只判断容器类型，不判断流**：只含音轨的 `.mp4` 仍会被归为视频；扩展名写错且 MIME 无法识别的罕见视频会漏判。
- **第一级扩展名优先**：视频若被改名成 `.jpg`，会按图片处理，不会继续读内容；这样能保证图片永远不会出现在视频结果里。
- **不调用外部工具**：不再使用 ffprobe / ffmpeg。此前用 ffprobe 做最终兜底，但它会把静态图片解析成视频流（`image2` 解复用器 + mjpeg、AVIF 的 AV1 静帧），且每个文件都要启动一次子进程，收益小于代价。
- **识别失败不报错**：文件不存在、无权限、读取失败时返回空类别，由遍历逻辑继续处理下一个文件。

### 回归测试

- `TestFindAllVideosIgnoresImages`：1 个 `.mp4` + 2 张真实 JPEG（含大写 `.JPG`）+ 真实 `.avif` + 内容为 AVIF 的 `.jpg`，只返回 1 个视频
- `TestFindAllVideosIgnoresTextFiles`：纯文本 `.txt` / `.csv` 不会被当成视频
- `TestMediaDetectionFallsBackToMIMEType`：内容分别为 MP4 / JPEG / MP3 / AVIF、扩展名全部写成 `.txt` / `.dat`，验证第二级兜底能正确归类

## 开发过程记录

- **从 `find` 命令出发**：先把常用查询拆成「递归 / 仅当前层」×「文件 / 文件夹 / 媒体类型」的组合，每个函数只做一件事，返回 `[]string` 而不是流式回调，调用方组合起来更简单。
- **遍历容错**：`filepath.Walk` 回调中对 `err` 直接返回 `nil`，保证遇到无权限目录或损坏文件时继续扫描，而不是整体失败。
- **为什么先看扩展名**：实际使用中遇到过文件头不规范的 AVI，magic number 识别失败但文件可正常播放；扩展名判断既快又能覆盖这类文件。
- **视频识别的演进**：最初只靠 `filetype` + 少量无法识别的扩展名硬编码；遇到不规范 AVI 后改为「扩展名优先」；之后为解决 JPG / AVIF / TXT 误判，一度引入 ffprobe / ffmpeg 流探测和多种手工排除规则，规则越堆越多。
- **最终改为 mimetype 两级判断**：ffprobe 会把静态图片当成视频，而且需要外部进程；`filetype` 停更多年、无法识别 AVIF 等新格式。换成 mimetype 后，扩展名未命中时由内容识别兜底，同时删除了 ftyp brand 解析、纯文本判定、文档 / 压缩包 / 字体排除和全部 ffprobe / ffmpeg 代码，依赖也减少了一个。
- **测试**：`module_test.go` 使用 `t.TempDir()` 构造临时目录树，覆盖递归 / 仅当前层两种行为；媒体测试写入最小的真实文件头（PNG / JPEG / GIF、ID3 / RIFF-WAVE / OggS、MP4 / AVIF 的 ftyp 头），既验证扩展名判断，也验证内容兜底。

## 项目结构

```text
.
├── generic.go       # 文件夹、文件遍历
├── media.go         # 扩展名映射、mimetype 内容识别
├── image.go         # 图片识别与查找
├── audio.go         # 音频识别与查找
├── video.go         # 视频识别与查找
├── module_test.go   # 单元测试
├── go.mod
└── go.sum
```

## 测试

```bash
go test ./...
```

## 许可证

[Apache License 2.0](LICENSE)
