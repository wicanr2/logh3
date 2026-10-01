# 032 — Linux 隨包音樂工具驗收

日期：2026-10-01，工作項目：[#69](https://github.com/wicanr2/logh3/issues/69)。
規格：[072](../spec/072-linux-music-tools.md)，證據：[研究 360](../re/360-linux-bundled-music-tools.md)。
狀態：Linux 隨包工具技術抽驗通過，未宣稱正式交付完成。
完整包內檔案、220 共享依賴、198 套件授權／來源身份、原版 flags、
產物與解碼結果見 [`linux-music-tools.json`](../internals/data/linux-music-tools.json)。
JSON SHA-256：`b3ba7394ecedb3c2882b56f6cd5df0846d1d2a5d43e16b2cbca735eec21a6f58`。

## 已驗結果

- 包含固定 SF2、原生 importer／probe／啟動器、三個真實工具、全部共享依賴及完整授權，
  總檔案 419,718,884 bytes。不含原版 EXE／MDS 或生成曲庫。
- 乾淨 `python:3.12-slim` 沒有系統 FFmpeg／FluidSynth，PATH 為 `/none`，
  LD_PRELOAD／LD_LIBRARY_PATH 故意指向錯誤檔案；原生入口清理干預後仍可啟動。
  中文／空白路徑搬移通過，載入器解析的全部動態依賴在包內，缺件明確拒絕。
- 三種原版 FM 路徑實際由隨包 probe 擷取，VGM／固定封裝 OGG 全檔符合既有基準；
  一首 GM 的完整 f32le PCM 與 071 相同。這是依賴封裝抽驗，不擴大原版硬體時序聲明。
- 四首由正式播放層完整讀出 PCM 約 2.39 秒，標準產品測試全綠；
  新增壞 JSON 邊界與正式曲庫測試另乾淨重跑通過。
- 198 套件對應 157 個去重來源套件。三個主要工具的同版本 Debian `.dsc` 已實讀並保存雜湊，
  尚未交付全部來源 archive，不把完整著作權索引當成 GPL／LGPL 來源碼交付完成。

## 重生入口

```sh
tools/go.sh test ./cmd/music-tool -count=1 -v
tools/go.sh build -o workplace/music69-linux/music-tool ./cmd/music-tool
tools/go.sh version -m workplace/music69-linux/music-tool > workplace/music69-linux/launcher-buildinfo.txt
bash tools/audio.sh sh /src/tools/scripts/music_linux_collect.sh /work/music69-linux/stage
LOGH3_EXTRA_RO=/home/anr2/cht/logh3 tools/py.sh \
  /work/tools/scripts/music_linux_finalize.py \
  --audio-image '<本輪實讀 image ID>' --git-head '<本輪實讀 Git HEAD>'
LOGH3_TIMEOUT=25m tools/py.sh /work/tools/scripts/music_linux_check.py
tools/gui.sh test ./cmd/gin3 -run TestBundledLinuxMusicImportPlayerDecode -count=1 -v \
  > workplace/music69-linux/player-test.log 2>&1
LOGH3_NET=1 tools/py.sh /work/tools/scripts/music_linux_source_probe.py
tools/py.sh /work/tools/scripts/music_linux_receipt.py
```

收集器只准新目錄，不覆蓋前一份包。音色庫來源準備與原生 importer／probe 前置入口
見[研究 356](../re/356-player-music-library.md)、[驗證 031](031-native-fm-import.md)。
所有分析／建置在容器內，玩家入口仍是原生程式。
驗收器把已完成暫存包移到「封裝 測試」目錄，以無效 PATH 與載入環境干預實跑。
三種 FM 路徑與一首 GM 的固定子集輸出放專屬 `imported/`，不覆蓋舊對拍收據。

## 界限

暫存包保留完整共享依賴、各套件版本／來源身份、著作權與授權全文；
不含原版 EXE／MDS 或生成曲庫。MIT 音色庫隨包且驗固定 SHA-256。
來源索引不代表對應來源碼 archive 已完成正式交付，其他平台與人耳／循環仍待驗收。
