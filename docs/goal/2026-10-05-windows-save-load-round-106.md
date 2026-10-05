# 第106輪：Windows 正常保存與讀回

日期：2026-10-05。工作項目沿 [#69](https://github.com/wicanr2/logh3/issues/69)，保持 OPEN。
開工來源 `91d85785b937145f060066cf9bb29edc643a7c05`；工作樹乾淨。
最新 Issue 快照為 89 張、6 OPEN／83 CLOSED；verify 269 條、0 斷言不成立。

知識路由命中 Wine／Xvfb／正常玩家路徑與存讀檔，入口為
`~/.codex/knowledge-base/sources/claude/retro-cht/retro-game-playtest.md`。
文件職責依 `~/.codex/knowledge-base/local/project-document-responsibilities.md`。
沿用逆向 remake 技能及既有保存規格 005／006／007／009／067。

- 使用第103輪來源重建的 Windows 玩家程式，預設片頭、真實 X11 輸入，無腳本或局面注入。
- 正常新遊戲編艦隊，公開保存到中文及空白路徑；退出後重新啟動，從主選單讀回及重存。
- 原版資料與玩家程式唯讀；對照同來源 Linux 公開輸入及全部六檔，保留載入重算的實際差異。
- 保存完整視窗、差分、操作時戳、來源身份、原版輸入身份與容器版本。
- 新增收據與核對器，不改凍結的第103／104／105輪文件及來源索引。

本輪只驗 Wine 的 Windows 玩家存讀範圍，不宣稱 Windows 實機、新增原版整回合同態或正式發行完成。
人耳與完整音樂循環、macOS、版號／dist-all 及其他玩法閘門保持待完成。

正常退出觀測發現顯示層漏接 Game.Quit，已先開 [#90](https://github.com/wicanr2/logh3/issues/90)。
沿研究341既有原版證據審查117至READY，只補退出接線；修正前失敗批次與回歸保留。
