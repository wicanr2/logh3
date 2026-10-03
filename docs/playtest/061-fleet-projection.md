# 061：typed艦隊投影、下方畫面與結果等待

2026-10-04。沿[097 READY](../spec/097-fleet-projection.md)、[研究409](../re/409-fleet-projection.md)。
完整092仍DRAFT；本驗證為受控原版純規則及下方合成器，沒有正常Game或GUI聲明。

| 範圍 | 原版與remake結果 |
|---|---|
| 決定性 | 30案61份RAM雙跑逐byte相同，正規化report相同 |
| 投影 | 420分隊的+3E／40／42／44、10個鏡頭word及完整排序一致 |
| 下方場景 | 傷害前後兩個完整操作計畫由typed來源生成，逐操作一致 |
| 全圖 | 每張307200像素，合計614400，差異0 |
| 下方區域 | (8,208)..(631,431)，每張139776像素，合計279552，差異0 |
| 未閉合裁切 | 目前座標+104產生Y376，65001艘的131塊被拒絕，來源保持、沒有部分計畫 |
| 結果等待 | 11案44份RAM／色號雙跑相同，108原指令；既有Game共用predicate11案通過 |
| 標準產品套件 | 18套件：13PASS／5無測試，正常入口既有回歸通過 |

投影比較使用固定已提交state的原版入口；受控幾何注入、seed及SS:SP皆在收據。
下方畫面的兩個原版來源沿第70輪收據；上方保留原版入口，畫布640×480，沒有縮放或排除區。
PNG使用有效RGB色盤，兩側及差分均保存，傷害後圖已目視。
單元測試通過不代表整場交戰、其他語言輸入或保存通過。

## 重跑入口

```sh
bash tools/fleet_projection_oracle.sh a
bash tools/fleet_projection_oracle.sh b
bash tools/gui.sh xrun python3 tools/scripts/fleet_projection_receipt.py
bash tools/gui.sh xrun python3 tools/scripts/fleet_scene_plan_receipt.py
bash tools/go.sh test ./internal/rules -run TestFleetProjection -count=1
bash tools/fleet_wait_oracle.sh a
bash tools/fleet_wait_oracle.sh b
bash tools/gui.sh xrun python3 tools/scripts/fleet_wait_receipt.py
bash tools/gui.sh xrun go test ./internal/game -run TestEngagementWaitNative -count=1 -v
bash tools/gui.sh xrun env LOGH3_FLEET_SCENE_OUTPUT=/src/workplace/<全新目錄> go test ./internal/screen -run TestFleetScene -count=1 -v
bash tools/gui.sh test ./cmd/... ./internal/... -count=1
```

先在容器內建立全新輸出目錄；原版工具拒絕覆寫，兩個凍結器只接受完全相同的重生結果。
不要刪舊收據以通過重跑。需要新批次時使用獨立診斷命名並記錄新來源。

## 收據與輸出

| 檔案 | SHA-256 |
|---|---|
| `docs/internals/data/fleet-projection.json` | cd191199bf5297fd25d41891dc89fe4108709c6426bde78fa44ce95aa8f471f0 |
| `docs/internals/data/fleet-scene-plan.json` | 4988a5b95d2167bc14166a8fb173fb26bc0b6d4b9f6e7f070c86c9bcc3b126f4 |
| `docs/internals/data/fleet-wait.json` | 8ff102acff0607b23524e0b8678fc3eef3817aa171a6ce779d5d58c56392c566 |
| `workplace/round72-standard.log` | 8faee270dd1344006ff58ea06ff1f70a4699503391703f4bf447ff3c3cf4f9c6 |
| `workplace/round72-scene-reviewed.log` | 08080c433777d19a4161af0af92e524d9a92a0caf89fa5d8c5566ceeb81b2511 |
| `workplace/round72-wait-game.log` | ee324a4e0285831e71315c1c470508b5586019cf23c1110889a5b2153d1e65db |
| `workplace/projection48-r72-reviewed-remake/comparison.json` | dde41be690231339c1dc6855d2ae4c146a3989c162e439a75007bc1e40c67495 |

正式PNG在`workplace/projection48-r72-reviewed-remake/`，兩個場景各原版／remake／差分。
第一張原版與remakePNG相同，SHA-256 aa033dfdaa1875fdcc3d248bb15736f7ab9362f24f21cf81e7ef9e1286480d5f；
第二張相同，296ecd6d7767bf42e898e34f9f479485df4ce3a7b0510b340ed48aecf2dce058。
差分皆為零，PNG SHA-256 b63c42743ea731077221620de1aa9508a3dc1b90c401f179ce6e49df12442769。

typed投影與下方計畫沒有Game caller。098只核結果等待predicate，完整起點、裁切正式合法域與場景接線仍缺。
097／098維持READY，不由本頁升格整份092或關閉#48／#79。
