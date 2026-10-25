# 第164輪：定案proposal入庫及正式接線

使用者明確定案163 proposal並要求進入GitHub開工。沿核准構圖實作，不重新做視覺選擇。

## 目前狀態表

| 項目 | 狀態 | 依據與下一步 |
|---|---|---|
| 定版 | 已凍結 | docs/spec/157-poster-galaxy-frame.html與163原檔SHA-256完全相同 |
| 規格 | 限定CONFORMED | 157先READY後實作；來源Linux正常入口及核准藝術層已驗，非DOS parity |
| GitHub | 提案與實作已推送 | feat/approved-poster-frame，93d362e4原樣保存核准HTML、324cb899接入正式程式；遠端分支SHA已核對，main仍44a8cae1，不移動tag／Release |
| 正式程式 | 已接線及驗證 | 九張核准藝術層，88%viewport與輸入同源，偵測實際空白，游標／工具列／說明在最上層 |
| 正常GUI | 已抽驗 | 帝國與同盟各23張；三組／五語／F1／F7／動畫／保存讀回及重啟，18秒以上相位更新不增加背景上傳 |
| 最後切線 | 已修正 | 空白補畫納入小數邊緣像素，三組正常入口再跑5張；不透明藝術區零差，透明合成最大通道差3，小數遊戲邊界最大1 |
| 測試 | 通過 | 20產品套件16PASS／4無測試；素材manifest、SHA拒絕、輸入命中與空白保護測試另通過 |
| 新增名言輪替 | DRAFT待來源裁決 | 使用者要求五組兩人名言、約60秒切換；已整理五語與來源及可操作HTML，尖端逐字／頁碼unknown，等待是否接受網路轉錄 |
| 既有修改 | 保留 | 145至163未提交內容及使用者文字檔不覆蓋；提交只選本輪及必要依賴 |

入口：[目前脈絡](../../CONTEXT.md)、[正式規格157](../spec/157-poster-galaxy-frame.md)、[定案來源](../internals/data/poster-approved.json)。

## 可重現入口與限制

- 正式實作：`cmd/gin3/poster_approved.go`，唯讀載入`assets/hd/poster/approved/manifest.json`核尺寸與雜湊；缺新素材退回既有海報／4:3及其相同命中變換。
- 藝術匯出：`tools/scripts/poster_approved_export.mjs`，鎖版Chromium1194，來源為核准HTML；正常GUI工具為`tools/scripts/poster_approved_gui.py`，差分工具為`tools/scripts/poster_approved_check.py`。
- 證據：`workplace/poster164/{empire,alliance,final-seams}/RESULT.json`及`ART-CHECK.json`；公開截圖與摘要位於[驗證137](../playtest/137-approved-poster-frame.md)及唯一來源JSON，未公開遊戲檔或完整版。
- 名言草案：[五組候選及來源](../../tools/scripts/poster_quote_candidates.json)、[預覽生成器](../../tools/scripts/poster_quote_preview.py)。生成`workplace/poster164/quotes-preview.html`，五語、五組與實際60秒計時在`poster_quote_preview_browser.mjs`檢查，未進production path。
- Windows／macOS這次未重建或真機驗收，版號與dist-all保持；本機v1.6.2完整包仍是歷史戰機／多艦裝飾。
- 收尾核對：README、157、137及164所有相對連結均在Git索引；已用README到157作正對照。整專案find無root-owned檔案或誤建.md目錄，輸出均1000:1000；本輪Go／HD／Chromium容器均退出自刪，沒有新增工具映像。
- GitHub連接器建立draft PR回403 `Resource not accessible by integration`。分支提交與推送成功且已核遠端，改提供[分支](https://github.com/wicanr2/logh3/tree/feat/approved-poster-frame)及[比較頁](https://github.com/wicanr2/logh3/compare/main...feat/approved-poster-frame)，未轉移主機憑證或修改連接器權限。
