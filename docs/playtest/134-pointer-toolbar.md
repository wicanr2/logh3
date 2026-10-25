# 134：游標回應、hover與工具列字型／AI

驗收對象為[154規格](../spec/154-pointer-and-toolbar.md)。唯一進度入口是[152狀態表](../goal/2026-10-24-pointer-toolbar-round-152.md)，機器收據放[來源與交付索引](../internals/data/pointer-toolbar152.json)。本次只驗證remake修正，不新增DOS原版一致性聲明。

## 程式與交付

Windows更新包為`dist-all/v.1.6.1-20261025/release/LOGH3-v.1.6.1-20261025-windows-amd64-patch.zip`。它只含引擎及授權文字；關閉遊戲後將`gin3.exe`覆蓋到既有v1.6.0完整版資料夾，沿用原啟動器、素材、海報與玩家存檔。完整版本由實際EXE的`-version`核對，ZIP內README為UTF-8 BOM及CRLF。

來源凍結1616項，其中638項非vendor。台北時間跨日前的v.1.6.1-20261024僅作候選與Linux驗證來源，未交付或發布；跨日後重新從乾淨凍結來源建置25日最終包。兩份凍結來源逐檔SHA-256全同，只有建置版本字串與更新說明日期不同。正式Windows收據必須綁25日實際ZIP解出的EXE，不能用24日收據代替。

## 根因與修正

- 軟體游標原本參與整張HD背景的雜湊，移動會觸發CPU重播及1920×1440材質上傳。現在背景與最上層游標分離，沿相同游標形狀、裁切及畫序繪製。正式Game.Frame及原色重播保持。
- 主指令列原hover沿JSON舊座標，與正式commandBar少40像素；部分舊按鈕的DrawPanel角色又是空白。正式點擊幾何及可見按鈕角色現在同時提供反光，系統下拉列沿同一份commandBar與列高。
- 工具列字型與設定頁共用現有FontStyle，實際v1.6.0素材提供原版點陣、原版強化、思源宋體及思源黑體。使用者選擇現有Noto／思源黑體，未新增圓體。原版顯示仍用原字形，設定頁仍用專用黑體。
- 工具列AI直接選原版或強化0至3，沿既有SetAILevel，與設定頁共用數值。字型及AI沿既有本次執行契約，沒有新增永久偏好。下拉選單可用Esc或右鍵取消，不點穿底下遊戲。

## 正常玩家路徑

Linux同盟、Linux帝國與最終Windows Wine同盟各從正常主選單開始，選陣營、通過防拷問答、關閉月報後操作。未使用`-script`、direct-entry、局面或seed注入；`-frame-trace`只讀取正常視窗狀態。各31張完整視窗原圖與24次移動樣本，逐張PNG、視窗尺寸、程式及驗證器雜湊保存於交付索引。

覆蓋項目：

- 主系統按鈕及存進度列hover、游標移動時背景上傳數保持。
- 工具列選思源黑體及AI強化3，設定頁與讀檔後數值一致。
- 繁體、簡體、日文、英文及韓語各開啟字型與AI選單。
- Esc、右鍵取消、F1說明、F7原版／HD、原版hover。
- 海報開關與中央像素、海報上方工具列、側欄防誤觸。
- 正常公開存檔、選擇讀檔陣營、六表讀回。同盟的Linux、Windows及v1.6.0基準六表SHA-256全同。

帝國第一次驗證器使用y188選擇讀檔陣營，實際第一列從`181+highlight_dy19=200`開始，因而停在StateLoadFaction。保留`workplace/pointer152/final-empire/FAILED.json`；改沿`battle-command-menus.json`的正式列高與loadfaction.go原點計算列內座標後乾淨重跑。產品程式沒有因此改動。

## 獨立核對與效能限制

固定不含游標的區域取原生完整視窗比較。HD按鈕及下拉列的基準變色像素均為0；修正後各為916及398。原版顯示同樣產生反光。三條正式路徑分別核對960×720普通畫面與1280×720海報中央，中央691200像素均零差，沒有遮罩或縮圖。

同一24點移動路徑，v1.6.0基準Draw中位32.85ms，凍結來源Linux樣本11.93ms；背景上傳20至20，移動不再重建HD材質。較早候選樣本4.30ms只作調查記錄，不作正式固定延遲承諾。容器共用主機排程、CPU配額與軟體OpenGL，數字不代表所有電腦的性能。

GUI收據中的`latency_ms`從Python啟動xdotool前計時，包含程序啟動與排程。較早收據的scope文字稱「X11指標事件至draw_start」過強；本次不能用該數字宣稱輸入延遲已降低。修正後驗證器已改正文案，舊收據保留原位元組。硬體端到端延遲與Windows真機操作仍待使用者實際體驗。

## 重現入口

全部分析、建置與GUI在既有Docker映像內執行。Linux沿`logh3-ebiten:2`建置，正常視窗沿`logh3-hd-preview:20261011-r1`與`tools/with-xvfb.sh`；Windows沿`eob-remake-wine-verify:ubuntu-noble-20261008-r2`，由`tools/player_wine.sh`建立有界Xvfb並沿用poster150的Wine環境。未建立新映像。

```sh
tools/gui.sh test ./internal/... ./cmd/...
tools/gui.sh xrun python3 tools/scripts/pointer_windows_patch.py --version v.1.6.1-20261025
```

打包器拒絕覆寫現有版本，重現時需先另選合規、尚不存在的候選版本。GUI實際命令、來源與執行檔雜湊保存於BUILD及各RESULT，舊版與修正版驗證器完整快照分別保存在`workplace/pointer152/driver-final.py`及`driver-final-v2.py`。

獨立收據由下列命令核對ZIP精確檔案集合、CRC、manifest、PE amd64、BOM／CRLF、來源、截圖、hover、海報中央及六表：

```sh
LOGH3_GUI_IMAGE=logh3-hd-preview:20261011-r1 \
LOGH3_XVFB_SCRIPT=/src/tools/with-xvfb.sh \
tools/gui.sh xrun python3 tools/scripts/pointer_toolbar_receipt.py --check
```

最終標準20套件為16 PASS、4無測試。游標透明合成測試允許兩次整數混色的最多2階差，背景原像素與原色重播仍精確一致。其餘測試包含游標背景別名、較早游標與裁切、快取失效、覆蓋按鈕、正式列界線及工具列輸入攔截。

Windows採Wine，未作Windows／macOS真機驗收；本次沒有重打三平台完整版或重剪影片。既有v1.6.0交付與公開v1.5.0保持，未commit、push或發布Release；#48正式戰術讀回重建保持OPEN。
