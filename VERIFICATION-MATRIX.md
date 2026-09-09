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
| G1 | Ebiten 視窗版走玩家路徑 | `same-state` | 0 | `tools/gui_check.sh` | `cmd/gin3` 畫出來的色號陣列與 `r0`／`cb1`／`key1`／`cbAlly`／`keyAlly`／`main1e`／`load1e`／`vs3`／`vs4`／`vs5` 逐位元組相同。**「視窗開得起來」不算數**——比的是同一條玩家路徑跑完的畫面 |

## 畫面

| # | 畫面 | 等級 | 未核准差異 | 收據 | 備註 |
|---|---|---|---:|---|---|
| 1 | 主選單（含反白與游標）| `same-state` | 0 | [`docs/playtest/001-main-menu.md`](docs/playtest/001-main-menu.md) | **五個狀態**逐位元組一致：反白第 0–3 列各一、游標在選單外一個 |
| 2 | 開頭動畫 | — | — | — | 進行中：374 格已切分、編碼、調色盤與播放速度都解出來了（[`docs/internals/19`](docs/internals/19-ibm16-350.md)、[`docs/re/110`](docs/re/110-intro-playback-timing-and-palette.md)），已接進 `cmd/gin3` 的玩家路徑。整段的畫面對拍——見 4z |
| 4x | **開頭動畫的調色盤與播放速度** | `same-state`（調色盤）| 0 | 收據 `introvram`／`introtime2`／`introtime10`；[`docs/re/110`](docs/re/110-intro-playback-timing-and-palette.md) | `BU16IBM.RGB` 的值 ×4 ＝ 6 位元 DAC 值，48 個位元組與收據逐一相符（`TestIntroPaletteMatchesOracle`）。播放速度：PIT 除數 `0x0800`（582.6 Hz），一格 **73 個計時器中斷** ＝ 125.3 ms（約 8 fps），`-tick 20000` 下三次量到 1,459,999／1,460,000／1,459,999 道指令，換 tick 步數跟著縮放 |
| 3a | 正常玩家路徑（選單→防拷→本體）| `same-state` | 0 | [`docs/playtest/005-player-path.md`](docs/playtest/005-player-path.md) | 三個畫面都是前一步的輸入帶進去的，不是 direct-entry |
| 3 | 主選單（互動）| — | — | — | 進行中：反白、滑鼠→選項對應、游標都已對拍；點擊已能推進到下一畫面；還缺**切換過程**（反白從 A 列搬到 B 列會留殘留）|
| 4b | 防拷問答（輸入中的數字）| `same-state` | 0 | 收據 `typed138`；[`docs/re/56`](docs/re/56-typed-digits-are-full-width.md) | 打進去的數字是 `GRAPH.IMG` 的**全形**數字（`A2AF` 起），不是半形——先前搜不到是因為只試了半形的打包方式 |
| 4 | 防拷問答 | `same-state` | 0 | [`docs/playtest/003-copy-protection.md`](docs/playtest/003-copy-protection.md) | 全畫面 640×480 逐位元組一致（框、頭像、三行字、輸入框） |
| 4y | **開頭動畫的第 0 格** | `same-state` | 0 | 收據 `introseq`（`-watch 27616-2BD05`）；[`docs/re/108`](docs/re/108-intro-plane-buffers-located.md) | 收據錄到第 0 格 plane 0 的 12 筆寫入（di 9,406 起每次 +80，值 `C0 80 20 30 70 10 30 08 14 0C 02 06`），與 `asset.IntroDecoder` 產生的**逐位元組相同**，其餘 18,148 個位元組都是 0（`TestIntroFirstFrameMatchesOracle`）。畫面畫在頁面第 61 列也量過（`docs/spec/020`）|
| 4z | **開頭動畫的解碼器**（`IBM16.350` 374 格）| — | — | [`docs/re/71`](docs/re/71-intro-chunks-are-cut-at-frame-boundaries.md)、[`docs/re/72`](docs/re/72-intro-decoder-solved.md)、[`docs/internals/19`](docs/internals/19-ibm16-350.md) | 解碼器對得起來（逐格長度相符、整檔吃光、兩處場景黑幕累加歸零、解出來是船艦）。**先前「還不是畫面對拍」的兩個成因都解掉了**：CRTC 顯示起點 dosgolem 已補（[`docs/re/73`](docs/re/73-intro-planes-not-combined-on-dosgolem.md)），而「怎麼比都對不上」是因為原版在第一次搬圖之前就換過緩衝區、第 1–116 格從來沒被解過（[`docs/re/108`](docs/re/108-intro-plane-buffers-located.md)）。第 0 格現在逐位元組對上了，見 4y；整段動畫的畫面對拍還要先決定要不要照抄原版那個緩衝副作用（[`docs/spec/020`](docs/spec/020-intro-animation.md)）|
| 4w | **戰術選單「資料」（三個分頁）** | `same-state` | 0 | 收據 `s3data`／`s3data0`／`s3data_planet`／`s3data_enemy`；[`docs/re/111`](docs/re/111-battle-data-screen-layout.md)、[`docs/spec/021`](docs/spec/021-battle-data-screen.md) | 三張分頁的視窗 (16,60) 447×360 全部**逐位元組一致**（各 0／160,920，含原版自己畫的游標）：`TestBattleDataMatchesOracle`（艦隊）、`TestBattleDataTabsMatchOracle`（行星、敵艦隊）。文字照原版的字串引擎排——一格 16 像素、半形兩個併一格，立體邊與加粗整格一起做。已接進玩家路徑（`TestBattleDataOpensFromMenu`）。**內容也對拍過**：`TestBattleDataContentMatchesOracle` 用遊戲自己從原版資料檔算出來的列去比收據，0／160,920。士氣是艦艇記錄 `+0x06`（`-poke` 驗過，收據 `s3morale`）。行星分頁的內容也對上（名字 `+0x00`、經濟力 `+0x1A`、防衛力 `+0x14`、陣營色 `+0x11`）。補給欄的判準反組譯出來了（`0x140A5`：不是玩家陣營整段遮蔽，否則 M 看 `+0x3C`、F 看 `+0x40`、七格看 `+0x20 + 艦種 × 4`，一律「32 位元非不非零」）——**沒有允許差異** |
| 4v | **艦隊詳細資料**（「資料」點一列開出來的）| `same-state` | 0 | 收據 `s3detail`；[`docs/re/112`](docs/re/112-battle-fleet-detail.md)、[`docs/spec/022`](docs/spec/022-fleet-detail.md) | 版面量完：視窗與「資料」同一個、頭像 (355,102) 72×93、文字十四列列距 20、三欄（標籤 x 48／176／304，數值靠右到 160／288／416）。**上半段十二個數值的來源全部反組譯出來**（`0x14BA1`–`0x14F60`）：艦隊士氣值 `+0x3C`、三個速度 `+0x4D`／`+0x4E`／`+0x4F`、統率力在提督記錄 `+0x0B`、其餘八格是艦隊能力 `+0x3D`–`+0x47`；顏色是參數，色 5 ＝ 副官在那一項比提督高。下半段五行也對完（艦種列同時給名字索引與每艘幾件；耐久力 ＝ 兩個半軸的平均；三個速度是各分隊艦艇 `+0x1F`／`+0x21`／`+0x22` 取最小值）。**已接進玩家路徑**，版面與內容兩層對拍都是 0／160,920，沒有允許差異 |
| 4u | **「資料」的戰場顯示模式** | — | — | 收據 `s3detail_back`；[`docs/re/113`](docs/re/113-battle-data-display-mode.md) | 量到骨架：戰場整片色 3 ＋ 色 6 的虛點格線、單位是 16×16 的方塊（外圈 6、內圈 1、內部 12×12 畫艦種字母，旗艦印艦隊編號）、右側狀態欄換成選中分隊的八行資料。**12×12 的字形來源、怎麼切回正常顯示都未解**，remake 不畫 |
| 4t | **戰場的命令選單** | `same-state` | 0 | 收據 `s3cmdsq3`／`s3cmdfl3`／`s3cmdoff`；[`docs/re/117`](docs/re/117-battle-command-menus.md)、[`docs/spec/023`](docs/spec/023-battle-command-menus.md) | 觸發解出來了：點**自己陣營**的分隊（艦艇 `+0x00 ＝ ds:8F1D`）就開命令選單——旗艦開艦隊選單（移動／戰鬥／索敵），其餘按艦種 0–4／5／6／7 開四張分隊選單之一。可選旗標的規則、不可選的畫法（字色 8、游標壓上去也不畫反白帶）也量了。版面收斂成一條公式：框 ＝ (格x×16+1, 格y×20+1)、大小 ＝ (內寬×16+30)×(項目數×20+38)。三張選單的框逐位元組一致（0／4,836、0／6,076、0／4,836，含原版自己畫的游標），已接進玩家路徑（`TestCommandMenuOpensOnOwnUnit`）。**右側分隊面板與每一項的行為還沒做**（`docs/spec/023`「還沒做的」）|
| 4r | **戰場的「移動」命令（分隊層）** | `same-state` | 0 | 收據 `s3cmdmove`、`s3cmdrange`；[`docs/re/119`](docs/re/119-battle-move-order.md)、[`docs/re/121`](docs/re/121-move-target-checks.md)、[`docs/spec/024`](docs/spec/024-battle-move-order.md) | 下命令之後寫進去的目的地與「點到的畫面座標 ＋ 視野位移」一個像素不差（收據 `s3cmdmove`：點畫面 (200,200)、視野 (1040,540) → 世界 (1240,740)），沒有取整也沒有夾到範圍內。邊界（世界 x 0–2031、y 0–2023）與地形判斷（`1588:04EA` ＝ `rules.BattleTerrain.Passable`）照原版。**原版選點時畫的兩個範圍圖形只是提示**——點擊的檢查鏈裡沒有距離判斷。不合法時的訊息框已經照原版畫、照原版關（見 4h），已知差異剩兩項（不畫範圍圖形、旗艦那一項要先逆向抵達朝向的選擇器） |
| 4g | **行星命令的四項行為** | — | — | [`docs/re/134`](docs/re/134-planet-commands.md)、[`docs/internals/29`](docs/internals/29-planets-and-economy.md) | 四支讀掉三支半。「補給」「補充」與艦艇版同骨架，只換中心（`+0x46`/`+0x48`）、陣營（`+0x11`）與半徑（`PlanetQueryRadius`）；存量在行星身上——`+0x3C`/`+0x40` 飛彈與艦載機、`+0x20` 起七格每個艦種一份的補充存量（解掉 internals/29 原本標「未解」的那一段）。補充補到**建制滿編 `+0x0A`**（不是修理的 `+0x0C`）。「要塞砲」只指定目標：左鍵 `+0x50` ← 目標、`+0x45` ← 3，右鍵 `FFFF`/0；射程是一個**橢圓**（`3870:03A3` → `39D7:035E`）：長半軸 ＝ 防衛力 ÷ 200 ＋ 80、短半軸 ＝ 長半軸 ÷ 4（伊謝爾倫 12800 → 144／36），顏色 4 XOR 底色，已經畫出來。三項接上玩家路徑，七支測試。**「情報」的 22×10 視窗沒做**（`1373:1CBC` 未讀）、要塞砲的開火結算沒找到、射程圈與候選高亮沒畫 |
| 4f | **行星的命令選單** | — | — | [`docs/re/133`](docs/re/133-planet-command-menu.md)、[`docs/internals/44`](docs/internals/44-battle-command-menus.md) | 開哪一張與可選旗標全解出來：行星 `+0x12`（要塞旗標，81 顆裡 8 顆）決定三項 `ds:42EA`（內寬 2）還是四項 `ds:4318`（內寬 3，多「要塞砲」）；不是自己的行星只剩「情報」；自己的、有要塞、`+0x45` 不是 0 也不是 3 → 關掉「要塞砲」。四項那張的前三個名稱在**字串裡**就補了半形空白。remake 已經接上（選單內寬改成逐張指定，既有三份逐位元組收據重跑仍全同）。**四項的行為都還沒逆向**（`1373:1C71`／`1930:0DA2`／`1AAB:0A33`／`2264:0049`），原版的行星命中判定（`341B:039D`）也沒讀，所以這一列不能標對拍 |
| 4e | **戰場的「索敵」「移動（艦隊）」「戰鬥」** | — | — | [`docs/re/132`](docs/re/132-scout-order.md)、[`docs/re/120`](docs/re/120-battle-move-range.md)–[`122`](docs/re/122-arrival-facing-and-fleet-step.md)、[`docs/spec/027`](docs/spec/027-fleet-move-and-battle-order.md) | 艦隊選單三項全接上玩家路徑。「索敵」＝ 清 `+0x3A`／`+0x3B`、把 `+0x38` 那一筆偵察隊擺到旗艦位置（`+0x1E` ← 3、`+0x19` ← 0x80）、再接分隊「移動」的選點。「移動」與「戰鬥」走同一支 `2583:0003`：點目的地 → 繞著它轉游標選抵達朝向 → 寫 `+0x60`/`+0x62`（備份）、`+0x64`/`+0x66`（目的地）、`+0x4C`（朝向），「戰鬥」另外把 `+0x48` 低四位寫 1；右鍵任一段整個作廢。**畫面沒有做**——兩個範圍圖形（內圈公式已知、外形 `165F:004D` 未讀）、選朝向的四臂十字、偵察隊的航跡線（`1727:0324` 未讀）都要收據；艦隊層的移動引擎也還沒接進回合流程 |
| 4d | **戰場的「補給」「修理」命令** | — | — | [`docs/re/131`](docs/re/131-supply-and-repair.md)、[`docs/spec/026`](docs/spec/026-battle-supply-repair.md)；收據 `s3form` 的艦艇池 | **數字全解出來了**：半徑 ＝ 艦種 5／6 那一列 `+0x06`（最高速度）× 6.4，而原版同一個 6.4 算了兩次（收候選 `<<21÷5` ＝ 192.0、確認 `×0x66666` ＝ 191.99982，差 12／65536，兩個都照抄）；候選是同陣營、沒出局、在半徑內，補給排除艦種 5／6／7、修理只排除 6／7；補給補到「艦種列 `+0x10`／`+0x14` × **現有**艘數」，修理把 `+0x08` 補到 `+0x0C`。拿收據那 148 筆真記錄逐筆對過公式，玩家路徑也接上了。**畫面**：候選本來就沒有高亮（`docs/re/140` 讀完整支模態迴圈的否定結論），範圍圈與右側面板都畫了——面板與「結束」鈕與收據 `s3sup` 的 8×8 格**逐位元組相同**（`docs/re/143`）。沒有候選時的兩句訊息已接上玩家路徑（見 4h）。**選中標記不存在**——面板跟著滑鼠走（`1930:0769` 命中換人才重印），有目標時多印三列，與收據 `s3sup2` 的八列（含「結束」鈕）逐位元組相同；「修理」那一版（`1AAB:08FB`，收據 `s3rep`）也對上了。`s3sup2` 順帶驗到搬的量：補給艦飛彈 8850 → 5900、戰鬥機 23600 → 11800，正好等於目標補滿所需 |
| 4h | **訊息框**（`3826:0005`）| `same-state` | 0 | 收據 `s3msg`；[`docs/re/137`](docs/re/137-message-box.md)、[`docs/re/138`](docs/re/138-window-frame-tiles.md)、[`docs/internals/47`](docs/internals/47-window-frame-tiles.md) | 框 ＋ 那行字 ＋「確認」按鈕，整個 22 × 6 格 **42,240 個像素逐位元組相同**。框是 `FRAME.BLK` 的圖塊（起始碼 1），色號 0 的像素填視窗顏色的高四位；字在內容格左上角往下 2 像素、字色 7、陰影 3 往上一列；按鈕是內容格的旗標（`0x08` 上下邊、`0x10`／`0x20` 左右端，色號只有 7 與 3），而且**裝飾最後畫**（先畫會差兩個像素）。關閉走 `3826:023B`：模態，只認按鈕上的左鍵，關掉之後命令作廢。三個呼叫端（分隊「移動」、艦隊「移動」／「戰鬥」、「補給」／「修理」沒有候選）都接上玩家路徑 |
| 4i | **行星選單的「情報」視窗**（`1373:1CBC`）| `same-state` | 0 | 收據 `s3plinfo`；[`docs/re/144`](docs/re/144-planet-info-window.md) | 自己一個 24 × 12 格的視窗（格 (3,4)，內容 22 × 10）。整個 **92,160 個像素逐位元組相同**，含框、內容與壓在左上角的滑鼠游標。分行靠格式字串自己帶的 `\n`（與補給面板「剛好佔滿一列」是兩種習慣）。收據的做法：那顆行星原本被旗艦壓在同一格點不到，把座標 poke 到空位再點 |
| 4j | **要塞砲的射程橢圓與瞄準**（`2264:0049`／`39D7:035E`）| `same-state` | 0 | 收據 `s3fort`／`s3fort2`；[`docs/re/145`](docs/re/145-fortress-gun-aim-and-fire.md) | 兩張只差游標位置，橢圓是 XOR 上去的，所以逐像素比出來的差異必須剛好等於兩條曲線的**對稱差**——**786 個像素逐一相同**（扣掉兩個游標框，那兩塊本來就移動了）。驗到的是：焦點在行星佔位框中心、中心 ＝ 焦點 ＋ 沿瞄準方位走 ⌊√(a²−b²)⌋、長半軸吃防衛力**上限**、傾角 ＝ `Atan256(游標 − 焦點)`、曲線是中點圓演算法壓扁再旋轉 |
| 4k | **要塞砲開火的結算**（`181B:0385` → `2264:0248`）| `same-state` | 0 | 收據 `s3gunfire2`；[`docs/re/148`](docs/re/148-fortress-gun-fire-receipt.md)、規格 [`docs/spec/028`](docs/spec/028-fortress-gun-fire.md) | 同一次執行在要塞砲那一步的前後各倒一份 1 MB。**候選集合 33 支的射程判定逐一相同**（17 內、16 外）、**17 支扣掉的艘數逐一相同**（連兩個 `rand()` 都對上）、艦艇 `+0x0C`（掉 實扣²÷滿編）與 `+0x0E`／`+0x12`（按剩餘比例縮減）逐一相同、兩個計分器各 ＋989 ＝ 逐支 ⌊實扣÷20⌋ 的和 |
| 4c | **戰場的「編隊」命令** | `same-state` | 0 | 收據 `s3form`；[`docs/re/130`](docs/re/130-formation-chooser.md)、[`docs/internals/46`](docs/internals/46-battle-formation-chooser.md)、[`docs/spec/025`](docs/spec/025-battle-formation-order.md) | 分隊選單第二項開的選格子畫面，整支 `2BE1:0328` 讀完。收據同時倒出 `ds:D072` 那 41 筆：**41 格的世界座標 41／41 完全相同**（槽位偏移依艦隊朝向旋轉再加位置），有人的六格與「發令者自己那格放回可選」的規則一致。標記是 `SPRITE.SPR` 第 11 塊、色號 0 透空，35 個空格子有 **30 個逐位元組相同**（另外五個底下坐著分隊或被游標壓到，所以也量到分隊畫在標記之後）。選下去寫艦艇 `+0x05`、艦隊分組表 `[3]`、`+0x1E ← 0x7F` 並立刻算出新目的地。**「透空」與「色號 0 換成 3」這份收據分不開**（沒有空格子壓在格線上）|
| 4s | **戰場的分隊面板** | `same-state` | 0 | 收據 `s3cmdfl3`／`s3cmdsq3`；[`docs/re/118`](docs/re/118-battle-unit-panel-layout.md)、[`docs/internals/45`](docs/internals/45-battle-unit-panel.md) | 點自己陣營的分隊時右側換上來的八列。版面走原版格式字串算出來的半形格位置（格原點 (480,242)、欄距 16、列距 20）；值的來源全部反組譯過（士氣值旗艦讀艦隊 `+0x3C`、其餘讀艦艇 `+0x06`，低於 50 換色 2；行動名與顏色查 `ds:691C`／`ds:6930`；光線砲 ＝ 艘數 × 艦種列 `+0x0B`，飛彈與戰鬥機印艦艇 `+0x0E`／`+0x12` 的 32 位元量，顏色是拿現有量比「滿量」與「門檻」兩個欄位）。文字區 (480,242) 128×160 兩張各 **0／20,480**，沒有允許差異。按艦種換的那一行（運補／修護／索敵）規則已讀出，缺收據未做 |
| 5 | 遊戲主畫面（796 年 1 月）| `same-state` | 0 | [`docs/playtest/004-main-game.md`](docs/playtest/004-main-game.md) | 全畫面逐位元組一致；地圖、頭像、文字全部從原版檔案重建 |
| 5a | **同盟開局**（主選單第 1 列 → 防拷 → 月初會計報告 → 訂造報告 → 建造計劃 → 戰略地圖）| `same-state` | 0 | 收據 `cbAlly`／`keyAlly`／`shipAlly`／`planAlly`／`main1eAlly`；[`docs/spec/011`](docs/spec/011-opening-report.md)、[`docs/re/61`](docs/re/61-work-staff-is-faction-pairs.md)、[`docs/re/62`](docs/re/62-unit-prices-differ-by-faction.md) | 五張整張 640×480 逐位元組一致。署名、頭像、金額、右側首都摘要、星圖的首都特別色與**七列的船艦單價**都跟著陣營走 |
| 5a2 | **同盟的命令階段十二張**（含三個選單）| `same-state` | 0 | 收據 `*Ally` 十五份；[`docs/re/62`](docs/re/62-unit-prices-differ-by-faction.md)、[`docs/re/63`](docs/re/63-security-divisor-is-the-intel-staff.md)、[`docs/re/64`](docs/re/64-command-phase-follows-the-player.md) | 每一張整張 640×480 逐位元組一致，走的是同一條玩家路徑只換主選單那一列。署名、頭像、提督清單、星系清單順序、兩張地圖的著色、左下角摘要與貨幣單位都跟著陣營走 |
| 5d | **提督詳細資料與經歷**（出擊命令 → 待命中提督清單 → 點一列 → 經歷）| `same-state` | 0 | 收據 `adm1e`／`adm2e`／`adm3e`／`adm15e`／`keireki1e`；[`docs/spec/012`](docs/spec/012-admiral-detail.md)、[`docs/re/65`](docs/re/65-admiral-detail-screen.md) | 五張整張逐位元組一致。姓名、軍階、頭像、現在位置、旗艦、等級與十個能力值都跟著提督走；`ADMIRAL.DT*` 十四個能力欄位的名字就是這張畫面給的。「經歷」印的是 `KEIREKI.DAT` 的人物傳 |
| 5h | **艦隊編成**（提督詳細資料 → 出擊）| `same-state` | 0 | 收據 `fleet1e`／`fleet2e`；[`docs/spec/013`](docs/spec/013-fleet-organisation.md)、[`docs/re/66`](docs/re/66-fleet-organisation-screen.md) | 兩張整張 640×480 逐位元組一致。艦隊能力九格是提督與副官**逐項取大**，副官贏的印成色 5；副官 ＝ `ADMIRAL.DT*` 的 `+0x07` |
| 5i | **艦隊編號清單**與**副官選單**（艦隊編成的兩顆按鈕）| `same-state` | 0 | 收據 `fnum`／`fnum3`／`fadj`／`fadj3` ＋ 兩個反例 `fnum2`／`fadj2`；[`docs/re/67`](docs/re/67-fleet-org-buttons.md)、[`docs/re/68`](docs/re/68-hotspot-origin-rule.md) | 四張整張逐位元組一致；反白條跟著滑鼠走，由 `hover` 版面算。另外兩張是**反例**：點在畫出來的框內、熱區外一個像素，畫面與 `fleet1e` 完全相同 |
| 5j | **艦隊組成**（艦隊編成 → 確認）| `same-state` | 0 | 收據 `fok`；[`docs/spec/015`](docs/spec/015-fleet-composition.md)、[`docs/re/69`](docs/re/69-region-ship-strength.md) | 整張 640×480 逐位元組一致。9×9 格盤鋪的是 `SPRITE.SPR` 圖塊；「待命部隊」＝ 星域記錄 `+0x30` 的艦艇艘數 ÷ 一單位幾艘 |
| 5k | **陣形清單**（艦隊組成 → 陣形）| `same-state` | 0 | 收據 `form1`／`form2`／`form3`；[`docs/re/70`](docs/re/70-formation-list-and-decide.md) | 三張整張逐位元組一致。八列圓錐／紡錘／半月／Ｕ字／Ｍ字／Ｔ字／分散／密集，反白條跟著滑鼠走 |
| 5l | **艦隊建成**（艦隊組成 → 決定）| `same-state` | 0 | 收據 `dec1`、`decwatch`；[`docs/re/70`](docs/re/70-formation-list-and-decide.md)、[`docs/internals/33`](docs/internals/33-fleet-table.md) | 整張逐位元組一致。右下角清單多出「第１艦隊　萊因哈特　100」；左下角的首都摘要留在畫面上不被擦掉 |
| 5m | **對戰的交棒視窗**（主選單第 2 列 → 防拷 → 帝國／同盟輪流）| `same-state` | 0 | 收據 `vs3`／`vs4`／`vs5`／`vs6` ＋ 反例 `vs3` 的框內熱區外那一點；[`docs/spec/016`](docs/spec/016-versus-mode.md)、[`docs/re/76`](docs/re/76-versus-turn-order.md) | 四張整張 640×480 逐位元組一致，走的是完整玩家路徑（確認 → 月初報告 → 關報告 → 系統指令／實行 → 換邊）。**兩邊各按一次「實行」才推進三天**；交棒那一刻的底板不上玩家色，按下確認之後又塗回來 |
| 5n | **戰場畫面**（面板 ＋ 地形 ＋ 格線）| `same-state` | 0 | 收據 `s3ai6`；[`docs/re/82`](docs/re/82-battle-field-has-no-units.md)、[`docs/internals/34`](docs/internals/34-battle-screen.md)、[`docs/internals/36`](docs/internals/36-map-cmp.md) | **涵蓋整張畫面扣掉右上角的小地圖**（面板 ＋ 448×360 的戰場 ＋ 格線 ＋ 右側狀態欄的框與九行文字），該範圍逐位元組一致。小地圖 128×128 **逐位元組相同**（`TestMinimapMatchesOracle`）——原本剩的 18 個像素是分隊與行星（`docs/re/89`），五層（地形、格線、行星、分隊、視野框）都補上之後一個都不差，主對拍因此不再排除小地圖。整張畫面現在只排除兩處：收據上 dosgolem 畫的滑鼠游標，以及小地圖下緣 y=203 那一列——那一列是 `2376:10EA`／`110B`／`1132` 這組 16×20 的貼圖常式畫的，三趟的列數是 19／20／20，多出來的那一列只有部分 plane 被寫到；內容對應線性 `0x70EE6` 的 1bpp 資料（`docs/re/103`），它在畫哪一個元件還沒對出來。背景是 `ANI_PNL0.A98` 畫在 y=40，地形是 `MAP26.CMP` 配 `BLOCK_2.BLK` 的 28×18 格，格線在世界座標每 240 像素一條。戰鬥停在輸入階段時雙方不動，這張不含單位；單位另外驗（收據 `s3fight`，見 5o）。狀態欄的框是 `FRAME.BLK` 的圖塊：`0C0A:6B6D` 用 `3603:0006(色 0x97, 格 (29,9), 內容 8×8 格, 旗標 1)` 建的視窗，實際佔 10×10 格（`docs/re/142`）。原本量出來的八層與 26 個角落修正像素都刪了——那些只是圖塊裡本來就長那樣的像素，三份收據各 11,520 個像素逐位元組相同 |
| 5o | **戰場上的單位** | `same-state` | 0 | 收據 `s3fight`；[`docs/re/85`](docs/re/85-units-are-sprite-tiles.md) | 分隊畫成 `SPRITE.SPR` 的 16×16 圖塊，左上角對齊世界座標。視野內比得出圖塊的八個分隊，每一塊 16×16 逐位元組相同——落點錯一個像素整塊就對不上，所以座標換算也被驗到。**畫哪一塊已經解完**：塊號存在艦艇記錄 `+0x2F`，由 `25DB:03E8` 依陣營基底（帝國 20、同盟 50）與顯示層級 `+0x17` 算出來；兩份收據十六個樣本與畫面全部相同（`rules.BattleSprite`）。`+0x17` 是索敵殘留、每輪衰減，判定「看到」的條件還沒解，所以 remake 的 `Deploy` 一律給 2 |
| 5p | **戰術選單** | `same-state` | 0 | 收據 `s3menu`（`menu-open.bin`）；[`docs/re/101`](docs/re/101-battle-tactic-menu-layout.md)、[`docs/re/98`](docs/re/98-battle-tactic-menu.md) | 戰鬥每一回合結束時開在游標處的七項選單。外框 (225,221) 126×178（相對游標 (−15,−19)），反白帶 (240,240) 96×20 色 2，字身比帶頂低 2 列。框與主選單同一組（`ui.MenuFrame`）——連四個角的修正像素都成立，這是那組修正的第二個樣本。文字是遊戲既有的立體邊規則：本體色 7、立體邊色 3（字模每列左移一位再上移一列）。**選單矩形 22,428 個像素全部相同**，含原版自己畫在第一項上的滑鼠游標（`TestBattleMenuMatchesOracle`）|
| 5q | **戰場移動** | `same-state` | 0 | 收據 `s3move`／`s3budget`／`s3tgt`；[`docs/re/102`](docs/re/102-battle-movement.md) | 一回合的行程 ＝ 艦隊 `+0x4E × 8` 像素，沿方位每次走 8，行程餘額（`+0x56`／`+0x58`，16.16）扣掉實際距離，剩不到 8 就停；每一步先問地形（`2583:040C` → `1588:04EA`）。十八支艦隊的 `+0x4E` 與行程全部對得上。remake 用收據的起點與目標跑兩個回合，落點 (1420,1637)、(1306,1434) **與原版逐像素相同**（`TestBattleMoveMatchesOracle`）。`+0x4E` 自己的輸入（艦種列 `+6`）還沒對出來。**額度的每一筆也逐位元對上了**（`docs/re/129`）：一次「實行」先轉一步再前進一步（轉向只在 |角差| > 回轉速度 那條扣 8.0），前進扣的是 `3310:0006` 算出來的實際長度，而且**先扣再檢查落點**——被擋掉的一步照樣扣。收據 `s3fltseq` 的 160.0 −8.0 −7.6157 −18×7.8102 ＝ 3.801086 與原版相同，鎖在 `rules.AdvanceFleet` 的測試裡。艦隊層的移動引擎本身還沒接進玩家路徑（remake 目前是逐分隊各自走） |
| 5r | **戰後不改星域歸屬** | `same-state` | 0 | 收據 `s3own`／`s3win`；[`docs/re/105`](docs/re/105-battle-does-not-capture-region.md) | 星域記錄 `+0x12`（`ds:A79C` ＋ 26 × 126 ＋ 0x12 ＝ 線性 `5FFAA`）在整場戰鬥中一次都沒有被寫，**連用 `-poke` 把同盟四十個分隊的存活艘數寫成 0、讓戰鬥提早結束也一樣**（`+0x13` 從 12 退回 8 證明戰鬥真的結束了）。remake 的 `EndBattle` 不動星域歸屬因此是照原版。星域什麼時候易主還沒解 |
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
| `FRAME.BLK` | **已解出**（33 張 16×20 四平面圖塊，版面同 `SPRITE.SPR`；四組視窗外框各八張，色號 0 ＝ 視窗底色）| [`docs/internals/47`](docs/internals/47-window-frame-tiles.md) |
| `BLOCK_*.BLK` | **已解出**（各 128 張 16×20，與 `FRAME.BLK` 同一種版面；`BLOCK_2` 已用在戰場地形，`BLOCK_0` 的圖塊 2／3／4 在收據 `s3menu` 上逐位元組命中面板分隔線）| [`docs/internals/36`](docs/internals/36-map-cmp.md)、[`docs/internals/47`](docs/internals/47-window-frame-tiles.md) |
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
