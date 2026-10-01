# 360 — Linux 音樂工具完整依賴與隨包載入器

日期：2026-10-01，工作項目：[#69](https://github.com/wicanr2/logh3/issues/69)。
沿用 `logh3-audio:2`，映像檔身份及三個工具 SHA-256 見
[前輪完整收據](../internals/data/native-fm-import.json)。沒有新建替代音訊 image。

## 平台契約與實物

GNU libc 可明示啟動自身載入器，程式路徑必須完整；來源：
[GNU 動態載入器入口](https://sourceware.org/glibc/manual/latest/html_node/Dynamic-Linker-Invocation.html)。
`--library-path` 可指定共享依賴目錄，`--inhibit-cache` 不用系統 cache，
`--list` 列出實際解析結果；來源：[Linux ld.so 手冊](https://man7.org/linux/man-pages/man8/ld.so.8.html)。
同 image 的 glibc 2.36-9+deb12u14 `--help` 也實際列出這三項選項，非猜測舊版本支援。

`ldd` 三工具共有 220 個獨立共享檔名，完整依賴 248,793,704 bytes。
含 private `libpulsecommon-16.1.so` 與 glibc 載入器；只複製工具本體不足。
FFmpeg 7:5.1.9-0+deb12u1、FluidSynth 2.3.1-2、AdPlug 2.3.3+dfsg-2，
與前輪音樂雜湊驗收使用相同實物，沒有更換合成／編碼版本。

## 可丟棄探針結果

將工具與共享依賴複製至 `workplace/music69-linux/probe/`，原始 image 檔案保持。
在既有 `python:3.12-slim` 容器以另一掛載前綴 `/work/workplace/` 啟動：
系統 `/usr/bin/ffmpeg`／`fluidsynth` 均不存在，PATH 設為 `/none`。
隨包載入器的 `--list` 證實 FFmpeg 214、FluidSynth 64、opl2wav 7 項解析全部在
隨包目錄內；版本／用法入口成功。完整結果保存本機 `probe-result.json`，後續正式收據綁其雜湊。
首次探針誤將 opl2wav 用法錯誤預期為 exit 1；原始碼是 exit 2，修正助手後同 image 重跑通過。

等級：**已證實**，限此工具版本、Linux amd64 與乾淨容器依賴解析。
尚未證明完整隨包匯入、其他 Linux 發行版、Windows／macOS 或實機聆聽完成。

## 正式啟動器與匯入結果

原生 Go 啟動器與收集／合併器已依 READY 072 實作。完整版暫存包包含固定 MIT SF2、
220 共享檔案與 198 套件完整著作權，總檔案 419,718,884 bytes。
搬移至中文／空白路徑後，在同一無系統音樂工具的容器清理注入的 LD_*，
三個工具啟動及全部依賴解析成功；缺 libc／未知名稱拒絕，不依賴系統補件。
三種 FM 觸發路徑的 VGM／OGG 全檔與 071 一致，一首 GM 完整 PCM 相同；
正式播放層再完整讀出四首 PCM，產品測試全綠。完整身份與收據見
[驗證 032](../playtest/032-linux-music-tools.md)。

三個主要工具同版本 Debian `.dsc` 已實讀，URL 與 SHA-256 一併存入結構化收據。
198 套件合併為 157 個來源套件，完整來源 archive 仍待正式交付；不聲稱三份 `.dsc`
涵蓋全部共享依賴。可丟棄大檔副本清除，保留探針結果雜湊與正式可重現建置入口。

## 授權與交付停止線（正式封包仍未完成）

原版輸入與生成曲庫不得放進可公開工具包。音色庫保持已確認 MIT，
固定 archive／SHA-256 與完整授權沿用[研究 356](356-player-music-library.md)。
各共享依賴保留自身套件版本、來源套件／版本及完整著作權文件，不改成 remake 授權。
FFmpeg 含選配 GPL 組件時整體採 GPL，且散布須提供精確對應來源；來源：
[FFmpeg 官方授權說明](https://ffmpeg.org/legal.html)。
建置器的套件／著作權索引不等於對應來源碼已隨正式交付，公開封包 gate 須另核對。

2026-10-01 接續：157 個同版本來源集合的完整取得、離線重驗與本機來源 tar 已完成，
見[研究 361](361-linux-music-source-closure.md)；正式版交付與其餘平台門檻仍未完成。
