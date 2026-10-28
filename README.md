# Finder

用 Go 模拟 Unix `find` 命令，递归扫描指定目录并把符合条件的文件 / 文件夹以**绝对路径**字符串切片的形式返回。

适合作为媒体库整理、文件迁移、批量转码等场景的基础组件：调用方只需要拿到路径列表，不需要关心遍历、隐藏目录或格式识别的细节。

## 特性

- 纯 Go 实现，遍历基于 `filepath.Walk` / `os.ReadDir`，单文件错误会被忽略，不中断整体扫描
- 同时提供「递归全部层级」和「仅当前目录」两种粒度的查询
- 图片、音频通过文件头（magic number）识别，不依赖扩展名，可识别被改名的文件
- 视频采用「扩展名 → 文件头 → ffprobe → ffmpeg」四级判断，兼顾速度与准确率
- 返回结果均为绝对路径，便于直接交给后续处理流程

## 环境要求

- Go 1.26.1 及以上
- 依赖 [`github.com/h2non/filetype`](https://github.com/h2non/filetype) v1.1.3（仅用于文件头识别）
- 视频识别兜底需要 `ffprobe` 或 `ffmpeg`，二选一即可，不安装也能正常工作（见下文「视频识别策略」）

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

### 图片

1. `.avif` 通过扩展名直接判定（其 magic number 比较特殊，`filetype` 无法识别）
2. 其余格式读取文件前 261 字节，交给 `filetype.IsImage` 判断 magic number

### 音频

读取文件前 261 字节，交给 `filetype.IsAudio` 判断 magic number。

### 视频（四级判断，按开销从小到大）

1. **扩展名**：忽略大小写匹配常见视频扩展名（`.mp4`、`.mkv`、`.avi`、`.mov`、`.wmv`、`.flv`、`.rmvb`、`.ts`、`.m2ts`、`.webm` 等 50 余种），命中即返回
2. **文件头**：读取前 261 字节，交给 `filetype.IsVideo` 判断 magic number
3. **ffprobe**：执行 `ffprobe -select_streams v:0`，输出 `video` 即认为包含视频流
4. **ffmpeg**：ffprobe 不存在时打印一次告警，改用 `ffmpeg -i`，从 stderr 中匹配 `Video:` 判断

之所以把扩展名放在第一步，是因为实际使用中遇到过文件头不规范的 AVI：magic number 识别失败，但文件本身确实可以正常播放。先看扩展名可以避免这类误判，同时这也是最廉价的一步。真正的流探测（3、4）计算和 IO 开销最大，放在最后，只有前两步都没命中时才会执行。

> ffmpeg 在没有输出文件时会以非 0 退出码结束，所以第四步只解析输出内容，不依赖退出码。
> 两个告警通过 `sync.Once` 只输出一次，避免扫描大目录时刷屏。

## 开发过程记录

- **从 `find` 命令出发**：先把常用查询拆成「递归 / 仅当前层」× 「文件 / 文件夹 / 媒体类型」的组合，每个函数只做一件事，返回 `[]string` 而不是流式回调，调用方组合起来更简单。
- **遍历容错**：`filepath.Walk` 回调中对 `err` 直接返回 `nil`，保证遇到无权限目录或损坏文件时继续扫描，而不是整体失败。
- **为什么用 magic number**：只信任扩展名会漏掉改名文件，也容易被误改后缀的文件骗过；读取文件头是成本较低的折中方案。`filetype` 只读取固定长度的头部，不需要完整加载文件。
- **视频识别的演进**：最初只靠 `filetype` + 少量无法识别的扩展名硬编码。遇到不规范 AVI 后改为「扩展名优先、文件头其次、外部工具兜底」的分层策略，既保证常见场景的速度，也覆盖了文件头异常的文件。
- **外部工具兜底**：`ffprobe` 是首选，因为输出结构化、易解析；考虑到存在只安装了 `ffmpeg` 的环境，增加了 ffmpeg 回退分支，并在 `PATH` 中找不到可执行文件时给出告警而不是静默跳过。
- **测试**：`module_test.go` 使用 `t.TempDir()` 构造临时目录树，覆盖递归 / 仅当前层两种行为；媒体测试使用真实扩展名的空文件验证路径筛选逻辑。

## 项目结构

```text
.
├── generic.go       # 文件夹、文件遍历
├── image.go         # 图片识别与查找
├── audio.go         # 音频识别与查找
├── video.go         # 视频识别与查找（含 ffprobe / ffmpeg 兜底）
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
