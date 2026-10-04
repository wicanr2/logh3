# 084：正常全滅、敗退與公開存讀檔

2026-10-05，第95輪，GitHub #79。
父提交 `1bd9dfc1dfd26ee4e2b5d575a67f2f63f0149ec4`，正式程式無修改。
入口：[研究429](../re/429-battle-lifecycle-player.md)、
[凍結來源](../internals/data/battle-lifecycle-player.json)、
[重播與凍結工具](../../tools/scripts/battle_lifecycle_player_receipt.py)。

| 核對 | 結果 |
|---|---|
| 正常入口 | 片頭、主選單、防拷138、S3同盟、戰略實行與第一戰場 |
| 固定條件 | 首次戰場實行前設定seed12345678一次；沒有局面注入 |
| 正常戰鬥 | 50次艦隊命令、19次交戰、6745項輸入，旗艦6→0，先全滅再敗退 |
| 獨立雙跑 | 正式bootstrap d／e兩次；38筆逐擊及2筆報告的typed分隊、完整六表、文字／頭像／seed相同 |
| 六表比較 | 每筆44624 bytes，40筆共1784960 bytes／次重播；兩張報告色號全圖相同 |
| 公開存讀 | 正常選單保存與載入，六檔44560 bytes與載入後保存相同；Game、bootstrap雙跑及GUI通過 |
| 真實GUI | 測試10 Hz、1032操作、64次鏡頭回授；63張19353600像素與Game計畫及X11全圖零差 |
| 報告畫面 | 實際全滅報告已目視，文字及視窗無溢出；敗退及公開存讀末端有完整收據 |

## 限制

這是remake正常玩家重播與GUI驗證，未新增原版正常流程配對。
原版單擊與回收契約保持，不能將本輪升格為原版same-state。
旗艦全滅造成艦隊回收及提督失蹤；沒有驗到提督死亡判定。
一般非旗艦分隊自然死亡／空組搬移、燃料到期、自然升級與占領GUI仍缺。
正式60 Hz長流程未新增驗收，第93輪失敗保留。
109／110 READY、062 DRAFT、#79 OPEN。

## 重跑

全部命令在既有 `tools/gui.sh xrun` 容器內執行，image `logh3-ebiten:2`、UID/GID1000、無網路。
新run須使用新的英數revision及輸出目錄，生成器拒絕覆寫。
下列列出成功批次的完整命令；重跑時將b／d及輸出a換成新名稱。

```sh
python3 tools/scripts/battle_lifecycle_player_receipt.py --prepare-probe b
export LOGH3_FLEET_CHAIN_OUTPUT=/src/workplace/lifecycle79-r95-normal-b
go test -overlay=workplace/lifecycle79-r95-b-probe-overlay.json ./internal/game \
  -run '^TestBattleLifecycleNormalProbe$' -count=1 -v
python3 tools/scripts/battle_lifecycle_player_receipt.py --prepare-replay d
go run -overlay=workplace/lifecycle79-r95-replay-d-overlay.json \
  workplace/lifecycle79-r95-replay-d.go \
  -input workplace/lifecycle79-r95-normal-b/normal-1/inputs.json \
  -out workplace/lifecycle79-r95-bootstrap-d
```

第二次獨立重播使用同一生成來源，輸出改為 `lifecycle79-r95-bootstrap-e`。
GUI準備器使用固定輸出名；保留本輪輸出後，在獨立工作副本重生。

```sh
python3 tools/scripts/battle_lifecycle_player_receipt.py --prepare-gui
go run -overlay=workplace/lifecycle79-r95-gui-overlay.json \
  workplace/lifecycle79-r95-gui-plan.go \
  -input workplace/lifecycle79-r95-normal-b/normal-1/inputs.json \
  -out workplace/lifecycle79-r95-gui-plan
go build -overlay=workplace/lifecycle79-r95-gui-overlay.json \
  -o workplace/lifecycle79-r95-gin3 ./cmd/gin3
python3 workplace/lifecycle79-r95-gui-driver.py \
  --plan workplace/lifecycle79-r95-gui-plan/plan.json \
  --out workplace/lifecycle79-r95-gui-a --scale 1 \
  --binary /src/workplace/lifecycle79-r95-gin3 \
  --input-binary /src/workplace/order79-r93-input
python3 tools/scripts/battle_lifecycle_player_receipt.py --check
```

X11輸入器由 `tools/scripts/sound_tail_input.go` 建置；擷取器由
`tools/scripts/fleet_player_gui_capture.go` 建置至 `workplace/round79-capture`。
兩者使用容器內 `go build -o <輸出檔> <來源>`，實際binary及來源雜湊均入索引。
純驗證工具未改正式產品，標準18套件13PASS／5無測試沿第93輪有效結果。
