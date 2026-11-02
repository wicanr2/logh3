# v.1.6.5-20261102 三平台完整版

2026-11-02，第172輪。修正內容見[驗收143](143-new-game-first-battle.md)，機器收據見[v165-full-delivery.json](../internals/data/v165-full-delivery.json)。

從分支`fix/new-game-first-battle`的提交`caca2b8e`乾淨建置，建置時工作樹沒有未提交修改。來源1665檔；652個共用素材與v1.6.4逐檔相同，只有程式與說明檔改變。推廣片沿用v1.6.4，本版未重錄。

## 封包

| 平台 | 檔案 | bytes | SHA-256 |
|---|---|---:|---|
| Windows amd64 | `LOGH3-v.1.6.5-20261102-windows-amd64.zip` | 372490553 | `afe51a9ecb06939ff0fa6cf7763a0dc6723c498578df5043131bb3ab50a34e62` |
| Linux x86_64 | `LOGH3-v.1.6.5-20261102-linux-x86_64.AppImage` | 372476408 | `70d017ff94aa948fa00f610944dc2726148655e00d716463e16dcf29c228e9d2` |
| macOS通用 | `LOGH3-v.1.6.5-20261102-macos-universal.zip` | 382393230 | `628aac23a6f780a1f245294806ace2cfe12518b80b785a3c5c161088babf93ac` |

位置在`dist-all/v.1.6.5-20261102/full/`，只限本機，未公開上傳。

## 驗收

九條都是實際封包的正常入口，輸入只有滑鼠與鍵盤。AppImage走自動素材與預設XDG存檔，Windows用ZIP解出的`gin3.exe`。

| 路徑 | 平台 | 結果 |
|---|---|---|
| 新遊戲帝國：星域15開戰、九回合、存讀檔、再開戰 | AppImage、Wine | 通過 |
| 新遊戲同盟：星域13開戰、九回合、存讀檔、再開戰 | AppImage、Wine | 通過 |
| 同盟完整循環：三組海報、五語、五組名言300秒、F1／F7、存讀、重啟 | AppImage | 通過 |
| 帝國抽測：名言切換、v2／v3、F1／F7、存讀、重啟 | AppImage | 通過 |
| 同盟抽測，同上 | Wine | 通過 |
| 讀S1同盟：P選點、64點行軍、音效開啟打到第3回合 | AppImage | 通過 |
| 前線駐留、換提督、當地領船、C旗艦、出擊、存讀 | AppImage | 通過 |

新遊戲兩陣營的七個存檔檔案在AppImage與Wine之間逐位元組相同，進場鏡頭都看得到玩家旗艦。標準測試21套件16 PASS、5無測試、0失敗。

## 與v1.6.4的存檔差異

存檔回歸路徑是新遊戲直接存檔。五個表與v1.6.4逐位元組相同，`WORK`只差`+0C..14`：這九個byte是規格159刻意寫入的座位控制值，同盟列實際改變5個byte，帝國列改變4個。回歸基準由`tools/scripts/new_game_player_baseline.py`產生：v1.6.4的六表，只換這九個byte，其餘必須相同。

第一輪封包在目視時發現同盟進場鏡頭有誤，未交付，刪除後從修正後的提交重打。兩輪都用同一版號，該版號沒有tag或Release。

## 限制

- Windows只在Wine驗證，沒有實機。音效關閉，圖形用原版模式。
- macOS核對架構、簽章與素材，沒有真機。
- Linux用extract-and-run，沒有測FUSE。
- 對戰模式的第一場只有測試涵蓋，沒有走GUI。
- 不新增DOS同態聲明。#48仍開啟。
- 尚未打tag或發布。
