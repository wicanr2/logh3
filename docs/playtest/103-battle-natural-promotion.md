# 第115輪：正常功績升級接續

目前分類：remake正常存檔接續的自然升級、軍階UI與公開存讀已驗。沿#79及108，062 DRAFT。
正式Go、AI、存檔格式、dosgolem與原版素材未修改。

## 存檔與輸入

- 開工HEAD：`501ae687ad2b6973b117a99e5865244e4d277d6d`，工作樹乾淨。
- fixture為第114輪正式入口公開保存的六檔44560 bytes及SAVEDATA.IDX 528 bytes。
  來源由`battle-reinforcement.json`逐檔雜湊核對，索引SHA-256為
  `ba00445ab34f4f28665122098e690b15fc52db318e36ec19d72b0898611f34cb`。
- fixture複製到本輪可寫存檔目錄，再以公開取進度的S0列及雙人模式載入，不改原始遊戲資料。
- 副官23等級4、自然功績2652，符合108的2500門檻。功績來自已驗正常戰鬥及勝後占領。
- 遊戲內TURN95為796年10月16日；真正交戰TURN98為10月28日。雙席均為人類，沒有CPU政策變更。
- 第一次戰術實行前固定一次十六進位seed`12345678`；這是新的存檔接續測試，不宣稱跨第114輪存讀的骰序連續。
- image為`logh3-ebiten:2`，ID為`sha256:6ccf04cfd896019d83b92641b9c072f65e1403cc4ef41e0afdb09b5ee1bab8f8`，Go1.24.13、UID/GID1000、網路關閉。
- dosgolem保持`f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c`且工作樹乾淨，本輪沒有改動執行器或新跑原版整回合。

原版升級證據沿108及既有66案、三表／全局重算與報告契約。本輪不新增原版完整八步同態。

## 失敗保持與目前路徑

- a生成成功，但package game的測試import bootstrap造成循環匯入，遊戲未執行。
  b按UseSaveDir相同的ParseSaveIndex契約配置fixture；獨立正式入口仍使用bootstrap正式API。
- b正常載入至TURN95，艦隊7作戰選單停用占領。星域8仍屬中立，符合既有選單門檻，沒有產品崩潰。
- c不下新命令正常雙席實行，回到命令畫面，零戰場與升級；不把探針期待戰場的失敗當產品缺陷。
- d公開移動7由8直往11，同盟16由31往13。探針以迴圈次數作上限，錯跑到TURN3796，沒有戰場與升級。
  這是探針上限失當，不能宣稱遊戲失敗。原始命令及輸出保留，未改產品。
- e只讀觀測器引用不存在的狀態名稱，編譯失敗，遊戲未執行；f改沿既有狀態及12回合上限。
- f定位第一回合的實際帝國路線為8→13→11，13當时屬敵方，`PlanStep`中止7的命令並停駐。
  同盟16正常依序到20、18、21、16；13易主後，移動命令也受相同門檻阻擋。這符合051及正式規則，沒有新增產品缺陷。
- g改公開7→9→11，再占領16；同盟仍正常16→13。真正相遇在TURN98、星域16，首個戰術回合自然升級。
  h／i補公開軍階視窗，各102檔逐檔全同，全部輸入與升級觀測相同。

## 自然升級及正常續接

| 項目 | 結果 |
|---|---|
| 副官 | 23，庫典森；第8艦隊的副官 |
| 升級前後 | 等級4→5，功績2652→152，扣除已證實門檻2500 |
| 通知 | 能力提升，統率力80、防禦92、回轉95，等待180 |
| 軍階 | 上尉保持；4與5同屬既有軍階區間，不能稱軍階晉升 |
| 升級觀測 | TURN98、星域16、戰術回合1，固定seed後依正常規則取亂數 |
| 正常末端 | 回到命令畫面，副官等級5、功績1128；後續功績來自正常戰鬥 |
| 戰術操作 | 9次正常命令、20擊交戰；零勝利／勝後占領報告，不宣稱新勝利 |
| 存讀 | 六檔44560 bytes，助手、正式入口、公開保存、公開載入重存及GUI全同 |

正式bootstrap b／c各145檔逐檔全同，共311項正常輸入計畫，沒有fixture之外的局面注入。
唯一升級通知的頭像、文字、等待、完整六表及seed與正常助手相同；升級報告及軍階視窗完整色號圖也相同。
存檔fixture由正式`bootstrap.UseSaveDir`讀索引，再從公開選單載入；Game助手的環境配置不能代替這條證據。

正式60 Hz真實X11完成311項操作與60張原生640×480全圖，18,432,000像素零差，8212次更新TPS均60。
報告、公開艦隊視窗、保存與載入均在實際GUI擷取。21次額外右鍵鏡頭回授完整保留；不宣稱相同實體鍵鼠時序。
末筆實際輸入137.305秒；助手26556筆輸入中的Tick總和450.829624秒不是GUI時長。
已目視能力提升通知及副官等級5的上尉欄位，文字未溢出或裁切。音訊裝置與人耳未在本輪驗收。

標準18套件13 PASS／5無測試，Game126.270秒、gin3 60.872秒。
108限定既有成長規則及本正常升級／存讀樣本CONFORMED；062完整八步、正常原版整回合同態與自然非旗艦死亡仍待驗。

本輪唯一來源由[核對器](../../tools/scripts/battle_natural_promotion_receipt.py)產生，入口為
`docs/internals/data/battle-natural-promotion.json`。保留所有成功、失敗批次、原始命令與工具生成檔的SHA-256、大小及輸入時間戳。
父12372來源、164原始輸入及520正式來源全核；正式資料、Go、AI及存檔格式未改。
工作根擁有權掃描零root-owned檔案及Markdown誤建目錄，工具與收據抽樣均1000:1000。
本輪一次性容器均退出刪除，專用dosgolem工作樹乾淨；未重建工具鏈或清理其他專案資源。

## 工具

`tools/scripts/battle_natural_promotion_probe.py REV --command meet-alliance16`
在新批次產生唯讀觀測overlay與正常輸入。REV限英數字，拒絕覆寫既有來源。
Go探針以專案`tools/gui.sh xrun`執行，`LOGH3_FLEET_CHAIN_OUTPUT`指定獨立本輪輸出。
正式重播使用`tools/scripts/battle_natural_promotion_replay.py REV --input <inputs.json>`，加`--gui`執行真實X11。
同一REV的正式計畫與GUI產生器使用不同檔名；每次輸出拒絕覆寫。

驗證命令在Docker內執行，沿既有包裝器：

```sh
python3 tools/scripts/battle_natural_promotion_probe.py h
LOGH3_FLEET_CHAIN_OUTPUT=/src/workplace/promotion79-r115-h go test -overlay=/src/workplace/promotion79-r115-h-overlay.json -run '^TestBattleFinishNormalProbe$' -count=1 -v ./internal/game
python3 tools/scripts/battle_natural_promotion_probe.py i
LOGH3_FLEET_CHAIN_OUTPUT=/src/workplace/promotion79-r115-i go test -overlay=/src/workplace/promotion79-r115-i-overlay.json -run '^TestBattleFinishNormalProbe$' -count=1 -v ./internal/game
python3 tools/scripts/battle_natural_promotion_replay.py b
python3 tools/scripts/battle_natural_promotion_replay.py c
python3 tools/scripts/battle_natural_promotion_replay.py b --gui
go test ./internal/... ./cmd/... -count=1
python3 tools/scripts/battle_natural_promotion_receipt.py --check-only
python3 tools/scripts/battle_natural_promotion_receipt.py --write
python3 tools/scripts/battle_natural_promotion_receipt.py
```

GUI外層為`LOGH3_TIMEOUT=45m bash tools/gui.sh xrun ...`；其餘批次為`bash tools/gui.sh xrun ...`，預設15分鐘。
新REV會建立獨立輸出，h／i及b／c原收據不重用作新的探索目錄。

#79保持OPEN，正常自然非旗艦死亡與原版整回合同態仍待驗。#48直接確認題保持待答。
