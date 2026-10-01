# 363 — Windows GM 同版本來源建置與依賴對照

日期：2026-10-01，工作項目：[#69](https://github.com/wicanr2/logh3/issues/69)。
狀態：來源試建中；尚未替換正式工具。
前置證據：[Windows 技術驗收 034](../playtest/034-windows-music-pcm.md)、
[固定 Linux 來源 033](../playtest/033-linux-music-sources.md)。

已證實：官方 FluidSynth 2.3.1 SDK 可產生十一首 GM，但 ZIP 不附完整依賴著作權／授權。
SDK 使用 float DSP；本輪先保持 2.3.1 與 float DSP、相同 FluidR3_GM，
以完整固定來源試建，原 SDK 與曲庫保留作對照。
GLib 2.74.6 的 Windows iconv 在來源內；proxy-libintl fallback 鎖在
Git commit `c03e1a74b17fa7ec467e110130775409e4828a4c`，需要額外取得完整來源。
其他必要依賴 libffi／PCRE2／zlib／libsndfile 已在 Debian 固定來源集合內。

沿用 `bash tools/music_cross.sh <命令>`，image
`sha256:0745d96e4029b152f468e2fee63c01e20bd859a369efe80de681e95cf17f0541`。
試建目錄在既有 `workplace/music69-windows/`，來源唯讀、網路預設關閉、輸出 UID/GID 1000:1000。
不修改原版／既有 FM 工具或官方 SDK，不宣稱此處 DSP 合成為原版 MPU 音色。

來源鎖入口：[windows-music-source-lock.json](../internals/data/windows-music-source-lock.json)。
探針配方：`tools/scripts/music_windows_gm_probe.py`，由上述容器包裝器執行；
試建拒絕覆蓋 `gm-build`，從乾淨 archive 解包，GLib 禁止下載 fallback。

第一個試建失敗是 libffi 不必要的 autoreconf：系統 libtool 缺私有宏
`LT_SYS_SYMBOL_USCORE`。固定 release 的 README 明確要求直接用已提供 configure，
只有 Git 來源才跑 autogen；配方改用 release 入口，失敗副本保留至 gm-attempt-1。
這不是 libffi runtime、遊戲或音樂缺陷，也不以改源碼掩蓋宏缺失。

第二次試建 libffi／PCRE2 成功，但 Debian zlib 的 `Files-Excluded` 明列 `win32`，
缺少 CMake 需要的 `win32/zlib1.rc`。保留至 `gm-attempt-2`，改用完整同版本上游
`https://zlib.net/fossils/zlib-1.2.13.tar.gz`，1,497,445 bytes，SHA-256
`b3a24de97a8fdbc835b9833169501030b8977031bcb54b3b3ac13740f846ab30`。
已核對上游 Windows 資源檔與完整 LICENSE；不自行刪去 DLL 版本資源來讓建置過關。

固定來源取得／離線重驗入口：
`tools/py.sh /work/tools/scripts/music_windows_sources.py --download`（明確下載時設定
`LOGH3_NET=1`），接著同一命令移除 `--download` 與網路設定重驗。
既有 Linux 快取只讀取已驗證來源；另取得 MinGW 包裝來源與完整上游來源。
快取損壞直接拒絕，不用下載覆蓋；本收據不代表來源簽章信任鏈或正式發行已完成。
邊界測試入口：`tools/py.sh /work/tools/scripts/music_windows_sources_test.py`；
覆蓋缺項／撞名／未審 URL、損壞快取於網路前拒絕、連結／離線缺檔及下載大小／雜湊原子拒絕。

第三次試建 zlib DLL 已成功，但 CMake 安裝 `libzlib.dll.a` 至 `lib/`、
`zlib.pc` 至 `share/pkgconfig/`，內容仍宣告 `-lz`，GLib 因而拒絕依賴。
保留至 `gm-attempt-3`；第四次沿用上游 `win32/Makefile.gcc`，
它原生產出 `libz.dll.a` 與對應 pkg-config，不改上游程式碼。
第四次六個元件（含 FluidSynth）全部編譯成功；float DSP／檔案渲染／OpenMP 開啟，
維持 2.3.1／44100 Hz／0.8 合成增益／相同 SF2。未使用的音訊裝置與 DLS backend 關閉。
完整命令、建置設定與輸出身份見本機 `gm-build/BUILD.json`。

隔離包探針：`bash tools/music_cross.sh python3 /tools/scripts/music_windows_gm_package_probe.py`；
以實際 PE DLL 匯入表收取全部非系統依賴，保留完整上游／Debian／執行期授權原文，
輸出至 `gm-source-probe/封裝 測試`，官方 SDK 與既有 FM 包均保留。
完整來源 archive 另由固定來源鎖管理，不把這個試驗包宣稱為正式授權交付。

波形探針入口：`bash tools/music_wine.sh python3 /tools/scripts/music_windows_gm_wave_probe.py`。
三首 `gm-source-mid/*.mid` 必須與官方 SDK 完整原生匯入收據的 MIDI 雜湊相同；
比較 BGM_0M／END_2M／OPEN_M，BGM_0M 另做兩側各一次獨立重跑。
以同一 SF2／44100 Hz／0.8 及 `-ni -F` 實跑，保存完整 WAV、PCM 雜湊、幀數及差異統計。
OGG 異常的隔離入口：
`bash tools/music_wine.sh python3 /tools/scripts/music_windows_gm_encoding_probe.py`；
同 WAV 比較新舊 FFmpeg、CPU 路徑及零增益濾鏡，不以調大音量掩蓋格式轉換錯誤。

已證實：中文音色庫路徑在 FluidSynth 2.3.1 `main(char **argv)` 變成 `??`；
找不到音色庫時 CLI 仍返回成功並產生抖動噪聲。英文路徑的三首 WAV／PCM 與 SDK
逐位元組相同，BGM_0M 兩側各自重跑亦相同；同 WAV 的五種編碼路徑音量正常。
原生 SDK 舊 CLI 重跑英文音色庫仍正常，因此不能將低音量歸因於編碼器或新 DSP。
GLib 2.74.6 `glib/gwin32.c` 明確以 `GetCommandLineW()` 取得完整 Unicode 後轉 UTF-8；
[官方 API 文件](https://docs.gtk.org/glib-win32/func.get_command_line.html) 同時說明系統碼頁的 `?` 替代。
FluidSynth 內建 `fluid_fopen()` 已接受 UTF-8，缺口只在 Windows CLI 參數入口。
原始 `src/fluidsynth.c` SHA-256：`13a26a2c7bd62a4f7d041c38471ad2dc40bd7dc7f48c918b64d9a1413e2ceb04`。

可丟棄 Unicode 補丁：[music_windows_fluidsynth_utf8.patch](../../tools/scripts/music_windows_fluidsynth_utf8.patch)。
使用上述建置／試驗包探針的 `--utf8` 選項，另外建立 `gm-build-utf8` 與
`gm-source-utf8-probe/封裝 測試`；保留未修正版與低音量產物，不接正式工具。
Unicode 路徑 API 交叉核對探針：`tools/scripts/music_windows_path_probe.c`。

追加勘誤：第一版 Unicode 探針證明「只修 main 入口足夠」不成立。
CLI 音色庫判別另用 `win32_ansi_to_utf8()`，把已轉 UTF-8 的參數再當 ANSI 轉換；
libsndfile 1.2.0 `sf_open()` 的公開 Windows 入口也從 CP_ACP 轉碼，
先前讀到的 UTF-8 契約只屬內部 `psf_open_handle()`，不能外推公開入口。
Unicode API 探針已證實同路徑的 GLib 判別、GetFileAttributesW／CreateFileW／_wfopen 均成功，
因此下一版移除 CLI 重複 ANSI 轉碼，檔案渲染改用既有 `sf_wchar_open()`。
規格 075 返回 DRAFT；保留第一版補丁、建置與波形失敗收據，不將推測冒作完成。
本機完整來源試驗封裝入口：`tools/py.sh /work/tools/scripts/music_windows_sources.py --bundle`；
先無網路驗快取與工具／授權，再產生並逐項重讀來源 tar，不包含原版／衍生曲庫／SF2／執行檔。

第二版探針結果：中文／空白工具、SF2、MIDI 與 WAV 輸出三首，
WAV／PCM 與 SDK 逐位元組相同，BGM_0M 兩側各自重跑相同。
SDK 的三首原始 WAV 振幅正常；第一批低音量來自音色庫未載入，不是編碼器缺陷。
規格 075 以追加證據恢復 READY；正式匯入器新增工具明示錯誤拒絕，待完整曲庫與播放器驗收。
完成收據產生入口：`LOGH3_EXTRA_RO=/home/anr2/cht/logh3 tools/py.sh /work/tools/scripts/music_windows_gm_source_receipt.py`。
逐項重驗固定來源、工具／補丁／授權、十一首 MIDI／OGG、三首 WAV／PCM、載入失敗拒絕與產品日誌。

收尾追加（2026-10-02）：第二版 READY 後，中文／空白路徑十一首原生 GM、正式播放層完整
解碼與正常玩家狀態選曲通過；十一首 Vorbis 音訊封包與 SDK 相同，三首 WAV／PCM 相同。
未修版 SDK 的中文 SF2 載入失敗已被新匯入器拒絕，未提交 OGG。
十四組固定來源與兩組額外來源、本機完整來源 tar、十四個程式／DLL 引用閉合、
49 份完整授權／著作權／作者文字及五組邊界拒絕已核對，標準產品測試通過。
完整雜湊與重生命令見[驗證 035](../playtest/035-windows-gm-sources-unicode.md)。
規格 075 CONFORMED 限該技術範圍；本文開頭與過程中的「試建中／未接入口」為當時歷史，
目前正式匯入器已接失敗拒絕；正式 Windows 合併封包、macOS、實機、人耳及循環仍待完成。
