# 053：原版艦隊戰鬥選擇與提交的可重現收據

2026-10-03，第63輪。來源[研究401](../re/401-fleet-combat-command.md)、
[固定JSON](../internals/data/fleet-combat-command.json)。
JSON SHA-256：e3b39582bfda07105d27837747d46bca19eaf2f1fb5d6ff383500870e4a6870d。

比較等級：受控原版獨立重播。沒有remake圖、差分圖或正常玩家路徑驗收，不能稱same-state通過。
原版畫布640×480，滑鼠虛擬640×400；全圖比較不裁切、縮放或遮罩。

| 範圍 | 案例數 | 結果 |
|---|---:|---|
| 目標選單、兩武器、灰色項目、右鍵、防禦、自動 | 7 | RAM／色號／色盤／PNG雙跑相同 |
| 演出前的提交邊界 | 5 | 上述四種提交／取消及防禦；seed及庫存未消耗 |
| 交戰結束、下一階段前 | 2 | 光線砲／飛彈；完整78-byte前後記錄及seed相同 |
| 返回戰場的後續末端 | 2 | RAM／完整圖相同；情報與標記更新另列 |

共16組、32次固定輸入重播，64組檔案比較相同；每張307200像素色號相同。
序列化state是否相同逐案另列；共同採固定a父快照，不由畫面相同宣稱完整state一致。

重跑工具：[原版包裝器](../../tools/fleet_combat_command_oracle.sh)、
[固定計畫](../../tools/scripts/fleet_combat_command_plan.py)、
[收據驗證](../../tools/scripts/fleet_combat_command_receipt.py)。
每案以新a／b目錄保存，已有輸出時拒絕覆寫。既有收據重讀直接驗JSON的files雜湊。

```sh
bash tools/fleet_combat_command_oracle.sh combat79-r63-target-mapped-a
bash tools/fleet_combat_command_oracle.sh combat79-r63-target-mapped-b
# 依計畫依序重建laser等武器案例，再重建submit／phase-end／finish。
bash tools/py.sh tools/scripts/fleet_combat_command_receipt.py workplace/round63-rechecked.json
```

完整祖先注入、實際旗標、工具映像與執行器HEAD保存在JSON，不能只抄上列兩條命令。
首點預移動與Y座標兩次探索失敗保留workplace；修正同一執行器後重跑，未改oracle。

目標選單、取消、灰色項目、光線砲演出與防禦返回戰場已檢視。
正式引擎未變，標準產品回歸另見WORKLOG；這份收據不能替代092的正常玩家UI／保存及傷害驗收。
