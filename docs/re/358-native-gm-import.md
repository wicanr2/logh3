# 358 — 原生 MDS 匯出與 GM 管線

日期：2026-10-01。工作項目：[#69](https://github.com/wicanr2/logh3/issues/69)。
規格入口：[069](../spec/069-native-gm-import.md)。
原始資料版面：[MDS 參考](../internals/17-mds.md)；固定音色庫來源：[356](356-player-music-library.md)。

## 本輪證據與範圍

已證實：22 首實物均以 `4D 31`（M1）開頭，長度至少 0x100，
0x100 起為完整四位元組事件。原始位址空間為 MDS 檔案偏移，未引用 IDA 位址。
輸入雜湊與舊 MIDI 雜湊由容器讀實物，驗收收據會記錄逐首值，
原始 bundle 仍為 CONTEXT 記載的 `de905aa0b5a666aa9d31dce36f43a5584f1842950c39d7e1b1f97c72e0d79673`。

已證實：既有 `tools/scripts/mds_export.py` 以固定四 bytes 解析，
分軌只看首 byte FE；正常實物使用 FE FE FE FE。
保留有音符的軌，按 FC 開頭或檔頭通道選通道，輸出 format 1／每拍 48 ticks。
音符關閉比開啟先排，同 tick 的速度在最前；相同優先序維持事件原順序。
兩份 END_2 的小節引數採既有超過 90% 小節引數 ≤2 判定，否則一律前進一小節。

既有近似的實際語意：`bar_length` 對未解 raw 事件累加 opcode，
`timeline` 對非 FB／FC raw 事件也累加 opcode；E4 因此累加 224，
不是第二 byte 的步進。原生移植先保留這項現有輸出語意，明示為 legacy approximation，
不得從「MIDI 一致」推論原版時間完全正確。FA 直接當 BPM 及 FB／FC 不展開反覆
同樣沿用既有近似，本輪不重新解讀原版音樂驅動。

GM 合成沿用 `fluidsynth -ni -F <wav> -r 44100 -g 0.8 <SF2> <mid>`，
OGG 沿用 `ffmpeg -v error -y -i <wav> -ar 44100 -c:a libvorbis -q:a 5 <ogg>`。
實物工具版本為 FluidSynth 2.3.1、FFmpeg 5.1.9，映像檔及 SF2 的完整身份見研究 356。
正式玩家程式不得呼叫 Docker、Python、shell 或搜尋玩家 PATH 代替隨包工具。

FM 必須繼續用原版 OPL 擷取，不能因移植 GM 而轉成合成 GM。
本輪只接完整管線的 GM 階段；三平台封包仍須供應合成器、編碼器、音色庫與依賴。
新工具在 FM 尚未接齊時要求明示 `-gm-only`，不將部分曲庫說成全部完成。

## 驗收回填

22 首 MIDI 逐 byte 一致，十一首真實 GM 產物通過原生解碼與正式播放器完整讀取，
三首相同命令的完整 PCM 相同。錯誤路徑拒絕與完整產品回歸通過；
雜湊、工具／程式身份及輸入參數見[驗證 030](../playtest/030-native-gm-import.md)、
[`native-gm-import.json`](../internals/data/native-gm-import.json)。
其餘管線與聆聽限制保留，本輪沒有重跑 dosgolem 或變更 FM 基準。
