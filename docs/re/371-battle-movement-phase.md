# 371 — 第②步移動階段

日期：2026-10-02；工作項[#79](https://github.com/wicanr2/logh3/issues/79)，
範圍：[第二十九輪](../goal/2026-10-02-battle-movement-round-29.md)。
狀態：**取證中，尚未授權正式移動改動**；第一階段完成範圍沿用規格079／對拍042。
正式移動草稿：[規格080](../spec/080-battle-movement-phase.md)。
事實摘要：[艦艇池移動段落](../internals/64-ship-pool.md#移動階段額度與停止條件)、
[分級索引與原版局面JSON](../internals/data/battle-movement-phase.json)。

既有骨架：[研究339](339-battle-turn-eight-steps.md) §2.2；
艦隊步進：[研究294](294-short-move-strafe-and-arrival-facing.md)、
分隊隊形：[研究126](126-battle-formation-slots.md)。
原版`181B:00FA`先叫`1C03:0006`；每輪逐艦隊`1C03:006A`、一般分隊`1C61:0153`，
再索敵分遣隊；任何回非0就重畫並再跑整份清單。已知骨架不等於每個欄位已驗。

原版身份`GIN3.COM` SHA-256 `770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918`。
dosgolem專用副本固定`f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c`、乾淨；Go image為`golang:1.24-bookworm`。
起點`eng-t5.state` SHA-256 `c6ce5af3a8d52acc415ac1029a29a0905e9c95e02ac0e0ba547e71300fd947c1`，
固定序列沿研究367，第6回合入口前固定種子`0x12345678`。不選種子碰運氣。

原版重播入口：[battle_movement_oracle.sh](../../tools/battle_movement_oracle.sh)。
只使用既有固定檢查點、等價玩家按鍵與明示seed；原始memory／完整狀態／PNG留本機。
收據將分列原版CS:IP、dosgolem實體線性位址、工具版本、步數及完整旗標；
沒有重生及雙跑核對前不宣稱第二階段對拍完成。

## 本輪重生：原版三個邊界均可重現

`bash tools/battle_movement_oracle.sh <entry|init|end> battle79-move-<模式>-<a|b>`。
六份收據均由dosgolem自行產生，沒有替換或搬用其他模擬器畫面。
固定輸入：`-load-state /src/workplace/fcmd/eng-t5.state`，三次點擊
`1140000000:240:200,1195000000:240:200,1210000000:240:200`，
按住3000000、tick20000；`lin:5E00C@1211893242=78 56 34 12`。
完整旗標保存在JSON；兩次只有輸出目錄名稱不同。

| 邊界 | 執行前步數／原版CS:IP | 1 MiB記憶體SHA-256 |
|---|---|---|
| 第二階段入口 | 1213409021／181B:00FA | `9c4a4c28a1f53056962f613b1b643e5a5ba199d9231df291918d9499111be38d` |
| 初始化後首次艦隊步進 | 1213433817／1C03:006A | `939f4b1fef6e840a313cad6feeb6996a02d0571772e21b0dce69f2a849c25813` |
| 進入第三階段 | 1226638239／181B:0262 | `4716f2a83e1807d83cc521feb2846d5a495a4e7767fcddfa2334f355c51e16b3` |

每個模式雙跑的完整memory-in／memory-out、307200 bytes畫面、768 bytes色盤及PNG相同。
save-state在標示步數保存；最終report是下一條指令，故分別停在00FB／006B／0263。
兩次序列完整RAM種子皆保持`78563412`，不能把這個局部無RNG階段外推整回合。
7支艦隊清單順序為12、16、17、28、29、30、31；48一般分隊加7索敵隊。
各筆初始化、結束raw records與池位址在JSON，未遮罩任何欄位。
此處的「相同」只比較原版兩次重播，**尚無remake同狀態比較**。

## IDA 原始指令審查

兩個切段均從上表入口RAM產生：256 bytes `.COM`前綴加段起點0x1400 bytes。
IDA Pro9.4，原位移0對應IDA `0x10200`；每筆同列保留原CS:IP、runtime linear、bytes、
原運算元、語意、等級與來源。輸入原始雜湊如下：

- `re79-movement-1C03.com`：`9f3141e00160ea1970c5fc803d9ecfc9ffb1e2b6ea3f1630d3c500a1e17ac818`。
- `re79-movement-1C61.com`：`af8dbadc328c12d37b3fdde79b2929b0d5199fd446c7c7360797a35e3552749f`。

入口以push bp及控制流到retf審查，非靠自動函式名判斷。
`1C61:01D8`只有0153內部Action分派區塊，**不是獨立函式入口**；
首次導覽匯出保留unknown，不以這個位址查獨立callers。
語意由受版控JSON自動合併，沒有從未分級散文匯入。

已審查匯出：`workplace/ida/re79-movement-1C03-reviewed.json`
SHA-256 `11a8a4fe8d04408e88107b2387764fa485de180aa228a371e0aa1f0893cf91db`；
`workplace/ida/re79-movement-1C61-reviewed.json`
SHA-256 `bb7ad097c87306111513eda4e34b75a9d822cc257affe22773eb46f5b681e0e1`。
0229在該次匯出仍標unknown；本輪閱讀完整0231–029B後，確定它比較現在座標與艦隊座標，
只有不同才同步現在／目的地並重畫，回1；相同回0且不重寫目的地。
這項靜態新結論為proven，尚待下一批受控呼叫收據及更新分級匯出。

### 已證實的額度與停止條件

- 艦隊0041–004F把+52/+54抄到+56/+58，0055呼叫分隊初始化0005。
- 分隊004E讀+21巡航速度，0059–0060左移19位；006E–00B0有號比較舊額度，
  嚴格0<old<8.0才累加，否則覆蓋。啟用索敵隊00BE–0148採相同算法。
- 03AA一般分隊到點或餘額不足8.0：03F3 Action設3、03F8–0403清餘額、回0。
  索敵隊未到點的不足一步分支不改Action／餘額。
- 029C步長由艦隊+49 bit0與分隊+16 bit1選8／12／16／24，門檻始終8.0。
- 0444雖051C計算實際位移距離，最後062D `26 83 6f 24 00`、
  0632 `26 83 5f 26 08`固定扣8.0，**不是扣實際位移距離**。
- 0153的第0分隊、未啟用索敵隊先走0229；只有其餘分隊再依完整Action分派。
  舊研究126描述的01D8分派是這個後半部，不涵蓋旗艦路徑。

## 二十案常式受控診斷

入口：[battle_movement_probe.go](../../tools/scripts/battle_movement_probe.go)。
透過專用dosgolem包裝器執行（設定原版唯讀掛載與workplace:/out），例如：

```sh
DOSGOLEM_ORIG=/home/anr2/cht/logh3/org_game/extracted \
DOSGOLEM_EXTRA_MOUNT=/home/anr2/cht/logh3/workplace:/out \
bash /home/anr2/cht/dosgolem-logh3/tools/go.sh run \
  /out/battle_movement_probe.go -output /out/battle79-move-probe-a.json
```

執行前用`tools/py.sh`把同名tracked Go腳本複製到workplace；不改dosgolem工作樹。
完整RAM重設後另建堆疊、DS=ES=54B3、tick0；每案Restore且seed預先固定。
所有額外注入線性位址與完整bytes都在收據。14案比較舊額度−1、0、1、
8.0−最低位、8.0、8.0＋最低位、1000.0，各驗索敵隊未啟用／啟用。
另6案比較一般／索敵隊的自主移動不足額度或抵達停止。
兩次完整JSON相同，SHA-256：
`b900703920bd33e031ff62f3b4afddc64bbb2ed594355f8c5b27b14edf140f4b`。
全fleet／unit records除指定額度或Action欄位外均相同；seed不變。
不走繪圖／碰撞步進的診斷不能代替正常玩家路徑或完整移動驗收。

自動核對／摘要入口：
`bash tools/py.sh /work/tools/scripts/battle_movement_receipts.py`。
它核對三邊界雙跑、7艦隊／55分隊初始化、20案完整records及seed，再產生分級JSON。

## 現行差異與下一步

目前`AdvanceBattle`使用8.8餘額並依實際距離扣除；正式回合先跑完一支艦隊，
再跑各分隊整段。兩者與上述原版契約不同。`BattleUnit`註解仍有舊的現在／目的地
位移對調；本研究以原始+34/+36現在、+28/+2A集結點為準，不把欄位名稱當證據。

下次直接用新階段snapshot／原始bytes補旗艦同步、2583:0451避障、索敵隊歸隊／回收及
多種受控命令。規格080維持DRAFT；不在production或golden test猜補，也不重開已驗第一階段。


## 2026-10-02第三十四輪追加勘誤

181B:00FA在011E先印DS:3CF0移動文字，0126才初始化額度，詳研究375及083。新35邊界原版雙跑與Game逐圈比較見045第三十四輪續證，不重開原資料常式，不外推其他戰場。
