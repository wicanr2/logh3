# 第164輪：定案proposal入庫及正式接線

使用者明確定案163 proposal並要求進入GitHub開工。沿核准構圖實作，不重新做視覺選擇。

## 目前狀態表

| 項目 | 狀態 | 依據與下一步 |
|---|---|---|
| 定版 | 已凍結 | docs/spec/157-poster-galaxy-frame.html與163原檔SHA-256完全相同 |
| 規格 | READY | 157記核准比例、三組、靜態單艦、銀河、文字、輸入逆映射及驗收 |
| GitHub | 準備提交 | 獨立功能分支先保存定版檔；不移動既有tag或Release |
| 正式程式 | 開始接線 | 匯出核准藝術層，88%viewport與輸入同源，保留實際空白及最上層UI |
| 既有修改 | 保留 | 145至163未提交內容及使用者文字檔不覆蓋；提交只選本輪及必要依賴 |

入口：[目前脈絡](../../CONTEXT.md)、[正式規格157](../spec/157-poster-galaxy-frame.md)、[定案來源](../internals/data/poster-approved.json)。
