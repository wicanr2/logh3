# 368 — 戰場自動補給的跨艦隊排序

日期：2026-10-02；工作項[#79](https://github.com/wicanr2/logh3/issues/79)。
父提交與輪次範圍：[第二十六輪](../goal/2026-10-02-battle-preparation-round-26.md)。
原版／執行器／工具身份沿用[研究367](367-battle-preparation.md)，本輪不修改執行器。

## 問題與原始範圍

`1930:0510`與`1AAB:04BD`在同艦隊比較分隊編號，在不同艦隊優先來源艦隊，
其餘回0。平手輸出次序尚未知，不能假定穩定排序或艦隊編號序。
追查`3C6A:000D`僅為確認玩家庫存分配；不擴張成完整標準函式庫考古。

重播工具：[battle_preparation_sort_probe.go](../../tools/scripts/battle_preparation_sort_probe.go)。

## 固定身份、命令與位址空間

原版`GIN3.COM` SHA-256 `770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918`；
完整1MiB入口記憶體SHA-256
`f70507aaa771a0b97d425037d06696c90385238c6e294a39a103a836960a6808`。
dosgolem專用工作副本HEAD `f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c`、乾淨；
remake父提交`1a036681d4c2ebf136c9ebeb481bb14ceb60e3ea`、本輪研究／工具未提交。
Go 1.24.13及IDA Pro 9.4的image與digest沿用研究367；此次重新inspect相同。
執行器／原版檔皆未修改，輸出UID/GID為1000:1000。

探針先驗原版及入口雜湊，再完整注入；停IRQ，設DS=ES=`54B3`，每案Restore。
所有額外注入的線性位址與完整bytes保留在本機收據，事前固定種子
`lin:5E00C = 78 56 34 12`，不是靠重擲挑通過結果。
排序直接呼叫`3C6A:000D(arrayOff,arraySeg,n,2,callbackOff,callbackSeg)`；
參數次序由`1930:0324..0336`的push／far call確認。
行星三案呼叫自動補給／修理入口，艦隊數清零跳過士氣，限制在資源常式診斷。

在容器內把受版控探針複製至`workplace/battle_preparation_sort_probe.go`，再各跑一次：

```sh
DOSGOLEM_ORIG=/home/anr2/cht/logh3/org_game/extracted \
DOSGOLEM_EXTRA_MOUNT=/home/anr2/cht/logh3/workplace:/out \
bash /home/anr2/cht/dosgolem-logh3/tools/go.sh run \
  /out/battle_preparation_sort_probe.go -output /out/oracle/battle79-sort/run-a.json
```

第二次只換新輸出`run-b.json`；禁止覆寫。兩份完整收據75案相同，SHA-256
`370cfe8b8c7ca2efc1d23e8db6f4dab646be29cda33b2b3123b737a3fc49cb99`；
75案種子bytes都未變。欄位匯出工具：
[`battle_preparation_sort_receipt.py`](../../tools/scripts/battle_preparation_sort_receipt.py)，命令：

```sh
bash tools/gui.sh xrun python3 tools/scripts/battle_preparation_sort_receipt.py \
  --output docs/internals/data/battle-preparation-sort.json \
  --golem-head f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c --golem-dirty ''
```

首次輸出後，重驗必須換新的本機輸出，再逐位元組核對受版控JSON。
[解出欄位](../internals/data/battle-preparation-sort.json) SHA-256
`9df9ec35eb56af3e9445cff7be6359b4107803a2528fcfe2d922ccd0a7c8292f`。
原始完整注入、記憶體與資料庫留本機；Git只保存解出的欄位及重播方法。

## 排序規則：已證實

舊筆記把`3C6A:000D`簡稱為qsort；名稱不是演算法證據。本次原始指令證實為梳狀排序：

1. `n <= 1`直接返回；初始間距`gap = n`。
2. 每輪`gap = floor(gap * 10 / 13)`；零改1，9／10改11。
3. 從清單首筆起，逐一比較相隔`gap`的元素；只有比較結果為正才交換。
4. 間距大於1或本輪有交換，繼續縮減間距並重跑；間距1且無交換才結束。

`006B／0070／0075`為乘10除13，`0081／0086／0094`為9／10改11，
`00DE`為比較結果<=0跳過交換，`0105／010E`為終了條件。
兩bytes元素使用`011E..0157`交換常式；`F37C = 元素寬度/2 = 1`，逐word交換。
候選上限與16位乘積外溢不外推；本輪清單大小0、1、2、3、7、8、12、13、14、15、16、55。

比較器完整規則：同艦隊依分隊byte升冪；不同艦隊，來源指標所屬艦隊優先，
其餘回0。**沒有其他艦隊槽號升冪規則**。不可以換成穩定排序：較遠元素交換會改變平手相對次序。
三種事前固定清單為混合、反向、全部平手；兩比較器各36案，共72案。
可丟棄診斷模型與72案原版輸出全同；模型不是正式remake接線，更不是正常玩家路徑收據。

例如混合7筆，原清單`[0,1,2,3,4,5,6]`輸出`[4,0,5,6,1,2,3]`；
槽0／4屬來源艦隊16，其餘艦隊17／18／19。其他艦隊平手仍被較遠交換重排。
全部平手55筆保持原序；不是所有有平手的清單都保持原序。

## 行星來源指標：已證實與剩餘未知

來源指標是原版DS資料空間的遠指標，不能只保存成一次算出的艦隊編號。
`A42E／A430`在自動補給逐筆艦艇更新，`A440／A442`只在種類6更新。
行星比較器不改讀行星擁有者，而是沿用這些指標所指艦艇的`+01`。

| 固定案例 | 原版結果 |
|---|---|
| 最後艦艇是艦隊16的索敵分遣隊；初始補給來源指向艦隊17 | 行星100枚飛彈全給艦隊16；A42E指向最後艦艇 |
| 本次清單沒有修護艦；舊A440指向艦隊17 | 行星20艘全補艦隊17；A440不變 |
| 本次最後修護艦屬艦隊16，庫存為0；舊A440指向艦隊17 | 行星20艘全補艦隊16；A440仍更新至這艘修護艦 |

案例中的來源指標原始bytes為`9e008b66`（池槽2）及`ec008b66`（池槽3），
原版pool基底`668B:0002`、每筆78bytes；完整指標與資源欄位可回查JSON。
**未知**：首次沒有種類6時的來源初始化、手動命令／換戰場／池槽重用對持續指標的影響。
本輪沒有把這些未知猜成重設、快取艦隊或選最近分隊。

## 非破壞性IDA證據

切段皆前置256零bytes。新增3C6A與1261各切`0x700`bytes，既有1930／1AAB沿用研究367。
IDA程式碼基底`0x10200`；例如原版`3C6A:006B`對應IDA`0x1026B`、
dosgolem線性`0x3C70B`，三種空間分列。

| 段 | `.COM` SHA-256 |
|---|---|
| 3C6A | `19baaa058f8bc637f775b900a255a558ea3b3e6c2f767bd63c4ea7b687db62ce` |
| 1261 | `1f9ebfbe839bb5e2780c0eaac9599d6b9ae671230060e59dbc263d5dd7f8dad3` |

審查索引[battle-preparation-sort-semantics.json](../internals/data/battle-preparation-sort-semantics.json)，
SHA-256 `14f2c3dc4d2774dbec7b8a29cf7a03cdc95c2731bb6931e6cc17c3c21b5efeb9`。
透過`battle_rules_ida.py`自動核對切段雜湊／原始bytes並合併分級語意；
未審查的交換常式仍醒目標unknown，不用工具名稱冒充已證實語意。
正式分級輸出在`workplace/ida/re79-sort-reviewed-{段}.json`，
對拍方法與輸出雜湊入口：[驗證040](../playtest/040-battle-preparation-sort.md)。

額外確認`1261:000B`沿`9898`字節表掃`9894`支戰場艦隊，跳過艦隊`+04=FF`，
把收到的顯示旗標傳給`0067`；不是掃全圖32支。不因此宣稱士氣報告UI已對拍。

## 實作停止線與保留失敗

排序未知已解，來源更新／保留條件已解；完整第一階段仍DRAFT。
除了首次來源狀態、士氣報告續接與整步正式入口，原版人類實際獲得資源時會呼叫
`1930:0AA2`等結果顯示，正式UI也須驗證。不能只接數字或只接電腦案例後宣稱整步完成。
負缺口夾0已有2026-09-19使用者裁決，見[研究349](349-1930-supply-family.md)與規格063，
不重問或重開原版無號溢位的取捨。

跨規格回填另外命中知識路由`local/re-resolution-backlinks.md`並已載入。
欄位JSON內`resolution_backlinks`保存不可變定位、輸入雜湊、舊文件、必要訂正入口與
不受影響理由；匯出程式會在任何入口缺失時拒絕生成。較早349／367已追加勘誤，
規格079已回填；規格029仍保留AI配額平手的驗證限制，不改已驗的AI路徑。

首次IDA自動邊界因既有資料項擋住建立，無JSON；內部日誌明示「IDA未能建立函式邊界」。
在新的切段資料庫先清除該範圍項目，再建立函式，即取得`000D..011D`並以原始retf審查。
這是一次性匯出工具問題；未修改原版或正式資料庫，不把首次exit code當成功。

## 2026-10-02／研究369續證

本篇列出的首次來源、手動命令、池槽重用及結果／等待已由
[研究369](369-battle-preparation-lifecycle.md)補齊指定原版範圍。
換戰場／重新開始／讀檔重設仍未知，完整第一階段仍未接線。
原75案及其雜湊保持不變，不用新收據覆寫舊收據。
