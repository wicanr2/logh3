# 034 — Windows 原生曲庫與 FM 無損階段驗收

日期：2026-10-01，工作項目：[#69](https://github.com/wicanr2/logh3/issues/69)。
契約：[規格 074](../spec/074-cross-platform-fm-pcm.md)，過程：[研究 362](../re/362-windows-music-tools.md)。

已證實（proven）：Windows 原生匯入器重新執行原版，完整十一首 FM 的 VGM、
合成 WAV 與 -1.00 dB 後 f32le PCM 全數符合 Linux 固定基準。
十一首 OGG 都與 Linux 不同；這是使用者允許的平台差異，沒有宣稱有損解碼結果相同。
BGM_0F、END_2F、OPEN_F 三種觸發路徑在獨立暫存狀態再跑一次，
VGM／WAV／增益後 PCM／OGG 都與同一 Windows 工具首次輸出相同。
Linux 三路抽驗亦通過新無損閘門與既有 OGG 全檔基準。

正式播放器完整解碼 Windows FM／GM 各十一首，正常玩家狀態的選曲測試通過；
新增曲庫檔案與收據雜湊、完整無損欄位及錯平台／未完成批次檢查。
無損產物變更、缺檔及缺鎖的拒絕測試通過。
標準產品命令 `tools/gui.sh test ./internal/... ./cmd/...` 全部通過，部分套件使用快取。
正式解碼與正常選曲另以 `-count=1` 重跑，約 11.35 秒；標準 cmd/gin3 約 30.43 秒。

## 收據與身份

正式技術收據：[windows-fm-pcm.json](../internals/data/windows-fm-pcm.json)，SHA-256：
`4fc93a59c4a9b1f65b9e1a0114bc919126471b0d13fcef9102fdc29df93bac5a`。
舊門檻反例與 GM 探針：[windows-music-probes.json](../internals/data/windows-music-probes.json)，
SHA-256：`a8fca6e5966018bd60d0aa52a6bd1756343150aceaec19a78aabf3818665fb2e`。
十一首唯一無損基準：[fm-import-lock.json](../internals/data/fm-import-lock.json)，
SHA-256：`ac3805f650740002ae5aeadf11955d8e06b5db4d66d2f72e7cc6c1d16a95d939`。

原版 bundle SHA-256 沿用 `de905aa0b5a666aa9d31dce36f43a5584f1842950c39d7e1b1f97c72e0d79673`。
專用 dosgolem clean HEAD 為 `f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c`，
Windows probe：`d9e2243757e5c78ef70ed890c706abd4f2ab1ae76ce341bd259bf73dc23b9f90`。
新匯入器：`bb3d765eb75e3a1d1a0ba4f6619bb0aefdcfb9cc79e795f192829bd75334344b`；
Go 1.24.13、來源父提交 456098c、工作樹有本輪修改。完整建置資訊保存於 JSON，未冒稱乾淨提交。
Windows 新工具的十三個程式／DLL 逐檔與引用閉合檢查通過；
不是 SDK 與不同 C++ runtime 混裝，亦未加入原版遊戲、音樂或音色庫到 FM 試驗包。

交叉建置 image：`sha256:0745d96e4029b152f468e2fee63c01e20bd859a369efe80de681e95cf17f0541`；
Wine image：`sha256:7d3c7ff554db39300e3e7a92a7b79c2d5fcd647289596d22fa10899baac61201`。
Linux 音訊、Go、GUI、Python 的身份、完整來源及輸出雜湊、每曲 flags／poke／工具命令均在收據。
Windows 使用 C.UTF-8 保證 Wine 中文／空白 argv 正常，網路關閉，原版目錄唯讀。

## 重生入口

先按[規格 073](../spec/073-linux-music-source-closure.md) 備妥固定來源快取；
本輪編譯配方為 `tools/Dockerfile.music-cross` 與 `tools/scripts/music_windows_fm_probe.py`。
保留舊批次後，從乾淨來源／prefix 重建，不覆寫原版或既有收據。

```sh
bash tools/music_cross.sh --build
bash tools/music_cross.sh python3 /tools/scripts/music_windows_fm_probe.py
bash tools/music_cross.sh --probe
bash tools/audio.sh bash /src/tools/scripts/music_fm_pcm_baseline.sh
LOGH3_EXTRA_RO=/home/anr2/cht/logh3 tools/py.sh tools/scripts/music_fm_lock.py
```

生成的 Go 表格在 `workplace/music69-fm/fmdata.go`，核對後放回 `cmd/music-import/fmdata.go`。
以 `tools/gui.sh xrun` 在容器設定 GOOS=windows、GOARCH=amd64、CGO_ENABLED=0，
建置 `cmd/music-import`；將新程式、固定 probe、prefix/bin 的必要工具與 DLL
複製到獨立中文／空白測試目錄。確切包內檔案清單與 SHA-256 見收據。
完整實跑與重跑命令如下，收據另存 Linux 回歸的完整命令：

```sh
LOGH3_TIMEOUT=3h bash tools/music_wine.sh wine 'Z:\out\pcm-gate\封裝 測試\music-import.exe' \
  -source 'Z:\orig' -output 'Z:\out\fm-full' -fm-only
LOGH3_TIMEOUT=1h bash tools/music_wine.sh wine 'Z:\out\pcm-gate\封裝 測試\music-import.exe' \
  -source 'Z:\orig' -output 'Z:\out\fm-repeat' -fm-only -tracks BGM_0F,END_2F,OPEN_F
tools/gui.sh test ./cmd/gin3 \
  -run 'TestWindows(FM|GM)ImportPlayerDecode|TestPlayerPathMusicSwitches' -count=1 -v
tools/gui.sh test ./internal/... ./cmd/...
LOGH3_EXTRA_RO=/home/anr2/cht/logh3 tools/py.sh tools/scripts/music_windows_fm_receipt.py
```

`tools/music_wine_entry.sh` 對 Xvfb／Wine 設清理 trap。驗證腳本要求完成十一首、
三路獨立重跑與產品測試，才寫通過收據；不讀進行中批次冒充完成。

## 邊界

OPL 串來源已重現，硬體時間與合成仍為 hardware-spec approximation；OPEN_F 保留
substituted 標記。GM 沿用既有近似，官方 SDK 的完整來源／授權核對仍缺，部分峰值 1.0。
Wine 原生 CLI 與 Linux 正式播放器解碼，不等於 Windows 實機音效裝置、人耳或循環接點驗收。
macOS、完整來源／授權封裝與正式 dist-all 交付仍未完成；#69 保持開啟。
原版衍生 VGM／OGG 與中間 WAV／PCM 都只留本機 workplace，不加入本輪 Git。
