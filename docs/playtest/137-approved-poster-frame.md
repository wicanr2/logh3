# 定案銀河海報正式接線

使用者核准的163 HTML已保存於[規格157](../spec/157-poster-galaxy-frame.md)。正式程式使用其匯出藝術層，中央遊戲縮至88%，與滑鼠逆映射共用同一變換。上方休伯利安朝左，下方伯倫希爾朝右，船體固定，只循環短光束。

這是新增的remake美術，沒有DOS對應畫面。本紀錄驗證Linux來源程式、核准提案與操作，不宣稱原版同態、動畫船型逐形精確或三平台真機通過。

## 正常遊玩驗證

兩陣營各從主選單、正常輸入138保護答案進入戰略畫面，沒有direct-entry、狀態注入或固定種子。經設定頁啟用海報，完成三組切換、五語、工具列、F1說明、F7原版／HD、動畫開關、公開六表保存、讀回及重啟。主選單空白判斷不成立時不畫艦艇。

- 帝國與同盟各23張完整1280×720 X11視窗，完整逐步觀測與六檔雜湊留在`workplace/poster164/{empire,alliance}/RESULT.json`。
- 各觀測18秒以上，船體與背景固定；短光束相位更新時，整張HD背景上傳次數保持。關閉動畫時相位固定0。
- 最後修正小數縮放邊界的一像素切線後，以相同入口重跑三組海報及實際滑鼠設定切換。記錄位於`workplace/poster164/final-seams/RESULT.json`。
- 標準命令為`tools/gui.sh test ./cmd/... ./internal/...`，20產品套件，16通過、4無測試；另有定案素材完整載入、竄改拒絕、按鈕命中及空白區保護測試。

完整狀態、二進位與輸出雜湊見[唯一來源](../internals/data/poster-approved.json)；[實際v1](../images/poster-runtime-v1.png)、[v2](../images/poster-runtime-v2.png)、[v3](../images/poster-runtime-v3.png)。三張均為來源程式正常入口的視窗截圖，沒有裁切或重畫。

## 核准圖比較

輸出與核准HTML同為2560×1440。實際視窗是1280×720，確認Ebitengine螢幕縮放採2:1最近鄰的奇數像素取樣，再比較畫面。中央即時遊戲內容與HTML靜態圖不是同狀態，明確排除；短光束及實際游標另列確切矩形。

| 組別 | 不透明藝術區像素 | 不同像素 | 透明合成最大通道差 | 小數遊戲邊緣最大差 |
|---|---:|---:|---:|---:|
| v1 | 343632 | 0 | 3 | 0 |
| v2 | 399442 | 0 | 2 | 1 |
| v3 | 399442 | 0 | 2 | 1 |

透明文字漸層與艦艇抗鋸齒以Chromium和GPU的整數捨入分開統計；不把差異稱為逐像素相同。兩側人物、腳邊與銀河的不透明區均符合核准圖。程式與比較器均保留畫面差分，不以模糊濾鏡掩蓋切線。

重現藝術層：在`workplace/poster164/`放入核准HTML為`proposal.html`，以`LOGH3_POSTER_BATCH=poster164 LOGH3_POSTER_BROWSER=poster_approved_export.mjs bash tools/scripts/poster_closeup_capture.sh`匯出。工具使用既有鎖版Chromium1194容器；無外網請求，九張圖的尺寸、來源與雜湊由正式manifest核對。正常GUI工具是`tools/scripts/poster_approved_gui.py`，藝術區比較是`tools/scripts/poster_approved_check.py`，均只在專案Docker工具鏈執行。

## 範圍限制

本次未定新版號、建立Release或重新打包。既有v1.6.2完整包仍是舊裝飾。Windows與macOS尚未重建或真機驗證這次海報。使用者後續提出五組名言輪替，來源標準等待答覆，草案不作正式功能完成聲明。
