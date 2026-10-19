# 第142輪：全局提督能力編輯

使用者要求在系統設定加入提督編輯，配合現行遊戲樣式。可調整十項基本能力、三項內部能力，以及旗艦擊毀換艦率、全軍移動加成與全軍攻擊加成。使用者核准草稿，先答攻擊50%，隨即更正為30%；全局共用，新遊戲也套用。正式契約為[149](../spec/149-admiral-editor.md)。

## 本輪目前狀態

| 項目 | 2026-10-19狀態 |
|---|---|
| 目前產品 | 工作基底`26338ceb0dc8c299a2401f9653a4041e595dd51d`，149限定CONFORMED，原始碼與驗證完成 |
| 介面 | 正式設定頁增加入口，保留兩陣營色、立體框與三分頁配置 |
| 基本與內部能力 | 十三欄均可逐人編輯；套用、取消、還原及派生重算已接 |
| 特殊能力 | 逐人換艦率；兩位名將移動50%、攻擊30%；坎普與波布蘭任提督／參謀時本艦隊戰鬥機攻擊30%，其他人預設無此加成，均可編輯 |
| 交付物 | 正式程式、可重生工具與本機GUI收據；沒有新發行包 |
| 正式功能 | 全局稀疏覆寫存入既有偏好檔，新局／讀檔套用；六表格式保持 |
| 既有Issue | 開工快照102張，只有#48 OPEN；282條verify無斷言不成立。本輪未寫入Issue |
| 最近測試 | 最後標準19套件15 PASS／4無測試；r6兩陣營、四項特殊能力、無效值保護、取消／套用／重啟及S1讀回通過 |
| 交付限制 | 本輪不重新打包、不發布；#48保持OPEN，沒有新增DOS全流程或其他平台實機聲明 |

## 已有用途證據

依[內部能力表](../internals/08-admiral-dt.md#沒有名字那三格在做什麼)及[研究287](../re/287-admiral-remaining-fields.md)，三欄位的消費用途已記錄，但原版欄名仍未知。本輪核對既有研究及正式程式，沒有重新執行原版或新增原版parity聲明。

- `+20`：稅率、治安分段、AI艦隊類別評分、低完整度時的戰術判斷。
- `+21`：內政參謀評分、稅率、戰術AI權重。
- `+22`：攻擊目標選擇，較高值增加選中最高評分目標的機率。

介面使用用途說明，不把自訂標籤當成原版欄名。三欄均可編輯，輸入與重算範圍依149。[名將規格140](../spec/140-hero-tactics-and-release.md)原先只有戰術協同與50%移動增幅；149的獨立攻擊倍率屬remake新增規則。

## 核准草稿原圖

以下保留核准當時的原圖，攻擊欄的「待定」已由使用者另行裁決為30%。原圖不重畫、不當作正式執行截圖。

- [帝國基本](../images/admiral-editor-approved-empire-basic.png)、[內部能力](../images/admiral-editor-approved-empire-hidden.png)、[特殊能力](../images/admiral-editor-approved-empire-special.png)。
- [同盟特殊能力](../images/admiral-editor-approved-alliance-special.png)、[設定入口](../images/admiral-editor-approved-empire-settings.png)。

## 初版草稿重生與驗證入口

[產生器](../../tools/scripts/admiral_editor_preview.py)在既有`logh3-hd-preview:20261011-r1`映像執行：

```sh
LOGH3_GUI_IMAGE=logh3-hd-preview:20261011-r1 LOGH3_XVFB_SCRIPT=tools/with-xvfb.sh tools/gui.sh xrun python3 tools/scripts/admiral_editor_preview.py
```

輸出只在既有`workplace/`。`manifest.json`記錄素材／研究／目前程式與產生器SHA-256、瀏覽器版本及操作檢查。草稿不代表正式玩家路徑、保存相容性或原版同狀態驗收。

最終五張草稿的瀏覽器檢查通過，涵蓋82名提督、三分頁、數值微調／輸入、取消確認、還原、頁內套用、設定頁往返、選中提督可見及控制矩形。已目視兩陣營特殊能力、帝國基本／內部能力及設定入口；產品測試結果與環境修正記於[WORKLOG](../../WORKLOG.md#2026-10-19-第142輪提督編輯介面草稿)。

## 正式功能與最後驗收

[驗收130](../playtest/130-admiral-editor.md)、[唯一來源索引](../images/admiral-editor.json)及[收據核對器](../../tools/scripts/admiral_editor_receipt.py)保存最後結果。正式截圖：[帝國](../images/admiral-editor-empire.png)、[同盟](../images/admiral-editor-alliance.png)、[內部能力](../images/admiral-editor-internal.png)。

最後批次`workplace/admiral-editor-gui-r6/`從正常片頭至S1操作兩次啟動，核對執行中ELF、十一張完整960×720畫面、無效武器能力1拒絕、取消、全局套用與重啟讀回。新增戰鬥機加成依使用者確認本人艦隊30%，並讓參謀觸發；兩角色取較高值，僅強化模式武器2套用。正常GUI驗坎普及波布蘭預設30%並修改保存，坎普65%跨重啟與讀檔保持。README補齊可編輯範圍、操作、預設與五張正式圖。

最後二進位為`workplace/admiral-editor-gin3-r6`，僅本機驗證產物。先前r4、r5-r2索引保存在各自目錄的`exported-receipt.json`，r5-r2另存`standard-tests.log`，不把舊版升格。新增畫面：[坎普](../images/admiral-editor-kemp.png)、[波布蘭](../images/admiral-editor-poplan.png)。
