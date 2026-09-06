# 驗證矩陣

完成度的**唯一數字來源**。只有拿得出對拍收據的項目才能標通過；
「remake 單元測試全綠」「目視相似」「離線素材渲染成功」都不算。

比較等級定義見 `AGENTS.md`。

## 畫面

| # | 畫面 | 等級 | 未核准差異 | 收據 | 備註 |
|---|---|---|---:|---|---|
| 1 | 主選單（靜態）| `same-state` | 0 | [`docs/playtest/001-main-menu.md`](docs/playtest/001-main-menu.md) | 逐位元組一致；沒有互動、反白、游標 |
| 2 | 開頭動畫 | — | — | — | 未開始 |
| 3 | 主選單（互動）| — | — | — | 未開始：反白、鍵盤／滑鼠、選項切換 |
| 4 | 遊戲主畫面 | — | — | — | 未開始 |
| 5 | 結局 | — | — | — | 未開始 |

## 資料格式

| 格式 | 狀態 | 文件 |
|---|---|---|
| `GRAPH.IMG`（中文字型）| **已解出**（假名區的字碼對應未確認）| [`docs/internals/06`](docs/internals/06-font-graph-img.md) |
| `.RGB`（16 色調色盤）| **已解出** | [`docs/internals/05`](docs/internals/05-palette-and-fade.md) |
| `SPRITE.SPR` / `3DBLOCK.SPR` / `COM_FLAG.SPR` | 未開始 | — |
| `FRAME.BLK` / `BLOCK_*.BLK` | 未開始 | — |
| `PLADAT.DAT` / `ATTR.DAT` / `KEIREKI.DAT` | 未開始 | — |
| `ADMIRAL.DT*` / `STELLAR.DT*` / `SHIP.DT*` / `FLEET.DT*` / `PLANET.DT*` | 未開始 | — |
| `MAP*.CMP`（32 份）| 未開始 | — |
| `*.MDS`（音樂，22 份）| 未開始 | — |
| `*.VO5`（結局劇情，50 份）| 未開始 | — |
| `COMANI.BMG` / `FACE.AMG` / `IBM16.350` | 未開始 | — |

## 執行器（dosgolem）覆蓋

| 項目 | 狀態 |
|---|---|
| 跑到主選單並停在互動迴圈 | 通過（6000 萬道指令，可重現）|
| 未實作的 BIOS／DOS 服務 | 0 種 |
| SingleStepTests | 未在本專案重跑（語料未下載）|
