# 第154輪：海報模式上下裝飾

使用者希望海報模式中央遊戲上下留白加入header與footer裝飾，後續明確指定上方帝國與同盟戰機交戰、下方多艘雙方旗艦，採輕量循環動畫。前期金屬外框A／B方向已被新指示取代，不再要求選擇。

## 目前狀態表

| 項目 | 狀態 | 依據與下一步 |
|---|---|---|
| 開工 | 已核對 | 保留既有修改及v1.6.1三平台，沿CONTEXT與最新Issue快照 |
| 海報與輸入 | 已查證 | 中央1920×1440、兩側各320；工具列在中央頂端，v1名言在兩側下方，需保留 |
| 留白安全區 | 已查證並接線 | 不用粗估40／64常數；由無游標RGB背景的完整黑列取邊界，當前HD為上41及下418起，按畫面更新並限高 |
| 裝飾風格 | 已確認方向 | 上方戰機交火，下方多旗艦，輕量循環；依使用者補充繼續，不採前期A／B飾條 |
| 旗艦素材 | 已查證 | 沿flagships.json、engagement-panel.json與HD atlas，去重後16種帝國及12種同盟外形，共28種；同型不冒稱獨特新外形 |
| 戰機素材 | 已保存 | fighters.png原圖及完整提示、SHA-256、透明度與來源已保存於assets/hd/poster，程式負責動作 |
| 正式接線與驗收 | 156限定CONFORMED | 標準16 PASS／4無測試；正式AppImage兩陣營及Windows Wine共69張正常操作、三組／五語、留白與側欄零差、靜態、存讀及重啟通過 |
| 封包與預覽 | 已完成 | v.1.6.2-20261025三平台同源完整版；Linux AppImage、Windows ZIP、macOS通用ZIP，18秒無聲循環實錄及README原圖已保存 |

工作區為`workplace/poster154/`，保留前期效果稿、失敗擷取、正式來源與正常操作收據；A／B草稿未進產品，正式圖與預覽取新AppImage正常入口。

交付沿用三平台本機完整版範圍，新增`v.1.6.2-20261025`，由`tools/scripts/poster_decoration_release.py`凍結同一來源；Linux沿既有Type-2 AppImage工具鏈，不覆寫v1.6.1。

驗收入口：[正常操作與封包136](../playtest/136-poster-combat-decoration.md)、[機器收據](../internals/data/poster-decoration154.json)。

入口：[目前脈絡](../../CONTEXT.md)、[海報151](../spec/151-poster-widescreen.md)、[三組與名言152](../spec/152-poster-sets-and-quotes.md)、[正式AppImage驗收135](../playtest/135-v161-full-appimage.md)。
