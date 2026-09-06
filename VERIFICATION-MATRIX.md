# 驗證矩陣

完成度的**唯一數字來源**。只有拿得出對拍收據的項目才能標通過；
「remake 單元測試全綠」「目視相似」「離線素材渲染成功」都不算。

比較等級定義見 `AGENTS.md`。

## 畫面

| # | 畫面 | 等級 | 未核准差異 | 收據 | 備註 |
|---|---|---|---:|---|---|
| 1 | 主選單（含反白與游標）| `same-state` | 0 | [`docs/playtest/001-main-menu.md`](docs/playtest/001-main-menu.md) | **五個狀態**逐位元組一致：反白第 0–3 列各一、游標在選單外一個 |
| 2 | 開頭動畫 | — | — | — | 進行中：374 格已切分（[`docs/internals/19`](docs/internals/19-ibm16-350.md)），mode 10h 640×350；每一格的編碼未解 |
| 3a | 正常玩家路徑（選單→防拷→本體）| `same-state` | 0 | [`docs/playtest/005-player-path.md`](docs/playtest/005-player-path.md) | 三個畫面都是前一步的輸入帶進去的，不是 direct-entry |
| 3 | 主選單（互動）| — | — | — | 進行中：反白、滑鼠→選項對應、游標都已對拍；點擊已能推進到下一畫面；還缺**切換過程**（反白從 A 列搬到 B 列會留殘留）|
| 4b | 防拷問答（輸入中的數字）| — | — | 收據 `typed138` | **未通過**：位置量到了，但那套數字字型不是 `GRAPH.IMG` 的半形字型，來源沒找到（`docs/re/09`）|
| 4 | 防拷問答 | `same-state` | 0 | [`docs/playtest/003-copy-protection.md`](docs/playtest/003-copy-protection.md) | 全畫面 640×480 逐位元組一致（框、頭像、三行字、輸入框） |
| 5 | 遊戲主畫面（796 年 1 月）| `same-state` | 0 | [`docs/playtest/004-main-game.md`](docs/playtest/004-main-game.md) | 全畫面逐位元組一致；地圖、頭像、文字全部從原版檔案重建 |
| 5b | 船艦訂造報告（796 年 1 月）| `same-state` | 0 | 收據 `ship1`；同一個繪製器換規格 | 由主畫面熱區 (432,320) 點進去 |
| 5c | 次月份船艦建造計劃 | `same-state` | 0 | 收據 `plan1` | 由訂造報告熱區 (432,300) 點進去 |
| 5e | 次月份預算計劃 | `same-state` | 0 | 收據 `budget1`；[`docs/re/12`](docs/re/12-budget-plan-and-buttons.md) | 由主畫面熱區 (432,300) 點進去；正常玩家路徑有回歸測試 |
| 5f | 待命中提督資料 | `same-state` | 0 | 收據 `admirals1`；[`docs/re/12`](docs/re/12-budget-plan-and-buttons.md) | 三層視窗（底板＋建造計劃＋清單）；正常玩家路徑有回歸測試 |
| 5d | 建造計劃（改過預算的 34 個局面）| `same-state` | 0 | [`docs/re/11`](docs/re/11-build-plan-rest-amount.md) | 用 `-poke` 把百分比寫進原版記憶體造的局面，含手點到不了的「七列全 0%」與「某列吃滿 100%」；每張 640×480 逐位元組一致 |
| 6 | 結局 | `reference-only` | — | 收據 `ending-empire`、`ending-alliance` | 兩個結局都能在執行器裡跑到並擷取（`ENDING.EXE 9 1` / `9 2`）；50 張 CG 已解出（[`docs/internals/18`](docs/internals/18-vo5.md)），文字與職員表已取出（[`docs/internals/16`](docs/internals/16-ending-exe.md)）。remake 這邊還沒畫 |

## 資料格式

| 格式 | 狀態 | 文件 |
|---|---|---|
| `GRAPH.IMG`（中文字型）| **已解出**（假名區的字碼對應未確認）| [`docs/internals/06`](docs/internals/06-font-graph-img.md) |
| `.RGB`（16 色調色盤）| **已解出** | [`docs/internals/05`](docs/internals/05-palette-and-fade.md) |
| `SPRITE.SPR` | **已解出**（75 張 16×16） | [`docs/internals/10`](docs/internals/10-sprite-spr.md) |
| `3DBLOCK.SPR` | **已解出**（48×64，4 plane 逐列交錯，四塊等距方塊）| [`docs/internals/21`](docs/internals/21-3dblock-spr.md) |
| `COM_FLAG.SPR` | 未解（每 8 bytes 有 4 bytes 是 0，不是逐列交錯）| [`docs/internals/21`](docs/internals/21-3dblock-spr.md) |
| `STRMAP.BIT`（星域地圖）| **已解出**（320×300；記憶體版本與原版畫面 0 差異）| [`docs/internals/12`](docs/internals/12-strmap-bit.md) |
| `FRAME.BLK` / `BLOCK_*.BLK` | 未開始 | — |
| `STRMAP.RGN`（星域區域遮罩）| **已解出**（32 區域；底圖＋遮罩＋顏色重建 0 差異）| [`docs/internals/13`](docs/internals/13-strmap-rgn.md) |
| `MAP*.CMP` | 未解 | — |
| `PLADAT.DAT` | **已解出** | [`docs/internals/data/pladat.json`](docs/internals/data/pladat.json) |
| `KEIREKI.DAT`（人物傳）| **已解出**（82 篇，位移表 ＋ Big5 內文）| [`docs/internals/15`](docs/internals/15-keireki-dat.md) |
| `ATTR.DAT` | 未開始 | — |
| `ADMIRAL.DT*` | **版面已解出**（欄位語意確認 `+1` ＝ 陣營、`+11` ＝ 占領能力）| [`docs/internals/08`](docs/internals/08-admiral-dt.md) |
| `STELLAR.DT*` / `SHIP.DT*` / `FLEET.DT*` / `PLANET.DT*` | 未開始 | — |
| `*.MDS`（音樂，22 份）| **已解出**（22 首轉成 MIDI；速度數值的單位未確認）| [`docs/internals/17`](docs/internals/17-mds.md) |
| `*.VO5`（結局圖，50 張）| **已解出**（逐 byte 的相等即重複；50/50 位元組完全用完）| [`docs/internals/18`](docs/internals/18-vo5.md) |
| `ENDING.EXE`（結局文字、職員表）| **已解出** | [`docs/internals/16`](docs/internals/16-ending-exe.md) |
| `FACE.AMG`（.AMG 圖庫）| **已解出**（83 張逐像素對得上原版畫面）| [`docs/internals/09`](docs/internals/09-face-amg.md) |
| `IBM16.350` ＋ `IBMFLS16.350`（開頭動畫）| 影格切分已解出（374 格，長度總和逐位元組吻合）；每格的編碼未解 | [`docs/internals/19`](docs/internals/19-ibm16-350.md) |
| `COMANI.BMG`（戰鬥動畫）| 目錄已解出（253 段，位移逐筆接得起來）；每段的編碼未解 | [`docs/internals/20`](docs/internals/20-comani-bmg.md) |

## 遊戲規則

| 規則 | 狀態 | 證據 |
|---|---|---|
| 船艦建造預算分配（箭頭 ±1%、餘額不足不能加）| **已實作並對拍**（`internal/rules/buildplan.go`）| [`docs/re/11`](docs/re/11-build-plan-rest-amount.md)，35 個局面 |
| 金額 ＝ 建造費 × 百分比（捨去）| **已實作並對拍** | 同上 |
| 船艦數 ＝ 金額 ÷ 單價 × 10（七個單價已量出）| **已實作並對拍** | 同上 |
| 部隊 ＝ 船艦數 ÷ 定員（七個定員已量出）| **已實作並對拍** | 同上 |
| 餘額 ＝ 建造費 − Σ 各列實際花掉（零頭退回）| **已實作並對拍** | 同上 |
| 預估軍事費 ＝ 國庫 ＋ 預估徵稅額 | **已實作**（`internal/rules/budget.go`）| [`docs/re/12`](docs/re/12-budget-plan-and-buttons.md) |
| 船艦建造費 ＝ 預估軍事費 × 百分比（捨去）| **已實作** | 同上 |
| 剩餘軍事費 ＝ 預估軍事費 − 船艦建造費 | **已實作** | 同上 |
| 其他 | 未開始 | — |

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
