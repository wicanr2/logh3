# 驗證矩陣

完成度的**唯一數字來源**。只有拿得出對拍收據的項目才能標通過；
「remake 單元測試全綠」「目視相似」「離線素材渲染成功」都不算。

比較等級定義見 `AGENTS.md`。

> 收據全數在 dosgolem `35d4cc3` 之後重做過（補上 VGA read mode 1，
> [`docs/re/49`](docs/re/49-lists-were-a-vga-read-mode-gap.md)）：178 份裡有
> 155 份變了，因為原版本來就有、但畫不出來的那些字現在都出現了。

## 顯示層

| # | 項目 | 等級 | 未核准差異 | 收據 | 備註 |
|---|---|---|---:|---|---|
| G1 | Ebiten 視窗版走玩家路徑 | `same-state` | 0 | `tools/gui_check.sh` | `cmd/gin3` 畫出來的色號陣列與 `r0`／`cb1`／`key1`／`cbAlly`／`keyAlly`／`main1e`／`load1e` 逐位元組相同。**「視窗開得起來」不算數**——比的是同一條玩家路徑跑完的畫面 |

## 畫面

| # | 畫面 | 等級 | 未核准差異 | 收據 | 備註 |
|---|---|---|---:|---|---|
| 1 | 主選單（含反白與游標）| `same-state` | 0 | [`docs/playtest/001-main-menu.md`](docs/playtest/001-main-menu.md) | **五個狀態**逐位元組一致：反白第 0–3 列各一、游標在選單外一個 |
| 2 | 開頭動畫 | — | — | — | 進行中：374 格已切分（[`docs/internals/19`](docs/internals/19-ibm16-350.md)），mode 10h 640×350；每一格的編碼未解 |
| 3a | 正常玩家路徑（選單→防拷→本體）| `same-state` | 0 | [`docs/playtest/005-player-path.md`](docs/playtest/005-player-path.md) | 三個畫面都是前一步的輸入帶進去的，不是 direct-entry |
| 3 | 主選單（互動）| — | — | — | 進行中：反白、滑鼠→選項對應、游標都已對拍；點擊已能推進到下一畫面；還缺**切換過程**（反白從 A 列搬到 B 列會留殘留）|
| 4b | 防拷問答（輸入中的數字）| `same-state` | 0 | 收據 `typed138`；[`docs/re/56`](docs/re/56-typed-digits-are-full-width.md) | 打進去的數字是 `GRAPH.IMG` 的**全形**數字（`A2AF` 起），不是半形——先前搜不到是因為只試了半形的打包方式 |
| 4 | 防拷問答 | `same-state` | 0 | [`docs/playtest/003-copy-protection.md`](docs/playtest/003-copy-protection.md) | 全畫面 640×480 逐位元組一致（框、頭像、三行字、輸入框） |
| 4z | **開頭動畫的解碼器**（`IBM16.350` 374 格）| — | — | [`docs/re/71`](docs/re/71-intro-chunks-are-cut-at-frame-boundaries.md)、[`docs/re/72`](docs/re/72-intro-decoder-solved.md)、[`docs/internals/19`](docs/internals/19-ibm16-350.md) | 解碼器對得起來（逐格長度相符、整檔吃光、兩處場景黑幕累加歸零、解出來是船艦）。**還不是畫面對拍**——與執行期緩衝區還沒對上（多半是 dump 落在一格中間），remake 也還沒把動畫接上去 |
| 5 | 遊戲主畫面（796 年 1 月）| `same-state` | 0 | [`docs/playtest/004-main-game.md`](docs/playtest/004-main-game.md) | 全畫面逐位元組一致；地圖、頭像、文字全部從原版檔案重建 |
| 5a | **同盟開局**（主選單第 1 列 → 防拷 → 月初會計報告 → 訂造報告 → 建造計劃 → 戰略地圖）| `same-state` | 0 | 收據 `cbAlly`／`keyAlly`／`shipAlly`／`planAlly`／`main1eAlly`；[`docs/spec/011`](docs/spec/011-opening-report.md)、[`docs/re/61`](docs/re/61-work-staff-is-faction-pairs.md)、[`docs/re/62`](docs/re/62-unit-prices-differ-by-faction.md) | 五張整張 640×480 逐位元組一致。署名、頭像、金額、右側首都摘要、星圖的首都特別色與**七列的船艦單價**都跟著陣營走 |
| 5a2 | **同盟的命令階段十二張**（含三個選單）| `same-state` | 0 | 收據 `*Ally` 十五份；[`docs/re/62`](docs/re/62-unit-prices-differ-by-faction.md)、[`docs/re/63`](docs/re/63-security-divisor-is-the-intel-staff.md)、[`docs/re/64`](docs/re/64-command-phase-follows-the-player.md) | 每一張整張 640×480 逐位元組一致，走的是同一條玩家路徑只換主選單那一列。署名、頭像、提督清單、星系清單順序、兩張地圖的著色、左下角摘要與貨幣單位都跟著陣營走 |
| 5d | **提督詳細資料與經歷**（出擊命令 → 待命中提督清單 → 點一列 → 經歷）| `same-state` | 0 | 收據 `adm1e`／`adm2e`／`adm3e`／`adm15e`／`keireki1e`；[`docs/spec/012`](docs/spec/012-admiral-detail.md)、[`docs/re/65`](docs/re/65-admiral-detail-screen.md) | 五張整張逐位元組一致。姓名、軍階、頭像、現在位置、旗艦、等級與十個能力值都跟著提督走；`ADMIRAL.DT*` 十四個能力欄位的名字就是這張畫面給的。「經歷」印的是 `KEIREKI.DAT` 的人物傳 |
| 5h | **艦隊編成**（提督詳細資料 → 出擊）| `same-state` | 0 | 收據 `fleet1e`／`fleet2e`；[`docs/spec/013`](docs/spec/013-fleet-organisation.md)、[`docs/re/66`](docs/re/66-fleet-organisation-screen.md) | 兩張整張 640×480 逐位元組一致。艦隊能力九格是提督與副官**逐項取大**，副官贏的印成色 5；副官 ＝ `ADMIRAL.DT*` 的 `+0x07` |
| 5i | **艦隊編號清單**與**副官選單**（艦隊編成的兩顆按鈕）| `same-state` | 0 | 收據 `fnum`／`fnum3`／`fadj`／`fadj3` ＋ 兩個反例 `fnum2`／`fadj2`；[`docs/re/67`](docs/re/67-fleet-org-buttons.md)、[`docs/re/68`](docs/re/68-hotspot-origin-rule.md) | 四張整張逐位元組一致；反白條跟著滑鼠走，由 `hover` 版面算。另外兩張是**反例**：點在畫出來的框內、熱區外一個像素，畫面與 `fleet1e` 完全相同 |
| 5j | **艦隊組成**（艦隊編成 → 確認）| `same-state` | 0 | 收據 `fok`；[`docs/spec/015`](docs/spec/015-fleet-composition.md)、[`docs/re/69`](docs/re/69-region-ship-strength.md) | 整張 640×480 逐位元組一致。9×9 格盤鋪的是 `SPRITE.SPR` 圖塊；「待命部隊」＝ 星域記錄 `+0x30` 的艦艇艘數 ÷ 一單位幾艘 |
| 5k | **陣形清單**（艦隊組成 → 陣形）| `same-state` | 0 | 收據 `form1`／`form2`／`form3`；[`docs/re/70`](docs/re/70-formation-list-and-decide.md) | 三張整張逐位元組一致。八列圓錐／紡錘／半月／Ｕ字／Ｍ字／Ｔ字／分散／密集，反白條跟著滑鼠走 |
| 5l | **艦隊建成**（艦隊組成 → 決定）| `same-state` | 0 | 收據 `dec1`、`decwatch`；[`docs/re/70`](docs/re/70-formation-list-and-decide.md)、[`docs/internals/33`](docs/internals/33-fleet-table.md) | 整張逐位元組一致。右下角清單多出「第１艦隊　萊因哈特　100」；左下角的首都摘要留在畫面上不被擦掉 |
| 5b | 船艦訂造報告（796 年 1 月）| `same-state` | 0 | 收據 `ship1`；同一個繪製器換規格 | 由主畫面熱區 (432,320) 點進去 |
| 5c | 次月份船艦建造計劃 | `same-state` | 0 | 收據 `plan1` | 由訂造報告熱區 (432,300) 點進去 |
| 5e | 次月份預算計劃 | `same-state` | 0 | 收據 `budget1`；[`docs/spec/003`](docs/spec/003-budget-panel.md) | 由主畫面熱區 (432,300) 點進去；十個數字**從財政狀態算出來**，不是量好畫死的 |
| 5g | 月份會計報告 | `same-state` | 0 | 收據 `account1`；[`docs/re/13`](docs/re/13-account-report.md) | 由預算計劃 (472,325) 點進去；獨立驗證了三條錢的公式 |
| 5f | 待命中提督資料 | `same-state` | 0 | 收據 `admirals1`；[`docs/re/12`](docs/re/12-budget-plan-and-buttons.md) | 三層視窗（底板＋建造計劃＋清單）；正常玩家路徑有回歸測試 |
| 5d | 建造計劃（改過預算的 34 個局面）| `same-state` | 0 | [`docs/re/11`](docs/re/11-build-plan-rest-amount.md) | 用 `-poke` 把百分比寫進原版記憶體造的局面，含手點到不了的「七列全 0%」與「某列吃滿 100%」；每張 640×480 逐位元組一致 |
| 5h | 命令階段主畫面（796 年 1 月）| `same-state` | 0 | [`docs/playtest/006`](docs/playtest/006-command-phase.md) | 全畫面逐位元組一致。右鍵兩下關掉月初報告才到得了（[`docs/re/16`](docs/re/16-window-list-right-button.md)）；收據用 `main1e`（游標搬過一次，沒有殘留）|
| 5i | 戰略命令選單（情報收集／徵稅計劃／生產計劃）| `same-state` | 0 | [`docs/playtest/006`](docs/playtest/006-command-phase.md) | 收據 `cmd1e`；主畫面當底層，只量疊上去那一層 |
| 5j | 戰略情報報告 | `same-state` | 0 | [`docs/playtest/006`](docs/playtest/006-command-phase.md) | 收據 `info1e`；游標最後一列壓在第一個分頁鈕的上緣立體邊上，所以視窗的純色塊要畫在分頁鈕之後（[`docs/re/53`](docs/re/53-cursor-tail-under-a-tab-bevel.md)）。地圖跟著**治安**換色（色表在 `strmap-regions.json` 的 `colour_796_01_security`），頭像是 `FACE.AMG` 第 13 張 |
| 5l | 一覽表：艦隊資料／提督資料／敵將資料 | `same-state` | 0 | 收據 `lst1e`／`lst2e`／`lst3e`；[`docs/re/50`](docs/re/50-admiral-lists.md) | 三張共用同一份清單版面。提督與敵將那兩張的十六列**從原版資料檔算出來**：姓名、軍階（`ADMIRAL` 的 `+6` 分段）、幕僚職記號（`WORK` 的 `+0x37`）|
| 5k | 遊戲設定 | `same-state` | 0 | [`docs/playtest/006`](docs/playtest/006-command-phase.md) | 收據 `settings1e`；那張畫面同時逼出兩個字型層的錯：Big5 次常用字的 408 碼斷層、以及半形格是兩個 8×15 的字（[`docs/re/51`](docs/re/51-settings-screen-two-font-bugs.md)）。標籤是**粗體＋陰影**畫的（[`docs/internals/25`](docs/internals/25-text-rendering.md)），不是另一套字型 |
| 5m | 徵稅計劃 → 次月份預算計劃 | `same-state` | 0 | 收據 `tax1e`；[`docs/spec/003`](docs/spec/003-budget-panel.md) | 與 `budget1` 同一個視窗版面，底下換成命令階段底板；數字與右側面板都是算出來的 |
| 5n | 生產計劃 → 次月份船艦建造計劃 | `same-state` | 0 | 收據 `prod1e`；[`docs/spec/004`](docs/spec/004-production-panel.md) | 44 個數字**全部從 `rules.BuildPlan` 算**，含「建造中」那兩欄（上一份計劃買到的船）|
| 5o | 取進度 | `same-state` | 0 | [`docs/playtest/006`](docs/playtest/006-command-phase.md) | 收據 `load1e`；8 個未使用欄位（色 8）＋ 三個劇本 |
| 5p | 存進度 | `same-state` | 0 | [`docs/playtest/006`](docs/playtest/006-command-phase.md) | 紅條那張（`TestSaveBarMatchesOracle`）與正常玩家路徑那張都 0 點。這張逼出了「命中判定用畫面座標、游標繪製用滑鼠座標」的抵銷（[`docs/re/52`](docs/re/52-two-mouse-coordinate-conventions.md)）。收據 `save1e`；滑鼠所在那一列是紅色反白條，刷的時候把游標蓋掉 |
| 5q | 出擊命令 → 待命中提督資料 | `same-state` | 0 | [`docs/playtest/006`](docs/playtest/006-command-phase.md) | 收據 `sortie1e`；清單十七列，左下角那塊被重寫成星系摘要 |
| 5w | 月底船艦接收報告 | `same-state`（兩塊）| 0 | [`docs/re/55`](docs/re/55-ship-delivery-report.md) | 第十次「實行」跨月時跳出來。收據 `turnmonth2`；對話框 (144,80)–(495,399) 與「戰略模式處理中」那條 (320,40)–(639,79) 逐位元組。十四個數字、標題的年月、署名的姓名與軍階**全部算出來**；整張比不了的三處差異（兩個星域換色、首都色、捲軸箭頭）在 `docs/re/54`／`55` 列了成因 |
| 5r | 結束（確認對話框）| `same-state` | 0 | [`docs/playtest/006`](docs/playtest/006-command-phase.md) | 收據 `quit1e`；「您確定要結束這個遊戲嗎？」＋ 兩個按鈕 |
| 5s | 一覽表 → 星系資料 | `same-state` | 0 | [`docs/playtest/006`](docs/playtest/006-command-phase.md) | 收據 `star1e`；三十二列的星域名與艦隊數是算出來的，列的順序來自資料段 `ds:1DB2` 的排列表（[`docs/re/50`](docs/re/50-admiral-lists.md)）。地圖用第三組星域色（`colour_796_01_star_list`），選中的星域畫成色 7 |
| 5t | 兩張下拉選單（系統指令、一覽表）| `same-state` | 0 | [`docs/playtest/006`](docs/playtest/006-command-phase.md) | 收據 `sysmenu1e`、`lstmenu1e`；系統指令五項（實行／取進度／存進度／設定／結束）、一覽表四項 |
| 3b | 命令階段的正常玩家路徑（12 張）| `same-state` | 0 | [`docs/playtest/005`](docs/playtest/005-player-path.md) | 右鍵兩下關掉月初報告 → 命令列三格 → 選單每一項；每一張都是前一步的輸入帶進去的，不是 direct-entry |
| 5v | 右側面板（日期、軍事費、總稅收、首都摘要、防衛力、經濟力、成長率、徵稅額、稅率）| `same-state` | 0 | [`docs/spec/002`](docs/spec/002-capital-panel.md)、[`docs/re/32`](docs/re/32-capital-panel.md) | 文字**從遊戲狀態算出來**。`main1e`（796/1/1）整張 640×480 逐位元組；`turn260`／`turn320`／`turn450`（796/1/4，三份收據逐位元組相同）、`turnnext`（796/2/1）、`s2empire`（798/11/19）與 `s3rgn`（**799/3/1**，S3「巴米利恩會戰」）右側面板逐位元組——同一份規格換五組狀態畫出五張不同的畫面，橫跨三年 |
| 5u | 星域摘要面板（右鍵選星域，16 個局面）| `same-state` | 0 | [`docs/spec/001`](docs/spec/001-region-summary-panel.md)、[`docs/re/31`](docs/re/31-region-panel-layout.md) | 文字**從規則算出來**，不是量好畫死的。796/1 那 15 份整張比，跳過游標的外接矩形（原版右鍵之後游標只補回隔列）；含七個 `-poke` 造的局面。`s3rgn`（**799/3/1**，S3「巴米利恩會戰」，星域 25）狀態讀那一格存檔，比面板那塊 320×80 |
| 6 | 結局（單張 CG）| `same-state`（plane 2／3）| 0 | 收據 `ending-empire`、`ending-alliance` | remake 畫得出來，plane 2／3 與原版逐位元組相同；plane 0／1 是多張圖疊出來的，疊法未解。播放順序（哪張配哪句、換場）也未解 |

## 資料格式

| 格式 | 狀態 | 文件 |
|---|---|---|
| `GRAPH.IMG`（中文字型）| **已解出**（假名區的字碼對應未確認）| [`docs/internals/06`](docs/internals/06-font-graph-img.md) |
| `.RGB`（16 色調色盤）| **已解出** | [`docs/internals/05`](docs/internals/05-palette-and-fade.md) |
| `SPRITE.SPR` | **已解出**（75 張 16×16） | [`docs/internals/10`](docs/internals/10-sprite-spr.md) |
| `3DBLOCK.SPR` | **已解出**（48×64，4 plane 逐列交錯，四塊等距方塊）| [`docs/internals/21`](docs/internals/21-3dblock-spr.md) |
| `COM_FLAG.SPR` | **已解出**（64×45，四個 plane 以「一列一個 plane」交錯；內容是五個戰鬥指令字樣 `BEEM`／`MISSILE`／`FIGHTER`／`ATTACK`／`DEFENCE`，第一個是原版錯字）| [`docs/internals/21`](docs/internals/21-3dblock-spr.md) |
| `STRMAP.BIT`（星域地圖）| **已解出**（320×300；記憶體版本與原版畫面 0 差異）| [`docs/internals/12`](docs/internals/12-strmap-bit.md) |
| `FRAME.BLK` / `BLOCK_*.BLK` | 未解（全彩非遮罩；32×32／320×128／640×64 等切法都對不齊）| — |
| `STRMAP.RGN`（星域區域遮罩）| **已解出**（32 區域；底圖＋遮罩＋顏色重建 0 差異）| [`docs/internals/13`](docs/internals/13-strmap-rgn.md) |
| `MAP*.CMP` | 未解；**不是星圖的來源**（走到主畫面為止一份都沒被讀）| [`docs/re/17`](docs/re/17-star-region-owner.md) |
| `PLADAT.DAT` | **已解出**（檔頭 ＝ 每星域行星數；`+0x11` 擁有者、`+0x14` 防衛力、`+0x1A` 經濟力）| [`docs/internals/data/pladat.json`](docs/internals/data/pladat.json) |
| `KEIREKI.DAT`（人物傳）| **已解出**（82 篇，位移表 ＋ Big5 內文）| [`docs/internals/15`](docs/internals/15-keireki-dat.md) |
| `ATTR.DAT` | 版面已觀察（128 bytes，只有 `00`／`0F`；前 8 個全 0，接著三組「六個 `0F` ＋ 兩個 `00`」，其餘全 `0F`）；語意未解 | — |
| `ADMIRAL.DT*` | **版面已解出**（欄位語意確認 `+1` ＝ 陣營、`+11` ＝ 占領能力）| [`docs/internals/08`](docs/internals/08-admiral-dt.md) |
| `STELLAR.DT*` | **版面已解出**（32 筆 × 126 bytes；擁有者、行星數、經濟力、成長率、防衛力、稅率、艦隊、未安定度已對出，檔案裡 `+0x13` 之後仍有未對過的欄位）| [`docs/internals/22`](docs/internals/22-stellar-dt.md) |
| `*.MDS`（音樂，22 份）| **已解出**（22 首轉成 MIDI；速度數值的單位未確認）| [`docs/internals/17`](docs/internals/17-mds.md) |
| `*.VO5`（結局圖，50 張）| **已解出**（逐 byte 的相等即重複；50/50 位元組完全用完）| [`docs/internals/18`](docs/internals/18-vo5.md) |
| `ENDING.EXE`（結局文字、職員表）| **已解出** | [`docs/internals/16`](docs/internals/16-ending-exe.md) |
| `FACE.AMG`（.AMG 圖庫）| **已解出**（83 張逐像素對得上原版畫面）| [`docs/internals/09`](docs/internals/09-face-amg.md) |
| `IBM16.350` ＋ `IBMFLS16.350`（開頭動畫）| 影格切分與**壓縮語法**已解出（374 格逐格對上長度表、整檔吃光，`docs/re/36`）；格子對應到哪個位移未解 | [`docs/internals/19`](docs/internals/19-ibm16-350.md) |
| `COMANI.BMG`（戰鬥動畫）| 目錄已解出（253 段，位移逐筆接得起來）；每段的編碼未解 | [`docs/internals/20`](docs/internals/20-comani-bmg.md) |

## 介面覆蓋

| 項目 | 狀態 | 證據 |
|---|---|---|
| 資料段裡的熱區清單 | **25 份全部掃出**（含每個按鈕的中文標籤）| [`docs/internals/23`](docs/internals/23-hotspot-lists.md) |
| 清單 ↔ 畫面的對應 | 25 份裡只確定 2 份（主畫面、建造計劃）| 同上 |

## 執行期的資料表

| 表 | 位址 | 一筆 | 證據 |
|---|---|---|---|
| 星域 `STELLAR.DT*` | `ds:A79C` | 126 bytes | [`docs/internals/22`](docs/internals/22-stellar-dt.md) |
| 行星 `PLADAT.DAT` | `5313:0002` | 82 bytes | [`docs/re/30`](docs/re/30-region-economy-from-planets.md) |
| 提督 `ADMIRAL.DT*` | `ds:B77C` | 28 bytes | [`docs/re/33`](docs/re/33-planet-growth-and-tax.md) |
| **艦隊** | `ds:C074` | **122 bytes × 32 筆** | [`docs/re/40`](docs/re/40-fleet-table.md)；`+0x01` 陣營、`+0x04` FF ＝ 未使用、`+0x06` 所在星域 confirmed，其餘未解 |

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
| 曆法：一回合三天、一個月十回合、一年十二個月 | 通過 | [`docs/re/25`](docs/re/25-calendar.md)；`ds:8F18` 回合計數器對三份記憶體傾印＋三張畫面，另加 `-poke` 成第 118 回合 → 796/12/25（`docs/re/32`）|
| 月底交付「建造中」的船艦 | 數量算得出來（`docs/spec/004`）；交付之後放到哪裡未定 | [`docs/re/21`](docs/re/21-month-end-ship-delivery.md) |
| 治安 ＝ (200 − 未安定度) / 2 | 通過 | [`docs/re/24`](docs/re/24-security-formula.md)；六個帝國星域 ＋ 四個 `-poke` 局面（`docs/re/31`）|
| 治安數字的四段顏色（`<=40` 色 6、`<=80` 色 7、`<=120` 色 5、其餘色 2）| 通過 | [`docs/re/31`](docs/re/31-region-panel-layout.md)；40/41 那一對把門檻釘死 |
| 艦隊 `a：b` ＝ 記錄的 `+0x28`／`+0x2B`（帝國玩家）| 通過 | [`docs/re/31`](docs/re/31-region-panel-layout.md)；`-poke` 成 12：99 |
| 情報報告地圖的治安色（表 `ds:1780` ＋ 分段算式）| 通過 | [`docs/re/27`](docs/re/27-security-map-colours.md)；32 個星域逐一相符，除數 85 用 poke 實測定下來 |
| 星域經濟力 ＝ 該星域行星 `+0x1A` 相加（載入與每個月底各重算一次）| 通過 | [`docs/re/30`](docs/re/30-region-economy-from-planets.md)；星域 0 的四次累加逐值相符 |
| 行星每月的成長：階梯表查成長率（`ds:4A68`，32 列）| 通過 | [`docs/re/33`](docs/re/33-planet-growth-and-tax.md)；五個實測點逐一相符 |
| 月底逐行星成長 ＋ 逐行星扣稅（捨去在每顆行星上）| 通過 | [`docs/re/33`](docs/re/33-planet-growth-and-tax.md)；十個帝國星域結算後的經濟力與量到的 796/1 逐一相符 |
| 進國庫的稅 ＝ (成長後、扣稅前的星域經濟力) × 稅率 / 1000 | 通過 | [`docs/re/35`](docs/re/35-two-taxes.md)；十個帝國星域的增量逐一相符，合計 10741 |
| 未安定度的開局值怎麼算出來 | 骨架 ＋ 四支子函式都從程式讀出來了，**一個實測點都還沒有** | [`docs/re/33`](docs/re/33-planet-growth-and-tax.md)、[`docs/re/34`](docs/re/34-unrest-formula-skeleton.md) |
