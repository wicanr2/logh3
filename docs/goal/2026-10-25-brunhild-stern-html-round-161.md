# 第161輪：伯倫希爾艦尾取景

使用者只要求修正伯倫希爾艦尾看似被切斷的問題，沿160的休伯利安、銀河及人物版面修正HTML。

## 目前狀態表

| 項目 | 狀態 | 依據與下一步 |
|---|---|---|
| 原因 | 已量測 | 船體最左緣約195.60，低於裁切左界217.60；四角漸層又覆蓋內側12像素 |
| 修正 | 已完成HTML並目視 | 伯倫希爾寬度100%改92%，移除原向左16像素偏移，保留往下24像素 |
| 範圍 | HTML提案 | 保留160其他內容，正式遊戲及交付保持 |
| 驗證 | 已通過 | 456305個有效alpha像素的左右輪廓落在安全區；下方以外668400像素零差 |

入口：[目前脈絡](../../CONTEXT.md)、[160提案](2026-10-25-galaxy-frame-html-round-160.md)。

## 交付與證據

- [HTML提案](../../workplace/poster161/proposal.html)、[構圖截圖](../../workplace/poster161/proposal-frame.png)、[完整頁面](../../workplace/poster161/proposal-page.png)、[手機截圖](../../workplace/poster161/proposal-mobile.png)。
- [來源](../../workplace/poster161/SOURCE.json)、[瀏覽器收據](../../workplace/poster161/BROWSER.json)、[稽核](../../workplace/poster161/AUDIT.json)。HTML為12,721,788 bytes，SHA-256 `fc2d322e99f2f314dcc9c9e3658b76f371593ee303c73aadbf47e6eb6efa2111`。
- [生成器](../../tools/scripts/poster_stern_html.py)、[瀏覽器檢查](../../tools/scripts/poster_stern_browser.mjs)，沿[既有擷取入口](../../tools/scripts/poster_closeup_capture.sh)。沒有新增或改寫圖片。

```sh
LOGH3_GUI_IMAGE=logh3-hd-preview:20261011-r1 LOGH3_XVFB_SCRIPT=/src/tools/with-xvfb.sh LOGH3_TIMEOUT=5m LOGH3_CPUS=2 tools/gui.sh xrun python3 tools/scripts/poster_stern_html.py
LOGH3_POSTER_BATCH=poster161 LOGH3_POSTER_BROWSER=poster_stern_browser.mjs bash tools/scripts/poster_closeup_capture.sh
```

範圍滑桿改為1%一步，避免92%被瀏覽器捨入成90%。瀏覽器實際解碼船體透明層，以alpha至少32的像素經目前圖片及群組矩陣計算，左右範圍為245.868至1034.244，位於含漸層及留邊的235.6至1044.4安全區內。上下仍採原有局部特寫，不將此檢查稱為整艦全部入鏡。

原有單艦、銀河、五語、大小比較、武器、本體固定、設定下載及手機無溢出抽驗通過，零JavaScript例外及外網請求。構圖已目視，左後方圓弧及艦首尖端均可見。與160相比，下方以外668400像素零差，所有內嵌素材完全相同。

Python／shell語法、收據及來源雜湊、1000:1000及整專案root-owned／誤建.md目錄檢查通過。HEAD保持44a8cae1875a9af39ff9f23e645d7b8f96eeb5ff；未commit、push、打包或發布。正式程式與舊提案保持，既有Docker映像沿用，一次性容器退出自刪。
