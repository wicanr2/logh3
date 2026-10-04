# 083：S3自然勝負與正常玩家重播

2026-10-05，第94輪，沿#79及[110 READY](../spec/110-battle-finish.md)。
父提交`601476e810a8bd6bd36b5a237f0435714f3036f9`；正式程式無修改。
來源入口為[研究428](../re/428-battle-finish-natural.md)及[凍結收據](../internals/data/battle-finish-natural.json)。

| 核對 | 結果 |
|---|---|
| 正常玩家Game | 新局、防拷138、公開S3載入、同盟座位、戰略實行及第一戰場；50次艦隊命令、19次交戰 |
| 自然勝負 | 星域26、第九戰術回合自然出現一則敵方敗退報告，返回命令畫面；零占領報告 |
| 固定條件 | 進戰場後、首次實行前設定seed12345678一次；沒有位置、艘數、士氣、情報、功績或勝負注入 |
| 獨立重播 | 正式`bootstrap.Load`／`Game.New`兩次，6737項公開輸入均完成 |
| 報告狀態 | 文字、頭像、完整六表44624 bytes、seed1360112191及307200-byte畫面相同 |
| 末端保存 | 兩次重播六份正式檔案逐byte符合玩家探針；保存由診斷API執行 |
| 真實GUI | 正式cmd/gin3、開頭動畫、主選單、防拷、S3及首戰，1023個公開操作完成；額外三次鏡頭回授完整記錄 |
| GUI畫面 | 62張各307200像素，共19046400像素；正式Game與計畫色號一致，原生X11全圖零差，含自然敗退報告及戰後畫面 |
| GUI保存 | 六份診斷保存與計畫逐byte相同；未把診斷保存稱為玩家存檔選單驗收 |

## 比較等級與限制

這是remake正常玩家重播與GUI驗證。本輪沒有新增原版正常流程收據，不能標為原版`same-state`。
單人CPU沿[061](../spec/061-passive-battle-computer.md)保持停留；沒有為勝負啟用原版CPU或變更AI設定。
本次零占領報告，因此不補占領GUI的完成聲明。
自然死亡、燃料到期、功績升級與完整原版整回合同態仍依#79原判準。
110保持READY、完整062保持DRAFT、#79保持OPEN。

## 重跑入口

全部工作使用既有`logh3-ebiten:2`與`tools/gui.sh`，非root、無網路。
輸出批次已存在時拒絕覆寫；重跑探針先選新的英數revision與輸出目錄。

```sh
bash tools/gui.sh xrun python3 tools/scripts/battle_finish_victory_probe.py c
```

生成`workplace/finish79-r94-c-overlay.json`後，在同一容器內執行：

```sh
export LOGH3_FLEET_CHAIN_OUTPUT=/src/workplace/finish79-r94-normal-c
go test -overlay=workplace/finish79-r94-c-overlay.json ./internal/game \
  -run '^TestBattleFinishNormalProbe$' -count=1 -v
python3 tools/scripts/battle_finish_victory_receipt.py --prepare-replay
go run -overlay=workplace/finish79-r94-replay-overlay.json \
  workplace/finish79-r94-replay.go \
  -input workplace/finish79-r94-normal-c/normal-1/inputs.json \
  -out workplace/finish79-r94-bootstrap-a
```

第二次獨立重播使用新的`finish79-r94-bootstrap-b`。
GUI準備入口為凍結器`--prepare-gui`；生成正常bootstrap計畫及真實X11驅動器。

```sh
python3 tools/scripts/battle_finish_victory_receipt.py --prepare-gui
go run -overlay=workplace/round79-fleet-gui-overlay-finish94.json \
  workplace/finish79-r94-gui-plan.go \
  -input workplace/finish79-r94-normal-c/normal-1/inputs.json \
  -out workplace/finish79-r94-gui-plan
go build -overlay=workplace/round79-fleet-gui-overlay-finish94.json \
  -o workplace/finish79-r94-gin3 ./cmd/gin3
python3 workplace/finish79-r94-gui-driver.py \
  --plan workplace/finish79-r94-gui-plan/plan.json \
  --out workplace/finish79-r94-gui-a --scale 1 \
  --binary /src/workplace/finish79-r94-gin3 \
  --input-binary /src/workplace/order79-r93-input
```

X11輸入與擷取沿既有工具；重建來源為`tools/scripts/sound_tail_input.go`與
`tools/scripts/fleet_player_gui_capture.go`，以`go build -o <輸出檔> <來源>`建置。
擷取器預期路徑為`workplace/round79-capture`；更換路徑時同步改測試驅動器。
本輪實際binary及來源均有完整雜湊，沒有依賴主機library。

測試overlay將更新頻率設為10 Hz，為真實放開輸入留出觀測窗口。
規則、公開輸入、正式renderer與玩家入口保持；不宣稱正式60 Hz長流程時序已通過。
GUI完整命令、狀態、額外捲動、圖像與逐區統計保留於本輪來源收據。
自然敗退原生X11圖已目視，文字完整、頭像及視窗無溢出。
來源完整性由凍結器`--check`核對。
