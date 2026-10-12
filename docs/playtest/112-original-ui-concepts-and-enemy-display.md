# 第126輪：原版介面、陣營概念與敵軍顯示

日期：2026-10-11。唯一狀態表在[計畫126](../goal/2026-10-11-original-ui-ai-promo-round-126.md)，來源雜湊與核對數字在[來源索引](../internals/data/original-ui-concepts.json)。本輪沒有新封包或新版影片完成宣稱。

## 目前已驗範圍

| 項目 | 結果 | 限制 |
|---|---|---|
| 原框、原字模HD | 當幀原框重播；字模中心與四邊中點保留，透明邊界轉角固定覆蓋率；裁切與後畫遮擋測試通過 | 新陣營色未接正式renderer |
| 正常X11視窗 | 帝國、同盟各34張，完整960×720視窗；68張共47001600像素零差 | 比較正式Game產生的預期HD畫面，不是DOS同狀態對拍 |
| 敵軍顯示 | `TestEnemyVisibilityPresentationIsolation`、`TestEnemyVisibilityHoverAndSettings`、`TestEnemyVisibilityDoesNotChangeNormalTurn`通過；設定畫面實際點鈕及F7通過 | 戰場內真實X11 F7與AI比較實錄仍待驗 |
| 標準測試 | `bash tools/gui.sh test ./internal/... ./cmd/...`：19套件15PASS／4無測試 | 只證明remake回歸，不代表原版全遊戲parity |
| HTML第2版 | 四款、四畫面；1440寬三組、390寬一組，各481項通過 | HTML是外觀概念及位置試點，不是可遊玩的正式遊戲 |

## 已確認的配色

使用者先選原版配色與立體框，要求字形更接近原版，再要求帝國、同盟各兩款HTML。最新選擇是帝國B改亮藍與高雅黃金、同盟A改淡黃色與淡赤紅。

- 帝國B使用亮藍`#1c5393`、藍框`#88b4e2`、金色文字`#fff0c9`。沿B的明暗及立體層次。
- 同盟A使用淡黃`#e7d7a3`、淡赤紅選取`#d89386`、主要操作`#daa398`。亮底內文改深褐`#3c2e24`；黑底亮字與框的亮邊保留。
- 原鮮紅實際色為`RGB(227,32,32)`，已核當幀像素並置換設定清單、開關及戰略選取，沒有只改主要按鈕。
- 原版星圖、人物、按鈕矩形、字形與4:3保持。HTML預設記錄帝國B與同盟A；頁內再選只影響本頁，不修改使用者決策或遊戲。

本機入口為`workplace/hd126/FACTION-HD-CONCEPTS.html`，10,008,838 bytes，全部圖片內嵌，不需網路。第一版HTML、來源及瀏覽器結果保留為`-r1`。兩款修訂財政畫面及同盟設定畫面已目視；淡赤紅選取列與深色內文可辨。

## 重現入口

使用既有`logh3-hd-preview:20261011-r1`映像，不另建工具鏈。`tools/audio.sh`把專案及原版唯讀掛載，只讓`workplace/`可寫；Chromium使用384 PID、3 GiB及4分鐘外層逾時。

```sh
LOGH3_AUDIO_IMAGE=logh3-hd-preview:20261011-r1 LOGH3_MEM=3g LOGH3_PIDS=384 LOGH3_TIMEOUT=4m \
  bash tools/audio.sh python3 /src/tools/scripts/faction_hd_concepts.py
LOGH3_AUDIO_IMAGE=logh3-hd-preview:20261011-r1 LOGH3_MEM=3g LOGH3_PIDS=384 LOGH3_TIMEOUT=4m \
  bash tools/audio.sh python3 /src/tools/scripts/faction_hd_concept_check.py
```

生成器依四份現行介面JSON及`workplace/hd126/classic-{empire,alliance}/`的正常Game畫面製作。瀏覽器以容器內Chrome偵錯協定等待實際圖片解碼，核四款×四畫面、原色切換、試點矩形位置／尺寸、對齊線、原尺寸與Esc、分陣營選擇及手機無水平溢出。

正常X11模型來自`tools/scripts/hd_art_model.go`，重播用`tools/scripts/hd_art_gui.py`；最終輸出在`x11-empire-r2/`與`x11-alliance-r2/`。設定窗沿單一`EnemyVisibilityButtonBounds()`，沒有改六個原設定欄位。原設定視窗SHA測試只還原這塊已明列的新按鈕區，其他原像素保持原SHA。

## 擷取與環境問題

- 首批GUI使用scale3，但擷取工具只支援1至2，分類為驗證參數問題。改1.5後，兩陣營同路徑完整乾淨重跑。
- 帝國最終批次有一次擷取早於非同步重畫。原圖、差圖及145795差異保留為`-attempt-1`，同輸入等候重畫後為零差；不重送操作、不裁切游標、不放寬門檻。68張統計排除失敗attempt檔。
- Chromium的128 PID不足；本輪以384覆寫，包裝器預設仍128。虛擬時間的`--dump-dom`早於資料圖片解碼，改用容器內loopback協定等完成訊號，沒有加未鎖版套件或開外網。

## 仍待交付

新陣營色要接正式呈現並重跑正常視窗。敵軍顯示需補真實戰場開關錄影。影片需在相同可查初始條件下展示remake原版規則模式及強化模式的實際戰略／戰術AI，加入HD雙陣營世界觀片頭，再重錄正常S1讀檔及實際交火、保留遊戲聲音，最後重建新封包。

目前原版規則模式只代表remake既有政策；沒有新增DOS整場AI同態證據。第125輪包、359.2秒實錄與90秒影片保持原雜湊及歷史範圍。#99保持OPEN，134與135保持READY。文件及產物寫入均為1000:1000；容器與工作根清理結果另列來源索引及WORKLOG。
