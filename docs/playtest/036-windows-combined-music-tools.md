# 036 — Windows 同包曲庫、Unicode OPL 與中文 ZIP 搬移

日期：2026-10-02，工作項目：[#69](https://github.com/wicanr2/logh3/issues/69)。
契約：[規格 076](../spec/076-windows-combined-music-tools.md)；調查與首次失敗：[研究 364](../re/364-windows-combined-music-tools.md)。
已證實（proven）：同包 Windows 原生入口從唯讀原版生成二十二首，中文檔案路徑、
正式播放層與 ZIP 搬移抽驗通過。範圍限 Wine 原生 CLI 與 Linux 播放層技術驗收。

## 身份與結果

唯一結構化收據：[windows-combined-music-tools.json](../internals/data/windows-combined-music-tools.json)，
SHA-256 `ced26a8c7172e7c09722bf068b8f7f57699df61cd5a6470536f07c3ecf4857f3`。
它綁定實際來源／工具／輸出／日誌、完整命令、兩側 Go buildinfo、固定來源、首次失敗及重生腳本。
原版 bundle SHA-256 為 `de905aa0b5a666aa9d31dce36f43a5584f1842950c39d7e1b1f97c72e0d79673`；
dosgolem 為乾淨 `f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c`，原版未改。

匯入器從乾淨提交 `9a2ce97df1e222a4928dd782ddf210033213a96e`、Go 1.24.13 建置，
實際 `vcs.modified=false`，程式 SHA-256：
`048cc00e40144b5e59cc635c60f8fc994d5814a8a806ba8447c89461f064ba38`。
新 C++ 來源是本輪尚未提交的修改，來源 SHA-256：
`2e2b6afba914dd18b023d5c813dfe7e39fa915467bd8ccb8a4b7f0cdab74298e`；
其 `opl2wav.exe` SHA-256：`6ae83b64f238cf5073ef4e9601bc71ddd9567fd018d894316f037de83217d5e0`。
不可把匯入器的乾淨建置資訊套用到新 C++ 來源。

| 驗證 | 結果 |
|---|---|
| 預設入口、不指定曲目或階段 | GM 十一首、FM 十一首真實生成；兩份階段收據 complete，同一個匯入器及工具包 |
| FM 十一首 | 完整 VGM、WAV、-1.00 dB 後 f32le PCM 符合 074 Linux 固定基準；OGG 全檔符合先前 Windows 批次 |
| GM 十一首 | MIDI、時長、峰值及 Vorbis 音訊封包符合 075；不以容器或三個標頭差異判音訊不同 |
| 中文 OPL 探針 | BGM_0F／END_2F／OPEN_F 完整 WAV／PCM 與舊英文路徑相同，首曲重跑相同；受版控腳本獨立重生亦通過 |
| 搬移後正式 CLI | 新輸出生成 BGM_0M、BGM_0F、END_2F、OPEN_F；GM 音訊封包、FM 無損及同平台 OGG 與完整批次相同 |
| 正式播放器／正常玩家狀態選曲 | 二十二首完整解碼通過，目標測試 9.242 秒 |
| 標準產品套件 | 全部通過；cmd/gin3 44.373 秒，其餘部分使用 Go 快取 |
| 封裝邊界 | 六組通過：撞名內容／大小、缺工具、未知引用、路徑穿越、非普通／連結輸入、輸出連結及既有包保留等條件 |

合併包十九個程式／DLL，七個同名重疊內容一致，PE x86-64 與非系統引用閉合。
包內固定 MIT FluidR3_GM、十四組加兩組來源的完整 tar、既有補丁、新 C++ 原始碼與建置身份均保存。
六十份完整授權／著作權／作者文字包含工具與 Go／解碼套件；未加入原版遊戲或生成曲庫。
固定來源與音色庫身份沿用[驗證 035](035-windows-gm-sources-unicode.md)，完整來源 tar SHA-256：
`e94d2c6fc4f013bb6684bf1f33eaad963ea103daed27baef1d848273f0cdbce4`。

## ZIP 與反例

本機技術包 `workplace/music69-windows/combined-utf8-tools-technical.zip`：91 檔，275,685,648 bytes，
SHA-256 `d17cece03aba875f40d343405ec99d85861a60881d3134080e59fb989cf2202b`。
所有檔名在中文根目錄，91 項均有 UTF-8 旗標；README.txt 有 UTF-8 BOM／CRLF。
CRC、解壓名稱及每檔大小／雜湊核對通過；搬移至 `combined-utf8-relocated/中文 搬移/封裝 測試` 後實跑。
這是 workplace 技術包，尚非正式版號／dist-all 交付。

首次 `combined-full/生成 曲庫` 的 GM 完成；FM 首曲 VGM 通過，但舊 opl2wav 把中文路徑
轉成 `??`，AdPlug 開檔失敗。失敗輸出、收據、日誌及首次來源腳本均保留；076 曾回 DRAFT。
舊工具 SHA-256 `909ebabe8ec03080e88dc8f9290a40d1e56d7c4f57a8d226e639888eab1197a2`；
舊 ZIP SHA-256 `cd24ec2470497e8dc164e4f29fa4c689afc3df48a7d0cd9f2ef777bcc773fdc9`。
新工具僅修 Unicode 命令列與檔案邊界，未改 DSP／取樣／update；三路探針通過、重新 READY 後才接正式來源。
新包與完整批次另存 `combined-utf8-*`，沒有把失敗收據改成通過。

Linux 分支以相同建置容器及實際 pkg-config 旗標核對，非 Windows 預處理內容相同，
SHA-256 `5526758c8ff74ba89daead8c08e4aa36654b96482d1bb186f1dc3d7c61ecd224`。
僅移除空白行與行首尾空白，不改字串內容；這不等於 Linux 新版本編譯／實跑。
先前 runtime image 缺 g++／pkg-config，以及手寫標頭路徑漏 libbinio，均為驗證環境／腳本失敗，日誌保留。

## 重生入口

前置工具、固定來源、GM／FM 基準及完整音色庫依[驗證 034](034-windows-music-pcm.md)／
[035](035-windows-gm-sources-unicode.md)準備；所有輸入核對 SHA-256，原版唯讀。
三路探針輸入取自固定原版 VGM：BGM_0F、END_2F、OPEN_F，鎖與完整擷取方法見
[規格 071](../spec/071-native-fm-import.md)。不自製替代 VGM。
既有收據不可覆寫；下列封裝／建置命令要求新的目標目錄，重跑時先保存前一批與命令身份。
實際容器影像身份在 JSON；沿用 `logh3-audio-cross:20261001-r1`、
`colonization-windows-verification:20261001-r1`、`logh3-ebiten:2`，未另建工具鏈。

```sh
# 已保存三路探針輸入；這次獨立重生另存，不覆寫首次探針。
bash tools/music_cross.sh python3 /tools/scripts/music_windows_opl_unicode_probe.py --build \
  --inputs /out/opl-unicode-probe/inputs --output-dir /out/opl-unicode-replay
bash tools/music_wine.sh python3 /tools/scripts/music_windows_opl_unicode_probe.py --wave \
  --output-dir /out/opl-unicode-replay
# 生產來源建置要求原始三路 WAVE.json 已通過。
bash tools/music_cross.sh python3 /tools/scripts/music_windows_opl_unicode_build.py
bash tools/music_cross.sh python3 /tools/scripts/music_windows_bundle.py --utf8
bash tools/music_cross.sh python3 /tools/scripts/music_windows_bundle_archive.py --utf8
LOGH3_TIMEOUT=3h bash tools/music_wine.sh wine 'Z:\out\combined-utf8-stage\封裝 測試\music-import.exe' \
  -source 'Z:\orig' -output 'Z:\out\combined-utf8-full\生成 曲庫'
LOGH3_TIMEOUT=1h bash tools/music_wine.sh wine 'Z:\out\combined-utf8-relocated\中文 搬移\封裝 測試\music-import.exe' \
  -source 'Z:\orig' -output 'Z:\out\combined-utf8-relocated-output\生成 曲庫' \
  -tracks BGM_0M,BGM_0F,END_2F,OPEN_F
bash tools/py.sh /work/tools/scripts/music_windows_bundle_test.py
bash tools/gui.sh test ./cmd/gin3 -run 'TestWindowsCombinedImportPlayerDecode|TestPlayerPathMusicSwitches' -count=1 -v
bash tools/gui.sh test ./internal/... ./cmd/...
LOGH3_EXTRA_RO=/home/anr2/cht/logh3 bash tools/py.sh /work/tools/scripts/music_windows_bundle_receipt.py
```

可回查日誌在 `workplace/music69-windows/`，檔名與 SHA-256 均在唯一 JSON：
`combined-full.log` 是首次失敗；`combined-utf8-full.log`、`combined-utf8-relocated.log`、
`combined-player-test.log`、`combined-product-tests.log`、`combined-utf8-boundary-tests.log` 是修正後驗收。
完成收據腳本要求階段 complete、實際工具／輸出相符及測試 PASS，未完成不產生通過收據。

## 限制

跨平台 FM 維持使用者裁決：VGM／WAV／增益後 PCM 相同，允許各平台 OGG 差異。
本輪同 Windows OGG 相同，不外推有損解碼跨平台一致或人耳確認。
合成時間仍為 hardware-spec approximation，OPEN_F 仍有既有替換暫存 MDS 診斷前提。
部分 GM 峰值 1.0，未宣稱無 clipping。來源 archive 已取得，不宣稱 PGP 信任或執行期逐套重建。
Windows 實機、人耳／循環、macOS、正式版號／dist-all、本作品授權與原始碼交付仍待完成，#69 保持開啟。
