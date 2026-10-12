# 第110份驗收：三劇本、S1實錄影片與HD介面試樣

2026-10-11，第123輪。交付版號`v.1.0.1-20261011`。本頁沿[計畫123](../goal/2026-10-11-hd-interface-and-s1-promo-round-123.md)，不新增原版同狀態聲明。

## 修正範圍

出貨資料已有S1「亞姆力札會戰」、S2「諸神的黃昏」、S3「巴米利恩會戰」。個人存檔索引尾三格空白，會使載入程式用空格覆蓋出貨目錄。129先審READY，再補空格回退；保存後重新整理清單也走同一函式。非空個人記錄、前八格、六檔格式及原版資料保持。

`TestBundledScenariosSurvivePersonalIndex`修正前三案均停在讀檔畫面，修正後通過。Linux及Windows Wine實際新包另驗個人第0格、S1、S2、S3，載入回合分別為346、93、346、380；保存後再次載入S3並保存，四案均為380。這是公開CLI存讀抽驗，正常X11路徑另由實錄證明，不混稱。

## 三平台封包

三平台完整版與引擎包、兩份來源包使用同一完整版號，位於`dist-all/v.1.0.1-20261011/{full,release,promo}/`。前版`v.1.0.0-20261010`保持。

- Linux完整版：11條公開入口、兩種AI、戰鬥、存讀及600 tick配樂抽驗通過。
- Windows：Wine六入口及新包三劇本存讀通過；不等於Windows真機驗收。
- macOS：Intel／Apple Silicon通用架構、系統依賴、最低11.0、ARM ad-hoc程式頁及ZIP權限核對通過；未真機執行。
- 標準測試19套件，14PASS／5無測試。逐ZIP檔案、manifest、SHA-256及完整／引擎程式身份由交付核對器驗證。

## 推廣影片

影片90秒、1920×1080、30fps，共2700幀。左帝國旗、右同盟旗，以「這一戰由你指揮」及「下一頁歷史，由你寫下」串起故事。讀檔清單實際停留供閱讀S1、S2、S3，再選S1，進入戰場、選艦隊、下令、移動並交火。兩段保留正常戰鬥動畫，包括光線砲、命中與爆炸。

畫面來自兩份正常X11鍵鼠錄影，沒有direct-entry、局面注入或`-script`。前段用正式Linux完整版；交火錄影用同工作樹建置程式，兩個程式hash各自保存。成片跨回合剪接，配樂後配，不宣稱90秒連續完成或現場音效同步。完整4:3實錄置於中央1280×960，不裁掉遊戲資訊。

`PLAN.json`保存九段來源與時間點；`TRACE.json`、`EVENTS.json`、原始影片及程式hash可回查。首尾為故事卡，讀檔與回合輸入停留屬實錄。凍結偵測須依來源逐段分類，不以字幕或旗幟動態掩蓋遊戲區。

技術檢查及目視結果：九張分鏡與八張交火抽樣通過；遊戲區逐秒抽樣34張皆不同。無超過0.25秒黑段及兩秒靜音段，平均−23.6dB、峰值−8.3dB，字幕零裁切。凍結區依來源列為故事卡、讀檔／報告閱讀、正常命令停留與遊戲演出停格，沒有宣稱全片無靜止。影片SHA-256為`3df53903e42fb364262dac5616dda390820aeec477d145ce5e34c822e2776f09`。

旗幟識別參考[銀英傳官方商品頁](https://ginei.club/infoDetail.php?iKey=1224)，照片來自[帝國旗商品圖](https://st.cdjapan.co.jp/pictures/l/04/43/NEOGDS-191915.jpg)及[同盟旗商品圖](https://st.cdjapan.co.jp/pictures/l/04/42/NEOGDS-191914_102.jpg)。故事背景參考[官方亞姆力札活動介紹](https://ginei.club/infoDetail.php?iKey=384)。圖片、遊戲音畫與衍生素材未取得公開散布權，影片及完整版只留本機。

## HD試樣與限制

素材盤點列PNG尺寸、來源hash及相同檔案內容。不同編碼的同像素圖、共享圖塊與完整消費端仍待核對，不按檔案數宣稱獨立畫作工量。動畫重複幀仍保留停留節奏。

`HD-INTERFACE.html`提供A「典雅艦橋」與B「清晰指揮台」。可切換會計、預算、七種艦型建造，操作箭頭、參謀、結束、Esc及本機試樣保存讀回。它是獨立HTML，未接正式遊戲規則與存檔。4:3亦僅暫作試樣，尚待使用者選型。

Chromium 154.0.8037.92的1440、800及390像素視窗通過DOM點擊、保存重設再讀回、Esc與排版檢查；七種艦型十四個按鈕可用，箭頭命中區至少44×44，無水平溢出。兩款桌面及窄視窗截圖已目視。星圖以SVG視窗顯示同一張原圖的地圖區，原PNG保持。瀏覽器映像、47份套件下載SHA-256及安裝清單記在交付的`smoke/hd-preview/TOOLCHAIN.json`；精確deb仍保存在映像內。

HD範圍包含操作介面、片頭、艦隊交火、飛彈、戰機、命中、爆炸及登陸。正式介面與逐幀動畫重製仍開放；本輪影片使用現行動畫細化。#48登陸演出、#69人耳／平台／循環、#66 MIDI I/F及#16原版欄位證據保持。

## 重現入口

- `tools/gui.sh test ./internal/... ./cmd/...`
- `tools/scripts/bundled_scenarios_smoke.py`，Linux走`tools/gui.sh xrun`，Windows走`tools/player_wine.sh`。
- `tools/scripts/s1_gameplay_record.py`與`tools/s1_record_display.sh`，輸入計畫及事件紀錄留在原始錄影旁。
- `tools/scripts/s1_story_promo.py prepare|encode|audit`，走`tools/audio.sh`，使用既有音畫映像與Noto字型掛載。
- `tools/scripts/hd_interface_inventory.py`及`hd_interface_browser_check.py`，映像見`tools/Dockerfile.hd-preview`。
- `tools/gui.sh xrun python3 tools/scripts/s1_finalize.py`核對並集中交付。精確命令、來源與輸出SHA-256見[唯一索引](../internals/data/s1-gameplay-promo.json)。

入口：[專案索引](../internals/00-README.md)。
