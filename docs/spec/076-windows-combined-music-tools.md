# 076 — Windows FM／GM 同包工具與預設完整入口

狀態：**CONFORMED**（2026-10-02；限 Windows Wine 同包原生入口、中文 ZIP 搬移與 Linux 正式播放層技術驗收）。
工作項目：[#69](https://github.com/wicanr2/logh3/issues/69)，證據：[研究 364](../re/364-windows-combined-music-tools.md)。
這是規格 071／074／075 既有入口的工具合併與封裝驗收，不改音源、門檻或遊戲規則。

## 契約

1. 同一個 `music-import.exe` 的預設入口生成 GM／FM 各十一首，不加診斷選曲旗標。
   `tools/` 放四個原生工具及已驗依賴，`soundfont/` 放固定 MIT FluidR3_GM；
   玩家程式不搜尋 PATH，也不呼叫 Docker／Python／shell。
2. GM 與 FM 元件僅取自兩份固定技術收據；同名工具／DLL 若雜湊不同，封裝即拒絕。
   全部普通檔案、PE x86-64 與非系統引用須核對；不得使用未修改 SDK 或另一組執行期。
3. 用乾淨提交建置匯入器，保留 Go buildinfo、實際程式雜湊與來源 HEAD／dirty 狀態。
   dosgolem 仍須為規格 071 的 clean 固定 HEAD；不從 DOSBox 取得原版收據。
4. 包內保留完整第三方授權、來源／工具／補丁身份與完整固定來源 tar。
   README.txt 使用 UTF-8 BOM／CRLF，說明工具授權獨立與生成曲庫只能留本機。
   包不含原版檔案或原版衍生音樂；來源 archive 已取得不等於簽章信任或逐套重建完成。
5. 建置先驗全部輸入與輸出目錄擁有權，拒絕覆蓋既有包，建立內容與 SHA-256 清單。
   ZIP 若有使用中文路徑，逐項驗 UTF-8 旗標、解壓後逐檔雜湊，搬移後再實跑。
   驗證包保存於 workplace；正式交付須另通過版號與 dist-all 契約。
6. 原版來源唯讀，合併包的預設入口使用獨立輸出；FM 逐曲驗 VGM／WAV／增益後 PCM，
   GM 沿用既有相同 MIDI／工具／音色庫參數與載入失敗拒絕。不可把舊產物複製成新批次。
7. Windows opl2wav 從 Unicode 命令列取得 UTF-8 參數，僅檔案邊界轉 UTF-16。
   自訂 CFileProvider 使用原 libbinio 的 binifstream 讀取 API 與既有析構清理，
   透過 `_wfopen` 開啟輸入／輸出；AdPlug factory、取樣、update、DSP 與 WAV bytes 不改。
   轉碼或開檔失敗明確返回錯誤，Linux 維持原入口；保留舊工具及補充來源／建置身份。

## 驗收與停止線

中文／空白工具與輸出路徑，單次預設入口真實生成二十二首；兩份階段收據均完成且
同一原生入口雜湊，實際工具均取自同一個包。FM 全部符合無損基準，GM 與 075
音訊封包相同；正式播放器完整解碼二十二首、正常玩家狀態選曲與標準產品套件通過。
缺件、撞名、不同雜湊或未審依賴必須明確拒絕，不以系統安裝補洞。
來源包／授權／UTF-8 文字／輸入未改／輸出擁有權與 Docker 清理均核對。
CONFORMED 限 Windows Wine 原生 CLI 與 Linux 正式播放層；Windows 實機、人耳、
循環、macOS、完整正式發行封包／版本／本作品授權與原始碼交付仍待驗收，不關 #69。

## 首次完整入口的反例

GM 十一首完成；FM 的 BGM_0F 已擷取並通過固定 VGM 雜湊，opl2wav 卻收到
`Z:\out\combined-full\?? ??\...\BGM_0F.vgm` 而失敗。故同名 DLL 與靜態引用閉合
不足以證明中文輸出路徑可用，返回 DRAFT。不得改用英文輸出作為本契約通過證據。
須確認 Windows 命令列與 AdPlug／WAV 檔案入口的編碼契約，再審補充 READY。

追加審查：已核對官方 Unicode／寬字元開檔契約及固定 AdPlug／libbinio 檔案 provider。
可丟棄探針三路固定 VGM 的中文輸入／WAV 輸出與舊英文路徑 WAV 逐位元組相同，
完整 PCM 相同、首路重跑相同；沒有 DSP 改動。依上述第七條恢復 READY，
後續必須以相同中文路徑條件從乾淨輸出完整重跑，不能把首批失敗收據升格。

## 完成收據

修正版獨立輸出真實生成 GM／FM 各十一首；全部 FM 符合固定 VGM／WAV／增益後 PCM，
GM 音訊封包符合 075。正式播放器完整解碼二十二首、正常玩家狀態選曲與標準產品套件通過。
中文 ZIP 逐項 UTF-8／雜湊核對，解壓搬移後一首 GM／三種 FM 路徑通過；六組封裝邊界拒絕通過。
完整命令、程式／輸出身份與限制見[驗證 036](../playtest/036-windows-combined-music-tools.md)。
首批失敗留存，不宣稱 Windows 實機、人耳、循環或正式交付完成，#69 保持開啟。
後續：[規格 077](077-music-source-delivery.md) 已補齊本作品授權全文與可重建來源套件的技術準備；
正式版號／dist-all 交付仍待完成。
