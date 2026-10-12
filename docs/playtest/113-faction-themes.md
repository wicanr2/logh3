# 113：正式陣營介面與戰略星圖主題

2026-10-11。使用者核准帝國B亮藍／黃金及同盟A淡黃／淡赤紅，追加戰略地圖配合陣營主題。本次已接正式遊戲呈現，保留原版字模、立體框、視窗、按鈕底緣與操作流程。範圍為remake顯示驗證，沒有新增DOS原版same-state收據。

## 結果與入口

| 項目 | 結果 |
|---|---|
| 正式介面 | 帝國亮藍面板及金色文字；同盟淡黃面板、深褐內文及淡赤紅主要指令／選取 |
| 戰略星圖 | 帝國藍色星域與金色航路；同盟淡黃星域與柔和赤紅航路。16種狀態色分開，星系位置、徽章、航路幾何、敵我標記及選取提示保持 |
| 正常實際視窗 | 兩陣營各34張，共68張960×720完整視窗，47001600像素零差。由片頭、主選單、防拷正常操作，不用腳本入口或狀態注入 |
| 核准概念比較 | 四畫面×兩陣營，共13302720個UI像素。7張零差；同盟命令畫面有7個已分類亮邊像素，未分類差異為零。星圖的新主題不列入舊概念UI比較 |
| 狀態與保存 | 正常新局／載入及雙玩家交棒依當前陣營切換；原索引畫面、色盤、原色重播與六個DT0保存檔保持 |
| 回歸 | 標準19套件：15 PASS／4無測試；兩主題快取隔離、16狀態色、未知主題回退、覆蓋順序及資料隔離通過 |
| 離線預覽 | `workplace/hd126/FACTION-HD-GAME.html`，3828004 bytes。兩陣營×財政／建造／星圖／設定，1440與390寬各34項瀏覽器檢查通過 |

正式來源在[陣營主題資料](../../assets/hd/interface-themes.json)，核准原PNG保存於同目錄，載入時核對雜湊。原HTML不覆寫。唯一凍結來源在[本輪索引](../internals/data/faction-themes.json)，包含程式、素材設定、模型PNG、真實視窗PNG、完整差圖、輸入及測試日誌。第112份來源索引保持不變。

## 呈現邊界

`Game.Frame`只附當前陣營識別，主選單、片頭、防拷及結局保留中性。HD呈現依實際畫序維護UI範圍及按鈕角色，只置換已核原色。人物、戰術地圖、戰鬥演出及敵我標記沿既有畫法；索引像素、AI、規則、亂數與保存格式不變。未知主題或缺主題檔退回原框原色。

星圖使用132的320×300底圖、32區域遮罩及當幀色號。主題加入快取鍵，帝國→同盟→帝國完整還原，不借用上一方的圖。徽章區及動態敵我標記保留原色，星系、英文標籤及航路座標不移動。

概念頁以HD圖片13像素鄰域決定深色內文，正式呈現以原邏輯填色決定。HD游標與原索引游標的輪廓不同，在同盟命令畫面的游標交接處留下7個亮邊像素差異，HD範圍`(1371,1140,3,4)`。報告保存全部差異及色對，沒有遮罩；不宣稱核准概念與正式版全圖零差。

## 重現

建置及測試沿`logh3-ebiten:2`，GUI與瀏覽器沿`logh3-hd-preview:20261011-r1`，精確映像ID見來源索引。原版bundle SHA-256為`de905aa0b5a666aa9d31dce36f43a5584f1842950c39d7e1b1f97c72e0d79673`。Git HEAD為`523ced1346358b1d3a9982148c569a69af953a5f`，工作樹包含本次及先前未提交修改；程式身份以來源逐檔SHA為準。

```sh
bash tools/gui.sh xrun bash -c 'go run tools/scripts/hd_art_model.go -out workplace/hd126/theme-empire-r4 -faction 0'
bash tools/gui.sh xrun bash -c 'go run tools/scripts/hd_art_model.go -out workplace/hd126/theme-alliance-r4 -faction 1'
bash tools/gui.sh xrun bash -c 'go build -o workplace/hd126/gin3-faction-r4 ./cmd/gin3'
bash tools/gui.sh test ./internal/... ./cmd/...
```

重跑時用新的revision及輸出目錄，不覆寫本次收據。正常X11重播以帝國為例，同盟替換兩個陣營目錄名：

```sh
LOGH3_AUDIO_IMAGE=logh3-hd-preview:20261011-r1 LOGH3_MEM=3g LOGH3_PIDS=384 LOGH3_TIMEOUT=8m \
  bash tools/audio.sh bash /src/tools/s1_record_display.sh python3 /src/tools/scripts/hd_art_gui.py \
  --binary /work/hd126/gin3-faction-r4 --assets /work/package92-v.1.1.0-20261011-hd/common \
  --out /work/hd126/theme-x11-empire-r4 --plan /work/hd126/theme-empire-r4 \
  --capture /work/hd125/capture --art /src/assets/hd --scale 1.5
LOGH3_AUDIO_IMAGE=logh3-hd-preview:20261011-r1 LOGH3_MEM=3g LOGH3_PIDS=384 LOGH3_TIMEOUT=2m \
  bash tools/audio.sh python3 /src/tools/scripts/faction_hd_check.py --revision r4 --receipts
```

後者核對已核准HTML及全部UI差異，並核68個最終視窗、15套件通過與瀏覽器結果，產生本機來源索引。瀏覽器沿既有核對器的Chrome偵錯協定類別，只連容器內loopback，實際等待圖片解碼；四畫面及兩個陣營按鈕皆有切換回饋，手機無水平溢出。

`TestHDFactionThemeNormalLoadAndSaveIsolation`先正常關閉月初報告再保存，驗完整六表、色盤及索引重播不變，再從S1公開讀檔切換另一陣營。初次測試在報告未關閉時保存，被既有閘門正確拒絕；更正測試輸入後通過，未放寬產品保存限制。`TestHDFactionThemeVersusHandover`驗正常雙玩家交棒。

## 擷取問題與限制

最終帝國`lists-moved`與同盟`strategy-moved`各有一次取圖早於游標重畫，分別468及472像素不同。原圖、差圖與數字保存為`attempt-1`；同一輸入等待完成後為零差，沒有重送輸入、裁切游標或放寬門檻。68張統計只計最終收據。

新HD世界觀片頭、戰略／戰術AI比較、戰場真實F7展示、新S1交火影片及新封包仍未完成。#99保持OPEN。前版`v.1.1.0-20261011`包與影片保持原雜湊。所有新預覽與核准圖片含原版衍生內容，沿本機及目前private儲存庫範圍，未新增tag或公開Release。

本次使用者授權內的介面與星圖工作完成；不外推Windows／macOS真機、人耳、DOS全場AI同態或全遊戲parity。清理及擁有權結果見WORKLOG與`workplace/hd126/FACTION-CLEANUP.json`。
