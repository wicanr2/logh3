# 372 — 移動單步與避障

日期：2026-10-02；[#79](https://github.com/wicanr2/logh3/issues/79)；
[第三十輪](../goal/2026-10-02-battle-movement-round-30.md)。
狀態：原版102案取證完成；[規格081](../spec/081-battle-fleet-step.md)限定艦隊單步修正。
[草稿080](../spec/080-battle-movement-phase.md)的完整交錯移動仍未完成。
先前已驗額度及邊界由[研究371](371-battle-movement-phase.md)管理，不重新開啟。

輸入為原版`GIN3.COM`，SHA-256
`770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918`；
第二階段入口1MiB RAM SHA-256
`9c4a4c28a1f53056962f613b1b643e5a5ba199d9231df291918d9499111be38d`。
dosgolem固定`f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c`及乾淨工作樹；
原版CS:IP、工具IDA位址、實體線性位址分開記錄。
## 證據身份與重播

原版常式診斷以既有完整入口RAM建立新堆疊，DS=ES=54B3、tick=0；每案Restore。
沒有替身或改原版指令。VGA／DOS檔案服務是新建立的，故只驗資料、返回值與呼叫鏈，
不把這些常式輸出當正式UI畫面。全部注入、完整前後記錄與呼叫參數保存在
[分級JSON](../internals/data/battle-step.json)，可由`tools/scripts/battle_step_receipts.py`重生。

| 產物 | SHA-256 |
|---|---|
| 73案原計畫 | `1f2b79495ac5fddc3582f8abb6af3b0757646212916205507534526051332d26` |
| 73案原版a／b | `f4a75544c6381f1c0226629c1f0766850785bbac26a840f0c43bedfe14b15e08` |
| 102案計畫 | `d14d311df1efd314bc671c922a73b56d704bbdb1f37da1b1c9f0ab2666bb3f1f` |
| 102案原版a／b | `4cbe15151b3abcdc0da7d7a3b115ac22c600c44e31314e3f15efba9247d3395c` |
| 分級JSON | `054f1aef81a139da6ed85d8979fa72e647574c3984c6a73a3649a48c4eeba00c` |

102案包含原73案，前73份完整結果相同；舊計畫重生也逐位元組相同。
新增29案為通行起點的兩案牆面、兩案北向阻擋、兩組七案完整Action、
兩案抵達回轉旗標、九案碰撞邊界。兩組Action有七筆相同輸入重複核對，
因此102是執行案例數，不是102種獨立狀態。所有案例seed維持`78563412`。

重播入口（主機只呼叫既有Docker包裝器；新輸出名稱不得與已有收據相同）：

```sh
bash tools/py.sh /work/tools/scripts/battle_step_plan.py --extended --output battle79-step-plan-extended.json
bash tools/gui.sh xrun cp tools/scripts/battle_step_probe.go workplace/battle_step_probe.go
DOSGOLEM_ORIG=/home/anr2/cht/logh3/org_game/extracted DOSGOLEM_EXTRA_MOUNT=/home/anr2/cht/logh3/workplace:/out bash /home/anr2/cht/dosgolem-logh3/tools/go.sh run /out/battle_step_probe.go -plan /out/battle79-step-plan-extended.json -output /out/battle79-step-extended-a.json
DOSGOLEM_ORIG=/home/anr2/cht/logh3/org_game/extracted DOSGOLEM_EXTRA_MOUNT=/home/anr2/cht/logh3/workplace:/out bash /home/anr2/cht/dosgolem-logh3/tools/go.sh run /out/battle_step_probe.go -plan /out/battle79-step-plan-extended.json -output /out/battle79-step-extended-b.json
bash tools/py.sh /work/tools/scripts/battle_step_receipts.py
```

正式run前核對專用dosgolem Git HEAD與clean；不改執行器、不靠換seed取得局面。
原始完整RAM由[研究371](371-battle-movement-phase.md)的階段包裝器重生。

## 原始指令審查

IDA Pro 9.4、image `ida-pro-9.4-idapython:locked-v1`；每份切段前置256零bytes，
原位移0對應IDA`0x10200`。CS:IP與runtime線性位址另列，函式名稱只作導覽。
原始匯出與9函式／15指令分級索引自動合併，未審查指令仍帶unknown警示。

| 原CS | 切段SHA-256 | 分級匯出SHA-256 |
|---|---|---|
| 1C03 | `9f3141e00160ea1970c5fc803d9ecfc9ffb1e2b6ea3f1630d3c500a1e17ac818` | `e45f293cfef92ec0740a73761b859e6d4d9c8f737cd2ea26fd74074cfc10cf02` |
| 2583 | `037d7f0703caa9b4c8e4d15a28ae2f4e6893e160d8f3d223f161b353e5d4bf14` | `d8ce4ca25aa22eab9580adbee74a53b4e552e50bca5bb5f9d53236ef14d29caf` |
| 17A0 | `d60892aefb89793d6d957f44e8f320c00cc649388111e61bb731958b51cec0e4` | `90ed44eb4c002bad5902fd88bc215847895f2fcee23312687be369df92624d63` |

**proven**：`1C03:006A`先更新一般分隊追隊bit，額度至少8.0且完整Action為00／01／10／11
才進步進；兩種移動清+49 bit0。正常未抵達轉向保存AL並設bit，耗盡最後一步仍返回1。
抵達轉向不設bit，且受外層額度閘門限制。候選實際距離在通行檢查前扣掉，斜向可略負。
關鍵指令：`00B3 26 80 24 fd`、`0105 7d 09`、`020F 88 46 03`、`021E 26 08 47 49`。

**proven**：`2583:040C`先問地形，阻擋回FFFF；再掃敵分隊，兩軸差均<16才重疊。
同陣營、種類7及Action bit7略過；池索引加1，無碰撞轉成0。
來源：`041C`／`0433`呼叫鏈、`17A0:0718..072D`篩選、`0750 40`與`07A6/07AA`邊界。

**proven**：`2583:0451`依候選方位±16交替搜尋，包含重複0／128候選；全部受阻恢復舊座標。
艦隊在`1C03:0350`／`0513`傳舊座標及同兩格指標，故初始向量為0。
不能假設零向量角度是0：實測及既有反正切表均為64，先試下方。
`fleet-wall-from-open`從(992,1000)改成(992,1008)；北向牆面則依序試到(1008,1000)。
遇敵佔位的艦隊不避障，位置不變但仍扣額度。38案完整艦隊記錄及逐次通行候選已由正式rule互證。

## 分隊與索敵隊的有限原版事實

以下`proven`只指本批常式，不聲稱已接正式分隊引擎：

- 旗艦／未啟用索敵隊現在座標已等於艦隊時返回0，保留原集結點；不同才同步目前與集結點，返回1。
- 步長8／12／16／24均最後固定扣8.0；目標距離3、4時直接到點，仍扣8.0。
  斜向與地圖夾界結果完整保留；不是以Speed整段扣實際距離。
- 跟隊7F依回轉／追隊bit取8／12／16／24；Action3不受這兩bit影響，固定8。
  自主0第一次到點仍保持0，下次呼叫才改3並清額度。
- 普通分隊受阻會以欲前進的候選方位繞行。索敵隊遇地形FFFF改3、清餘額並停；
  遇敵而Action不是3時艦隊+3B改4、清餘額且停，Action保留；Action3可繞敵。
- 已到集結點的索敵隊回收後艦隊+3A=FF／+3B=0、Action=FF，返回0；未啟用索敵隊仍優先同步位置。
  圖示／小地圖呼叫被記錄，但冷VGA診斷不是玩家畫面完成證據。

## 正式修正與反例

[驗證043](../playtest/043-battle-fleet-step.md)分開原版資料同狀態、正常玩家路徑與GUI入口抽樣。
舊規則16案有3案失敗：轉向耗盡額度誤返回0，以及兩種移動在零額度仍免費轉向。
修正後38案完整艦隊／追隊bit／AL／通行序列與9案碰撞通過。

正常移動另抓出保存鏈缺口：單步寫回的+5C/+5E被`Slot.SyncFleets`的戰略Fleet舊座標覆蓋。
原失敗及欄位差保留；規格081補清資料流後同時同步戰略座標，原命令重跑存讀相同。
PNG首次將16色6-bit DAC當RGB8，造成診斷圖過暗；依正式`Palette.RGB8`公式重轉，
原色號／DAC未改，這是驗證轉檔問題。完整八步／第二階段仍未完成，#79保持open。
