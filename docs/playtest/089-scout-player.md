# 正常索敵到期、冷卻及再派出

2026-10-05，第101輪，工作項目[#79](https://github.com/wicanr2/logh3/issues/79)。
判準沿[107 READY](../spec/107-scout-fuel.md)，原版定位與呈現修正見[研究435](../re/435-scout-player-report-cursor.md)。
完整身份與數值由[凍結來源](../internals/data/scout-player.json)及[核對器](../../tools/scripts/scout_player_receipt.py)保存。

## 已驗範圍

| 層 | 結果 | 限制 |
|---|---|---|
| 正常入口 | 預設片頭、帝國新局、138、公開S3及首戰 | 首次實行前只固定seed12345678，沒有局面注入 |
| 完整實行 | 七次完整戰術回合，燃料1／2／3、失聯FF／3、冷卻2／1／0 | 061被動CPU保持，沒有直接第8步入口 |
| 報告等待 | 第四次先失聯，提督12；確認後才回收 | 表與seed於各里程碑保留，不稱原版整回合同態 |
| 再派出 | 索敵池81保持，49筆名單不增；回收Action FF，再派出Action0 | 只抽帝國S3一條自然路徑 |
| 保存 | 冷卻診斷六檔，最後公開保存／讀檔 | 五檔相同；STELLAR三byte符合既有載入派生重算 |
| 正式Frame | 新回歸修正前137／58240、修正後0 | 舊孤立renderer無法驗正式游標順序 |
| 原版參考 | 完整報告矩形58240像素與16色零差 | nearby-state，全圖6834背景差異保留，無遮罩 |
| 真實GUI | 正式60 Hz、128操作、14張全圖4300800像素零差 | 一倍視窗；3次額外鏡頭回授手勢全記錄 |
| 回歸 | 標準18套件，13 PASS、5無測試 | 不外推macOS／實體卡／聆聽驗收 |

## 重生

所有命令在專案Docker包裝器內執行。輸出名稱必須全新，禁止覆寫舊收據。

```bash
bash tools/gui.sh xrun python3 tools/scripts/scout_player_prepare.py c
bash tools/gui.sh xrun go run -overlay workplace/scout79-r101-c-overlay.json tools/scripts/scout_player_probe.go -out workplace/scout79-r101-normal-j
bash tools/gui.sh xrun go run -overlay workplace/scout79-r101-c-overlay.json tools/scripts/scout_player_probe.go -out workplace/scout79-r101-normal-k
bash tools/gui.sh xrun go build -overlay workplace/scout79-r101-c-overlay.json -o workplace/scout79-r101-gin3-fixed ./cmd/gin3
bash tools/gui.sh xrun python3 tools/scripts/scout_player_gui.py d --plan workplace/scout79-r101-normal-j/plan.json --out workplace/scout79-r101-gui-d --scale 1 --binary /src/workplace/scout79-r101-gin3-fixed --input-binary /src/workplace/order79-r93-input
bash tools/gui.sh xrun python3 tools/scripts/scout_player_receipt.py --check
```

目前j／k逐檔相同；完整公開輸入、相對Tick、六表、typed分隊、seed、未縮放PNG、
原生X11擷取與差分皆在來源索引。GUI沒有使用script旗標或直接入口，觀測器只讀資料，
只有首次實行前的seed請求會寫測試亂數。玩家保存目錄是執行設定。
工具image為logh3-ebiten:2，網路關閉、UID1000及既有資源上限保持。

原版索敵281案、原始輸入雜湊、dosgolem版本及命令沿第83輪凍結來源；本輪未重生原版。
新鮮工作樹需先依該來源入口重生原版診斷，X11事件與擷取工具的編譯入口見
[既有GUI工具](../../tools/scripts/fleet_player_gui_prepare.py)、[X11輸入器](../../tools/scripts/sound_tail_input.go)及[擷取器](../../tools/scripts/fleet_player_gui_capture.go)。

## 尚未通過

正常原版完整整回合同態、其他自然死亡／空組搬移、自然升級、占領GUI仍缺。
本輪只補索敵路徑的正式60 Hz驗收，不宣稱已補全部戰術長流程。
107保持READY，完整062 DRAFT，#79保持OPEN。四項核准取捨及待答MIDI／一像素裁決不變。
