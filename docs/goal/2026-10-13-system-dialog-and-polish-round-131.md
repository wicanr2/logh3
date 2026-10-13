# 第131輪：系統視窗、字型與音樂、啟動進度

## 目前狀態表

本表接替[130](2026-10-12-hero-ai-release-round-130.md)的現行角色；既有發行與交付保持。

| 項目 | 目前狀態 |
|---|---|
| 系統游標 | 已修單一即時游標，正式Linux兩陣營原圖／HD正常GUI通過；141限定CONFORMED |
| 讀檔取消 | 主選單取消回主選單、戰略回戰略；戰術返回回歸通過，未補正常戰術GUI |
| 字型選擇 | 原版點陣、原版強化、思源宋體、思源黑體均可切換，預設原版強化；原圖保留原字形，選擇本次執行有效 |
| 音樂HD | 原11首MIDI資料不變，GeneralUser GS 2.0.3音色與混音；正常HD／FM／關閉通過；#69人耳及音樂性循環保持 |
| 啟動進度 | 560×220小視窗、七個實際階段與83張人物圖進度；缺檔錯誤視窗及Escape退出通過 |
| 本輪交付 | v.1.3.0-20261013三平台完整／引擎八包逐檔及SHA-256通過；Linux正常GUI、Wine抽樣及macOS結構通過，完整包只留本機 |
| 既有交付 | v.1.2.0-20261012 Release、完整版及影片保持原身份，沒有重製或重標舊推廣片；已發布tag不移動 |
| 保留修改 | 既有HD資產、decomp與本機交付保持，不覆蓋不相關修改 |

規格：[141](../spec/141-system-dialog-cursor.md)、[142](../spec/142-startup-progress.md)、[143](../spec/143-font-and-hd-music.md)。[驗證120](../playtest/120-system-dialog-polish.md)及[來源交付索引](../internals/data/system-dialog-polish.json)保存結果。診斷與試樣集中於 `workplace/modal131/`；遠端發布身份以交付索引的remote_release為準。
