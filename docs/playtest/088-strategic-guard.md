# 088：戰略殘留守軍與原池

2026-10-05，[#56](https://github.com/wicanr2/logh3/issues/56)，[115](../spec/115-strategic-assault-guard.md)限定CONFORMED。
原版來源、執行器、完整保存、注入、輸入與雙跑雜湊見[434](../re/434-strategic-guard-lifecycle.md)。
五個固定原版局面沿[433](../re/433-strategic-assault-edges.md)，輸出雜湊
`d22b0e4b93c7283c23c0c26c0e8dca859c487fabb9ffec5e6b5f6ef3ff373653`。

## 比較範圍

| 層級 | 驗證 |
|---|---|
| 原版生命週期 | 完整8-1保存續行至下一星域27返回；55筆池清單保留，兩跑相同 |
| 正式typed規則 | 空、近距、Out、遠距、地形五案指定欄位與seed零差；忽略地形仍36差異 |
| 正式原池規則 | 五案各完整六表44624 bytes、清單164 bytes與seed零差，共223940 bytes |
| 正式Game | 同五案完整六表、保留池序與seed零差；刻意污染typed數值仍讀原池 |
| 保存 | 五案正式SaveSlot六檔重讀，SHIP／FLEET／ADMIRAL／PLANET保持原版全表 |
| 清單契約 | 活艘數、Out、位置、飽和、索引副本、越界停止及EndBattle保留池身份 |
| 正常玩家抽樣 | 公開選單占領亞斯提並實行五回合；既有勝方67案、保存與報告等待保持 |

五案屬明示狀態注入的規則same-state，不稱自然完整原版回合或GUI全圖對拍。
新遊戲沒有Slot的既有typed分支保持，不外推原池全同。報告幾何、字型與新GUI未改驗收範圍。
跨星域池可進戰略入口；戰場勝方入口仍拒絕異星域池，沒有放寬110原有保護。

## 重建

使用既有`logh3-ebiten:2`與`tools/gui.sh xrun`，原版唯讀、UID1000、網路關閉。

```sh
bash tools/gui.sh xrun go test ./internal/rules ./internal/game -run 'TestStrategicGuard|TestOccupyAstarteRunsStrategicAssault|TestBattleFinish' -count=1 -v
bash tools/gui.sh xrun go test ./internal/... ./cmd/... -count=1
bash tools/gui.sh xrun python3 tools/scripts/strategic_guard_review.py --check
```

首次typed接線驗出FLEET7／SHIP88 bytes漏寫；沿已驗共用1512原池流程修正，原版輸入與門檻不變。
Game診斷初次缺Battle規格及錯誤存檔檔名，修正實際資源／SlotExt後乾淨重跑；不是產品缺陷。
第99輪DRAFT反例固定於提交1ff96c1；舊來源索引保持，新的現行來源在
[strategic-guard.json](../internals/data/strategic-guard.json)，以[核對器](../../tools/scripts/strategic_guard_review.py)核對。
