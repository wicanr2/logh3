# 第165輪：v1五組名言輪替

使用者同意164所提的附來源網路轉錄方案，沿既有五組五語候選與已檢視預覽實作。

## 目前狀態表

| 項目 | 狀態 | 下一步／依據 |
|---|---|---|
| 來源裁決 | 已確認 | 2026-10-26「ok 同意」；尖端逐字與頁碼仍unknown，不作官方譯文聲明 |
| 規格 | 擴充限定CONFORMED | [157](../spec/157-poster-galaxy-frame.md)先READY後實作，限Linux來源程式正常入口 |
| 正式程式 | 已接線 | 原字形及姓名位置，20張2560×400文字層與獨立manifest；原九張藝術層及manifest SHA保持 |
| 驗證 | 通過 | 標準20套件16PASS／4無測試；時間邊界／負向素材、同盟36張五分鐘完整循環與25語系／組別，帝國9張乾淨重跑；兩陣營六表與164基準相同 |
| 交付 | 同一功能分支 | feat/approved-poster-frame；不更動正式版號、三平台包、tag或Release |

入口：[目前脈絡](../../CONTEXT.md)、[候選及來源](../../tools/scripts/poster_quote_candidates.json)、[164定案及藝術基準](2026-10-25-approved-poster-implementation-round-164.md)。

## 重現與交接

- [驗證138](../playtest/138-poster-quote-rotation.md)、[唯一來源索引](../internals/data/poster-quotes.json)與[正常入口第二組截圖](../images/poster-quote-rotation.png)。
- `workplace/poster165/`沿既有工作區。將凍結HTML存為`proposal.html`、已同意清單存為`quote-data.json`後，以`LOGH3_POSTER_BATCH=poster165 LOGH3_POSTER_BROWSER=poster_quote_export.mjs bash tools/scripts/poster_closeup_capture.sh`匯出。`poster_quote_install.py`只安裝20張名言及quotes.json，原九張逐項核SHA。
- 正常GUI：既有`logh3-hd-preview:20261011-r1`配`tools/gui.sh xrun`及`tools/with-xvfb.sh`，執行`poster_quote_gui.py --binary /src/workplace/poster165/gin3 --out <新收據目錄> --faction 1`，不注入時間；帝國以`--faction 0 --probe`抽樣實際60秒。`poster_quote_check.py`比對25組文字與v2／v3無字區。
- 尖端逐字／頁碼仍unknown；引用與衍生譯文來源記在清單。原版遊戲檔、完整版及影片未入Git，本輪未新建Release或三平台包。
