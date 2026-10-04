# 082：艦隊命令保留旗艦行動碼

2026-10-05，第93輪，沿#79。027補充契約驗收；110保持READY，完整062 DRAFT。
事實及原始位址見[研究427](../re/427-fleet-order-flagship-action.md)，
逐檔SHA-256、原指令分級與版本見[來源收據](../internals/data/fleet-order-action.json)。

| 驗證 | 結果 | 範圍 |
|---|---|---|
| 原版艦隊命令八案雙跑 | JSON相同，全SHIP共439296 bytes及seed不變 | 原版局部診斷，選點互動替身 |
| 公開移動／戰鬥舊程式 | 兩案失敗，旗艦7F→00 | 證明回歸檢查能抓到缺陷 |
| 同命令修正後 | 兩案通過，所有分隊Action保留 | 正常Game命令操作 |
| 命令後保存 | 六檔重讀，所有SHIP+1E保留 | 同兩案Game保存 |
| 兩陣營S3正常玩家 | 帝國10次命令／8擊，同盟7次命令／6擊 | 到首次交戰，再回Input並保存 |
| 正式bootstrap獨立重播 | 2824／2730個公開輸入，兩案六檔逐byte相同 | 同版本Game.New與正常入口 |
| 標準測試 | 18套件，13 PASS／5無測試 | remake回歸，不是原版parity |
| S1延長探索 | 修正前0擊，修正後13次交戰、39次實際命令觀測 | 仍未自然勝利，探針如實FAIL |

正常Game與bootstrap只明示固定測試seed，沒有艘數、士氣、位置或功績注入。
S1延長探針採AI被動模式及公開讀檔槽；少數其他戰場在Tick内自然續接，
僅明記實際執行seed setter的時點，不宣稱每個戰場都再次固定seed。
診斷查詢不取RND，不改狀態；它與正式射擊命令觀測分開標示。

## 真實GUI

當前正式cmd/gin3重新建置，入口包括開頭動畫、主選單、防拷138、
公開S3讀檔、同盟座位、艦隊命令及交戰。種子由測試overlay明示設定。
第一批`order79-r93-gui-1`右鍵保持使鏡頭超過一格，回授腳本來回未收斂。
此批保留輸入與失敗trace，不列產品缺陷，不宣稱GUI通過。
後續乾淨重跑結果追加本節。

第2批固定短按及第3批沿用第81輪同步XTEST仍未收斂，兩批完整保留。
原始Update可見一次按住跨多次更新；短按工具的排入佇列回覆不等於遊戲已放開。
不再調參數。長流程交戰GUI在本輪保持未通過。

另以最小正常玩家樣本驗本次修正：`order79-r93-command-gui`正常入口到首戰，
直接操作可見己方旗艦，完成移動、戰鬥兩段命令，沒有鏡頭控制、局面或seed注入。
兩次完成後typed與原池Action均7F；六檔診斷保存成功，實際SHIP.DT0也核對7F。
三張原生X11畫面各307200像素，與正式Game繪圖零差，共921600像素，已目視抽查。
完整PNG、差分及bbox均保存。這只驗當前renderer與視窗相符，不是原版畫面對拍。
右側艦隊行動文字仍顯示「待命」；本輪驗SHIP+1E，沒有裁決這項FLEET文字更新語意。

準備與重跑入口為凍結器的`--prepare-replay`、`--prepare-gui-input`、
`--prepare-command-gui`，再按本機批次log命令建置同名overlay binary，
執行[GUI驗證器](../../tools/scripts/fleet_order_action_gui.py)及凍結器`--check-gui-save`。
輸出路徑固定，已有批次拒絕覆寫；重新研究時使用新輪次。

## 重跑

全部工作在既有Docker工具鏈執行。

```sh
bash tools/gui.sh xrun python3 tools/scripts/fleet_order_action_prepare.py
bash tools/fleet_order_action_oracle.sh a
bash tools/fleet_order_action_oracle.sh b
bash tools/gui.sh test ./internal/game -run '^TestFleet(MoveOrderTwoStages|BattleOrderSetsAction)$' -count=1 -v
bash tools/gui.sh test ./internal/... ./cmd/... -count=1
```

原版與GUI來源拒絕覆寫。雙跑的局部選點替身不能代替原版正常玩家收據。
自然死亡／到期／升級／勝利與占領GUI、原版整回合同態仍依#79判準。
來源核對入口為[凍結器](../../tools/scripts/fleet_order_action_receipt.py)的`--check`。
