# 308 — 費沙事件雙頁報告收據

## 結論

既有 [`296`](296-phezzan-event-trigger.md) 已證實 `1213:0096` 在事件本體
`1213:023B` **之前**，依人類座位呼叫 `1213:01B8` 顯示兩頁訊息；本輪以
同一個 pzx296 正常玩家路徑在兩個等待時點重錄 framebuffer，補齊框、文字與
軍事專任參謀頭像的畫面證據。帝國拿下、帝國觀看這一組的 416×140 overlay
已達局部 `same-state`。

## 原版收據

兩批都使用原版 `GIN3.COM` SHA-256
`770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918`，
dosgolem-logh3 `e0b6a03`，由新遊戲正常路徑進第一次「實行」。費沙行星與
星域 10 的明示注入完全沿用 `pzx296`；完整旗標見各批 `flags.txt`。

| 批次 | 停止步數 | 頁面 | screen.bin SHA-256 |
|---|---:|---|---|
| `phezzan-report-page1` | 223,000,000 | `雖然我方完全占領了費沙星系…` | `3447239e421f31cd34cae3cddc4ccc3e0bcf0775c36db7c52653d7e99b991a73` |
| `phezzan-report-page2b` | 227,000,000 | `同時，支持費沙自治領地…` | `ebc4850c543dae6bec77a2d3a739447808ffcdf489ea913c8d4c474c4ffcd26b` |

兩批 palette SHA-256 均為
`28b933fe81e79a380123b932302ba2dd8983bd3ac85a8f922e9e3b6ac1bd02e8`。
`1213:0096` 與 `1213:01B8` 各命中一次，`1213:023B` 尚未命中，證明畫面
確實位於事件本體之前。

## 幾何與資料

`3603:0006(buffer, 0x97, 7, 6, 0x18, 5, 1)` 對應內容格
`cell=(7,6)`、24×5、style 1、attribute `0x97`；外框加一圈後的像素區域為
`(112,160)`–`(527,299)`。本人頭像不是使用：`1213:01B8` 取觀看陣營的
`WORK +0x37` 軍事專任參謀，帝國新遊戲為提督 6，完整 80×100 FACE 畫在
`(432,180)`。

四種 owner／viewer 組合的兩頁原文仍以
[`docs/internals/57`](../internals/57-phezzan-event.md) 與
`data/phezzan-event.json` 為權威。本輪產生器只把其 Big5 原文轉成 renderer
可直接消費的 raw hex，不另建第二份文字來源。

## remake 比較

`TestPhezzanReportPagesMatchOracle` 以產品的 `FRAME.BLK`、`GRAPH.IMG`、
`FACE.AMG` 及 `PhezzanReportPanel.Draw` 重畫兩頁；各比較 overlay 58,240
像素，兩頁皆 **0 差異**。畫面外的「確認」提示不在此 overlay 矩形內，
本切片的 dismiss 契約由 `1213:01B8` 的模態等待與遊戲層左右鍵測試覆蓋，
不對該提示另作逐像素聲明。

