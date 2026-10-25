# 第162輪：人物腳邊的銀河切面

使用者指出萊因哈特與楊威利腳邊，兩側人物海報與下方銀河之間有明顯直線切面。只修銜接，沿161構圖。

## 目前狀態表

| 項目 | 狀態 | 依據與下一步 |
|---|---|---|
| 原因 | 已定位 | 舊漸層覆蓋範圍只在上下飾框矩形，底框上緣留下水平切面，人物原圖內側又是硬裁切 |
| 修正 | HTML完成並目視 | 改成全高海報內緣透明漸層，取消矩形覆蓋，以共同銀河作底 |
| 範圍 | HTML提案 | 保留船體大小、位置、人物與文字，漸層預設28像素以避開腳部主要輪廓 |
| 驗證 | Chromium通過 | 兩個全高720像素遮罩，沒有矩形覆蓋；中央363268像素零差、船體幾何及內嵌素材相同 |

入口：[目前脈絡](../../CONTEXT.md)、[161提案](2026-10-25-brunhild-stern-html-round-161.md)。

## 檢視與重現

- [HTML提案](../../workplace/poster162/proposal.html)、[腳邊近照](../../workplace/poster162/proposal-footer.png)、[完整構圖](../../workplace/poster162/proposal-frame.png)、[完整頁面](../../workplace/poster162/proposal-page.png)、[手機截圖](../../workplace/poster162/proposal-mobile.png)。
- [來源](../../workplace/poster162/SOURCE.json)、[瀏覽器收據](../../workplace/poster162/BROWSER.json)、[稽核](../../workplace/poster162/AUDIT.json)。HTML為12,722,427 bytes，SHA-256 `ac72a7cbc1751b97aaf80a4897f284e3f361ebefb2c629dd256cb589418857b7`。
- [生成器](../../tools/scripts/poster_seam_html.py)、[瀏覽器檢查](../../tools/scripts/poster_seam_browser.mjs)，沿[既有擷取入口](../../tools/scripts/poster_closeup_capture.sh)。

```sh
LOGH3_GUI_IMAGE=logh3-hd-preview:20261011-r1 LOGH3_XVFB_SCRIPT=/src/tools/with-xvfb.sh LOGH3_TIMEOUT=5m LOGH3_CPUS=2 tools/gui.sh xrun python3 tools/scripts/poster_seam_html.py
LOGH3_POSTER_BATCH=poster162 LOGH3_POSTER_BROWSER=poster_seam_browser.mjs bash tools/scripts/poster_closeup_capture.sh
```

舊四角矩形覆蓋在底框上緣突然開始，會把人物原圖及星空蓋出水平切面。現在沒有覆蓋矩形，改在人物圖片群組內側套720像素全高遮罩；漸層與Y無關，不再於底框上緣突然切換。預設左側189.6至217.6、右側1062.4至1090.4透明過渡，文字在獨立上層，原人物PNG與船體均未改寫。滑桿改2像素一步以保留預設28，仍可關閉。

Chromium核對全高遮罩、兩側漸層位置、移除矩形覆蓋及開關、實際艦首艦尾安全範圍、銀河來源、單艦方向、本體固定、五語圖內文字、控制與實際設定下載；桌面1600／手機390無水平溢出，零JavaScript例外與外網請求。完整構圖及腳邊近照已目視。

稽核確認所有內嵌素材與船體幾何均與161相同，中央363268像素零差。Python／shell語法、雜湊、1000:1000與整專案root-owned／誤建.md目錄檢查通過。正式程式、README、舊稿及交付保持，未commit、push、打包或發布；既有Docker映像沿用，一次性容器退出自刪。
