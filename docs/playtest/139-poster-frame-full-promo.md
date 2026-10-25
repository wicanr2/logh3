# v.1.6.3：定案銀河海報完整版與故事推廣片

使用者要求三平台重新打包，補拍並替換原推廣片的海報片段。本輪採`v.1.6.3-20261026`，沿已驗165程式與核准藝術層，不改遊戲規則。完整版與影片含原版及衍生音畫，只留本機。公開引擎Release與舊封包保持。

唯一進度為[166狀態表](../goal/2026-10-26-poster-full-promo-round-166.md)。海報來源及正常路徑見[137](137-approved-poster-frame.md)，五組五語名言來源及五分鐘完整輪替見[138](138-poster-quote-rotation.md)。尖端舊版逐字及頁碼仍為unknown，本輪不新增此項聲明。

## 三平台封包

同一凍結來源1655項，其中非vendor677項，共用原版資料、介面、音畫及字型650項。每包附完整manifest；包內不帶玩家preferences或存檔。`INPUTS.json`記來源、共用檔與產生器SHA，`BUILD.json`及[獨立交付索引](../internals/data/v163-full-delivery.json)記正式產物與驗證收據。

| 平台 | 本機`dist-all/v.1.6.3-20261026/full/`檔名 | bytes | SHA-256 |
|---|---|---:|---|
| Windows amd64 | `LOGH3-v.1.6.3-20261026-windows-amd64.zip` | 473815288 | `21c7e6c8f5a4ba1fffe5b18d6a82c443d2d23a76d5fcd8375fca28c5cc8b925c` |
| Linux x86_64 | `LOGH3-v.1.6.3-20261026-linux-x86_64.AppImage` | 473745912 | `96f519353fe1e94c405fd56dfd0a91a0da48f481e2abd5b59924bf917b5f1e1d` |
| macOS通用 | `LOGH3-v.1.6.3-20261026-macos-universal.zip` | 483767233 | `01382e106821989af7be0204345e6c92553254bfbcb7f3e30157eeca145ca11d` |

Windows與macOS從實際ZIP逐檔驗CRC、名稱、內容及權限；UTF-8檔名、README BOM與Windows CRLF另驗。Windows為PE64 amd64。macOS為Intel／Apple Silicon雙架構，最低11.0、SDK15.5，依賴與ARM ad-hoc簽章頁雜湊通過；未作macOS真機啟動。

Linux使用既有Type-2 runtime，SHA`1cc49bcf1e2ccd593c379adb17c9f85a36d619088296504de95b1d06215aebbf`，前綴944632 bytes。實際AppImage解開790項檔案與權限，逐項等同AppDir；ELF64 x86_64與15項相依通過。實測採`--appimage-extract-and-run`，素材自動尋址、預設XDG存檔，不傳`-orig/-data/-art/-save`。尚未驗桌面FUSE啟動。

## 正常玩家操作

標準命令`tools/gui.sh test -p 2 ./cmd/... ./internal/...`完成20產品套件，16PASS、4無測試。以下均從正常主選單與138答案進入，不注入遊戲狀態、seed或時間；trace只觀測。

| 收據 | 抽樣 | 結果 |
|---|---|---|
| `workplace/poster166/player-linux/RESULT.json` | 實際AppImage同盟15張，三組、五語、實際60秒切換、編輯器、F1／F7、存讀、重啟 | PASS |
| `workplace/poster166/player-empire/RESULT.json` | 實際AppImage帝國9張，60秒切換、v2／v3無名言、F1／F7、存讀、重啟 | PASS |
| `workplace/music69-windows/full166/player-windows-r2/RESULT.json` | 實際Windows ZIP解出後Wine9張，60秒切換、F1／F7、存讀、重啟 | PASS |
| `workplace/poster166/garrison-linux-r3/RESULT.json` | AppImage正常96步計畫、19張，前線駐留、更換提督、當地領船、E／T拒絕、C旗艦、出擊、六表存讀與重啟 | PASS |
| `workplace/poster166/diagnostic-linux/RESULT.json` | 實際AppImage開場、原版AI、強化AI三啟動，版本、素材與預設存檔路徑 | PASS |
| `workplace/music69-windows/full166/diagnostic-windows/RESULT.json` | 實際ZIP直接啟動與兩啟動器三模式，版本、HD曲庫及MT-32尋址 | PASS |

兩陣營六表各與164基準相同，同盟Linux與Windows逐表相同。前線19的庫存由`9000,3000,16000,2000,12000,3000,1000`減為`6000,3000,12000,2000,9000,3000,1000`，母星31庫存保持。新編第3艦隊提督41、巡洋艦C旗艦，正常出發至18；讀回後艦隊與六表一致。E／T拒絕不以孤立規則測試代替正常編成操作。

[README截圖索引](../images/readme-v163.json)保存11張實際AppImage原視窗、來源事件與SHA；三組海報、五語、名言第二組、設定及編輯器已掛README。完整1280×720，不重畫、不縮放。

## 推廣片

新片位於`dist-all/v.1.6.3-20261026/promo/LOGH3-v.1.6.3-20261026-promo.mp4`。沿[151已驗分鏡](../goal/2026-10-24-story-promo-round-151.md)，保留帝國與同盟國徽底圖、百年對峙、兩位主角動畫、亞姆力札危局、兩軍戰場、完整要塞砲及光線／飛彈／戰鬥機演出。

35段中，14段海報、語系、編輯與前線操作全部換成正式AppImage新錄；18段既有動畫與實戰中間片逐檔SHA保持，3張故事卡更新版號。沒有把舊片冒稱全片新錄。開頭12秒保留國徽與雙主角故事鋪陳。

v1取實際名言切換時刻前後各約1.5秒，不加速60秒計時。從首次選提督到當地領船、完成編成與出擊的47.46秒維持來源連續原速；字幕分段時不刪除其中操作。新片源完整1280×720縮放為1664×936置中，標題與字幕放在畫面外，人物、名言及上下船體沒有被字幕覆蓋或另行裁切。

配樂沿原BGM_0F，一首連續、不重啟、不循環，依既有演出重混40個音效。為容納正常整軍時間，只縮短展示及故事卡停留。正式影片126.883333秒、1920×1080、60fps、7613格、20226563 bytes，SHA-256為`2f417531977fe629b9421ec0bf3b6864b1a2fd66715d070e7a3813fc658ec263`。

AAC立體聲44100Hz，RMS0.08147、peak0.57017，對混音參考SNR37.0813dB，34個剪接點最低26.9381dB；非靜音且未削波。全片解碼無黑幀，46個新字幕安全框通過，18段舊畫面沿先前安全框。偵測器將國徽、靜態UI及占比小的名言／光束視為靜止，已分類為來源停留，沒有後製靜圖或延長影格。35段分鏡表、國徽、故事、海報、連續整軍、要塞砲、戰鬥機、片尾及名言切換前後已目視。這次未新增人耳驗收。詳細數字見本機`promo/AUDIT.json`及交付索引。

## 重跑入口與工具界線

以下命令在儲存庫根目錄執行，由既有包裝器啟動Docker。先讀[166狀態表](../goal/2026-10-26-poster-full-promo-round-166.md)及腳本來源路徑；準備腳本拒絕覆寫既有批次或版本。完整來源需本機合法資料與已驗基底，不能只靠公開Git重生完整版。

```sh
tools/go.sh run tools/scripts/garrison_promo_plan.go -out /src/workplace/poster166/garrison-plan

LOGH3_GUI_IMAGE=logh3-hd-preview:20261011-r1 \
LOGH3_XVFB_SCRIPT=/src/tools/with-xvfb.sh tools/gui.sh xrun \
python3 tools/scripts/poster_frame_promo.py audit

LOGH3_GUI_IMAGE=logh3-hd-preview:20261011-r1 \
LOGH3_XVFB_SCRIPT=/src/tools/with-xvfb.sh tools/gui.sh xrun \
python3 tools/scripts/poster_frame_receipt.py --check

LOGH3_GUI_IMAGE=logh3-hd-preview:20261011-r1 \
LOGH3_XVFB_SCRIPT=/src/tools/with-xvfb.sh tools/gui.sh xrun \
python3 tools/scripts/poster_frame_readme.py --check
```

封包入口為`poster_frame_release.py`的prepare／linux／windows／macos／appimage與實際解包命令。AppImage工具為`tools/appimage.sh`；macOS沿`tools/macos.sh`及`local_test_macos_build.sh`。錄製入口為`poster_frame_promo_gui.py`與`poster_frame_garrison_gui.py`，普通封包試玩為`poster_frame_gui.py`。影片依prepare／render／`poster_frame_mux.py`／audit／目視／publish順序，publish只安裝到本機，拒絕未目視或SHA不符的產物。

合併器對各MP4容器時長取毫秒四捨五入，原合成雖有7613格，尾端時間累積多4.667ms，平均幀率未精確60。mux工具先驗全部時間偏差小於半格，再把PTS與DTS同步對齊每格1000個1/60000秒；逐封包核視訊及AAC內容不變，沒有重編碼、增刪影格或加速。

Wine沿專案既有`eob-remake-wine-verify:ubuntu-noble-20261008-r2`映像與`tools/player_wine.sh`，設定`LOGH3_WINE_SCREEN=1280x800x24`。初次使用預設映像缺xdotool，已分類工具環境問題；沒有改產品規則。整軍初次座標整數截斷落在命中區外1像素，改以round對應實際88%畫面；軟體繪圖的回合過場另沿遊戲正常右鍵略過，乾淨重跑PASS，不是修遊戲資料或加測試旗標。

本輪不新增DOS同狀態、Windows／macOS真機、實機輸入延遲、FPS或新音樂人耳驗收聲明。#48同戰場讀回仍待完成。交付及清理最後狀態見166與交付索引。
