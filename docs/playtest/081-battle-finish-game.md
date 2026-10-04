# 081：戰場結束正式規則、報告與保存

2026-10-05，第92輪，沿#79與 [110 READY](../spec/110-battle-finish.md)。
父提交 `8b8b2d73a9e634537ad0c4b67daab53b8d024866`，本輪程式工作樹有修改。
原版、產品、測試、圖片與存檔的精確來源見 [正式收據](../internals/data/battle-finish-game.json)。
1046來源索引SHA-256 `46f54109e7bb664077e8e25081ce0836da82ab9daebf3923205b90972eee2af3`。
原版收據沿第89至91輪固定父保存與明示注入，本輪沒有修改或覆寫。

| 核對 | 結果 |
|---|---|
| 正式規則 | 60案結束判定與七案非零守軍，完整六表、清單與seed符合原版 |
| 正式Game | 同67案，固定勝負、敗方報告順序、占領續接及typed回寫均通過 |
| 真正占領callback | 兩案完整末端；唯一實際報告前六表與seed相同，返回後才加上限功績 |
| 報告等待 | 等待時拒絕保存與下一回合，左右鍵逐則返回，末端續接一次 |
| 報告畫面 | 50張，各58240像素，共2912000像素；索引色與RGB零差，16色相同 |
| 六檔保存重讀 | 六案，人類座位1／2／3與雙向勝方；完整六表保存相同，LoadSlot清續接狀態 |
| 正常入口回歸 | 帝國10次命令／7擊、同盟7次命令／5擊，正式bootstrap與玩家輸入及保存通過 |
| 標準測試 | 18套件，13通過、5無測試；無失敗 |

## 比較範圍

報告比較為 `same-state`，僅限預先固定矩形 `(32,180,416,140)`。
兩側使用相同報告內容、收件座位、頭像與游標；沒有遮罩或像素排除。
保存640×480原版圖、remake圖、差分圖、全圖與區域統計。
診斷renderer使用空白背景，全圖差異照實保存，不宣稱整張戰場same-state。
已目視雙方一張同盟人類占領報告，文字、頭像與邊界無溢出。

完整六表為ADMIRAL、FLEET、SHIP、WORK、STELLAR、PLANET，包含未用欄位。
原版清單 bytes 保持；Game另核艦隊與池順序，並驗後續同步不能蓋回舊存量。
六檔保存為第8-5步診斷，明示切入StateSaveGame；不替代正常勝利後選單。
正常玩家回歸只涵蓋既有首戰命令、交戰及保存，不包含自然勝利或占領GUI。

## 重現

原版雙跑入口、完整注入與時間序列沿 [078](078-battle-finish.md)、
[079](079-battle-finish-attack.md)及 [080](080-battle-finish-raw.md)。
缺原版收據的測試會skip，不能把skip當成通過。本輪零skip。
原版 dosgolem HEAD固定為 `f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c`，工作樹乾淨。
原始bundle與GIN3.COM雜湊、原始保存及RAM均在索引中。

以下命令均由既有Docker包裝器執行，使用 `logh3-ebiten:2`、非root、無網路。
輸出目錄只用於本輪驗證，重生既有收據前先另選批次，避免覆寫歷史。

```sh
bash tools/gui.sh xrun python3 tools/scripts/battle_finish_formats.py
bash tools/gui.sh xrun bash -c 'export LOGH3_PROMOTION_OUTPUT=/src/workplace/finish79-r92-render LOGH3_FINISH_OUTPUT=/src/workplace/finish79-r92-render; go test ./internal/rules ./internal/game -run "Test(BattleFinish|WinnerOccupation|Strategic|NormalFleetPlayerChain)" -count=1 -v > workplace/finish79-r92-targeted.log 2>&1'
bash tools/gui.sh xrun bash -c 'go test ./internal/... ./cmd/... -count=1 > workplace/finish79-r92-standard.log 2>&1'
bash tools/gui.sh xrun python3 tools/scripts/battle_finish_game_receipt.py
```

凍結器拒絕覆寫來源索引，會檢查原版雙跑、測試狀態、50組畫面及六案保存。
第89至91輪原始模型索引保留當時來源，不能以新production程式雜湊改寫歷史。
110尚未CONFORMED；正常自然勝利、占領GUI及完整062整回合同態仍待續。
