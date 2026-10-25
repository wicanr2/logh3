# 133：v1.6.0三組海報、五語介面與本機交付

本機產品交付通過。產品現況見[150唯一狀態表](../goal/2026-10-24-poster-v160-round-150.md)，影片依使用者後續要求沿[151故事重剪](../goal/2026-10-24-story-promo-round-151.md)。現行封包與影片的雜湊以[總交付索引](../internals/data/v160-delivery.json)為準。

使用者指定三組海報與五語遊戲介面，版本為`v.1.6.0-20261024`。[152](../spec/152-poster-sets-and-quotes.md)定義海報，[153](../spec/153-interface-localization.md)定義顯示翻譯與語意隔離。原版存檔、索引及操作位置保持；圖片內原文及未收錄內容不冒稱完整翻譯。

封包沿[本輪建置器](../../tools/scripts/poster_release.py)、既有Go及macOS工具鏈。同一凍結來源建Windows amd64、Linux amd64、macOS通用完整版，所有共用資料與海報需逐檔相同。正式ZIP取出後，正常[五語與三組鍵鼠](../../tools/scripts/poster_sets_gui.py)及[駐地整軍](../../tools/scripts/garrison_promo_gui.py)各自驗證，診斷入口不能代替正常操作。

150初版沿[製片器](../../tools/scripts/poster_promo_movie.py)，僅取本輪正式包正常實錄。151改沿[故事製片器](../../tools/scripts/poster_story_promo.py)，依先前完整影片手法接入雙國徽、1988主角動畫與已驗實戰。新版操作的v1名言、v2及v3群像、五語、駐地改選提督、領船、E／T拒絕、C旗艦及出發保留。配樂沿遊戲BGM_0F單曲連續，不因剪接重新開始，實戰音效依played觀測重混。

README截圖入口：[v1.6.0來源索引](../images/readme-v160.json)。取正常1280×720視窗，保留圖檔、來源收據、版本、原始畫面與雜湊。新海報原圖與含原版音畫的完整版、影片限本機；本輪授權新增README實錄不等於發布完整版或移動既有tag。

## 封包與來源

1610項凍結來源含632項非vendor來源；Windows與Linux沿`logh3-ebiten:2`，macOS沿`wizardry7-osxcross:go1.24.13-python3.13.15-15.5-r1`。三包615項共用檔逐檔相同。獨立核對ZIP CRC、完整檔名集合、manifest、SHA-256、執行權限、UTF-8檔名及中文說明BOM／CRLF。

| 平台 | ZIP大小 | SHA-256 |
|---|---:|---|
| Windows amd64 | 451562109 bytes | `008710aeea24fd93ed383ff35e58de80c2f1a28b378d672d258c6d8aad663aa6` |
| Linux amd64 | 451686846 bytes | `ff05f0dd4739d39c72de900e890f343555172a38e479819fcb60a294f41c85ab` |
| macOS通用 | 461380729 bytes | `a53357d22fd46de93fb516e0d027a1899d8e937f8879389de979ddfccca8c622` |

正式位置為`dist-all/v.1.6.0-20261024/full/`。macOS同包含Intel與Apple Silicon，最低11.0、SDK15.5，系統依賴及ARM全部簽章頁通過。Windows採Wine驗證，macOS未真機，不能以架構及簽章核對代替真機驗收。來源及建置結果由`full/BUILD.json`與[總交付索引](../internals/data/v160-delivery.json)保存。

## 遊戲與翻譯驗證

最終來源的標準`tools/gui.sh test ./internal/... ./cmd/...`共20套件，16 PASS、4無測試。譯表762項、字形5368項；UTF-8 TSV欄數、唯一key、五語非空、字形覆蓋、顯示複本不改原始frame、裁切及快取均有針對性測試。最終來源凍結後沒有再修改產品程式或字型。

正常玩家路徑以實際ZIP取出程式及資料，從主選單新局進入。未使用`-script`、狀態注入或seed注入；遙測只讀。Linux與Windows各54張，涵蓋五語設定、戰略星圖、F1說明、提督基本及特殊能力、三組海報、英文與韓文F7、左右海報防誤觸、存讀與退出重啟。另有帝國8張、原版renderer8張。兩平台十組開關海報中央比較共6912000像素，逐像素零差，沒有遮罩。

Windows首次獨立比較發現日文一張截圖保留舊星球hover，遙測位置與實際游標不一致。保留整份初次收據於`player-windows-stale-hover`，修改擷取器於視窗改寬後先移至鄰點再回固定位置，重新跑完整正常流程，沒有改產品程式或遮掉差異。正式核對只用重跑收據。

駐地整軍在兩平台各96步、19張，包含從Region19改選派特、第3艦隊、C旗艦、E／T拒絕、領取當地船艦、移動並抵達Region18、六表保存讀回及開關重啟。當地庫存由`9000,3000,16000,2000,12000,3000,1000`降為`6000,3000,12000,2000,9000,3000,1000`；母星保持`102000,30000,104000,36000,9000,6000,5000`。五語流程與駐地流程各自產生的六表，在兩平台逐檔相同。

另有兩平台共6條診斷入口，從`/tmp`執行程式及原版／強化啟動器，不指定資料路徑，確認自動找包內素材、HD、GM、MT-32及預設關閉的海報與非戰艦旗艦。診斷入口與正常GUI分開聲明。

翻譯驗收限本次762項及正常抽樣路徑。圖片內故事、商標、未收錄的原字串仍保留；不聲稱每張畫面每一行文字皆已翻譯，也未新增DOS同狀態parity。原版戰術存檔讀回重建仍屬既有#48。

## 第150輪初版與README

全新錄影剪為20段、91.883333秒，1920×1080、60fps、5513格，檔案13375916 bytes，SHA-256為`22c0e7c855d8c99a2d10da81fb4a09a13d4847e3ca646658744fa4c832ba7f0d`。實際操作皆取本輪正式ZIP，從改選提督到編成及下達移動保持原速連續；駐留等待與下回合抵達有明示剪輯。全片同首BGM_0F連續，AAC雙聲道44100Hz，RMS 0.08010、peak 0.46027，與連續參考音軌SNR 40.75dB，19個剪接點逐段通過。此為離線配樂，沒有混入遊戲音效，不冒稱人耳驗收。

無黑段，52項字幕文字安全矩形通過。凍結偵測保留於AUDIT，來源為片頭片尾字卡及真實介面停留，不將它們隱藏或宣稱零凍結。已目視20格CONTACT與完整第04、11、16抽樣格，人物臉部、五語名言、字幕及C旗艦編成無裁切。初次concat的毫秒時長四捨五入使平均幀率不精確，保留候選後重整固定60fps；不重錄或改遊戲速度。

README新增18張正式包完整1280×720原圖，逐張與正常GUI收據SHA-256相同，沒有縮圖、裁切或重畫。涵蓋兩陣營設定與星圖、三組海報、四個外語、提督編輯及前線改選提督／領船／編成。舊戰術與交火圖明標v1.5.0，沒有冒稱新版重拍。

## 第151輪：國徽故事版為現行影片

使用者明確要求沿先前完整拍攝手法，以雙方國徽作底圖、加入故事鋪陳，再串連新版內容。重新核對126雙旗片頭、[140影片契約](../spec/140-hero-tactics-and-release.md)、143正式分鏡與成片CONTACT，先保存150初版及其PLAN、AUDIT、校驗、交付索引與產生器，共10檔。保存映射為`workplace/poster151/previous/MAPPING.json`，舊收據的原路徑依映射回查，原位元組與雜湊不改。

現行影片仍為`dist-all/v.1.6.0-20261024/promo/LOGH3-v.1.6.0-20261024-promo.mp4`，已改為35段、123.566667秒、1920×1080、60fps、7414格。檔案18336447 bytes，SHA-256為`2ce661ff5fa66ac2498718e39a5cf81443db6adfe6f7b26b2b457a0540f78314`，取代同路徑的92秒初版作為目前入口。三平台完整包及18張README原圖保持，未重新建置遊戲。

| 時間 | 敘事與畫面 |
|---|---|
| 0–14秒 | 原雙國徽作底圖，百年對峙、1988動畫雙主角、亞姆力札補給危局，引出由玩家下令 |
| 14–33秒 | 三組海報與英雄親友，繁中及四外語，提督能力編輯 |
| 33–80.88秒 | 前線駐留、非戰艦旗艦、改選提督、當地領船、E／T拒絕、編成及出發抵達 |
| 80.88–119.57秒 | 兩軍戰術部署、完整要塞砲7.6秒，以及三武器完整原速交火合計28.083333秒 |
| 119.57–123.57秒 | 雙國徽與「下一頁歷史，由你下令」收束 |

雙國徽直接沿`workplace/hd127/promo-final/BACKGROUND.png`，逐幀微量星點沿先前風格，不另生成國徽。人物動畫沿使用者先前指定的1988版，保留原字幕及水印，不稱原生HD。新版操作沿150實際ZIP新錄；兩軍戰場、要塞砲與交火沿v1.5已驗來源。每段綁定舊PLAN、SOURCE-PLAYBACK、原RAW與演出收據SHA-256，三種武器皆有已見結果且無漏格的原速完整演出，不冒稱全片新版重拍。改選提督至下達出發保持原速連續，回合等待仍明示剪輯。

全片同首BGM_0F從零連續播放，無循環或剪接重啟；另依既有played觀測重混40處遊戲音效，不混入各素材的其他配樂。AAC雙聲道44100Hz、RMS 0.08160、peak 0.52711，與連續音樂及音效參考軌SNR 36.30dB，34個剪接點通過。無黑段、82項字幕安全矩形通過；凍結偵測保留並分類國徽／字幕停留、微量星點與來源相同的介面停留，不稱零凍結。已目視35格CONTACT及完整00／01／03／15／32／34格，國徽、故事銜接、動畫來源標記、名言、E／T拒絕、交火及片尾均無裁切。音訊為技術驗收，未作人耳驗收。

新產生器為[poster_story_promo.py](../../tools/scripts/poster_story_promo.py)，依序在既有HD容器執行`prepare`、`render`、`audit`，目視AUDIT所列抽樣後才`publish`至現行promo。舊92秒成片與全部收據保存於previous映射。最後以`poster_delivery.py receipt --refresh-promo`更新現行總索引；該模式要求舊索引備份雜湊相同，且除影片、校驗與產生器外的產品、封包、字型、圖片、正常GUI欄位必須全同，否則拒絕。

使用者另明確要求GitHub README海報截圖，已發布[五檔提交44a8cae1](https://github.com/wicanr2/logh3/commit/44a8cae1875a9af39ff9f23e645d7b8f96eeb5ff)，僅含README、三張完整海報實錄與[三圖索引](../images/readme-v160-posters.json)。線上README及三張PNG、JSON均實際下載核SHA-256相同，PNG解碼1280×720通過。[線上預覽入口](https://github.com/wicanr2/logh3#海報寬螢幕預覽)仍明列公開引擎為v1.5.0。原始海報、完整版、影片及其他未提交產品未公開，既有tag／Release保持。

## 重現入口

正式收據在`workplace/poster150/{player-linux,empire-linux,native-linux,garrison-linux,diagnostic-linux}/RESULT.json`及`workplace/music69-windows/poster150/{player-windows,garrison-windows,diagnostic-windows}/RESULT.json`。各收據保存程式與擷取器雜湊、輸入事件、完整視窗及唯讀狀態。可追溯的總雜湊與獨立判定在[總交付索引](../internals/data/v160-delivery.json)。

依序於既有HD容器執行`python3 tools/scripts/poster_delivery.py packages --check`、`images --check`及`receipt --check`。封包驗證不重建產品；來源、README原圖、實錄、安裝檔、影片或收據若改動，核對即失敗。正式包與影片的總清單在`dist-all/v.1.6.0-20261024/SHA256SUMS.txt`。
