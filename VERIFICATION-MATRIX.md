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
| 5g | 月份會計報告 | `same-state` | 0 | 收據 `account1`；[`docs/re/13`](docs/re/13-account-report.md) | 由預算計劃 (472,325) 點進去；獨立驗證了三條錢的公式 |
| 5f | 待命中提督資料 | `same-state` | 0 | 收據 `admirals1`；[`docs/re/12`](docs/re/12-budget-plan-and-buttons.md) | 三層視窗（底板＋建造計劃＋清單）；正常玩家路徑有回歸測試 |
| 5d | 建造計劃（改過預算的 34 個局面）| `same-state` | 0 | [`docs/re/11`](docs/re/11-build-plan-rest-amount.md) | 用 `-poke` 把百分比寫進原版記憶體造的局面，含手點到不了的「七列全 0%」與「某列吃滿 100%」；每張 640×480 逐位元組一致 |
| 5h | 命令階段主畫面（796 年 1 月）| `same-state` | 0 | [`docs/playtest/006`](docs/playtest/006-command-phase.md) | 全畫面逐位元組一致。右鍵兩下關掉月初報告才到得了（[`docs/re/16`](docs/re/16-window-list-right-button.md)）；收據用 `main1e`（游標搬過一次，沒有殘留）|
| 5i | 戰略命令選單（情報收集／徵稅計劃／生產計劃）| `same-state` | 0 | [`docs/playtest/006`](docs/playtest/006-command-phase.md) | 收據 `cmd1e`；主畫面當底層，只量疊上去那一層 |
| 5j | 戰略情報報告 | `same-state` | 0 | [`docs/playtest/006`](docs/playtest/006-command-phase.md) | 收據 `info1e`；地圖跟著**治安**換色（色表在 `strmap-regions.json` 的 `colour_796_01_security`），頭像是 `FACE.AMG` 第 13 張 |
| 5l | 一覽表：艦隊資料／提督資料／敵將資料 | `same-state` | 0 | [`docs/playtest/006`](docs/playtest/006-command-phase.md) | 收據 `lst1e`／`lst2e`／`lst3e`；三張共用同一份清單版面 |
| 5k | 遊戲設定 | `same-state` | 0 | [`docs/playtest/006`](docs/playtest/006-command-phase.md) | 收據 `settings1e`；標籤是**粗體＋陰影**畫的（[`docs/internals/25`](docs/internals/25-text-rendering.md)），不是另一套字型 |
| 5m | 徵稅計劃 → 次月份預算計劃 | `same-state` | 0 | [`docs/playtest/006`](docs/playtest/006-command-phase.md) | 收據 `tax1e`；與 `budget1` 同一個視窗版面，底下換成命令階段底板 |
| 5n | 生產計劃 → 次月份船艦建造計劃 | `same-state` | 0 | [`docs/playtest/006`](docs/playtest/006-command-phase.md) | 收據 `prod1e`（＋`prod1i` 補游標蓋住的那一格）|
| 5o | 取進度 | `same-state` | 0 | [`docs/playtest/006`](docs/playtest/006-command-phase.md) | 收據 `load1e`；8 個未使用欄位（色 8）＋ 三個劇本 |
| 5p | 存進度 | `same-state` | 0 | [`docs/playtest/006`](docs/playtest/006-command-phase.md) | 收據 `save1e`；滑鼠所在那一列是紅色反白條，刷的時候把游標蓋掉 |
| 5q | 出擊命令 → 待命中提督資料 | `same-state` | 0 | [`docs/playtest/006`](docs/playtest/006-command-phase.md) | 收據 `sortie1e`；清單十七列，左下角那塊被重寫成星系摘要 |
| 5r | 結束（確認對話框）| `same-state` | 0 | [`docs/playtest/006`](docs/playtest/006-command-phase.md) | 收據 `quit1e`；「您確定要結束這個遊戲嗎？」＋ 兩個按鈕 |
| 5s | 一覽表 → 星系資料 | `same-state` | 0 | [`docs/playtest/006`](docs/playtest/006-command-phase.md) | 收據 `star1e`；地圖用第三組星域色（`colour_796_01_star_list`），選中的星域畫成色 7 |
| 5t | 兩張下拉選單（系統指令、一覽表）| `same-state` | 0 | [`docs/playtest/006`](docs/playtest/006-command-phase.md) | 收據 `sysmenu1e`、`lstmenu1e`；系統指令五項（實行／取進度／存進度／設定／結束）、一覽表四項 |
| 3b | 命令階段的正常玩家路徑（12 張）| `same-state` | 0 | [`docs/playtest/005`](docs/playtest/005-player-path.md) | 右鍵兩下關掉月初報告 → 命令列三格 → 選單每一項；每一張都是前一步的輸入帶進去的，不是 direct-entry |
| 6 | 結局（單張 CG）| `same-state`（plane 2／3）| 0 | 收據 `ending-empire`、`ending-alliance` | remake 畫得出來，plane 2／3 與原版逐位元組相同；plane 0／1 是多張圖疊出來的，疊法未解。播放順序（哪張配哪句、換場）也未解 |

## 資料格式

| 格式 | 狀態 | 文件 |
|---|---|---|
| `GRAPH.IMG`（中文字型）| **已解出**（假名區的字碼對應未確認）| [`docs/internals/06`](docs/internals/06-font-graph-img.md) |
| `.RGB`（16 色調色盤）| **已解出** | [`docs/internals/05`](docs/internals/05-palette-and-fade.md) |
| `SPRITE.SPR` | **已解出**（75 張 16×16） | [`docs/internals/10`](docs/internals/10-sprite-spr.md) |
| `3DBLOCK.SPR` | **已解出**（48×64，4 plane 逐列交錯，四塊等距方塊）| [`docs/internals/21`](docs/internals/21-3dblock-spr.md) |
| `COM_FLAG.SPR` | 未解（每 8 bytes 有 4 bytes 是 0，不是逐列交錯）| [`docs/internals/21`](docs/internals/21-3dblock-spr.md) |
| `STRMAP.BIT`（星域地圖）| **已解出**（320×300；記憶體版本與原版畫面 0 差異）| [`docs/internals/12`](docs/internals/12-strmap-bit.md) |
| `FRAME.BLK` / `BLOCK_*.BLK` | 未解（全彩非遮罩；32×32／320×128／640×64 等切法都對不齊）| — |
| `STRMAP.RGN`（星域區域遮罩）| **已解出**（32 區域；底圖＋遮罩＋顏色重建 0 差異）| [`docs/internals/13`](docs/internals/13-strmap-rgn.md) |
| `MAP*.CMP` | 未解；**不是星圖的來源**（走到主畫面為止一份都沒被讀）| [`docs/re/17`](docs/re/17-star-region-owner.md) |
| `PLADAT.DAT` | **已解出** | [`docs/internals/data/pladat.json`](docs/internals/data/pladat.json) |
| `KEIREKI.DAT`（人物傳）| **已解出**（82 篇，位移表 ＋ Big5 內文）| [`docs/internals/15`](docs/internals/15-keireki-dat.md) |
| `ATTR.DAT` | 版面已觀察（128 bytes，只有 `00`／`0F`；前 8 個全 0，接著三組「六個 `0F` ＋ 兩個 `00`」，其餘全 `0F`）；語意未解 | — |
| `ADMIRAL.DT*` | **版面已解出**（欄位語意確認 `+1` ＝ 陣營、`+11` ＝ 占領能力）| [`docs/internals/08`](docs/internals/08-admiral-dt.md) |
| `STELLAR.DT*` | **版面已解出**（32 筆 × 126 bytes，32 個星域名；`+18` ＝ 擁有者，其餘數值欄位語意未知）| [`docs/internals/22`](docs/internals/22-stellar-dt.md) |
| `*.MDS`（音樂，22 份）| **已解出**（22 首轉成 MIDI；速度數值的單位未確認）| [`docs/internals/17`](docs/internals/17-mds.md) |
| `*.VO5`（結局圖，50 張）| **已解出**（逐 byte 的相等即重複；50/50 位元組完全用完）| [`docs/internals/18`](docs/internals/18-vo5.md) |
| `ENDING.EXE`（結局文字、職員表）| **已解出** | [`docs/internals/16`](docs/internals/16-ending-exe.md) |
| `FACE.AMG`（.AMG 圖庫）| **已解出**（83 張逐像素對得上原版畫面）| [`docs/internals/09`](docs/internals/09-face-amg.md) |
| `IBM16.350` ＋ `IBMFLS16.350`（開頭動畫）| 影格切分已解出（374 格，長度總和逐位元組吻合）；每格的編碼未解 | [`docs/internals/19`](docs/internals/19-ibm16-350.md) |
| `COMANI.BMG`（戰鬥動畫）| 目錄已解出（253 段，位移逐筆接得起來）；每段的編碼未解 | [`docs/internals/20`](docs/internals/20-comani-bmg.md) |

## 介面覆蓋

| 項目 | 狀態 | 證據 |
|---|---|---|
| 資料段裡的熱區清單 | **25 份全部掃出**（含每個按鈕的中文標籤）| [`docs/internals/23`](docs/internals/23-hotspot-lists.md) |
| 清單 ↔ 畫面的對應 | 25 份裡只確定 2 份（主畫面、建造計劃）| 同上 |

## 遊戲規則

| 規則 | 狀態 | 證據 |
|---|---|---|
| 船艦建造預算分配（箭頭 ±1%、餘額不足不能加）| **已實作並對拍**（`internal/rules/buildplan.go`）| [`docs/re/11`](docs/re/11-build-plan-rest-amount.md)，35 個局面 |
| 金額 ＝ 建造費 × 百分比（捨去）| **已實作並對拍** | 同上 |
| 船艦數 ＝ 金額 ÷ 單價 × 10（七個單價已量出）| **已實作並對拍** | 同上 |
| 部隊 ＝ 船艦數 ÷ 定員（七個定員已量出）| **已實作並對拍** | 同上 |
| 餘額 ＝ 建造費 − Σ 各列實際花掉（零頭退回）| **已實作並對拍** | 同上 |
| 預估軍事費 ＝ 國庫 ＋ 預估徵稅額 | **已實作**（`internal/rules/budget.go`），三組獨立資料 | [`docs/re/12`](docs/re/12-budget-plan-and-buttons.md)、[`13`](docs/re/13-account-report.md)、[`15`](docs/re/15-alliance-start.md) |
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

## 規則

只有拿得出「原版實測值當預期值」的回歸測試才算通過。

| 規則 | 狀態 | 證據 |
|---|---|---|
| 船艦建造預算分配（左欄 +2%、右欄 −2%、金額 ＝ 建造費 × 百分比捨去）| 通過 | [`docs/re/11`](docs/re/11-build-plan-rest-amount.md)；34 個 `-poke` 局面 |
| 預算計劃三條錢的公式 | 通過 | [`docs/re/12`](docs/re/12-budget-plan-and-buttons.md) |
| 星域擁有者（全屬同一方才算，混雜就是無主）| 通過 | [`docs/re/17`](docs/re/17-star-region-owner.md)；六個星域 |
| 星域防衛力 ＝ 該星域行星 `+0x14` 相加 | 通過 | [`docs/re/23`](docs/re/23-region-panel-fields.md)；六個星域對執行期 `+0x20` |
| 行星 `n/m` ＝ 玩家陣營行星數 / 總數 | 通過 | [`docs/re/23`](docs/re/23-region-panel-fields.md)；五個星域對畫面 |
| 徵稅額 ＝ 經濟力 × 該星域稅率 / 1000（捨去）| 通過 | [`docs/re/26`](docs/re/26-tax-and-treasury.md)；稅率是**每個星域各自的** |
| 總稅收 ＝ 自己擁有的星域徵稅額相加 | 通過 | [`docs/re/26`](docs/re/26-tax-and-treasury.md)；帝國 10341、同盟 11087 兩邊都對 |
| 月底軍事費的完整算式 | **未解** | [`docs/re/26`](docs/re/26-tax-and-treasury.md)；`軍事費 += 總稅收` 已確認，扣款那一步未追 |
| 星域相鄰圖與原版路徑表（32×32 中點矩陣 ×2）| 通過 | [`docs/internals/26`](docs/internals/26-star-region-graph.md)；1024 對逐步驗證 |
| 曆法：一回合三天、一個月十回合、一年十二個月 | 通過 | [`docs/re/25`](docs/re/25-calendar.md)；`ds:8F18` 回合計數器對三份記憶體傾印＋三張畫面 |
| 月底交付「建造中」的船艦 | 已量到、未實作 | [`docs/re/21`](docs/re/21-month-end-ship-delivery.md) |
| 治安 ＝ (200 − 未安定度) / 2 | 通過 | [`docs/re/24`](docs/re/24-security-formula.md)；六個帝國星域 |
| 情報報告地圖的治安色（表 `ds:1780` ＋ 分段算式）| 通過 | [`docs/re/27`](docs/re/27-security-map-colours.md)；32 個星域逐一相符（除數只框到 85–88）|
| 未安定度的開局值怎麼算出來 | **未解** | [`docs/re/24`](docs/re/24-security-formula.md) |
| 經濟力每月 +1002 的算式 | **未解** | [`docs/re/23`](docs/re/23-region-panel-fields.md) |
