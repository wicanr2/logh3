# 徵稅／月初預算箭頭與結算的完整畫面對拍

日期：2026-10-01。工作項 [#80](https://github.com/wicanr2/logh3/issues/80)。
比較等級：`same-state`，S1 帝國 796 年 1 月 1 日，正常玩家路徑編成第一艦隊後開徵稅計劃。
兩側邏輯畫布 640×480，原生大小比較，沒有裁切、遮罩或排除矩形。
規格：[003](../spec/003-budget-panel.md)。研究：[350](../re/350-plan-arrows-retain-dim-colour.md)。

## 來源與可重現入口

- 原版 bundle：`org_game/银河英雄传说3SP.zip`，SHA-256
  `de905aa0b5a666aa9d31dce36f43a5584f1842950c39d7e1b1f97c72e0d79673`。
- 原版 `GIN3.COM`：SHA-256
  `770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918`。
- 執行器：專用副本 dosgolem-logh3，HEAD
  `f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c`，工作樹乾淨，`golang:1.24-bookworm`。
  未修改執行器或原版資料。初始完整檢查點 `workplace/fcmd/fcmd-plan-tax.state`
  SHA-256 `f2f688370e91bb8434ca002018f600538613e7d41c4791f5872a8a225bf65e8d`。
- remake 擷取時 HEAD `7b1b227ad57a1acc3f3f74df085308f75e01c824` 加本輪未提交修正，
  image `logh3-ebiten:2`。核心檔 SHA-256：`game.go`
  `9b53e9a1d33752b5a28bbf1d3d66257ff632f998b348a5bcfde6770b16483ca1`；`planscreens.go`
  `6ab2ff144137d5da7649929086ca2ca2c84b72ecbf9cd8da6cedc247406cbaf3`。
- 分析與 PNG 差分：`python:3.12-slim`，全部在容器內。
  箭頭操作不含亂數判定；原版從完整檢查點還原 RNG，remake 從同一劇本正常入口走入。
  未重新抽種子或挑選結果。

```sh
bash tools/plan_arrow_receipts.sh
bash tools/gui_check.sh
tools/py.sh tools/scripts/plan_arrow_stats.py
tools/gui.sh test ./internal/... ./cmd/...
```

原版腳本共 14 份，其中徵稅 9 份、月初 5 份；全部重錄兩次，screen.bin 的 SHA-256 相同。
原版每份 `flags.txt` 記完整時間、按鍵、滑鼠及保持步數，`inputs.json` 記父檢查點雜湊。
徵稅單擊發生於第 462,000,000 步，保持 100,000 步，終點 490,000,000 步；
滑鼠座標換畫面座標為 y＋40。其他生命週期的時間與輸入由腳本固定。

remake 使用 `cmd/gin3` 正式入口與 renderer，`-script` 注入等價的輸入事件，
沒有跳過主選單、防拷問答、月初報告、艦隊編成或直接進畫面；`-shot-after 2`。
這是自動化玩家路徑，不宣稱人工操作手感或作業系統滑鼠事件已另行驗收。
八份的完整序列由 `tools/gui_check.sh` 單一來源保存，亦寫入 `remake-script.txt`。
腳本每次擷取自動更新 `remake-meta.json`，不沿用前次版本；只重跑本輪八份可用
`LOGH3_GUI_PREFIX=planarrow- bash tools/gui_check.sh`。
文字仍使用相同 Big5 素材；此次僅改四顆箭頭的色號，不改語言或輸入映射。

## 完整畫面結果

以下每張均比較 307,200 像素，相同 307,200、不同 0、最大 RGB 通道差 0，
差異邊界為空。表中雜湊是雙方相同的色號陣列；PNG 編碼不同，但解碼後 RGB 完全相同。

| 批次（`planarrow-` 前綴） | 狀態 | screen.bin／remake.bin SHA-256 |
|---|---|---|
| tax-up | 稅率＋ | `46a34b92b76a6ee54a0439349a01dab9a6eb670ec1a9fd73b8c561a828ae05ce` |
| tax-down | 稅率－ | `9e43598a042ce8cc30d6a57ad33f55a0d9b5a4dab6e267b2e5e15652756c70f1` |
| pct-up | 建造費＋ | `5a2de892f446394e62f3f09ac53c5bff87283e2f9a4b8dba59291b7e44540088` |
| pct-down | 建造費－ | `940c54e9920caeb399c87302472ed322127cf132ccc3755b578e8c0905b276f0` |
| hover | 懸停不轉色 | `e67311e83f45448c1553e95f66712a6be763afe3d65dd4c2df45d0b1489f9f4b` |
| moved | 點擊後移開仍保留 | `e19004cdd2c55d3e770ceac1a933bc740bf0635f57042f1c208203103cca1bad` |
| updown | 先加後減，各顆保留 | `06e7f2c6c9b24f99cfc73978e0f8360b440150fe4e936d624add455a532e5b3c` |
| reopen | 關窗再開重設 | `f707bf502f1607b7fa1ebf532f0dfc0c42c4131f9f16ae5a0f447b5e69faca64` |

每份另比較預算面板 (128,60,384,340) 130,560 像素、四顆箭頭各 320 像素，
結果也都是零差異。色盤前 16 色及 PNG 顏色相同；`pal.bin` SHA-256
`28b933fe81e79a380123b932302ba2dd8983bd3ac85a8f922e9e3b6ac1bd02e8`。

原版、remake 與差分圖保存在 `workplace/oracle/planarrow-<批次>/` 的
`screen.png`、`remake.png`、`diff.png`。`stats.json` 保存全部輸入／PNG／色號陣列雜湊
與逐區統計；`meta.txt`、`flags.txt`、`inputs.json`、`remake-meta.json` 保存版本及參數。
這些為可重生的本機素材，不進 Git。每份全黑差分圖 SHA-256
`33806e3d5c67f100a92295c93044bbb67d6cea90060327b69d682ebd283c87d2`。
例如 tax-up 原版 PNG SHA-256
`15b4ae196a62ab439bb480922c8108a2404c408e58256ab5313a54800cc05516`，remake PNG
`cdece9d19d480bbbc12923d07b61f0b6818c5e18a39b2fd8ca59571fdacbbd88`。

上界另有一份明示注入診斷：`planarrow-clamp`、原版
`-poke lin:5DB04@451000000=04`，remake 只把檔位設為 4 後點擊。
`TestPlanArrowRenderingMatchesOriginal/上界仍轉色` 整張相同，原版 screen.bin SHA-256
`5022a517d85f68acf0881a0d74fb172ba521a69e2576a467ef3133cd24a6b602`。
它不冒充正常入口證據。

## 修正與範圍

已證實原版點擊會將該箭頭的全部色號 9→8，各顆獨立保留，關窗再開重設。
原先 17–21 像素未解差異已消除，`planArrowSkip` 移除；三份原判準收據也改成整張比。
造成轉色的內部呼叫仍為 `unknown`，不影響此有限呈現修正。

全部產品套件測試通過，既有 40 份 GUI 驗收加本輪 8 份，共 48 份通過。
既有 40 份中四份仍沿用其原有規格列明的排除區；本輪八份無任何排除。
`tools/go.sh test ./...` 則因 Go image 缺 X11／ALSA 連結庫與既有 gitignore
探索測試引用過期符號而失敗，屬工具／探索環境問題；未修改無關檔案，改用既有 GUI
image 對所有正式產品目錄乾淨重跑成功。

第一階段的月初五份收據當時只有原版重播；其收稅時序缺口由 #82 接續處理，
後續結果如下。按住連發仍是 #60。

## #82 月初互動與真正結算

同日接續驗收：S1 帝國正常主選單 → 防拷問答 → 會計報告 → 預算；
四顆箭頭各一次，或不調整；預算結束 → 會計／訂造報告右鍵 → 命令階段。
比較等級固定 `same-state`，無記憶體注入、裁切、縮放或排除矩形。
原版工具與輸入雜湊沿用上節；月初完整檢查點 SHA-256
`ee6138b6abbc29e89de5f04a4d5480b3fcfbcc75c546ab440c8f37b2f69e6e6c`。
remake GUI 擷取為 HEAD `2a337de` 加本輪修正、`logh3-ebiten:2`；
每份 `remake-meta.json` 記擷取時工作樹、完整腳本及核心程式 SHA-256，
受版控索引見 [結算收據 JSON](../internals/data/budget-settlement-receipts.json)。

```sh
LOGH3_GUI_PREFIX=planarrow-budget bash tools/gui_check.sh
bash tools/py.sh tools/scripts/plan_arrow_stats.py
bash tools/go.sh test ./internal/game -run TestBudgetSettlementMatchesOriginal -count=1
```

十份 GUI 收據是 `planarrow-budget-base`、四份 `planarrow-budget-{tax-up,tax-down,pct-up,pct-down}`，
以及五份 `planarrow-budget-{unchanged,tax-up,tax-down,pct-up,pct-down}-post`。
每份色號及 RGB 均 307200 像素全同；面板及四顆箭頭逐區零差異、最大通道差 0、
差異邊界空；原版／remake／全黑差分圖、統計及輸入保存位置與前節相同。

五份 `-post` 的完整記憶體、畫面及色盤第二輪重播雜湊全部相同；
數值、七種建造艘數與原版記憶體雜湊見 [研究 351](../re/351-budget-settlement-after-reports.md)。
remake 逐筆核對軍事費、原定軍事費、實付、下期預估、稅率快照、比例、
32 個星域的經濟力／成長率／當期及前期稅率，以及 81 顆行星經濟力，全部一致。
報告期間未扣行星經濟力；訂造報告退場只扣一次；重複右鍵不再次收稅。
結算後存檔的行星 bytes 保留，從正常「取進度」入口讀回不重開待結算。

兩陣營開局、既有徵稅及原版 `monthwork2` 跨月結算回歸通過；
全部產品套件由 `tools/gui.sh test ./internal/... ./cmd/...` 通過。
標準 Go image 的 X11／ALSA 連結與歷史 `workplace/tmp` 失敗仍屬環境／探索輸出，
本輪再次核對，不歸為本次產品缺陷。

**限制**：整張星域／行星表仍有財政之外的 64／759 bytes 未分類差異，
五組位置相同，已另列 [#83](https://github.com/wicanr2/logh3/issues/83) 與
[差異索引](../internals/data/budget-table-differences.json)。本次只宣稱上述財政欄位與畫面一致，
不宣稱整表或全遊戲完成；對戰跨月、按住連發及其他未取證分支維持原有限制。
