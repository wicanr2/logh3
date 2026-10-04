# 063：交戰候選與命令保存

日期：2026-10-04，第74輪。沿[100](../spec/100-fleet-combat-candidates.md)及
[研究411](../re/411-fleet-command-data.md)。父提交dfccda34441a6271b18fd3594bfa340f177fa6ea。

## 範圍與來源

候選為受控原版資料對拍；Game提交為原池資料介面診斷。
正常bootstrap抽樣只驗S3首戰資料域與唯讀性，不是完整玩家武器操作收據。
沒有新畫面修改或同狀態像素宣稱；Game完整艦隊交戰尚未呼叫。

原版GIN3.COM／bundle身份、完整命令旗標、注入與工具映像沿
[候選收據](../internals/data/weapon-candidates.json)及[命令收據](../internals/data/fleet-combat-command.json)。
SHA分別59a5ad2bd68a54c888179be3d85a6b99c15f16a1892b7d2b05c4de0667ad601f與
e3b39582bfda07105d27837747d46bca19eaf2f1fb5d6ff383500870e4a6870d。
dosgolem f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c未改動；原版a／b輸入與report全部重核。
GUI用logh3-ebiten:2，映像身份沿完整索引；所有工作UID/GID1000，原版唯讀。

## 重播命令

```bash
bash tools/gui.sh xrun bash -c 'go test ./internal/rules ./internal/game -run "^Test(FleetCombat(Candidates|Command)|NormalSpottingInformationAndSave)" -count=1 -v > workplace/round74-command-final.log 2>&1; test_result=$?; tail -n 45 workplace/round74-command-final.log; exit "$test_result"'
bash tools/gui.sh test ./cmd/... ./internal/... -count=1
bash tools/gui.sh xrun python3 tools/scripts/fleet_command_data_receipt.py
```

第74輪標準命令輸出保存為workplace/round74-standard.log。凍結器會拒絕不同歷史紀錄，
重跑要另存新批次，再明示審查；不可覆寫既有對拍結論。

| 驗證 | 結果 |
|---|---|
| 原版候選48案 | 111次查詢AX、61次mask寫入零差；不讀probe期望 |
| 候選邊界 | 射程相鄰整數、兩word庫存、情報／Out、地形／分隊遮蔽、零艘數及重疊 |
| Game命令 | 0／1／2／FF／3五案，變動限+2C／+2E，彈藥與RND不動 |
| 保存 | SaveSlot→原六表→LoadSlot，目標池與武器全保留；FF／3保持既有目標 |
| 非法提交 | 十案拒絕，原池／分隊／RND不變，包含061單人CPU |
| 正常玩家資料域 | 主選單取進度、防拷138、S3帝國、實行到首戰；48分隊查詢不改畫面或原池 |
| 標準回歸 | 18套件，13PASS、5無測試；Game85.420秒、gin3 67.513秒 |

正常首戰固定事前seed D3DC167E；首戰380／799/3/1、7艦隊／55池、末端B6E4B78A保持。
Game提交診斷用第65輪near輸入RAM，SHA
b0b72e8a6d92a31b6dec3dae6b0016f3aad976fc216fb9b5ebf5187b94d16be1；傷害RND12345678未消費。

限定log SHA-256為d54a404c3e1aaae192e699b4b088c205138608988c4f68bee9e86a7225a9a0e3，
標準log為b8739f75b5f67600380412c3827f3b41944212fd2574b2dc4177bb2abd9fac4b。
[完整索引](../internals/data/fleet-command-data.json)共201來源，SHA-256
4b42447bfe2bbe899bc1c51987501b9b2c21cd95a1323452c38383f490acea80。

100限定資料鏈CONFORMED，原版自動選擇、完整武器UI與第④步演出仍未實作。
正常資料域抽樣不取代它們；#48／#79保持OPEN，完整056／092仍DRAFT。
