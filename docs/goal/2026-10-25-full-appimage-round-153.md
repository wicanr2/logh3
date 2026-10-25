# 第153輪：三平台完整版與Linux AppImage

使用者確認剛才試玩平台為Linux，要求以152輪游標、hover、字型及AI工具列修正重新打包三平台，Linux需AppImage。本輪不改遊戲規則或圖像，不重剪影片，也不發布GitHub Release。

## 目前狀態表

| 項目 | 狀態 | 依據與下一步 |
|---|---|---|
| 開工 | 完成 | HEAD44a8cae1，保留既有修改；Issue102張、#48 OPEN，282條verify零失效 |
| 產品來源 | 完成 | 1616項含638項非vendor與152全同，Windows／Linux內部程式逐位元組相同，沿16 PASS／4無測試 |
| 共用素材 | 完成 | 615項逐檔全同，包含三組海報、五語、四字型、音樂與原版資料 |
| Windows完整版 | 完成 | ZIP451581535 bytes，PE amd64，各三診斷及31張正常操作、五語、hover及六表存讀通過 |
| Linux AppImage | 完成 | AppImage451459576 bytes，固定runtime、741項AppDir與15個相依；無素材／存檔參數，正常31張及預設XDG六表存讀通過 |
| macOS完整版 | 完成 | ZIP461418440 bytes，雙架構11.0／SDK15.5、ARM4083頁簽章通過，未真機 |
| 交付與文件 | 完成 | full目錄三包與SHA-256，獨立收據通過；保留更新包、舊三平台及影片，未發布 |

交付入口沿專案`dist-all/v.1.6.1-20261025/full/`。AppImage包含完整素材，玩家偏好及存檔仍在既有XDG設定目錄；不寫唯讀AppDir。容器內採官方extract-and-run方式驗證，不能宣稱已驗實體桌面的FUSE掛載。[155封裝規格](../spec/155-appimage-full-package.md)、[135驗收](../playtest/135-v161-full-appimage.md)與[交付索引](../internals/data/v161-full-delivery.json)保存來源及限制。

重現入口：[三平台打包器](../../tools/scripts/full_appimage_release.py)、[AppImage工具](../../tools/appimage.sh)、[獨立驗證器](../../tools/scripts/full_appimage_receipt.py)。

入口：[目前脈絡](../../CONTEXT.md)、[152修正與限制](../playtest/134-pointer-toolbar.md)、[v1.6.0完整包](../playtest/133-poster-five-language-delivery.md)。
