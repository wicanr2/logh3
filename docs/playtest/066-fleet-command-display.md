# 066：交戰命令顯示修正

2026-10-04，第77輪。沿[#48](https://github.com/wicanr2/logh3/issues/48)、
[#79](https://github.com/wicanr2/logh3/issues/79)、[103 READY](../spec/103-fleet-command-display.md)及
[研究414](../re/414-fleet-command-display.md)。父提交e30f3869b58db2d484f6109931e5181e784026b7。

## 比較結果

原版七組受控狀態各獨立雙跑。完整RAM、色號、PNG與有效色盤相同。
快照解碼後完整機器欄位與DOS段相同；gob map序列化順序不同，原檔各自雜湊保留。
執行器f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c未改、工作樹乾淨。
bundle與GIN3.COM雜湊沿AGENTS.md及研究414。

| 同狀態顯示診斷 | 第76輪繪法差異像素 | 修正後差異像素 |
|---|---:|---:|
| 等待目標 | 6404 | 1 |
| 情報1摘要 | 6658 | 1 |
| 情報2摘要 | 6237 | 1 |
| 情報3摘要 | 6057 | 1 |
| 情報1武器彈窗 | 4670 | 1 |
| 情報2武器彈窗 | 4249 | 1 |
| 情報3武器彈窗 | 4069 | 1 |

七圖各比較307200像素。右側資訊、小地圖、頂底黑邊及有效16色色盤相同。
射程圓、兩個焦點橢圓、長十字、武器彈窗隱藏圖形、目前分隊亮標及游標已繪製。
完整差分仍包含(255,255)方向物件覆蓋亮標一像素，未核准差異各1，完整對拍未通過。
沿#89既有未初始化字鍵裁決，不修改正式順位、不加遮罩、不升格CONFORMED。

顯示fixture還原注入前已登錄的物件及快取，不能把目前Action bit7當作重新建物件的指令。
原池仍唯讀，沒有把原版像素注入remake。比較等級限受控顯示，不推廣正常全流程或其他typed欄位。

## 正常玩家路徑

正常主選單、防拷138、S3讀檔、戰略實行、戰術下令及實行，兩陣營各獨立雙跑。
帝國10次下令、7擊；同盟7次下令、5擊。每場執行前固定remake測試seed12345678。
沒有位置、命令或原版RAM注入，CPU位置保持061政策。

每陣營的命令等待及彈窗各一張完整圖，由正式bootstrap與Game.New重播相同公開輸入。
四圖色號逐位元組相同；完成交戰後返回戰場，六份保存與正常Game相同，重播兩次輸出亦相同。
正常畫面已目視。這項證據只證明正常remake路徑，不宣稱正常原版整段same-state。

公開純移動、彈窗取消、返回射程、移入面板清敵方摘要及再移回恢復畫面均通過回歸。
純顯示不改SHIP／FLEET／ADML／WORK或RND，連續Frame沒有累積XOR。
標準`tools/gui.sh test ./cmd/... ./internal/... -count=1`：13套件通過、5套件無測試。

## 重現與來源

- [原版來源索引](../internals/data/fleet-command-display-native.json) SHA-256：`d56ab5516e494b43987aab0d72210b71d2ff93f1b56c9c3576dfba08c255b1db`。
- [正式顯示索引](../internals/data/fleet-command-display-game.json)保存逐檔來源與產物雜湊、完整差分、原版排序物件及正常輸入。
  SHA-256為`830e5c5615be977c5e18cff23fe990c18cc6ee0ad01802b17562a8338ee33405`，255份來源。
- [原版重播器](../../tools/fleet_command_display_oracle.sh)保存每案完整旗標，拒絕覆寫。
- [Game凍結器](../../tools/scripts/fleet_command_display_game_receipt.py)的`--overlay`產生固定seed與唯讀stage方法；正式產品來源不改。
- [正常重播器](../../tools/scripts/fleet_command_display_replay.go)逐圖核Game命令畫面及六份保存。

```sh
tools/gui.sh xrun python3 tools/scripts/fleet_command_display_game_receipt.py --overlay
tools/gui.sh xrun env LOGH3_FLEET_CHAIN_OUTPUT=/src/workplace/round77-chain-c LOGH3_FLEET_DISPLAY_OUTPUT=/src/workplace/round77-final-d go test ./internal/game -run '^(TestFleetCommandDisplayNativeRegression|TestFleetCommandHoverTransitions|TestNormalFleetPlayerChain)$' -count=1 -v
tools/gui.sh xrun go run -overlay workplace/round77-display-replay-final.json tools/scripts/fleet_command_display_replay.go -input workplace/round77-chain-c/normal-0/inputs.json -out workplace/round77-bootstrap-final-a-0
```

另一陣營將兩個0改1；第二次使用chain-d／final-e及bootstrap-final-b-0／1的新目錄。
收尾使用`--freeze`產生索引；存在的歷史輸出先保持，重跑用新工作目錄。
修正前對照由父提交原fleetcombat.go及`--before`產生，保持已校正fixture，僅撤回本輪顯示繪法。
初始fixture漏圖、輸出目錄及重播階段錯誤的探針保留，正式結果以最終索引為準。

103保持READY，完整056／092／062仍DRAFT。真實X11交戰操作、音效接線／聆聽、正常原版完整同態與其他八步仍缺。
#48／#79保持OPEN，未宣稱整款remake或跨平台發行完成。
