# 035 — Windows GM 固定來源與中文路徑技術驗收

日期：2026-10-02（實跑於 2026-10-01），工作項目：[#69](https://github.com/wicanr2/logh3/issues/69)。
契約：[規格 075](../spec/075-windows-gm-unicode.md)，過程與勘誤：[研究 363](../re/363-windows-music-source-build.md)。

已證實（proven）：固定來源 FluidSynth 2.3.1 保持 float DSP，修正 Windows 命令列與
檔案渲染的 Unicode 轉碼後，包內工具、音色庫、MIDI 與輸出均可使用中文／空白路徑。
原生匯入器完成十一首 GM；每首 MIDI、時長、峰值及完整 Vorbis 音訊封包與未修改
官方 SDK 的英文路徑基準相同。這裡比較的是音訊封包，不將 OGG 容器全檔視為相同。

BGM_0M、END_2M、OPEN_M 三首另保存完整 WAV，與 SDK 逐位元組相同；PCM 長度及
全部樣本相同，BGM_0M 兩側各自獨立重跑亦相同。不把三首 WAV 抽驗外推十一首 WAV。
正式播放器完整解碼新十一首及正常玩家狀態選曲通過，獨立測試約 6.80 秒。
標準 `tools/gui.sh test ./internal/... ./cmd/...` 全部通過，cmd/gin3 約 38.35 秒，部分套件使用快取。

未修版 SDK 使用中文音色庫時返回 0 並明示載入失敗；新匯入器已拒絕本曲，寫入 failed
收據，沒有提交 OGG 或留下匯入暫存。修正檢查工具錯誤，不以音量閾值掩蓋失敗。

## 固定來源、封裝與身份

正式技術收據：[windows-gm-source-tools.json](../internals/data/windows-gm-source-tools.json)，SHA-256：
`49c07dac366671e95ec5b6589f83efaecf4ff2ac8df88ee0d45023a14cd8dded`。
固定來源鎖：[windows-music-source-lock.json](../internals/data/windows-music-source-lock.json)，SHA-256：
`c0f01816d77de17f041e58a6d6a30ee3fe82ff19b7c2dc3063018f0ec20b3767`。

十四組 Debian 來源與兩組額外來源完整取得，總計 126,093,999 bytes；固定 URL、大小、
SHA-256、描述檔的全部來源項目與離線重驗通過，五組邊界拒絕測試通過。
額外來源為固定 commit 的 proxy-libintl 及完整上游 zlib 1.2.13；沒有追蹤浮動分支。
本機 `workplace/music69-windows/source-delivery/windows-music-fixed-sources.tar` 共 112 項、
127,129,600 bytes，SHA-256：`e94d2c6fc4f013bb6684bf1f33eaad963ea103daed27baef1d848273f0cdbce4`。
逐項重讀大小／雜湊通過，包含完整來源、授權、建置配方與本地補丁；不含原版、衍生曲庫、SF2 或程式二進位。

隔離工具包共十四個程式／非系統 DLL，實際 PE 引用閉合通過，使用同一 GCC 12 SEH
執行期；另外保存 49 份完整授權／著作權／作者文字。這是來源與工具技術驗收，
尚未逐套重建編譯器執行期或驗來源 PGP 簽章信任鏈，也不宣稱正式發行授權封裝完成。
實際工具包包含固定 MIT FluidR3_GM，SHA-256：
`74594e8f4250680adf590507a306655a299935343583256f3b722c48a1bc1cb0`。

原生匯入器 SHA-256：`6ed783fc065429c0472fd2be550f216324b92ccdc34351dad6e1fdde818b903f`；
修正 FluidSynth.exe：`0e2125c9ef3c54a06e8676f6eed70740ee4bb10ddeec23a357b2a6ef0d9ac0ff`。
Go 1.24.13，來源父提交 `851a2f402a2d836b27e4df6e2c83f5294b5bdeb2`，工作樹含本輪修改。
原版 bundle SHA-256 為 `de905aa0b5a666aa9d31dce36f43a5584f1842950c39d7e1b1f97c72e0d79673`。
本輪不修改原版或 dosgolem；完整工具、程式、補丁、來源檔、日誌與輸出身份保存於 JSON。

## 重生入口

先按[驗證 034](034-windows-music-pcm.md) 備妥固定交叉工具鏈、FM 編碼器與官方 SDK 基準，
按[驗證 033](033-linux-music-sources.md) 備妥已驗來源快取，按[驗證 030](030-native-gm-import.md)
備妥固定音色庫。以下一律經既有 Docker 包裝器執行，預設無網路，原版與來源掛載唯讀。
交叉工具鏈 image 為 `sha256:0745d96e4029b152f468e2fee63c01e20bd859a369efe80de681e95cf17f0541`；
Wine image 為 `sha256:7d3c7ff554db39300e3e7a92a7b79c2d5fcd647289596d22fa10899baac61201`。
其餘 image 身份、完整配方與限制見收據與研究 363。

```sh
LOGH3_NET=1 tools/py.sh /work/tools/scripts/music_windows_sources.py --download
tools/py.sh /work/tools/scripts/music_windows_sources.py
tools/py.sh /work/tools/scripts/music_windows_sources_test.py
bash tools/music_cross.sh cp -- \
  /out/source-cache/proxy-libintl-c03e1a74b17fa7ec467e110130775409e4828a4c.tar.gz \
  /out/source-cache/zlib-1.2.13.tar.gz /out/
bash tools/music_cross.sh python3 /tools/scripts/music_windows_gm_probe.py --utf8
```

建置探針拒絕覆寫 `gm-build-utf8`；重生前保留既有失敗及通過批次，從乾淨 archive 建置。
Unicode 補丁先核對兩份上游原始檔 SHA-256，再試套與套用；不改 DSP、版本或 SF2。
以 `tools/gui.sh xrun` 在容器指定 GOOS=windows、GOARCH=amd64、CGO_ENABLED=0，
建置 `cmd/music-import` 至 `workplace/music69-windows/music-import-utf8-gate.exe`，再執行：

```sh
tools/gui.sh xrun sh -c 'GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -o workplace/music69-windows/music-import-utf8-gate.exe ./cmd/music-import'
bash tools/music_cross.sh python3 /tools/scripts/music_windows_gm_package_probe.py --utf8
LOGH3_TIMEOUT=1h bash tools/music_wine.sh wine 'Z:\out\gm-source-utf8-probe\封裝 測試\music-import.exe' \
  -source 'Z:\orig' -output 'Z:\out\gm-utf8-full\生成 曲庫' -gm-only
tools/py.sh -c 'from pathlib import Path; import shutil; p=Path("/work/workplace/music69-windows/gm-source-mid"); p.mkdir(exist_ok=True); [shutil.copyfile(Path("/work/assets/original/music/mid")/(n+".mid"), p/(n+".mid")) for n in ["BGM_0M","END_2M","OPEN_M"]]'
bash tools/music_wine.sh python3 /tools/scripts/music_windows_gm_wave_probe.py --utf8
tools/py.sh /work/tools/scripts/music_windows_sources.py --bundle
tools/gui.sh test ./cmd/gin3 \
  -run 'TestWindowsUnicodeGMImportPlayerDecode|TestPlayerPathMusicSwitches' -count=1 -v
tools/gui.sh test ./internal/... ./cmd/...
LOGH3_EXTRA_RO=/home/anr2/cht/logh3 tools/py.sh /work/tools/scripts/music_windows_gm_source_receipt.py
```

載入失敗拒絕批次位於 `gm-utf8-rejected/`，使用上述新匯入器、相同中文 SF2 與
未修版 SDK，完整工具命令與錯誤保存於 JSON 及 `gm-utf8-rejected.log`。
收據產生器重新核對完整來源、實際包內檔案、十一首產物、三首完整 WAV／PCM、
失敗拒絕與產品日誌後才寫完成狀態；不拿進行中或失敗批次當通過收據。

## 邊界

Wine 原生 CLI 加 Linux 正式播放層解碼，不能替代 Windows 實機音效裝置、人耳或循環接點。
GM 保留既有 MPU 音色近似，部分峰值 1.0 未聲稱無削波；沒有新原版硬體時序對拍。
Windows 合併 FM／GM 的正式工具與來源封包、macOS、正式 dist-all 交付及聆聽仍待完成。
本輪大檔、工具二進位、SF2、原版衍生音樂與 WAV／PCM 只留本機 workplace，#69 保持開啟。
