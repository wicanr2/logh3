# 驗證矩陣

完成度的**唯一數字來源**。只有拿得出對拍收據的項目才能標通過；
「remake 單元測試全綠」「目視相似」「離線素材渲染成功」都不算。

比較等級定義見 `AGENTS.md`。

## 畫面

| # | 畫面 | 等級 | 未核准差異 | 收據 | 備註 |
|---|---|---|---:|---|---|
| 1 | 主選單（含反白與游標）| `same-state` | 0 | [`docs/playtest/001-main-menu.md`](docs/playtest/001-main-menu.md) | **五個狀態**逐位元組一致：反白第 0–3 列各一、游標在選單外一個 |
| 2 | 開頭動畫 | — | — | — | 未開始 |
| 3a | 正常玩家路徑（選單→防拷→本體）| `same-state` | 0 | [`docs/playtest/005-player-path.md`](docs/playtest/005-player-path.md) | 三個畫面都是前一步的輸入帶進去的，不是 direct-entry |
| 3 | 主選單（互動）| — | — | — | 進行中：反白、滑鼠→選項對應、游標都已對拍；點擊已能推進到下一畫面；還缺**切換過程**（反白從 A 列搬到 B 列會留殘留）|
| 4b | 防拷問答（輸入中的數字）| — | — | 收據 `typed138` | **未通過**：位置量到了，但那套數字字型不是 `GRAPH.IMG` 的半形字型，來源沒找到（`docs/re/09`）|
| 4 | 防拷問答 | `same-state` | 0 | [`docs/playtest/003-copy-protection.md`](docs/playtest/003-copy-protection.md) | 全畫面 640×480 逐位元組一致（框、頭像、三行字、輸入框） |
| 5 | 遊戲主畫面（796 年 1 月）| `same-state` | 0 | [`docs/playtest/004-main-game.md`](docs/playtest/004-main-game.md) | 全畫面逐位元組一致；地圖、頭像、文字全部從原版檔案重建 |
| 6 | 結局 | — | — | — | 未開始 |

## 資料格式

| 格式 | 狀態 | 文件 |
|---|---|---|
| `GRAPH.IMG`（中文字型）| **已解出**（假名區的字碼對應未確認）| [`docs/internals/06`](docs/internals/06-font-graph-img.md) |
| `.RGB`（16 色調色盤）| **已解出** | [`docs/internals/05`](docs/internals/05-palette-and-fade.md) |
| `SPRITE.SPR` | **已解出**（75 張 16×16） | [`docs/internals/10`](docs/internals/10-sprite-spr.md) |
| `3DBLOCK.SPR` / `COM_FLAG.SPR` | 未解（不是 `SPRITE.SPR` 的版面） | — |
| `STRMAP.BIT`（星域地圖）| **已解出**（320×300；記憶體版本與原版畫面 0 差異）| [`docs/internals/12`](docs/internals/12-strmap-bit.md) |
| `FRAME.BLK` / `BLOCK_*.BLK` | 未開始 | — |
| `STRMAP.RGN`（星域區域遮罩）| **已解出**（32 區域；底圖＋遮罩＋顏色重建 0 差異）| [`docs/internals/13`](docs/internals/13-strmap-rgn.md) |
| `MAP*.CMP` | 未解 | — |
| `PLADAT.DAT` | **已解出** | [`docs/internals/data/pladat.json`](docs/internals/data/pladat.json) |
| `ATTR.DAT` / `KEIREKI.DAT` | 未開始 | — |
| `ADMIRAL.DT*` | **版面已解出**（欄位語意只確認 `+11` ＝ 占領能力）| [`docs/internals/08`](docs/internals/08-admiral-dt.md) |
| `STELLAR.DT*` / `SHIP.DT*` / `FLEET.DT*` / `PLANET.DT*` | 未開始 | — |
| `*.MDS`（音樂，22 份）| 未開始 | — |
| `*.VO5`（結局劇情，50 份）| 未開始 | — |
| `FACE.AMG`（.AMG 圖庫）| **已解出**（83 張逐像素對得上原版畫面）| [`docs/internals/09`](docs/internals/09-face-amg.md) |
| `COMANI.BMG` / `IBM16.350` | 未開始 | — |

## 執行器（dosgolem）覆蓋

| 項目 | 狀態 |
|---|---|
| 跑到主選單並停在互動迴圈 | 通過（6000 萬道指令，可重現；120M 步狀態不變）|
| 鍵盤（IRQ1 ＋ 埠 0x60）| 通過（送得進遊戲自己的 int 09h；遊戲不吃鍵是另一回事）|
| 滑鼠點擊（虛擬座標倍率依模式）| 通過（點在選單上會改變畫面）|
| 滑鼠事件回呼（`int 33h AX=000Ch`）| 通過（主選單只走這條路收按下；dosgolem `docs/spec/013`）|
| 走完開場 → 主選單 → 防拷 → 遊戲本體 | 通過（[`docs/playtest/002`](docs/playtest/002-into-the-game.md)，兩次逐位元組一致）|
| 未實作的 BIOS／DOS 服務 | 0 種 |
| SingleStepTests | 未在本專案重跑（語料未下載）|
