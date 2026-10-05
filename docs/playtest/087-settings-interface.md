# 087：設定入口與 BUG! 不可選

2026-10-05，第98輪，[#66](https://github.com/wicanr2/logh3/issues/66)。
範圍：出貨正常BUG!行為限定CONFORMED；MIDI I/F只有原版RE與收據，正式接線待裁決。
來源[432](../re/432-settings-interface.md)，契約[113](../spec/113-settings-bug-unavailable.md)，[完整來源索引](../internals/data/settings-interface.json)。

## 結果

| 層 | 驗證 |
|---|---|
| 原版正常入口 | GIN3.COM→帝國→138→兩次右鍵→系統指令→設定，沒有RAM注入 |
| 原版BUG! | 同父狀態／輸入雙跑，完整畫面及palette相同，除錯入口零命中，WORK230 bytes不變 |
| 正式bootstrap | Load／LoadIntro／Game.New及公開輸入，两跑各7檔相同 |
| 設定與BUG!全圖 | 各307200色號及RGB零差，設定視窗與BUG!熱區亦零差；沒有裁切、縮放或排除區 |
| 正式保存 | BUG!點擊前後公開SaveSlot生成的全部檔案逐byte相同，不宣稱原版整份存檔同態 |
| 正常GUI | cmd/gin3、60 Hz、真實X11鍵鼠；兩張各307200像素符合Game與原版，已目視原图／GUI／差分 |
| 標準測試 | `go test ./internal/... ./cmd/... -count=1`，18套件13通過／5無測試 |
| MIDI原版 | 七項開窗、無／330H選擇及右鍵取消各雙跑，screen／palette相同；兩選擇與取消完整RAM亦相同 |

比較等級為same-state，限相同設定值、底板內容與游標下的靜態畫面及不可選輸入。設定不呼叫RND，未以固定seed冒充完整玩法一致。完整世界、特殊除錯命令列與硬體MIDI偵測不在完成聲明內。
原版輸入坐標y加40對應本畫面繪製位置；Game與X11使用繪製位置。完整事件、命令、相對秒數與原圖SHA保存於索引所列收據。

## 重跑

使用[原版包裝器](../../tools/oracle.sh)，逐項讀取索引中批次的flags.txt。需要正常父快照時重走base-c完整命令，先在260000000步存狀態、到260000001停止；排程不放在主迴圈已退出的端點。

```sh
bash tools/gui.sh xrun bash -c 'LOGH3_SETTINGS_RECEIPTS=/src/workplace/settings66-repeat go test ./internal/bootstrap -run TestSettingsBugUnavailablePlayerPath -count=1 -v'
bash tools/gui.sh xrun go run tools/scripts/settings_interface_compare.go workplace/settings66-repeat workplace/settings66-repeat-comparison
bash tools/gui.sh xrun python3 tools/scripts/settings_interface_review.py --check
```

新批次必須用尚不存在的輸出目錄。GUI由[操作腳本](../../tools/scripts/settings_interface_gui.py)啟動當輪正式gin3與X11比較器；其路徑記在receipt.json，不用direct-entry或`-script`。
原版輸入唯讀，INI只驗RAM及寫入意圖，未實寫原始檔。MIDI軟體映射仍待使用者決定，#66 OPEN；061被動CPU與既有增強戰略AI契約保持。
