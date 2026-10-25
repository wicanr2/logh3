# 第166輪：新版海報三平台完整版與推廣片替換

使用者要求重新打包，補拍推廣影片並替換舊poster段落。沿既有完整版範圍，新修訂版`v.1.6.3-20261026`，舊封包與影片保留，僅本機交付。

## 目前狀態表

| 項目 | 狀態 | 下一步／契約 |
|---|---|---|
| 來源 | 已凍結 | 1655項來源、650項共用檔；165正式程式、核准銀河海報與五組五語名言；保留使用者檔案 |
| 三平台 | 已建置 | 同源Windows amd64 ZIP、Linux x86_64 AppImage、macOS Intel／Apple Silicon通用ZIP |
| 封包驗證 | PASS | 逐檔manifest、架構、素材與啟動器已驗；實際AppImage兩陣營及Wine存讀／重啟通過；獨立交付索引已產生 |
| 新拍片段 | 已通過 | 新版AppImage三組、五語、實際60秒切換、提督設定及駐地整軍出發；19張整軍截圖與公開六表存讀，無狀態／seed／時間注入 |
| 影片 | PASS | 35段126.883333秒1080p60、7613格，全部14段海報換新；18段原動畫／實戰保持；SNR37.08dB、無黑幀及字幕裁切，分鏡表及關鍵原尺寸已目視 |
| 交付 | 已安裝本機 | `dist-all/v.1.6.3-20261026/{full,promo}/`依專案契約，附manifest及SHA；原版資料／完整版／影片不入Git或公開Release |

新版海報取完整16:9畫面置入字幕安全範圍，避免字幕蓋住上下旗艦；不裁掉人物、名言或艦尾。v1片段選實際60秒切換附近，不加速計時或後製替換遊戲文字。其餘故事、主角動畫及實戰分鏡沿151，片頭及片尾標新版號。

實際AppImage兩陣營及Windows Wine共33張，連續前線整軍另19張；兩陣營六表與164各自基準一致，同盟跨平台相同。實際AppImage採extract-and-run，Linux桌面FUSE及Windows／macOS真機未驗；macOS雙架構、依賴及ARM簽章通過。#48同戰場讀回、DOS整段同態、尖端舊版逐字與新人耳驗收限制保持。

入口：[目前脈絡](../../CONTEXT.md)、[本輪驗收139](../playtest/139-poster-frame-full-promo.md)、[交付索引](../internals/data/v163-full-delivery.json)、[海報157](../spec/157-poster-galaxy-frame.md)、[五組名言驗證138](../playtest/138-poster-quote-rotation.md)、[既有故事片151](2026-10-24-story-promo-round-151.md)、[README實包截圖](../images/readme-v163.json)。
