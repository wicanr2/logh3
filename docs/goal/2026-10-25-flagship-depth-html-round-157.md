# 第157輪：上下旗艦景深隊列與人物內名言

使用者要求移除戰鬥機，上方同盟旗艦並列朝左，下方帝國旗艦並列朝右，以隊列由深到淺呈現。另指定兩側名言、楊威利與萊因哈特的名字全部移到人物圖片內。沿本輪既有HTML提案修改範圍，不接正式遊戲。

## 目前狀態表

| 項目 | 狀態 | 依據與下一步 |
|---|---|---|
| 陣營與方向 | 使用者已指定 | 上同盟朝左，下帝國朝右，無戰機 |
| 景深 | 使用者已澄清 | 以遠近大小、錯位與遮擋排列，艦體固定；不把顏色深淺當主需求 |
| 人物文字 | 使用者已指定 | 名言及姓名一起內嵌人物圖下緣，人物海報延伸到底 |
| HTML | 已完成供檢視 | 自包含檔workplace/poster157/proposal.html；上下各6艘，大小／景深／間距／取景、人物文字位置及五種海報語系可調 |
| 正式程式 | 不在本輪範圍 | v1.6.2與156舊HTML保留 |

入口：[目前脈絡](../../CONTEXT.md)、[上一份HTML156](2026-10-25-poster-closeup-html-round-156.md)。本輪模板、生成器與瀏覽器核對入口為tools/scripts/poster_flagship_rows.html、poster_flagship_rows.py及poster_flagship_rows_browser.mjs；擷取仍沿poster_closeup_capture.sh的既有Docker／Chromium入口。

最終HTML為13061896 bytes，SHA-256 4fadd1f487088c81edbdb26833029911e6cbe390125b4eab8b5cc3d05cb57bc1。SOURCE.json及BROWSER.json一致；上同盟朝左／下帝國朝右、遠近大小順序、本體固定、五語名言姓名在圖片內、控制項及設定下載通過，外網請求0。proposal-frame.png、proposal-page.png與proposal-mobile.png在同工作區；正式poster.go與poster_decoration.go雜湊保持。
