# 068：艦隊交戰的正常 X11 玩家路徑

2026-10-04，第79輪。沿[#48](https://github.com/wicanr2/logh3/issues/48)、
[#79](https://github.com/wicanr2/logh3/issues/79)及[102 READY](../spec/102-fleet-player-chain.md)。
正式程式父提交`838e9a125be40c411eca78d1635e52324fd1a5e4`；本輪只改驗證工具及文件。

## 比較範圍

正式`cmd/gin3`由預設片頭開始。以真實 X11 滑鼠略過片頭、選陣營，輸入`138`，
正常讀取S3、戰略實行、戰術下令、開交戰命令、經演出返回戰場。
沒有`-script`或直接呼叫Game輸入API來替代視窗操作。
兩陣營各驗一倍、兩倍，共四案。沒有新增原版收據，不稱原版正常整段same-state。
dosgolem仍為`f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c`，執行器未改。

測試overlay只觀測狀態、固定RND、擷取正式renderer及呼叫診斷SaveSlot。
不寫局面、位置、鏡頭或命令；每場實行前固定remake種子`12345678`。
完整公開輸入沿[066](066-fleet-command-display.md)正常玩家資料。
Game計畫用正式bootstrap、預設片頭及公開輸入建立，與實際視窗分開執行。

右鍵在正式app每個Update再捲一格。GUI實體按住時間與Game單次RightClick不同，
因此實際鏡頭以真實右鍵回授到計畫位置，再送後續輸入。額外手勢及相對時間全部記錄。
此方法驗等價鏡頭與正常操作結果，沒有聲稱兩側實體按住時序相同。
觀測器逐Update寫檔會增加測量成本，本輪不驗正式執行的效能或wall-clock一致性。
二倍視窗一次Draw可能含多個Update；鏡頭往返被邊界夾回時，先走離目標解除循環。
短暫自動階段若在放開之前已完成，須在每Update原始紀錄找到精確狀態才通過。
不靠低頻主機輪詢碰巧看見，也不猜測階段已發生。

## 實際畫面與保存

每案保存等待目標、彈窗及交戰返回三張圖，共十二張。
X11擷取用既有`github.com/jezek/xgb v1.1.1`及GetImage，查核visual，記錄window ID、
24-bit depth、32-bit像素排列、通道遮罩及位元組序。保留原生PNG及完整差分。
一倍視窗640×480；二倍1280×960，由獨立1600×1200 Xvfb完整容納。
原生畫面沒有裁切、遮罩或排除區。

| 抽樣 | 公開計畫動作 | 完整邏輯比較 | 完整原生比較 | 診斷保存 |
|---|---:|---|---|---|
| 帝國一倍 | 183 | 三張各0／307200 | 三張各0／307200 | 六份與Game及第77輪相同 |
| 帝國二倍 | 183 | 三張各0／307200 | 三張各0／1228800 | 六份與Game及第77輪相同 |
| 同盟一倍 | 159 | 三張各0／307200 | 三張各0／307200 | 六份與Game及第77輪相同 |
| 同盟二倍 | 159 | 三張各0／307200 | 三張各0／1228800 | 六份與Game及第77輪相同 |

邏輯差分採整數最近鄰縮回640×480。另逐點核所有原生像素，包含二倍每個2×2區塊，
與Game PNG整數最近鄰放大比較。十二張最大通道差均為0，差異bbox均為null。
GUI的307200-byte色號亦與獨立Game計畫逐位元組相同。
正常帝國仍10次下令／7擊，同盟7次下令／5擊；沒有改戰損、九回合或061被動CPU政策。
等待射程／四列摘要、彈窗及返回畫面已目視，文字與下方框線完整。

六份DT0由診斷SaveSlot保存，不能當作玩家從存檔選單操作的證據。
本輪未指定音樂目錄，不驗音效裝置或人耳。[067](067-fleet-sound.md)的技術音效驗證保持。

## 重現入口

所有命令沿既有`logh3-ebiten:2`，容器無網路、UID/GID1000、有資源與逾時限制。
需要自備原版及第77輪正常輸入；來源雜湊見下方索引。
準備工具的輸出必須尚未存在，舊收據不得覆寫。

```sh
tools/gui.sh xrun python3 tools/scripts/fleet_player_gui_prepare.py v3
tools/gui.sh xrun go build -overlay workplace/round79-fleet-gui-overlay-v3.json -o workplace/round79-gin3-v3 ./cmd/gin3
tools/gui.sh xrun go build -o workplace/round79-input-v2 tools/scripts/fleet_player_gui_input.go
tools/gui.sh xrun go build -o workplace/round79-capture tools/scripts/fleet_player_gui_capture.go
tools/gui.sh xrun go build -o workplace/round79-native-check tools/scripts/fleet_player_gui_native_check.go
tools/gui.sh xrun go run -overlay workplace/round79-fleet-gui-overlay-v3.json tools/scripts/fleet_player_gui_plan.go -input workplace/round77-chain-c/normal-0/inputs.json -out workplace/round79-plan-final-0
tools/gui.sh xrun go run -overlay workplace/round79-fleet-gui-overlay-v3.json tools/scripts/fleet_player_gui_plan.go -input workplace/round77-chain-c/normal-1/inputs.json -out workplace/round79-plan-final-1
bash tools/fleet_player_gui.sh my-gui-empire-1 0 1
bash tools/fleet_player_gui.sh my-gui-empire-2 0 2
bash tools/fleet_player_gui.sh my-gui-alliance-1 1 1
bash tools/fleet_player_gui.sh my-gui-alliance-2 1 2
tools/gui.sh test ./cmd/... ./internal/... -count=1
```

已有工具時只重跑新批次；須重建overlay時另用revision，並以驅動器`--binary`／`--plan`指定新來源。
每批保存`receipt.json`、`updates.jsonl`、`input-trace.jsonl`、正式app log、
三組renderer色號／PNG、X11原生PNG、差分、原生／邏輯統計及六份診斷保存。
Game與XTEST子程序有終止／等待處理，獨立Xvfb有EXIT trap，容器完成後刪除。

[GUI完整索引](../internals/data/fleet-player-gui.json)由
[凍結器](../../tools/scripts/fleet_player_gui_receipt.py)綁定正式父提交、Docker image ID、
overlay、執行檔、輸入、完整命令與每份產物SHA-256。
實際最終批次為`workplace/round79-gui-complete-<陣營>-<倍率>/`。
標準18套件13通過、5套件無測試。完整來源索引的雜湊另見CONTEXT及WORKLOG。
索引SHA-256為`0b1c1b79d13442420f9e9bec0e7f20b215ea72dfa262decc779f93b44ce7b6f0`，196份來源逐檔核對。

## 未完成閘門

原版七張命令完整圖各一像素未核准，仍依[066](066-fleet-command-display.md)及
[#89](https://github.com/wicanr2/logh3/issues/89)待裁決，沒有改順位或排除像素。
正常原版整段同態、聲音裝置／人耳、完整八步、macOS與正式發行驗收仍缺。
102／103／104保持READY，完整056／092／062保持DRAFT；#48／#79保持OPEN。
