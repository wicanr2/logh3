# 365 — 音樂工具的本作品授權與來源閉合

日期：2026-10-02，工作項目：[#69](https://github.com/wicanr2/logh3/issues/69)。
前置：[驗證 036](../playtest/036-windows-combined-music-tools.md)；契約：[規格 077](../spec/077-music-source-delivery.md)。

已證實（proven）：0f08832 的根目錄沒有 LICENSE，README 只宣告採 RRSAL、待轉公開前補全文。
使用者既定契約要求完整 RRSAL-1.0，只換占位符；本輪按實際受版控目錄填入原版素材排除與第三方元件。
保留英文參考譯本，不自行改條款；封裝器以完整範本模式核對只有占位符可變。
範本 SHA-256 `cb9f00c0494f0ae87a48dbe15e8911e2e3fc4c64bd47420357ac8cc7c96d8a5c`；
實際 LICENSE SHA-256 `ef3709a8581ce080ee02ee25bab9e0006e767e980440c1ac40eef6a480b34583`。
這不授權原版素材或公開私有研究集合；原版權利人列名沿用原專案製作／發行紀錄，未裁定現行權利鏈。

Git 受版控清單確認 assets/original/ 的解碼美術、地圖、字型及 MIDI／VGM／OGG，
docs/images/ 的原版素材渲染畫面及 docs/lore/ 的原版文字均不能加入可散布來源包。
來源包只取乾淨 logh3 的 cmd／internal／tools／模組檔，以及固定 dosgolem 的 Go／模組／LICENSE。
原生音樂入口的實際 Go 依賴已量測；Go mod vendor 另匯出 978 檔與十個鎖版模組，
包含完整授權／PATENTS。來源 kit 另附 Go 的完整 LICENSE／PATENTS、新 LICENSE 與封裝器。

首次包 1,912 檔，兩次確定性 tar 相同。首次建置比較混用了 -mod=readonly 與 -mod=vendor；
Go buildinfo 的依賴版本相同，一側帶 h1、另一側沒有，完整二進位因註記而不同。
目前 Go 1.24.13 的 /usr/local/go/src/cmd/go/internal/modload/build.go 在 vendor 模式不從模組快取
補齊模組資訊；兩份實際 buildinfo 留在 rebuild/，不能豁免完整二進位驗收。
[Go 官方模組文件](https://go.dev/ref/mod#vendoring)說明 vendor 透過 modules.txt 提供版本資訊，
建置改讀 vendor 而非網路／模組快取。改為目前來源另建對照樹、獨立匯出 vendor，
兩側逐檔 vendor 相同並用相同旗標，三種 logh3 二進位逐位元組相同。

追加反例：來源側 TestFMImportLocksAndRoutes 失敗，必要的 fm-import-lock.json 漏收。
這是來源交付缺件，規格 077 回到 DRAFT；固定 JSON 只含身份／雜湊，且與同一提交的
fmdata.go 由現有測試逐欄核對，補審後才恢復 READY。原包及失敗不改寫，改用
logh3-source-reviewed.tar、source-kit-reviewed／repeat，新增一份同 HEAD 的鎖檔。

修正版 1,913 檔，兩次 tar 全檔相同；Linux／Windows 匯入器、Linux music-tool、
Windows dosgolem-probe 四種重建程式與同條件對照相同。來源側使用空 GOMODCACHE、
GOPROXY=off、-mod=vendor；容器預設無網路，尾端確認模組快取仍空。
來源 CLI 五個測試事件通過、一項 TestFMProbeIdentity 明示 skip；另有四組來源封裝邊界通過。
全部十八個標準產品套件報告完整，部分測試快取。未重新外推原版對拍、播放或人耳確認。

首次收據檢查器截取 logh3/ 前綴多切一字，造成 /ref/ICENSE 查無檔案；修正檢查腳本
後以同容器、同命令重跑，這是驗證腳本問題，不是來源或產品缺陷。
完整輸入、輸出、腳本、日誌與重生入口見[驗證 037](../playtest/037-music-source-delivery.md)。
正式發行版號／dist-all、macOS、實機、聆聽／循環仍待完成，#69 保持開啟。

## 追加勘誤：編譯成功尚不足以通過正式版本閘門

已證實：前述四種 -buildvcs=false 二進位只證明編譯閉合。正式匯入器 checkProbe 要求
真實 f7523f9 提交與 modified=false；來源側 TestFMProbeIdentity 跳過使此缺口未被檢出。
保留舊包及 before-vcs JSON（SHA-256 `753e0d9539ba82c404d8a103bbcdbc13232d9f73bf66a0aea22d61f1249c512e`），
規格返回 DRAFT。固定提交的淺層 Git 與完整樹探針恢復真實身份後補審 READY，未手寫版本常數。

固定 dosgolem 樹共 665 檔；其中 222 個 Go 檔，其餘是文件、規格、JSON、腳本與既有 rich2.test。
後者為 5,120,025 bytes 的 Go 測試程式，Git blob `531dd7f6fdac355d16913dda17d99dc24ce84298`；
Go buildinfo 證實為 linux/amd64、github.com/wicanr2/dosgolem/apps/rich2.test，不是 DOS 原版。
它僅為保持完整提交樹留存，未執行。Go 來源未使用 go:embed，未收錄原版媒體。
Git pack 的 .rev 是合法反向索引，依[Git 官方格式](https://git-scm.com/docs/gitformat-pack)
保留同前綴 .pack／.idx／可選 .rev，不以封裝器的兩檔假設刪掉索引。

第二個 VCS 候選缺 .git/refs，Git 向上找到外層 logh3 儲存庫；reset 動到外層 index，
沒有改寫工作樹內容，核對後全部本輪修改仍在。日誌 source-rebuild-vcs-2.log 與失敗包保留。
修正加入真實 branch ref，操作前明設 GIT_DIR／GIT_WORK_TREE 並核對 show-toplevel；
不符立即停下，避免外層身份污染來源驗證。

最後 source-kit-vcs-root／repeat 共 2,361 檔、32,819,200 bytes，兩份 tar 完全相同。
五種二進位與真實來源基準相同；Linux／Windows probe 均真正 f7523f9／modified=false。
來源 CLI 六個測試通過、沒有跳過。重建的正常 FM 入口以隨包依賴產生 BGM_0F，
完整 VGM／WAV／增益後 PCM 符合固定鎖，Linux OGG 全檔亦相同；未宣稱人耳或實機通過。
唯一收據及各失敗日誌的 SHA-256 由[驗證 037](../playtest/037-music-source-delivery.md)回查。
