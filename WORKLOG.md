# WORKLOG — 銀河英雄傳說III SP

唯一流水帳。新条目加在最上面。

## 2026-09-06

- dosgolem 第一次接觸（詳 `docs/re/01-first-contact-gin3com.md`）：
  `GIN3.COM` 是 622-byte launcher；dosgolem 新增 `LoadCOM` 後原版路徑跑 35 道
  指令停在 EMS 偵測（`EMMXXXX0` 字元裝置開啟失敗，exit 3）。patch 版證實
  EXEC 鏈：`OPEN.EXE` → `GIN3PS.EXE`（本體）→ `ENDING.EXE`。
  缺口：字元裝置開啟、int 67h EMS、EXEC（4B00/4D）。
- 建立 GitHub repo `wicanr2/logh3`（private）並推送。
- 從 dos.zczc.cz 下載 js-dos bundle（SHA-256
  `de905aa0b5a666aa9d31dce36f43a5584f1842950c39d7e1b1f97c72e0d79673`），
  存入 `org_game/`（gitignore），解出至 `org_game/extracted/`。
- 建立專案骨架：`AGENTS.md`（參考 sangokushi 協作規範 + dosgolem 對拍方法論）、
  `CONTEXT.md`、`.gitignore`、`docs/{playtest,spec,re}/`、`workplace/`。
