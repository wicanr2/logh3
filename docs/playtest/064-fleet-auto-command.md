# 064：自動命令資料、原池與存讀

日期2026-10-04。範圍為[101](../spec/101-fleet-auto-command.md)，完整092／056仍DRAFT。
原版與重製使用固定相同資料、等價候選及相同RND條件；比較的是命令資料與局部規則。
不稱正常玩家入口或完整艦隊交戰same-state。

## 證據與命令

- [原版完整來源](../internals/data/fleet-auto-command.json)，SHA-256 23627c23a9d2ef93991c842e00fbf43df75f00d2043c78788454e868d3a19401。
- [Game驗證索引](../internals/data/fleet-auto-game.json)保存程式、測試、log及來源雜湊。
- [研究412](../re/412-fleet-auto-command.md)記注入、原始定位、兩種fixture錯誤與證據限制。

重播走`tools/combat_player_command_oracle.sh command79-r75-<案例>-<a|b>`，
案例與固定參數在`tools/scripts/combat_player_command_plan.py`。輸出存在即拒絕覆寫。
比較／凍結走`tools/py.sh /work/tools/scripts/combat_auto_receipt.py <新輸出檔>`。
原始state／RAM唯讀，567份本機來源由JSON雜湊索引，dosgolem版本為f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c。

```sh
bash tools/gui.sh test ./internal/rules ./internal/game -run '^TestFleetAutoCommand' -count=1 -v
bash tools/gui.sh test ./cmd/... ./internal/... -count=1
```

## 結果與範圍

| 抽樣 | 結果 | 限制 |
|---|---|---|
| 原版31案雙跑 | input／end RAM、screen、pal及PNG逐bytes相同 | 父state已注入；不稱正常玩家 |
| 純規則31案 | 115筆attack／defence／shots／百分比／量、三word、score、排列、目標、武器及seed符合 | 十案含非正常池或編制身分，只作局部規則 |
| Game21案 | 有效一般池命令及預估暫存符合，艘數／庫存／typed分隊不變，存讀保留 | 沿受控RAM，未由UI呼叫 |
| 原版缺beam錯誤 | 多候選可沿上一候選武器2，選中mask2目標；+3E成3001 | 自動原版行為保持；手動提交不放寬 |
| 錯誤拒絕 | 純規則十二案、Game六案不改SHIP／typed／seed | 缺原版skip，不用代用品 |
| 控制模式 | 雙人另一陣營仍用人類score；061單人被動CPU拒絕 | 不因此啟用增強戰術攻擊 |

原版比對期望直接取RAM／暫存器，不從probe模型或remake結果填入。
取消／防禦、零距離及其他既有收據不重生；101僅關閉自動命令資料子缺口。
正式命令UI、跨艦隊順位、逐擊動畫／音效及正常完整玩家路徑未驗，#48／#79保持OPEN。
