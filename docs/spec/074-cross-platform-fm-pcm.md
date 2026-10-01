# 074 — 跨平台 FM 無損合成驗收

狀態：**CONFORMED**（2026-10-01，限 Windows 原生 CLI 技術驗收與 Linux 回歸）。工作項目：[#69](https://github.com/wicanr2/logh3/issues/69)。
接續 [071](071-native-fm-import.md)，證據：[研究 362](../re/362-windows-music-tools.md)。

## 決定與證據審查

使用者選定「改驗 VGM／WAV／增益後 PCM 相同，允許各平台 OGG 差異」。
舊規格的跨平台 OGG 全檔相同門檻自此被取代；既有 Linux OGG 基準與歷史收據保留。
BGM_0F 的 Windows／Linux 完整 WAV 與 -1.00 dB 後浮點 PCM 已逐位元組相同，
同版本 Vorbis 的 260 個音訊封包不同，解碼最大絕對差為 0.0037054196。
此抽驗足以界定新門檻，不代表其餘十首已驗收。

## 正式契約

1. 原版 EXE／MDS、clean dosgolem HEAD、三種觸發路徑與完整 VGM 雜湊沿用 071。
2. 由既有 Linux 工具與已驗十一首 VGM 產生完整 WAV，並輸出 44100 Hz、立體聲、
   -1.00 dB 後的小端序 32 位元浮點 PCM。入口：
   `bash tools/audio.sh bash /src/tools/scripts/music_fm_pcm_baseline.sh`。
   `tools/scripts/music_fm_lock.py` 驗來源、VGM、WAV／PCM 與工具雜湊後更新唯一 JSON 鎖，
   Go 表格由該腳本產生；缺基準不得回退到只驗 VGM。
3. 每首先驗完整 VGM，再合成並驗完整 WAV，再由隨包 FFmpeg 以相同增益輸出 PCM 並驗雜湊。
   任一不符就拒絕該曲提交。PCM 為驗證中間產物，不改正式 OGG 播放格式。
4. OGG 沿用 q5、固定 serial／CRC 與完整解碼檢查。各平台可產生不同 OGG；
   收據記實際 OGG 雜湊，不冒填 Linux 基準。相同平台／同一工具的獨立抽樣重跑須相同。
   既有 Linux OGG 另留作回歸檢查，不宣稱跨平台解碼後的有損 PCM 相同。
5. 收據增加完整 WAV、增益後 PCM SHA-256、增益與取樣率及實際命令。
   中間 WAV／PCM 驗後可移除；正式輸出保留 VGM、OGG 與收據。

## 驗收範圍與停止線

十一首 Windows FM 必須由原生匯入器重新擷取原版並通過全部無損閘門；
三種觸發路徑獨立重跑，正式播放器完整解碼；補雜湊不符拒絕測試與標準產品測試。
Wine 技術驗收不等於 Windows 真機音效裝置或人耳確認。
完整依賴來源／授權封裝、macOS、聆聽與循環接點仍屬 #69 未完成範圍。

## 實際驗收

十一首 Windows FM 的完整 VGM／WAV／增益後 PCM 全數符合 Linux 鎖；
三種原版觸發路徑獨立重跑，四種產物雜湊與同平台首次輸出一致。
十一首 Windows OGG 皆與 Linux 不同，依使用者裁決接受並記錄實際雜湊。
Linux 三路回歸、正式播放器完整解碼 Windows 22 首、正常玩家狀態選曲、
無損產物變更／缺鎖／缺檔拒絕及標準產品測試通過。
完整收據、版本、重生與限制見[驗證 034](../playtest/034-windows-music-pcm.md)。
