# 057：交戰量表驗證

2026-10-04，第68輪；[研究405](../re/405-engagement-gauges.md)、
[READY 094](../spec/094-engagement-gauges.md)、[原版凍結資料](../internals/data/engagement-gauges.json)。
原版及工具身份、輸入／注入、地址基準與每份輸出SHA-256均在405及JSON。

## 比較範圍

局部原版常式相同參數的色號／幾何對拍，不是整場same-state。
雙側各16案，32×11×66＝23,232像素，相同23,232、不同0、最大通道差0、差異bbox空。
原版與remake局部PNG用同一份原版色盤呈現，沒有裁切隱藏量表內差異。
裁切區域由screen.EngagementGauge.Rect與原版凍結座標相互核對；矩形外寫入亦驗零。

原版雙跑：`bash tools/engagement_gauges_oracle.sh a`／`b`。
凍結：`bash tools/py.sh /work/tools/scripts/engagement_gauges_receipt.py`，重生SHA-256不變。

輸出目錄先確認1000:1000再建立，指定一次性空目錄執行：

```sh
bash tools/gui.sh xrun env \
  LOGH3_GAUGE_OUTPUT=/src/workplace/gauges48-r68-remake \
  LOGH3_GAUGE_GAME_OUTPUT=/src/workplace/gauges48-r68-game \
  go test ./internal/screen ./internal/rules ./internal/game \
  -run 'TestEngagementGauges|TestPlanetDefenceGauges' -count=1 -v
```

| 本機輸出 | SHA-256 |
|---|---|
| workplace/round68-gauge-artifacts.log | 2365670351b4b1b7f3c7605c49d791b20275e2ea5c8fd0c0772f59391673dfc2 |
| workplace/gauges48-r68-remake/comparison.json | 98756b078305f190f6f4d6a4d73a8983bbd637c20312f0a98bdafdbc35dfcf30 |
| workplace/gauges48-r68-game/faction1-aftertrue.png | a67aba68794b8537a43ed9ca14b0620f4fbc6f17c885344fb4dada849b139665 |

## 正式接線抽樣

規則assaultOne在反擊前後保存完整度及攻擊力；與既有反擊直接呼叫的戰損和seed相同。
Game.queueEngagements接typed Hit，保留獨立快照；正式drawEngagement畫雙側前後量表，
結果tick才切換，動畫關閉立即用後值。四張640×480診斷PNG已保存，包含正常renderer及COMANI。
這些圖使用診斷隊列，無正常玩家輸入到達聲明；Game色盤與原版不同，未宣稱正式RGB通過。
下方方塊仍缺，其他標題的分段生命週期待056／092閉合；#48與#79保持OPEN。

目標32例、非法輸入、規則快照與雙側Game抽樣皆通過且零skip。
標準命令 `bash tools/gui.sh test ./cmd/... ./internal/... -count=1`：18套件13PASS／5無測試。
workplace/round68-standard-tests.log SHA-256 `317013f4f540e824663237f8bad7f0d93fd8500271a5db60bf20de6b00313587`。
新檔1000:1000；專案容器清理及工作樹狀態記WORKLOG，不由單元測試升格整場CONFORMED。
