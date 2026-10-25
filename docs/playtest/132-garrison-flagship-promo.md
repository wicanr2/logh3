# 132：駐地出擊與非戰艦旗艦

日期：2026-10-24。入口為[唯一狀態表146](../goal/2026-10-24-garrison-flagship-promo-round-146.md)，規則依[150](../spec/150-non-battleship-flagship.md)。這是使用者指定的remake擴充，沒有新增DOS原版對拍聲明。

## 正常玩家路徑

從主選單開同盟新局，兩位提督各帶艦隊從母星前往友方前線駐留。執行下一回合後，兩支艦隊解散，提督與船艦留在駐地。此時才開啟「非戰艦旗艦」，先查看第一位提督，再改選第二位派特，編成第3艦隊。嘗試以T及E替換中央旗艦均被拒絕，改用C成功，再加入B及D。確認編成後下達移動，執行回合並確認實際抵達。

| 比較點 | 實際觀測 |
|---|---|
| 前線駐地 | 星系19；目標為相鄰友方星系18 |
| 改選提督 | 40改選41；新編第3艦隊，不宣稱既存艦隊的人事替換 |
| 駐地編成前庫存 | B／H／C／A／D／T／E為9000／3000／16000／2000／12000／3000／1000 |
| 駐地編成後庫存 | 6000／3000／12000／2000／9000／3000／1000 |
| 母星庫存 | 當地編成前後均為102000／30000／104000／36000／9000／6000／5000 |
| 旗艦 | C，4000艘；新增B及D各3000艘 |
| 支援艦拒絕 | T與E均顯示訊息，庫存與編制不改寫 |
| 保存 | 正常「存進度」再「取進度」，第3艦隊、提督41、星系18與C旗艦保持 |
| 偏好 | 新目錄預設關閉，點擊開啟後保存，退出再啟動保持開啟 |

[輸入模型](../../tools/scripts/garrison_promo_plan.go)只用公開Game輸入產生預期狀態；開關API在實際視窗中轉為設定按鈕。
[GUI腳本](../../tools/scripts/garrison_promo_gui.py)從正式主選單送真實X11鍵鼠，按下被程式觀測後才放開，沒有重送、狀態注入、seed注入或測試劇本旗標。`-intro=false`只略過開頭動畫，仍由主選單開始。設定、編成、移動、保存與讀回都走產品UI。

新增的`StrategicObservation`僅在可選frame-trace啟用時匯出複製資料，記錄所在地、當地與母星庫存及艦隊，不供應輸入或更動局面。實錄逐步與模型比對；獨立核對器再驗編制長度、艦種、數量、池欄位與庫存。

## 驗證狀態

- `tools/gui.sh test ./internal/... ./cmd/...`：19套件，15 PASS／4無測試。
- 艦種開關矩陣、手動旗艦與取消退船、低階轉移交易拒絕後不改寫、遊戲轉移及六表保存、偏好預設／未知鍵／重啟通過。擊破轉移仍沿原成功率與主開關。
- Linux正常視窗19個檢查點通過，原始錄影279.85秒。
- Windows第一次重啟在人物載入時超過90秒，原紀錄保留於`workplace/music69-windows/garrison146/verified-r1/`。以180秒載入上限及寫入中間檢查點重跑，`verified-r2`完整通過；沒有改遊戲程式。
- Linux及Windows Wine各96步、19個檢查點，共38張完整視窗。兩側均確認母星庫存不減、C旗艦保存讀回及開關重啟保持。
- [獨立核對收據](../internals/data/garrison-flagship146.json)已產生。1593項凍結來源與615項現行產品一致，ZIP逐檔manifest、UTF-8 BOM／CRLF、PE amd64及實際`-version`讀回通過。數值測試不代表Windows真機手感、人耳驗收或DOS parity。

## 工具與可重現入口

沿既有`logh3-ebiten:2`建置與標準測試，`logh3-hd-preview:20261011-r1`錄影及剪輯，`eob-remake-wine-verify:ubuntu-noble-20261008-r2`跑Windows。沒有新增工具映像。所有分析及執行都在Docker內，原版掛載唯讀，UID/GID為1000。

```sh
tools/gui.sh run tools/scripts/garrison_promo_plan.go -out workplace/garrison146-plan-final
LOGH3_GUI_IMAGE=logh3-hd-preview:20261011-r1 \
LOGH3_XVFB_SCRIPT=/src/tools/with-xvfb.sh \
tools/gui.sh xrun python3 tools/scripts/garrison_promo_gui.py \
  --binary /src/workplace/garrison146-gin3 \
  --plan /src/workplace/garrison146-plan-final/PLAN.json \
  --out /src/workplace/garrison146-recording-r2 --video
LOGH3_GUI_IMAGE=logh3-hd-preview:20261011-r1 \
LOGH3_XVFB_SCRIPT=/src/tools/with-xvfb.sh \
tools/gui.sh xrun python3 /src/tools/scripts/garrison_promo_movie.py
```

腳本要求新輸出目錄，拒絕覆寫既有收據。Linux錄影用`-ldflags '-X main.buildVersion=v.1.5.2-20261024'`建置的正式入口。Windows封包由[既有替換包建置器](../../tools/scripts/personnel_windows_patch.py)加`--version v.1.5.2-20261024`乾淨凍結及交叉建置，不含原版或HD素材。

## 影片與交付邊界

新片取自本輪正常視窗實錄。駐留及執行回合的等待段有剪接；駐地設定、改選提督、領船編成及下令的來源127.5至197.5秒連續呈現，僅更換章節文字。完整960×720畫布以最近鄰縮放到1280×960，放在1920×1080畫面中央，不裁切UI。

配樂沿已匯出的遊戲`BGM_0F.ogg`，來源SHA-256為`48c22ed0b5d9fec5d0b9bed70e7a06ae1ab3a07faef99da80ddfd21b4644739c`。全片單曲連續播放，只有首尾淡入淡出，沒有遊戲音效或配音。影片含原版及衍生音畫，只留本機；Windows引擎替換包沿用玩家既有v1.5資料與存檔。

首次音訊候選RMS僅0.00953，未通過音量閘門。保留該候選於`workplace/garrison146-movie/QUIET-CANDIDATE.mp4`，以`garrison_promo_movie.py --resume-mix`重做配樂。各鏡頭長度、字幕及字卡SHA-256保持，畫面不用重編碼。最終RMS為0.07944、峰值0.46027，無削波；AAC與連續參考音訊SNR為41.04dB，7個剪接點各自通過，沒有重啟配樂。

影片共8段87秒、5220格、H.264 1920×1080 60fps，AAC 44.1kHz立體聲，7638278 bytes。零黑段，76項文字安全邊界通過；目視8段縮圖，另檢查改選提督、C旗艦編成及抵達的完整畫面，無字幕或UI裁切。凍結偵測包含4秒首尾字卡及實際介面停留；最長8秒區間跨越換旗艦及加船，54及60秒畫面有124174個像素不同，不能把小範圍介面更新誤判成錄影完全停止。未作人耳驗收。

| 交付 | 路徑及SHA-256 |
|---|---|
| 影片 | `dist-all/v.1.5.2-20261024/promo/LOGH3-v.1.5.2-20261024-garrison-promo.mp4`；`bc622d1e53b5f7ce328a04fe044060c0833ad6652562016fcce5544b8e0c8803` |
| Windows替換包 | `dist-all/v.1.5.2-20261024/release/LOGH3-v.1.5.2-20261024-windows-amd64-patch.zip`；8044560 bytes；`c8fb79e0b8e60e46cc6929af48b874d2751ce3eefe50cd71d524f6c2390ecc05` |

封包只用`gin3.exe`替換原v1.5.0或v1.5.1同名檔案，保留原有資料、素材、啟動器及`%APPDATA%\logh3\save`。既有v1.5.1交付與公開v1.5.0 Release未改。格式、差異與全工作根擁有權檢查通過，容器已退出自刪，無新工具映像。未commit、push或發布。
