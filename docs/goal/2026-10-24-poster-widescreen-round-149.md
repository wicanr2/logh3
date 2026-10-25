# 第149輪：左右海報寬螢幕選項

使用者要求模仿`~/cht/zeliard`工具的呈現方式，新增左右海報選項，中央保留遊戲畫面成為寬螢幕；隨後指定左帝國、右同盟、人物各站一邊，並要求上網找圖再修改。

## 目前狀態表

| 範圍 | 狀態 | 驗收 |
|---|---|---|
| 參考及角色 | 已讀Zeliard呈現分層及官方經典動畫人物頁 | 左萊因哈特、右楊威利；網頁參考不當DOS證據 |
| 海報素材 | 完成，本機使用 | 官方參考經image_gen重構，左萊因哈特、右楊威利；圖檔、提示與來源雜湊齊全 |
| 16:9外框 | 151限定CONFORMED | 中央4:3完整保留；三組2073600像素零差，原版／HD均通過 |
| 開關及偏好 | 完成 | 預設關閉、保存重啟、缺圖退回；側欄左右點擊不改遊戲，鍵盤照常 |
| 正常操作及交付 | 完成，限所列平台 | Linux與Windows Wine各13張、Linux無HD另5張；19套件15 PASS／4無測試；本機Windows更新包已核對 |

本輪入口：[151](../spec/151-poster-widescreen.md)。工作區`workplace/poster149/`，新海報先保存在本機素材`assets/hd/poster/`，來源與生成提示記錄於該工作區及後續驗收索引。原有交付不覆寫，不擴張公開素材範圍。

驗證工具：[正常視窗與點擊邊界](../../tools/scripts/poster_mode_gui.py)、[顯示及偏好](../../cmd/gin3/poster.go)、[邊界與偏好測試](../../cmd/gin3/poster_test.go)。[來源索引](../internals/data/poster-mode149.json)保存生成提示、來源雜湊及驗收；本機Windows更新包沿[原替換包工具](../../tools/scripts/personnel_windows_patch.py)的`--with-posters`分支建立。

[獨立核對器](../../tools/scripts/poster_mode_receipt.py)在既有HD影像容器核對兩平台PNG及Windows更新ZIP；Wine映像不另安裝Pillow，Windows中央比較由此核對器完成。

目前Zeliard checkout未查得poster／bezel／widescreen命名的側欄實作；已核對其`cmd/zeliard-preview`及`internal/prototype/hd`的原生畫面／呈現分層，採同類外層合成方式，不宣稱移植了一支不存在的海報函式。參考工作樹僅唯讀。

## 本機交付

`dist-all/v.1.5.3-20261024/full/LOGH3-v.1.5.3-20261024-windows-amd64-poster-update.zip`，14586513 bytes，SHA-256 `e296296cc23cc9027c1d1b37ca58ed7c9eab0fb9ae66356d561b81eccdbedd3a`。將`gin3.exe`及`art`合併到既有v1.5.x完整版；設定頁右上「海報」切換，預設關閉。1595項凍結來源、617項非vendor來源與目前產品一致，最終EXE與Wine正常操作使用的檔案相同。包內README及素材說明皆UTF-8 BOM／CRLF，PNG、manifest、PE amd64及SHA-256已獨立核對。

同目錄`LOGH3-v.1.5.3-20261024-poster-preview.png`為1280×720正常Linux視窗原圖。完整素材、原版與既有三平台v1.5.2包保留；本輪未建立新的macOS完整包，未發布、commit或push。Windows為Wine驗證，非真機；沒有新增DOS parity聲明。
