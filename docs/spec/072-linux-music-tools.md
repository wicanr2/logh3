# 072 — Linux 原生隨包音樂工具

狀態：**CONFORMED**（2026-10-01，限 Linux amd64 工具封裝與匯入抽驗）。
工作項目：[#69](https://github.com/wicanr2/logh3/issues/69)，證據：[研究 360](../re/360-linux-bundled-music-tools.md)。
審查：同版本載入器與完整依賴在無音訊工具的乾淨容器已可啟動且解析全來自包內。
這是既有 071 匯入管線的依賴封裝，不改音源、規則、存檔、固定輸入或雜湊門檻。

## 契約

1. 建置器只在既有 audio image 內執行，工具與完整共享依賴沿用實物，不臨時掛主機 library。
   暫存封裝留 `workplace/`，正式交付仍須依 `dist-all` 契約與版號 gate。
2. 包內 `tools/{ffmpeg,fluidsynth,opl2wav}` 是原生 Go 啟動器；
   `tools/real/` 保存原第三方執行檔，`tools/lib/` 保存同 image 載入器與全部共享依賴。
   `tools/dosgolem-probe` 是 clean 固定 HEAD 的 Go 程式，不包另一種執行器。
3. 啟動器由自己的完整路徑取相鄰檔案，不搜尋 PATH、不呼叫 Docker／shell／Python。
   使用隨包載入器、`--inhibit-cache` 與字面值 `$ORIGIN/../lib`，適用搬移及中文／空白目錄。
   以程序替換保留退出碼與信號語意，既有匯入器逾時可直接終止真正工具。
4. 清除 `LD_*`、`GLIBC_TUNABLES` 等載入干預；載入參數來自包自身。
   只准三個工具名；完整 manifest 必須 Linux amd64、非空且有界。
   先驗載入器、真實工具及清單內共享檔案存在且為普通檔案；缺件不能偷偷依賴系統安裝。
5. 建置收據保留原 image 身份、工具／載入器／全部依賴雜湊、各檔所屬套件、
   來源套件／版本、完整著作權及 GPL／LGPL 等授權文字。
   音色庫從已驗快取或固定版本下載取得，驗 SHA-256 並隨包完整 MIT／README 聲明。
6. 不放原版 EXE／MDS／生成曲庫到工具包；使用者自備原版僅傳給 071 入口，來源唯讀。
   第三方程式與授權獨立保留，不聲稱 remake 授權覆蓋它們。

## 驗收與限制

乾淨容器沒有 FFmpeg／FluidSynth，無有效 PATH；全部動態解析指向包內，
移到中文／空白目錄後版本入口可用。缺共享依賴、未知工具名或壞 manifest 明確失敗。
正式原生匯入器抽驗 BGM_0F、END_2F、OPEN_F 三種 FM 路徑與至少一首 GM；
FM 的 VGM／OGG 全檔仍須符合原基準，GM 完整 PCM 與 071 同命令一致。
標準產品測試通過；保存命令、來源及產物雜湊，檢查 UID／GID 與容器清理。
不將 Linux 容器驗收外推所有發行版／實機；其他平台、正式對應來源碼封包與人耳／loop
仍待 #69 接續驗收，未達整項判準不得關閉。

## 實際結果

完成 220 共享檔案、198 套件授權／來源身份與 MIT 音色庫的暫存包，
總檔案 419,718,884 bytes。移至中文／空白路徑，在沒有系統音樂工具、PATH 無效且
LD_PRELOAD／LD_LIBRARY_PATH 被注入錯值的乾淨容器，三個原生入口成功。
全部動態依賴解析取自包內；缺 libc 與未知工具名明確拒絕。
BGM_0F、END_2F、OPEN_F 的 VGM／OGG 全檔符合 071 基準；BGM_0M 完整 PCM 相同。
四首再經正式播放層完整解碼，標準產品套件與補強邊界測試通過。
來源／命令／程式雜湊見[驗證 032](../playtest/032-linux-music-tools.md) 與結構化收據。
這個 CONFORMED 不包含 157 個來源套件完整 archive 的正式散布、其他平台或人耳驗收。
