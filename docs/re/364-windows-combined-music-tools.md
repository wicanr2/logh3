# 364 — Windows FM／GM 合併依賴與完整入口

日期：2026-10-02，工作項目：[#69](https://github.com/wicanr2/logh3/issues/69)。
前置：[驗證 034](../playtest/034-windows-music-pcm.md)、[驗證 035](../playtest/035-windows-gm-sources-unicode.md)。

已證實（proven）：兩個已驗工具包的十九個程式／DLL 可組成單一工具目錄。
七個重複檔名的內容完全相同：FFmpeg、GCC SEH、libstdc++、libwinpthread、
libogg、libvorbis、libvorbisenc；沒有不同 C++ 執行期混裝。
以現有交叉工具鏈的 objdump 核對 PE x86-64 與每個 DLL 匯入名稱，所有非系統引用閉合。
候選收據 `workplace/music69-windows/COMBINED-CANDIDATE.json` SHA-256：
`bfa0ced14cc765553405add8b2972e8d2612eb0af45f0c563e0d9fc51368a695`。
它只證明依賴候選與同名檔案，不能證明預設二十二首入口完成。

GM 工具及固定來源身份沿用 windows-gm-source-tools.json，FM 工具沿用 windows-fm-pcm.json；
完整固定來源 tar 包含十四組來源與兩組額外來源、FluidSynth 本地補丁與 opl2wav 建置配方。
合併工具另保留 AdPlug／libbinio 的完整 COPYING／AUTHORS／Debian 著作權。
未驗來源 PGP 信任鏈或逐套重建執行期，不能以來源 archive 已取得外推這兩項。

新原生入口從 clean `9a2ce97df1e222a4928dd782ddf210033213a96e`、Go 1.24.13 建置，
未修改正式程式，只準備合併包驗收。建置資訊在
`workplace/music69-windows/importer-combined-buildinfo.txt`，工具入口依[規格 076](../spec/076-windows-combined-music-tools.md)。
目前為合併技術驗收階段；原版、dosgolem 與音色庫不改。

封裝入口：`bash tools/music_cross.sh python3 /tools/scripts/music_windows_bundle.py`。
單一預設入口完整實跑，不使用 gm-only／fm-only／tracks：
`LOGH3_TIMEOUT=3h bash tools/music_wine.sh wine 'Z:\out\combined-stage\封裝 測試\music-import.exe' -source 'Z:\orig' -output 'Z:\out\combined-full\生成 曲庫'`。
Go／解碼套件授權取自實際 Go 建置容器，dosgolem 完整授權取自上述 clean 固定副本，
補充與 ZIP／中文搬移入口為 `tools/scripts/music_windows_bundle_archive.py`，保留原 BUNDLE.json。
同名雜湊／引用／路徑／非普通檔案拒絕入口：`tools/py.sh /work/tools/scripts/music_windows_bundle_test.py`。

路由命中重製驗收、dist-all 與 Windows ZIP 編碼契約，實際載入逆向重製技能、
retro-remake-end-to-end-playbook、retro-remake-dist-all-output、windows-zip-utf8-packaging
及 project-document-responsibilities。本階段使用既有 workplace，不建立正式交付目錄。

追加反例：首次完整預設入口的十一首 GM 通過，BGM_0F 的原版擷取與 VGM 雜湊通過，
但 opl2wav 的 CRT argv 將中文輸出路徑改成 `??`，AdPlug 回報認不出檔案。
完整批次 terminal exit 1，GM 收據 complete、FM 收據 failed；未宣稱二十二首完成。
固定失敗目錄 `combined-full`、工具包、ZIP 及日誌均保留；規格 076 回到 DRAFT。
Windows FM 先前英文輸出的通過結論仍限原範圍，不能外推中文檔案路徑。

平台前提依新命中的 retro-platform-spec-first 路由實際載入規則，再查官方契約：
[CommandLineToArgvW](https://learn.microsoft.com/en-us/windows/win32/api/shellapi/nf-shellapi-commandlinetoargvw)
從 Unicode 命令列提供參數並以 LocalFree 釋放；
[_wfopen](https://learn.microsoft.com/en-us/cpp/c-runtime-library/reference/fopen-wfopen?view=msvc-170)
使用寬字元檔名且不改二進位內容；
[WideCharToMultiByte](https://learn.microsoft.com/en-us/windows/win32/api/stringapiset/nf-stringapiset-widechartomultibyte)
提供 UTF-16／UTF-8 邊界轉換。這是平台規格，不另開遊戲硬體逆向。
固定 AdPlug 2.3.3 的 factory 可傳 CFileProvider；固定 libbinio 1.5 的 binifstream
預設使用 narrow fopen，binfbase 的 FILE 指標為 protected 且析構負責 close。
可丟棄探針在 Windows 以相同 binifstream 讀取 API、寬字元開檔及自訂 provider 接到
原有 factory；WAV 輸出亦用寬字元開檔，DSP／update／refresh／取樣與 WAV bytes 不改。
探針原始碼、輸入與命令保存在 `workplace/music69-windows/opl-unicode-probe/`，正式程式尚未改。

追加審查與實作：三路完整 WAV／PCM 與首路重跑相同後，規格 076 恢復 READY，
同一份來源接入 tools/opl2wav.cpp；另建 Unicode 工具與 combined-utf8-stage，不覆寫首批。
正式來源建置入口 `tools/scripts/music_windows_opl_unicode_build.py`，探針可重生入口
`tools/scripts/music_windows_opl_unicode_probe.py` 的 --build／--wave；都由既有 cross／Wine 包裝器執行。
修正版 ZIP 保留完整來源 tar，另附新 opl2wav.cpp、實際命令與三路波形身份。
非 Windows 預處理內容相同；初次嘗試在只有 runtime 的 audio image 編譯，缺 g++／pkg-config，
第二次建置 image 的手寫標頭路徑漏 libbinio，均屬驗證環境／腳本缺件；以實際 pkg-config
修正後同 image 重跑通過。未聲稱完成新的 Linux 原生建置或實跑，失敗日誌保留。
封裝邊界拒絕六組通過；正式播放器完整解碼二十二首與正常玩家狀態選曲約 9.24 秒。
完整預設 CLI 十一首 GM／十一首 FM 通過，標準產品套件通過；ZIP 解壓搬移後一首 GM
與三種 FM 路徑抽驗通過。最終收據入口 `tools/scripts/music_windows_bundle_receipt.py`，
完整身份與重生命令見[驗證 036](../playtest/036-windows-combined-music-tools.md)，#69 仍待正式交付與實機。
