# 第156輪：大型戰機與旗艦裁切的HTML提案

使用者否決155的小艦列陣，明確要求大型艦艇左帝國、右同盟，上方戰機也放大；允許局部裁切，只需能看見旗艦與戰機的形體。先整理成HTML供檢視，不接正式程式或打包。

## 目前狀態表

| 項目 | 狀態 | 依據與下一步 |
|---|---|---|
| 尺度與陣營 | 使用者已指定 | 大型局部特寫、左帝國／右同盟；艦尾與機翼可出框 |
| 動態 | 沿155使用者指示 | 本體固定，只讓飛彈／光束活動，不自動換批 |
| 圖源 | 已核對 | 直接取HD圖集993×396的原始艦型格，不沿192×76縮圖；戰機沿既有透明圖 |
| HTML提案 | 已完成供檢視 | 自包含單檔3593006 bytes，完整遊戲構圖、縮放／上下取景、旗艦／後景切換、武器及暫停控制、下載設定；輸出workplace/poster156/proposal.html |
| 正式程式 | 不在本輪範圍 | 使用者先要求HTML審閱，維持v1.6.2與既有原型 |

入口：[目前脈絡](../../CONTEXT.md)、[上一稿155](2026-10-25-poster-elegance-round-155.md)、[正式裝飾156規格](../spec/156-poster-combat-decoration.md)。模板及生成器為tools/scripts/poster_closeup.html與poster_closeup_html.py；HTML與原生瀏覽器擷取證據保存在workplace/poster156。

瀏覽器核對入口為tools/scripts/poster_closeup_capture.sh及poster_closeup_browser.mjs，沿既有Playwright映像內Chromium與原生CDP。離線載入、固定本體、縮放、裁切示意、後景切換、光束／飛彈／關閉、設定下載、1600桌面及390手機版已抽測；BROWSER.json綁最終HTML雜湊，proposal-page.png與proposal-frame.png供目視審閱。

最終HTML SHA-256為70e8ec9ef7e88b129e015e3e3cc2ca66a81e49c087b0a37477b69fde96dbbdcf，SOURCE.json與BROWSER.json一致，外網請求0、正式裝飾Go雜湊保持。只完成本輪要求的HTML提案；後續遊戲接線以使用者檢視結果為準。
