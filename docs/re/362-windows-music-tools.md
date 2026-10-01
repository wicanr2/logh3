# 362 — Windows 音樂工具來源建置與實跑

日期：2026-10-01，工作項目：[#69](https://github.com/wicanr2/logh3/issues/69)。
狀態：Windows 原生 CLI 技術驗收完成；正式封裝、實機與聆聽未完成。

已證實：既有 `fd2-build-mingw:latest` 是 Debian bookworm，MinGW GCC 12-win32，
`gcc-mingw-w64-x86-64` 為 `12.2.0-14+25.2`，開發標頭 `10.0.0-3`。
image ID：`sha256:4ed8c0dc9513b4cf2127ae988aee7a9e238d41ec5ef4f5e0a5c6cb146b09d356`。
既有 image 有編譯器，缺 autoconf、patch、CMake、Meson 等來源建置工具。
`logh3-audio-cross:20261001-r1` 將沿用該父 image 補缺件；入口
`bash tools/music_cross.sh --build`，工作命令 `bash tools/music_cross.sh <命令>` 預設網路關閉。
可重生來源配方為 `tools/Dockerfile.music-cross`，不修改其他專案工具鏈。
父工具鏈入口為 `/home/anr2/cht/fd2/remake/packaging/Dockerfile.mingw`；
它必須先產生已核對的 MinGW image，本專案包裝器拒絕身份不同的父 image。

原始輸入沿用[固定來源集合](../playtest/033-linux-music-sources.md)，唯讀掛載；
試建輸出放 `workplace/music69-windows/`，不屬正式交付。
官方工具鏈方法參考：[FFmpeg Windows 平台文件](https://ffmpeg.org/platform.html)、
[MinGW 的 CMake 交叉建置文件](https://www.mingw-w64.org/build-systems/cmake/)。
尚須以實物版本原始碼與輸出核對，不能由文件推定 FM 雜湊會相同。

## 已取得的 Windows FM 證據

既有 MinGW image 補齊建置工具後，cross image ID 為
`sha256:829642b4f69c06c48deeeb90b558b90b08e29c6cfeaa3162f120e63f4b51eadb`。
libbinio／AdPlug 的共享庫建置指定 no-undefined，保留動態 DLL；
來源含 Debian 修補，取自固定鎖後先驗全部雜湊，再由 dpkg-source 解開。
AdPlug 的 Automake 版本不合以 autoreconf 修正；opl2wav 改用原配方的完整
pkg-config 旗標補齊 binio 標頭路徑。兩次失敗日誌保存，不將它們列成遊戲缺陷。

Wine image 為 `sha256:7d3c7ff554db39300e3e7a92a7b79c2d5fcd647289596d22fa10899baac61201`。
中文程式路徑首次啟動失敗，設定 C.UTF-8 後同一個程式成功；兩處 EXE 雜湊完全相同。
這是 Wine argv 語系問題。`tools/music_wine.sh` 提供有界 Xvfb／Wine，結束關閉背景程序。
Windows probe 的 Go 建置資訊證實為 clean f7523f9，不改共用執行器或原版輸入。

已證實（proven）：原生 Windows 匯入器在 BGM_0F 通過 VGM 閘門，
固定 Ogg serial 後全檔雜湊失敗，正式入口拒絕提交。以同一份 VGM 做診斷，
Windows 與 Linux 的完整 WAV 都是 5,606,647 frames、44.1 kHz／立體聲，逐 byte 相同：
`b12de28b606c0266a7a5dc05ebee8c990029844ed02e2205ce17c53b93ec293e`。
PCM SHA-256：`c6378fa69a2ef7390c970ae5470a593f8832ab00f1435e618130ff392d37e4c9`。
相同輸入的 Ogg 頭與 packet 數相同，260 個音訊 packet 不同；
Windows 固定封裝後為 `fa77257c455ad2e050da72327d308907d9c0a637d8fe9651a070b5e66f73a5c3`，
Linux 恢復既有 `48c22ed0b5d9fec5d0b9bed70e7a06ae1ab3a07faef99da80ddfd21b4644739c`。
共同 Linux 解碼器讀出 11,213,294 samples，410,268 個不同；最大絕對差
0.0037054196、RMSE 0.0000258283。不是只有封裝序號不同，尚不能宣稱一致。

強推論（strong inference）：[libvorbis 1.3.7 的平台相容層](https://github.com/xiph/vorbis/blob/v1.3.7/lib/os.h)
在所有 _WIN32 分支改用舊式 rint、double 數學及 sqrt 近似 hypot。
MinGW 支援原有 Unix 數學介面；目前以可丟棄來源副本測試改選既有路徑。
仍須比較濾波後 PCM 及完整 OGG，未通過前不將此假設送入正式封裝。

## 後續實驗與使用者裁決

已證實：-1.00 dB 後 f32le PCM 44,853,176 bytes，Windows／Linux 都為
`6b31f0ed34da50e04b132d6b7569b034b09b46f73d04858413786701de6a093b`。
修改 libvorbis 平台數學路徑的可丟棄實驗仍有 232 個音訊封包不同，
不是可採用的完整 OGG 相同解法；正式工具保留原來源的數學選擇。
Windows 原生匯入器十一首 GM 均真實生成，MIDI 雜湊與 Linux 相同，
正式播放器完整解碼通過。官方 FluidSynth 2.3.1 ZIP 的 SHA-256 為
`7bf4ae8ae1856376266280bf680d465a96d3f66e2b875cdd7f3539b549c4d702`。
SDK 的 C++ runtime 與 FM 不同，試驗採獨立工具目錄；完整來源與授權仍未核對齊全。

2026-10-01 使用者選定「改驗 VGM／WAV／增益後 PCM 相同，允許各平台 OGG 差異」。
這是驗收契約修改，不是證明 OGG 或其解碼結果相同。
實作依 [規格 074](../spec/074-cross-platform-fm-pcm.md)，保留舊門檻失敗收據。
原探針合併收據為 [`windows-music-probes.json`](../internals/data/windows-music-probes.json)。
產生命令：`tools/py.sh tools/scripts/music_windows_receipt.py --current-cross-image <image ID>`。
當輪原編譯 image 為 829642b4，補釘父 image 與直接套件版本後的 image 為
`sha256:0745d96e4029b152f468e2fee63c01e20bd859a369efe80de681e95cf17f0541`，
原建置與新建置身份分開保存，不倒填成同一版本。

舊建置 metadata／來源副本／prefix 保存至 `workplace/music69-windows/pre-pcm-gate/`。
合併反例收據 SHA-256：`a8fca6e5966018bd60d0aa52a6bd1756343150aceaec19a78aabf3818665fb2e`。
反例統計重新加總的末位浮點差異以相對 1e-10／絕對 1e-12 檢核，
sample 數、不同數及實物 PCM 雜湊仍須完全一致。

## 正式無損閘門結果

規格 074 READY 後，為十一首建立 Linux WAV／增益後 PCM 鎖；
由乾淨來源重建 Windows 工具，不採數學路徑試驗。
原生匯入器先驗 VGM，再驗完整 WAV，再驗增益後 f32le PCM，任一不符拒絕提交。
十一首真實原版擷取全數通過，OGG 記實際雜湊而不冒填 Linux 基準。
三路重跑、Linux 回歸、22 首正式解碼與標準產品測試通過。
完整命令、雜湊與限制見[驗證 034](../playtest/034-windows-music-pcm.md)。
