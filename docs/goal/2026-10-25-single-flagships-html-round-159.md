# 第159輪：上下各一艘大型旗艦

使用者否決158隊列，指定上方只放休伯利安，下方只放伯倫希爾。下方依使用者提供的圖片反方向，艦首朝右，船體鋪滿下方。沿既有HTML原型修改，四角漸層、圖片內姓名與名言及輕量武器動畫保留。

## 目前狀態表

| 項目 | 狀態 | 依據與下一步 |
|---|---|---|
| 艦數與方向 | 使用者已指定 | 上方一艘休伯利安朝左；下方一艘伯倫希爾朝右，移除後景及隊列 |
| 伯倫希爾參考 | 已保存 | workplace/poster159/brunhild-reference.png，使用者提供的圖，作為白色船體與外形依據 |
| HTML | 瀏覽器抽驗通過，待使用者檢視 | workplace/poster159/proposal.html，12,051,863 bytes；上下各一艘，船體固定，只更新武器 |
| 漸層與人物文字 | 沿158 | 保留四角漸層、五語圖片內姓名及名言，名言紙本查證限制仍適用 |
| 畫面配置 | 提案完成 | 下方寬度100%，上方90%；保留88%／100%遊戲大小比較，未視為正式配置核准 |

入口：[目前脈絡](../../CONTEXT.md)、[158提案](2026-10-25-flagship-readable-html-round-158.md)。正式遊戲、舊提案及交付保持。

## 檢視入口與來源

- [HTML提案](../../workplace/poster159/proposal.html)、[構圖截圖](../../workplace/poster159/proposal-frame.png)、[完整頁面](../../workplace/poster159/proposal-page.png)、[手機截圖](../../workplace/poster159/proposal-mobile.png)。
- [SOURCE.json](../../workplace/poster159/SOURCE.json)、[BROWSER.json](../../workplace/poster159/BROWSER.json)、[AUDIT.json](../../workplace/poster159/AUDIT.json)綁定HTML SHA-256 `e1d74675fd15e6b12fcd2a6227e994989b0816b6d5f211a06ba068cf611be9c1`。
- [使用者參考圖](../../workplace/poster159/brunhild-reference.png)，SHA-256 `aa9772fdf23fd068862f11cd0d5782b6eb5291d43d848c690b031642053263e5`。
- [朝右伯倫希爾素材](../../workplace/poster159/brunhild-right.png)，2172×724 RGBA，SHA-256 `f1e4aa6724d100ced524035836879cea2ea3d22c067a6f1f3efda746c544e085`；[完整生成提示](../../workplace/poster159/brunhild-prompt.txt)。使用內建imagegen，參考使用者圖片重新呈現朝右的船體與透明背景，沒有用程式拉伸原圖。
- 休伯利安沿[158仰視圖](../../workplace/poster158/hyperion-low.png)，來源與限制沿[158記錄](2026-10-25-flagship-readable-html-round-158.md)。

生成工具原輸出為`/home/anr2/.codex/generated_images/01a16b43-4ac7-74b3-95be-186e30695ea3/exec-6279f6f7-395a-428d-b993-93fd73b6a30e.png`，已原樣複製至本輪工作區，保留原始透明層。新視角是依圖片重構的藝術提案，不稱動畫或DOS逐形證據，限本機。

## 工具與重現

[生成器](../../tools/scripts/poster_single_html.py)與[單艦繪製片段](../../tools/scripts/poster_single_js.txt)沿158已修好的SVG viewport及四角漸層，移除全部後景艦與隊列資料，改為各一艘大船體。依原比例等比放大並剛體轉向取景，允許上／下局部出框，不宣稱完整全艦均可見。下方留足艦首尖端空間，避免被接合漸層覆蓋。可調上下寬度、上下取景、遊戲顯示大小、文字位置與漸層；保留暫停及設定下載。

[瀏覽器驗證](../../tools/scripts/poster_single_browser.mjs)沿既有[有界Chromium入口](../../tools/scripts/poster_closeup_capture.sh)，使用原生CDP，不安裝新套件或新建映像。

```sh
LOGH3_GUI_IMAGE=logh3-hd-preview:20261011-r1 LOGH3_XVFB_SCRIPT=/src/tools/with-xvfb.sh LOGH3_TIMEOUT=5m LOGH3_CPUS=2 tools/gui.sh xrun python3 tools/scripts/poster_single_html.py
LOGH3_POSTER_BATCH=poster159 LOGH3_POSTER_BROWSER=poster_single_browser.mjs bash tools/scripts/poster_closeup_capture.sh
```

## 驗證與限制

Chromium核對上下各一艘、上方46BD朝左、下方45B1朝右、後景及隊列資料移除、下方主艦設定寬度跨滿範圍、四角漸層開關、五語文字在人物圖片內、4:3大小比較、武器開關、本體固定及實際JSON下載。桌面1600／手機390無水平溢出、零JavaScript例外及外網請求。構圖及朝右尖形艦首已目視確認。

Python／shell語法、素材及提案雜湊、1000:1000擁有權、整專案無root-owned或誤建.md目錄通過。正式poster.go、poster_decoration.go與158原稿雜湊保持，未跑未改動的全產品測試或打包。名言仍只有二手轉錄來源，未完成尖端舊版紙本逐字／頁碼核對。未commit、push或發布；一次性Docker容器退出自刪，既有遊戲與其他專案資源保持。
