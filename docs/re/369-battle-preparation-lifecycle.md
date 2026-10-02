# 369 — 自動補給來源生命週期與人類結果

日期：2026-10-02；工作項[#79](https://github.com/wicanr2/logh3/issues/79)。
前置排序證據：[研究368](368-battle-preparation-sort.md)。
本輪範圍：[第二十七輪](../goal/2026-10-02-battle-preparation-round-27.md)。
原版／工具固定身份沿用研究367／368；執行器專用工作副本f7523f9，未修改。

本輪起始未知：首次A440／A42E值、手動命令與換戰場／池槽重用、資源結果顯示及士氣續接。
靜態word搜尋只作IDA查詢線索；不是「無其他寫入者」證明。

重播入口：[原版工具](../../tools/battle_preparation_lifecycle_oracle.sh)，
例如 `bash tools/battle_preparation_lifecycle_oracle.sh supply-panel battle79-life-supply-panel-a`。
`boot`、`supply-panel`、`supply-end`、`repair-trace` 分別保存開機、補給面板、
補給後第一階段結束與修理探索；批次必須全新，探索不得覆蓋既有收據。
此工具走既有容器包裝器，固定執行器提交並拒絕不乾淨的工作樹。

來源局部診斷：[Go探針](../../tools/scripts/battle_preparation_lifecycle_probe.go)。
完整入口注入與額外bytes留本機；三個來源寫入端停在寫入後，不執行假返回。
其他案例只驗常式資源排序，明示跳過士氣；不代替完整玩家流程。

人工分級來源：[指令索引](../internals/data/battle-preparation-lifecycle-semantics.json)。
匯出仍透過`battle_rules_ida.py`核對輸入及bytes，自動合併等級與證據，不改原始名稱。

欄位核對／匯出：[收據工具](../../tools/scripts/battle_preparation_lifecycle_receipt.py)，
重驗另指定新的本機輸出；六案、六組第一階段收據與十五項分級不符即拒絕生成。

## 固定身份與位址基準

原版`GIN3.COM` SHA-256
`770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918`。
原版bundle來源／雜湊沿用專案CONTEXT，沒有改原始檔。
專用dosgolem HEAD `f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c`、乾淨；
remake父提交`6158b185595f96e4970c7b8af86c89dcae7ed042`，本輪只有研究／工具／規格修改。
Go 1.24.13，image `golang:1.24-bookworm`，IDA Pro 9.4使用
`ida-pro-9.4-idapython:locked-v1`；image digest與限制沿用研究367並重新核對。

檢查點位於**執行器副本**`workplace/fcmd/eng-t5.state`，262326bytes，SHA-256
`c6ce5af3a8d52acc415ac1029a29a0905e9c95e02ac0e0ba547e71300fd947c1`。
來源常式診斷的完整1MiB入口SHA-256
`f70507aaa771a0b97d425037d06696c90385238c6e294a39a103a836960a6808`。
原版DS=`54B3`，`A42E`／`A440`分別對應dosgolem線性`5EF5E`／`5EF70`。
池基底`668B:0002`、每筆78bytes；IDA切段加256零bytes，程式碼基底`0x10200`。
例如`1930:061F`分別是IDA`0x1081F`、dosgolem線性`0x1991F`，不可混成單一位址。

常式診斷先完整注入、停IRQ、設DS／ES，每案Restore；事前種子`0x12345678`。
原版正常控制流案例則在第一階段入口前一指令`1211893242`固定種子並明示注入局面。
兩類案例全部在本機收據保留完整注入bytes；沒有換seed挑結果。

## 來源：已證實範圍與限制

兩次未修改原版開機到主選單，完整1MiB相同，SHA-256
`153b32b19ac630602fbda27796a887fe8617ac34a87ab91e24f832e485ecd5da`。
兩個來源都是四個零byte；實體位址1為0。零來源不是池槽0：
比較器仍解參照`0000:0001`，在純常式診斷將該byte改16，優先對象就由艦隊0改艦隊16。
這個人工更改中斷向量的案例已停IRQ，只隔離解參照規則，不外推真實DOS初始化。

三個來源寫入端都取得局部執行證據：

| 入口 | 寫入後停止點 | 條件 |
|---|---|---|
| `1930:05EC` | `1930:0623` | 手動運補先保存引數，再找候選 |
| `1AAB:0583` | `1AAB:05BC` | 手動修理先保存引數，再找候選 |
| `1AAB:0190` | `1AAB:01E7` | 電腦前置修理只處理種類6及指定陣營 |

局部探針使用假far frame，停在寫入後，不執行返回、不冒稱正常手動UI已驗。
有效來源保留池槽3，第一次其艦隊byte為17，行星20艘補到艦隊17；
第二次明示將同一池槽艦隊byte改16，來源四bytes完全不變，20艘改補艦隊16。
這證明讀活指標，不證明原版何時真的配置或回收那個槽。
六案完整收據雙跑SHA-256
`f1a95e05dbdc0aec055c10ae07de89ae93ea1a76f1b4d8a6d09b59c280d82c27`，所有種子不變。

**未知**：換戰場／重新開始／讀檔是否另行重設持續來源。
靜態word掃描找到三類寫入，沒有直接其他命中不等於不存在間接／整塊寫入。
此未知仍阻擋完整第一階段生命週期實作，不自行猜成每場重設或跨載入永久保留。

## 人類結果與士氣續接：已證實

六組重播都沿`eng-t5`正常控制流走到第一階段，注入僅為建立所需診斷局面：

- 補給：人類池88改種類5並同步艦隊組成種類，來源2000枚飛彈／2000架戰機，池80兩格清零。
  實際轉入2000／2000、來源清零。面板顯示來源艘數4000，目標戰艦3000，
  缺額飛彈-10000／戰機-34000以紅字顯示，刷新後直接返回。
- 修理：人類池88改種類6並同步組成，來源20艘；池80艘數50、上限75。
  面板實際顯示戰艦70／損壞艘數5，來源剩餘0。沒有追加按鍵仍接回士氣、防衛力與移動。
- 士氣：只把人類艦隊12先前低士氣bit0設1，現有完整艘數不改。
  重算後士氣100、bit2設立，印恢復報告並等待；報告關閉後bit0／bit2清零，才回復防衛力。

| 事件 | 補給案例步數 | 修理案例步數 | 士氣案例步數 |
|---|---:|---:|---:|
| 結果面板刷新後 | 1212045377 | 1212415406 | — |
| 士氣等待入口`3765:006C` | — | — | 1213517110 |
| 防衛力入口`181B:0083` | 1213547771 | 1213521804 | 1231561041 |
| 第一階段結束／移動入口`181B:00FA` | 1213547864 | 1213521897 | 1231561134 |

階段前後資料及完成面板各兩跑，完整記憶體、640×480色號與色盤相同。
這些是**注入診斷局面下的原版正常控制流證據**，不是無注入玩家路徑，也不是remake畫面對拍。
沒有新增等待模式、時間精度或硬體wall-clock聲明。

## IDA審查與重播

1930及1AAB切段雜湊分別是
`9e448baf09b748601da82f9febb5d00f6c3161432e7f0a52bef0643461b385ae`、
`92f8d0a53dbb26d5c8ed27474e66560f73390a2933b69c526749ab28156252a2`。
新一次性副本`re79-life-reviewed-{段}.com`，使用`battle_rules_ida.py`與分級索引：

```sh
timeout 2m tools/ida.sh raw idat -A -L/work/re79-life-reviewed-1930.log \
  '-S/idascripts/battle_rules_ida.py /work/re79-life-reviewed-1930.json 1930 05EC:17D 0AA2:300 --index=/work/re79-life-reviewed-index.json' \
  /work/re79-life-reviewed-1930.com
```

1AAB改查`0190:CB 0583:197 08FB:138`；輸入／輸出名同樣換段值。
五個函式以原始retf及控制流核對，十五項分級自動合併，其餘仍醒目unknown。
`1AAB:024D`其實是0190裡面的call，探索時誤作入口，只產生尾段；正式索引／匯出明確排除。

Go探針先在容器複製至專案`workplace/`，再由執行器包裝器各跑一次：

```sh
DOSGOLEM_ORIG=/home/anr2/cht/logh3/org_game/extracted \
DOSGOLEM_EXTRA_MOUNT=/home/anr2/cht/logh3/workplace:/out \
bash /home/anr2/cht/dosgolem-logh3/tools/go.sh run \
  /out/battle_preparation_lifecycle_probe.go -output /out/oracle/battle79-life-probe-r2-a.json
```

第二次改新的`r2-b.json`。面板／終點各用原版工具的`supply-panel`、`supply-end`、
`repair-panel`、`repair-end`、`morale-panel`、`morale-end`模式，批次名`battle79-life-<模式>-a／b`。
完整flags、meta及注入保留本機；正式核對：

```sh
bash tools/py.sh tools/scripts/battle_preparation_lifecycle_receipt.py \
  --output workplace/battle79-life-recheck.json \
  --golem-head f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c --golem-dirty ''
```

欄位／輸出雜湊與範圍見[驗證041](../playtest/041-battle-preparation-lifecycle.md)。
知識路由另命中並載入`local/re-resolution-backlinks.md`；367／368／040保留歷史，
追加續證入口，生成器會檢查三份舊文件標記及三個原始入口。

## 保留失敗與下一閘門

重播工具首次把執行器快照誤查成remake目錄；修正路徑與雜湊檢查，原始檔沒改。
局部假堆疊首次連續線性寫越過FFFF，原版讀到另一段值；改逐word以16位位移回繞，
並要求來源等於指定參數才接受收據。舊`probe-a／b`保留但不作驗收，另產`r2-a／b`。
欄位初稿曾把士氣byte讀成word，已按原始型別改正；原版收據未受影響。
這些均為工具／驗證腳本問題，不是遊戲缺陷。

規格079只把已量測子契約標READY，完整第一階段仍DRAFT。
下一閘門只追查實際重新開始／載入／戰場入口的來源寫入，證據充分即回正式回合接線；
不擴張成整個記憶體管理／硬體驅動考古。#79保持開啟。

## 2026-10-02／研究370續證

指定重新開始、正常S3讀檔及達希利戰場入口的持續來源行為已由
[研究370](370-battle-preparation-reset.md)與三組雙跑補證；
第一階段正式接線、完整結果面板與指定資源欄位、士氣續接、正常實行及正式存讀
見[對拍042](../playtest/042-battle-preparation-first-phase.md)。本篇「完整第一階段仍DRAFT」
保存當時證據閘門；目前狀態以規格079為準，完整八步#79仍開啟。
