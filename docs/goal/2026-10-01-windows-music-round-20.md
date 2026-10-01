# 第二十輪 — Windows 音樂工具來源建置與實跑

日期：2026-10-01。已開工作項目：[遠端 #69](https://github.com/wicanr2/logh3/issues/69)。
前輪固定來源集合已提交；本輪 87 張 Issue、12 張 open，267 條斷言零失效。

沿用同版本來源、MinGW 及 Wine 隔離工具，補齊缺少的來源建置工具，試建 Windows
音樂工具並核對實際匯入結果。FM 完整 VGM／OGG 雜湊門檻保持；遇到差異先分類、
取得證據，不直接放寬門檻或將編譯成功當玩家入口驗收。
研究入口：[研究 362](../re/362-windows-music-tools.md)。

當輪後續裁決：已保留舊 OGG 門檻失敗證據，使用者選定改驗 VGM／WAV／增益後 PCM
相同，允許平台 OGG 差異；後續依 [規格 074](../spec/074-cross-platform-fm-pcm.md)
完成十一首 Windows 原生匯入、三種觸發路徑重跑、Linux 回歸與正式解碼。
本段取代開工快照的跨平台 OGG 雜湊門檻；原始試驗與失敗形成原因保留於研究 362。
