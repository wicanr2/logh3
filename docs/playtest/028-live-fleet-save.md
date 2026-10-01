# 028 — 現況艦隊存進度／取進度

規格：[067](../spec/067-live-fleet-save.md)。調查：[355](../re/355-live-fleet-save.md)。
工作項目：[Issue #86](https://github.com/wicanr2/logh3/issues/86)。
來源／輸出 SHA-256、工具版本、原始定位及分級索引：
[`live-fleet-save.json`](../internals/data/live-fleet-save.json)。

## 正常 GUI

重生：`bash tools/live_fleet_save_check.sh`。全程使用 `tools/gui.sh` 的
`logh3-ebiten:2` 容器、正式 `cmd/gin3` 入口與 renderer；無狀態注入或診斷入口。
輸入：主選單 → 防拷 → 關月初報告 → 出擊編隊 → 實行 → 存進度第 0 格 →
取進度第 0 格 → 存進度第 1 格。初始種子 `0xD3DC167E`，執行前固定，未重擲。

輸出目錄 `workplace/playtest/save86/`：`script.txt` 保存完整序列，`state.txt`
確認停在 `StateSaveGame`，`gui.png` 是實際 640×480 顯示層擷取；
`receipt.json` 記錄命令、HEAD、工作樹、來源／輸出雜湊與判準。

結果：帝國 1 支／同盟 3 支艦隊，兩次存檔的 FLEET 3904／SHIP 27456 bytes 全表零差異；
32 星域 × 兩方的艦隊計數與完整行星表全部保留。這是 remake 正常玩家存讀回歸，
不是原版與 remake 所有 FLEET／SHIP bytes 的同狀態對拍聲明。

## 原版欄位與內部續玩

`TestSyncFleetsPreservesOriginalTables`：出貨第 8／9／A 格，未操作的兩表全 bytes 不變；
改艘數時保留既有池索引與未知欄位。
`TestSyncFleetsSpeedsMatchOriginal`：以 dosgolem 綁定版本的 `runtime85-normal/command.bin`
抽取四支原版艦隊，保留其非零位置、命令與編制欄位作獨立資料輸入；物化後解析全部建模欄位不變，
15 組編制＋4 組索敵隊的三項速度均對上原版。此部分限指定欄位，並非正常玩家 GUI 的替代品。

`TestSaveLiveFleetStateLoop`：正常編隊／回合／系統存讀，保存的所有編制欄位與池身份讀回相同。
下一回合對照不經存讀的控制組，執行前兩側再設相同受控種子，玩法欄位、星域與行星整表一致。
控制組較晚才物化原始表，因此該續回合比較不要求初配池數字相同；第一次存讀仍嚴格比較池身份。
正式存檔沒有新增 RNG 欄位，正式遊戲不永久固定測試種子。

池滿、越界與重複引用會拒絕，兩表不變；索引 `00FF` 的 255 號槽可由解析、部署與計分讀到；
縮編後不部署後方舊槽。`tools/gui.sh test ./internal/... ./cmd/...` 最後乾淨重跑全部通過。

最初測試用載入列 y=122 點保存畫面，而保存首列為 y=142，屬測試熱區錯誤；
下一回合把較晚物化的控制組池數字也視為相同，屬比較條件錯誤。
兩者已按正式 UI 與初配身份的實際限制修正，沒有調整產品規則、原版黃金值或增加畫面排除區。
