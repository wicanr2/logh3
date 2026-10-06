# 107：完整版正常滑鼠操作修正

狀態：**修正與交付完成**。版本 `v.0.1.3-20261006`。契約[121](../spec/121-player-input-hotspots.md)，原因與歷史勘誤[441](../re/441-player-input-hotspot-space.md)。

重現資料在 workplace/input96-r120-old，正常片頭、帝國、防拷及 X11 實際點擊。變更／確認無作用，三種選單游標保持舊位置。此為 remake 缺陷，非原版行為對拍。

## 修正結果

- 會計、訂造及建造計畫在命中舊熱區前轉換40像素座標。畫面上的變更、確認、箭頭、參謀及結束可直接操作。
- 訂造確認沿既有關報告與月初結算路徑，建造結束及右鍵返回訂造報告。修改的七列百分比保留，正式保存再讀回通過。
- 三種下拉選單、訂造與建造畫面的游標跟隨實際位置。建造動態數字後補游標，避免數字蓋住箭頭。

原版規則、AI、音訊素材與存檔格式保持。這是remake輸入修正，未新增原版同態對拍、雙玩家命令或戰場差異驗證。

## 新版完整版

| 平台 | 本機ZIP | SHA-256 |
|---|---|---|
| Linux amd64 | [完整包](../../dist-all/v.0.1.3-20261006/full-local/LOGH3-v.0.1.3-20261006-linux-amd64.zip) | 42f7000dadad7cf736cda1abc39ea6dc65ca79a8a8e8c53871435d45e2ebc333 |
| Windows amd64 | [完整包](../../dist-all/v.0.1.3-20261006/full-local/LOGH3-v.0.1.3-20261006-windows-amd64.zip) | b3f75af3e24a2cc76c85e8483459e64b037fc3a3075aeee4d96520cca8e19e39 |
| macOS Intel／Apple Silicon | [通用完整包](../../dist-all/v.0.1.3-20261006/full-local/LOGH3-v.0.1.3-20261006-macos-universal.zip) | e92a998bbc9d56c4ccd3b1ad0ef1c657bad555d78d069a13ffe84124b7ff256b |

Linux已解壓資料夾的入口是 `full-local/LOGH3-v.0.1.3-20261006-linux-amd64/start-original.sh`；強化AI用同目錄的 `start-enhanced.sh`。不要再啟動v.0.1.2的gin3。預設存檔位置保持，舊存檔可沿既有取進度入口載入。

三平台含完整資料、22首配樂、28段音效、固定音色庫與授權。全部ZIP逐檔CRC、大小、SHA-256、manifest及來源tar核對；來源tar為建置時快照，之後補的存讀測試與註解另由索引保存。
Windows完成Wine四條入口，macOS完成雙架構、依賴、最低11.0及ARM程式頁簽章核對；Windows／macOS未真機驗收。macOS啟動與權限說明沿[106](106-macos-and-retro-promo.md)。

## 正常玩家路徑

原封包以真正X11輸入重現，沒有script、直接入口或局面注入。修正中二進位的帝國1.5倍與同盟2倍各24張全圖通過，再用新交付Linux包重跑相同兩條路徑，共48張、14745600原生邏輯像素零差。

路徑為片頭、主選單、防拷、會計變更、預算調整與返回、會計確認、訂造變更、建造兩列箭頭與結束、訂造確認、三種選單開啟、移動及左鍵選項。選項實際進入徵稅、提督一覽與設定，右鍵關閉也保留。
原始960×720與1280×960視窗PNG完整保存；比對在每個640×480邏輯像素的中心取樣，沒有遮罩或排除區域。早期統計標籤沿用「整數最近鄰」字樣，交付副本已勘誤並保留原始JSON雜湊，數字與PNG不變。帝國選單移動與同盟建造箭頭已目視確認。

標準 `bash tools/gui.sh test ./internal/... ./cmd/... -count=1` 完成18套件，13PASS／5無測試。275條worklist核對零失效，95項遠端仍4 OPEN／91 CLOSED。新Issue建立被自動審核拒絕，未作對外寫入。

第一次標準命令誤用 `./...` 納入workplace歷史prototype，已停止該容器並回復Go依賴檔。其後正式範圍測試暴露另一條舊座標測試，修正至畫面座標後同容器命令乾淨重跑通過。GUI早期失敗依序為模型漏每幀滑鼠更新、1.5倍輸入四捨五入落鄰格、2倍畫面超出Xvfb高度、音訊與GUI輸出撞名；均保留原始失敗資料，未歸為產品缺陷或改期望遮掉。

## 重建與收據

[Game預期畫面](../../tools/scripts/player_input_model.go)、[X11操作](../../tools/scripts/player_input_gui.py)、[擷取與差分](../../tools/scripts/player_input_capture.go)及[大視窗入口](../../tools/player_input_display.sh)沿既有GUI／playback Docker。執行與退出均有界，無持續背景容器。
[交付核對器](../../tools/scripts/player_input_receipt.py)集中[唯一來源](../internals/data/player-input-hotspots.json)，逐包及原生全圖核對。前兩版ZIP與影片保持，凍結來源按歷史提交66d7ad7回查，不要求舊Game來源與本輪修正相同。

本輪沿reverse-engineer-retro-game-remake技能的正常路徑與交接規範。完整包仍只留本機，不建立公開Release。推廣影片沿v.0.1.2已交付的72秒版本，本輪沒有重新製作影片。
