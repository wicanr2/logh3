# 116：目前引擎與音樂工具的可重建來源

狀態：**CONFORMED**。2026-10-05，第103輪，限定本規格來源封裝及指定Go入口重建／技術抽驗。
沿#69及[077](077-music-source-delivery.md)，結果見[091](../playtest/091-current-source.md)。
這是已核准來源交付流程的延伸，不改玩家規則、音色、跨平台FM門檻或來源資料格式。
logh3固定9e1d31830f3d5555cfc313e06f83b0d159186ff1；dosgolem固定f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c。

## 證據與契約

077／[037](../playtest/037-music-source-delivery.md)已證實Git archive、完整vendor、授權範本及單一提交Git物件的來源閉合。
目前go.mod／go.sum仍鎖同十個Go模組，cmd及internal沒有go:embed；沒有新增編譯期媒體輸入。
目前來源檔名檢查只有Go、工具腳本及Dockerfile，不含原版媒體。
Linux玩家入口另需既有logh3-ebiten:2的CGO／ALSA／X11工具鏈，不能用CGO=0推測可建置。
Windows玩家入口以GOOS=windows、CGO=0抽驗；若失敗，回本規格補證據，不猜補產品程式。

1. 沿077的tar及manifest逐檔欄位，以明示40位Git HEAD產生乾淨cmd／internal／tools、模組及FM鎖快照。
2. 完整vendor由鎖定模組的正式cmd／internal依賴圖離線重生，逐檔含授權。研究tools及workplace不加入此依賴圖；原版oracle探針另屬dosgolem模組。
   固定dosgolem與Go授權取自037已凍結來源包，全檔SHA-256先核。
3. LICENSE須與指定HEAD一致，RRSAL完整範本僅占位符替換。當輪封裝器附件另記來源及SHA，不偽稱已在來源父提交。
4. 只寫workplace下新批次。拒絕既有輸出、連結、越界路徑、重複tar項及特殊檔案；逐項重讀名稱、大小、mode及SHA。
5. tar順序、mtime、UID/GID固定，兩次封裝全檔相同。舊037包及索引不重寫。
6. 使用同一Go工具鏈、-trimpath、-mod=vendor及logh3的-buildvcs=false，以空GOMODCACHE及GOPROXY=off建置。
   目前工作樹另建對照，獨立vendor與來源包一致；Linux／Windows gin3、music-import及Linux music-tool全檔須相同。
7. dosgolem保留真正Git身份與-buildvcs=true。核根目錄、fsck、唯一提交及clean，再重建Linux／Windows probe。
8. 原生CLI六項測試以來源重建probe驗身份，不以skip過此門檻。Linux玩家程式正常片頭啟動，未要求全平台實機驗收。

## 限制及停止線

來源包只保證指定程式可重建，不包含玩家原版、解碼素材、生成曲庫或所有合成器／編碼器來源。
GUI執行所需原版與畫面規格是分開的本機輸入，不能把來源編譯成功稱為完整玩家封包。
Linux／Windows合成器與依賴來源仍由033／035／036固定集合提供。
macOS、正式版號／dist-all、公開發行、實機、人耳及循環接點仍未完成；#69保持OPEN。
本輪不建立正式交付目錄，也不改已有的忠實度裁決。
