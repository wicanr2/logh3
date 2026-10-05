# 第104輪：Windows 正常玩家入口

遠端 worklist 89張、6 OPEN／83 CLOSED，最新核對269條無失效斷言。
本輪工作沿用已開啟的 #69，不新增平行清單。

## 目標

- 使用第103輪來源重建的 Windows 玩家執行檔，在既有 Wine image 開啟正常遊戲視窗。
- 原版及畫面規格唯讀，保存完整命令、二進位身份、擷取畫面與退出狀態。
- 先驗預設片頭自然到標題，再以真實滑鼠及鍵盤抽測正常選單。
- 結果只限 Wine GUI，不宣稱 Windows 實機、人耳、正式封包或完整原版同態。

## 契約與入口

命中知識路由的「Wine／正常玩家路徑」，載入
`knowledge-base/sources/claude/retro-cht/retro-game-playtest.md`、重製技能及
`knowledge-base/local/project-document-responsibilities.md`。專案 Docker-only 規則優先。
目前來源與既有收據見 [091](../playtest/091-current-source.md)及
[116](../spec/116-current-source.md)。流水帳仍只寫 `WORKLOG.md`。

正式版號、macOS、人耳及循環接點仍未驗收，#69 保持 OPEN。
