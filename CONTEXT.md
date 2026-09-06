# CONTEXT — 銀河英雄傳說III SP 專案目前真相

本檔是目前真相與唯一剩餘工作。流水帳在 `WORKLOG.md`；協作規範在 `AGENTS.md`。

## 現況一覽

- 2026-09-06：專案建立。原版 js-dos bundle 已下載並解出（僅本機，gitignore）。
- 遊戲：《銀河英雄傳說III SP》（Micronet，1994），PC DOS 版，啟動程式 `GIN3.COM`。
- 原版 bundle：`org_game/银河英雄传说3SP.zip`，
  SHA-256 `de905aa0b5a666aa9d31dce36f43a5584f1842950c39d7e1b1f97c72e0d79673`，
  來自 <https://dos.zczc.cz/games/银河英雄传说3SP/>。
- `GIN3.COM` SHA-256 `770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918`。
- bundle 基準機器設定（`.jsdos/dosbox.conf`）：`machine=svga_s3`、`memsize=16`、
  SB16、autoexec 為 `mount c .` + `GIN3.COM`。
- 目前沒有任何已對拍畫面、已逆向格式、已驗證規則，也沒有 remake 程式碼。
- `GIN3.COM` 尚未在 dosgolem 執行器內跑過；DOSBox 擷取工具鏈（`tools/dosbox.sh`）
  尚未建立。

## 已被推翻的斷言

（無）

## 第一次接觸結果（2026-09-06，`docs/re/01-first-contact-gin3com.md`）

- `GIN3.COM` 只是 622-byte launcher，非遊戲本體（confirmed）。
- dosgolem 已可載入 .COM：`internal/machine/loader.go` 新增 `LoadCOM`，
  `cmd/probe` 依 MZ magic 自動分派（改動在 dosgolem 工作樹，尚未 commit）。
- 原版路徑跑 35 道指令後主動結束：開字元裝置 `EMMXXXX0` 偵測 EMS 失敗，
  印出 `EMS Driver not found...` 結束（exit 3）。CPU 無未支援指令。
- patch 版（跳過 EMS 檢查）證實 launcher 依序 EXEC：
  `OPEN.EXE` → `GIN3PS.EXE`（189,536 bytes 未打包 MZ，遊戲本體，尾巴選 `9 1`/`9 2`）
  → `ENDING.EXE`；目前 EXEC 是假成功。
- 阻塞缺口（依序）：字元裝置開啟、int 67h EMS（AH=40/42，需 ≥8 頁）、
  int 21h AX=4B00h EXEC、AH=4D 取回碼；潛在 AH=63h DBCS、int 2Fh AX=4F01h
  （疑似 DOS/V，假說）。

## 唯一 worklist

1. ~~計算並記錄原版 bundle 與 `GIN3.COM` 的 SHA-256~~（2026-09-06 完成，見上）。
2. ~~dosgolem 第一次接觸~~（2026-09-06 完成，見上節）。
3. dosgolem 實作 EXEC（int 21h AX=4B00h + AH=4D）：先寫 DRAFT spec 進
   dosgolem `docs/spec/`，READY 後才實作。目標：讓 launcher 原生路徑（含 EMS）
   跑進 `GIN3PS.EXE` 本體。
4. 評估 EMS（int 67h）最小實作範圍——launcher 需要 ≥8 頁回報才繼續。
5. 本體 `GIN3PS.EXE` 跑起來後，用 dosgolem `Save`/framebuffer 拍第一批
   same-state oracle；DOSBox 擷取鏈（`tools/dosbox.sh`）作為交叉驗證再建。
