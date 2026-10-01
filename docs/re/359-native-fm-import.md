# 359 — 原生 FM 擷取與可重現 OGG

日期：2026-10-01。工作項目：[#69](https://github.com/wicanr2/logh3/issues/69)。
原版 OPL 觸發證據：[研究 275](275-fm-capture-from-opl.md)。
封裝規格：[070](../spec/070-deterministic-ogg-framing.md)。

## 本輪來源

原版 bundle SHA-256：`de905aa0b5a666aa9d31dce36f43a5584f1842950c39d7e1b1f97c72e0d79673`。
執行器專用副本 clean，HEAD `f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c`。
沿用 Go 1.24.13、`golang:1.24-bookworm` 與 `logh3-audio:2`；
映像檔完整身份見[前輪收據](../internals/data/native-gm-import.json)。
未修改原版素材、執行器來源或主 dosgolem checkout。

本輪重新確認 `cmd/probe/observe.go` 的 VGM 開檔錨與缺錨不寫檔契約，
不是由 CLI help 的「要加 adlib」推論本作分支；GIN3 的既有例外見研究 275。
既有 `opl2wav` SHA-256 為
`fd7864ccb7cd1eeac5f56bd3ca4bdcf23750f59bb6f03eb66d258a11d15ab8e1`，
實際接 AdPlug 2.3.3，直接吃 VGM 成功。Dockerfile／C++ 中「不認得 VGM」
的舊註解與實物矛盾，本輪以真實實跑及研究 275 為準；未改用另一套合成器。

## 已證實的封裝差異

保存 BGM_0F VGM SHA-256
`fcb09b274c7c8e7d1d8cb4e380ae9279b39357093ee54a9663a097ce1b6ee3e2`，
由既有 AdPlug 合成、FFmpeg 依既有 -1.00 dB 增益／44100 Hz／Vorbis q5 重編兩次：
OGG SHA-256 分別為 `5830b9202f175886a0a84362999367f8b20f06155056185d45d4c84a379e548a`、
`a306571df34ffb162615063cb2323bf1bace47e5aaa9ba7f592e66932af6bb24`，
而保存基準為 `48c22ed0b5d9fec5d0b9bed70e7a06ae1ab3a07faef99da80ddfd21b4644739c`。
完整解碼 PCM 與保存基準相同；128 頁逐頁比較確認只有 serial 與 CRC 不同，
頁面旗標、granule、頁序、分段及 Vorbis 封包均相同，沒有裁切或遮罩音樂差異。

serial 是 Ogg 頁面偏移 14–17，CRC 在 22–25；來源為
[Xiph Ogg framing](https://xiph.org/ogg/doc/framing.html)。
用已保存串流的固定 serial `0x4107795A` 重算各頁 CRC，
可丟棄探針驗整份結果逐 byte 回到原有基準。這是封裝 metadata 的可重現修正，
不是複製舊 OGG、修改 Vorbis 內容或放寬音樂驗收條件。

## 完整入口實跑

目前執行器的 BGM_0F 固定序列 1.6B 指令完整擷取已完成，VGM 全檔與基準相同；
實際耗時約 209 秒，原始 lin poke 日誌位址符合契約。
其後由原生完整入口重錄十一首 FM，全部 VGM／OGG 全檔符合保存基準；
BGM_0F、END_2F、OPEN_F 再獨立重跑，三種路徑的 VGM／OGG 逐位元組相同。
十一首 GM 也從原版 MDS 真實合成；二十二首經正式播放器完整解碼及正常路徑選曲通過。
原生程式不呼叫 Docker／shell／Python，不複製保存的 VGM／OGG 到玩家輸出。
四份原版 EXE、每曲 MDS、probe clean HEAD／程式雜湊與固定 flags 均綁入收據；
來源 bundle 與已鎖定原版輸入的雜湊不變。全部參數與輸出見[驗證 031](../playtest/031-native-fm-import.md)。

## 證據等級與停止線

已證實：本輪固定原版 cue 擷取的完整 OPL 暫存器串／VGM、同工具合成編碼後
固定封裝的完整 OGG 與保存基準相同；三條觸發路徑可獨立重現。
FM 原版驅動 cue／暫存器串保留已證實級，
時間換算與合成聲音保持 hardware-spec approximation，不展開硬體逐週期逆向。
OPEN_F 是暫存副本替換 MDS 的診斷路徑，非新增正常玩家開場畫面驗收。
本輪完整曲庫匯入及正式解碼通過，不宣稱平台依賴封裝、實機或聆聽已完成。
