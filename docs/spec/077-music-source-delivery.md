# 077 — 音樂工具的本作品授權與可重建來源套件

狀態：**CONFORMED**（2026-10-02；限本作品授權、指定來源重建與正式 FM 抽驗）。
工作項目：[#69](https://github.com/wicanr2/logh3/issues/69)。
前置：[驗證 033](../playtest/033-linux-music-sources.md)、[035](../playtest/035-windows-gm-sources-unicode.md)、
[036](../playtest/036-windows-combined-music-tools.md)。本契約補交付來源，不改玩家規則或音訊門檻。

## 契約

1. `LICENSE` 僅使用已定案 RRSAL-1.0 完整範本，替換全部占位符；原版衍生素材與第三方元件
   明列排除，README 連到全文。私有研究集合不可因此宣稱可公開，來源包不收錄它。
2. 來源包以乾淨提交的 Git archive 為基礎，記錄 archive 的 Git comment HEAD 與 SHA-256。
   logh3 包含 Go 模組檔、cmd／internal／tools，以及測試必要的 docs/internals/data/fm-import-lock.json；
   JSON 只含固定身份／雜湊，不含原版素材。dosgolem 包含固定 HEAD 的完整來源樹及一層 Git 物件；
   不得手寫版本常數偽造 clean HEAD。逐項清單確認無原版媒體，唯一 .test 是其既有 Go 測試程式，
   只為保持提交樹身份留存，不執行、不作遊戲輸入。
   若所需非 Go 檔或嵌入資源未在此集合，建置應失敗並返回規格，不得用未記錄的工作樹補洞。
3. 當輪 `LICENSE`、來源封裝器與 README 附件另綁 SHA-256，明示來源父提交及新增附件，
   不偽稱它們已存在於乾淨父提交。已發布的來源與技術收據不得原地覆寫。
4. Go 依賴從同一個 go.mod／go.sum 的離線 `go mod vendor` 取得，保留 modules.txt、完整 LICENSE／PATENTS
   及每檔內容身份；不能把主機 Python／Go 快取直接當作玩家工具。包內不含原版 EXE／資料／音樂。
5. 目標與全部父目錄／輸入須確認形態及 UID/GID，拒絕連結、路徑越界、重複名稱、特殊檔案與既有輸出。
   tar 每筆使用確定的順序、權限、時間戳與擁有者；重讀全部 member 名稱／大小／SHA-256。
6. 在無網路容器將來源解壓到獨立目錄，使用空的 GOMODCACHE 與明示 `-mod=vendor`，
   建置 Linux／Windows 匯入器、Linux 隨包啟動器、Linux／Windows dosgolem-probe。
   同一工具鏈及 -trimpath，logh3 用 -buildvcs=false、dosgolem 用 -buildvcs=true，與目前固定來源逐位元組比較，
   避免把 Go VCS 註記差異或絕對路徑當成程式邏輯差異。
   logh3 的對照樹取自目前程式、獨立重生 vendor；兩側均使用 vendor，逐檔核對 vendor 內容相同。
   一般模組與 vendor 的 buildinfo 雜湊欄位不同，首次反例保留，不能因此免除完整二進位比較。
   dosgolem 的 Git 物件須通過 fsck、只有固定 HEAD 一個提交；reset index 後完整工作樹乾淨，
   兩側 buildinfo 皆為真正 f7523f9／modified=false。
   Git pack 保留一組同名前綴的 .pack／.idx，可附 .rev 反向索引；這是固定來源的 Git 格式，
   不得為了假設只有兩檔而刪掉現有索引。
   .git 須包含 HEAD、refs 與 objects，操作前明設 GIT_DIR／GIT_WORK_TREE 並核對 show-toplevel
   是解壓來源；若不是立即停止，禁止向上搜尋到外層專案後 reset 或以外層身份建置。
7. 來源自帶的原生 CLI 測試須通過；缺原版的項目明示 skip。基準與重建程式相同只證明
   來源足以重建指定入口，不等於新平台實機或完整正式版驗收。
   TestFMProbeIdentity 必須使用來源包建置的真實 Linux probe 執行，不得以 skip 過身份閘門。
   另以重建程式正常 FM 匯入一首，確認固定 VGM／WAV／PCM 與正式身份閘門皆通過。

## 驗收與停止線

RRSAL 範本比對只能有已審占位符替換；Go 依賴／授權與來源身份完整，獨立封裝重跑
tar 全檔雜湊相同；五種重建程式與對照相同，來源入口測試及一首正式 FM 匯入通過。
唯一 JSON 綁 archive、來源附件、建置命令、輸出、工具鏈與限制；收據與重生命令掛入既有索引。
技術來源包仍放 workplace；正式版號／dist-all、Windows 實機、聆聽／循環及 macOS 保留待驗。
本輪不重新宣稱原版逐週期音訊或完整玩法一致，不關 #69。

首次來源包、兩次封裝收據與測試失敗保留。鎖檔漏項補審後另建 source-kit-reviewed，
不能修改 source-kit／source-kit-repeat 或把原本失敗批次改成通過。

前階段的新包 1,913 檔、兩次 tar 相同；四種二進位與同條件對照相同，來源 CLI、
四組邊界與標準產品套件通過。完整命令、雜湊及一項明示 skip 見
[驗證 037](../playtest/037-music-source-delivery.md)，未因此完成正式發行或平台／人耳驗收。

再次追加反例：上述 -buildvcs=false 只證明編譯閉合，正式 checkProbe 仍會拒絕來源重建的
dosgolem。返回 DRAFT、用固定提交完整淺層來源建立原生探針，確認真正 Go buildinfo
恢復 f7523f9／clean 後補審 READY；舊包與 JSON 留在本機 before-vcs 記錄，不升格為正常入口。

最後候選補齊 Git refs，操作前明設根目錄且拒絕外層儲存庫；2,361 檔、兩次 tar 相同。
五種二進位與對照相同，真實 VCS 身份／乾淨狀態成立；CLI 六個測試通過、零跳過。
重建正常入口 BGM_0F 的 VGM／WAV／增益後 PCM 符合固定鎖，Linux OGG 全檔亦相同。
四組來源邊界及十八個產品套件報告通過；詳細身份與命令見[驗證 037](../playtest/037-music-source-delivery.md)。
