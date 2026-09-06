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

## 唯一 worklist

1. ~~計算並記錄原版 bundle 與 `GIN3.COM` 的 SHA-256~~（2026-09-06 完成，見上）。
2. 用 dosgolem 的通用探針（`cmd/probe`）對 `GIN3.COM` 做第一次接觸：能跑到哪、
   缺哪些 CPU 指令／DOS 服務／顯示模式。結果寫入 `docs/re/`。
3. 依第一次接觸結果決定：先補 dosgolem 覆蓋，或先建 DOSBox 擷取鏈拍第一批
   reference 畫面。
4. 建立 `tools/` docker 包裝器（go／dosbox），遵守 AGENTS.md 的容器規則。
