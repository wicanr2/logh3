# 075 — Windows GM 固定來源與 Unicode 路徑

狀態：**CONFORMED**（2026-10-02；限 Windows Wine 原生 CLI、固定來源與 Linux 正式播放層技術驗收）。工作項目：[#69](https://github.com/wicanr2/logh3/issues/69)。
證據：[研究 363](../re/363-windows-music-source-build.md)，沿用 GM 匯入規格 069。

## 問題與證據審查

已證實（proven）：固定來源 FluidSynth 2.3.1 float DSP 與官方 SDK，
同 MIDI／FluidR3_GM／44100 Hz／0.8 的三首 WAV／PCM 逐位元組相同。
但 Windows CLI 的系統碼頁 `argv` 將中文音色庫路徑換成 `??`，
載入失敗仍返回成功並產生抖動噪聲，現行匯入器只檢查退出碼而接受近乎靜音產物。
GLib 2.74.6 `g_win32_get_command_line()` 已提供 Unicode → UTF-8 契約；
FluidSynth `fluid_fopen()` 接受 UTF-8，但 CLI 的音色庫判別另有 ANSI 轉碼；
libsndfile 的內部檔案層接受 UTF-8，公開 `sf_open()` 在 Windows 仍從 ANSI 轉碼，
須改用同版本已提供的 `sf_wchar_open()`。
原始檔、固定來源、工具、命令與失敗產物身份由研究 363 與結構化收據保留。
首次 READY 審查的「只修參數入口足夠」被探針推翻；錯誤輸出拒絕仍有獨立證據，
Unicode 工具須先補齊全部轉碼端再重新達 READY，不將中間探針換入正式包。
追加審查：移除重複轉碼並改用既有寬字元開檔後，中文／空白工具、音色庫、MIDI 與輸出
三首 WAV／PCM 與 SDK 基準相同，BGM_0M 兩側重跑相同。路徑修正有完整證據，恢復 READY。

## 契約

1. 保持 FluidSynth 2.3.1、float DSP、相同 SF2／MIDI／44100 Hz／0.8，
   維持 OGG q5、既有曲目、播放器與選曲規則，不改原版 MPU 近似等級。
2. 以完整固定來源、大小與 SHA-256 建置 Windows 合成器，GLib fallback 固定 commit，
   禁止偷偷下載其他版本。非系統 DLL 由實際 PE 匯入表形成閉合；
   來源／著作權／完整授權／本地補丁可回查，不把編譯成功視為正式交付完成。
3. 僅修 FluidSynth Windows CLI：從 Unicode 命令列取得 UTF-8 `argv` 後走既有解析，
   失敗返回非零，正常清理時釋放參數。保留未修改上游來源與明示本地補丁，
   先驗原始檔 SHA-256，移除已不適用的第二次 ANSI 轉碼；檔案渲染端從 UTF-8 轉 UTF-16
   呼叫同版本 `sf_wchar_open()`，不改 library DSP 或 Linux 入口。
4. GM 匯入器檢查 FluidSynth 的成功退出輸出。若明示 `fluidsynth: error:` 或
   `not a SoundFont or MIDI file`，拒絕本曲、寫 failed 收據並清理獨立暫存；
   不以近乎靜音閾值掩蓋音色庫載入失敗，也不調大音量。
5. 包內工具／音色庫、原生匯入器與輸出搬到中文／空白路徑仍須可用。
   玩家入口維持原生程式，不呼叫 Docker／Python／shell，不搜尋 PATH。
6. 素材 → MDS parser → 相同 MIDI → 合成器 → OGG → 正式播放層抽驗。
   本改動不影響遊戲規則、存檔或 RNG，無須重新開原版硬體時序逆向。

## 驗收與停止線

中文／空白路徑完整十一首原生 GM 匯入；三首 WAV／PCM 與英文路徑未修改 SDK 基準相同，
BGM_0M 至少獨立重跑一次。原 SDK 的中文音色庫載入失敗須被匯入器拒絕。
正式播放層完整解碼新十一首，正常玩家狀態選曲及標準產品測試通過。
固定來源無網路全集合重驗、邊界拒絕、原版輸入未修改、UID/GID 與 Docker 清理均核對。
Wine 技術驗收不能外推 Windows 實機／人耳／循環接點；macOS 與正式 dist-all 交付仍待完成。

## 實作與驗收

第二版 READY 後，新匯入器已接工具明示錯誤拒絕。中文／空白路徑完整十一首 GM
生成及正式解碼通過；十一首 MIDI、時長、峰值與 SDK 相同，Vorbis 音訊封包相同。
三首 WAV／PCM 及 BGM_0M 兩側獨立重跑相同；未修版 SDK 返回成功的載入失敗
已被新入口拒絕，未提交 OGG。固定來源十四組加兩組額外來源、完整來源 tar、
工具 DLL 閉合、五組拒絕測試與標準產品套件通過。
重生命令、雜湊及剩餘限制見[驗證 035](../playtest/035-windows-gm-sources-unicode.md)。
首次 READY → DRAFT → 第二版 READY 的證據歷史保留，不將中間失敗改寫為通過。
