# 018 — 失蹤提督 fate pass 自動驗收

日期：2026-09-14
比較等級：控制流／狀態轉移的固定 seed 收據；畫面為 `reference-only`，
不宣稱逐像素一致。

## 原版

- 原版 `GIN3.COM` SHA-256：
  `770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918`
- 執行器：dosgolem-logh3 `e0b6a03`，工作樹乾淨
- 批次：`workplace/oracle/admiral-fate-survive/`
- 初始狀態：新遊戲正常路徑進第一次「實行」
- 注入：提督 0 狀態 `lin:602AC=80`；fate pass 入口前
  `lin:5E00C=00 00 00 00`
- 結果：`056A:0219` 命中一次；第一個 `rand()` 回 0；
  `056A:0311` 以提督 0、死亡旗標 0 命中一次

完整命令、事件時間、實際線性位址與 RNG 寫入見本機收據的
`flags.txt`、`meta.txt`、`report.txt`。本次沒有把 framebuffer hash 當畫面
驗收，原因見 [`docs/re/307`](../re/307-admiral-fate-pass-and-report.md)。

## remake

Docker 自動測試覆蓋：

- 原版兩階段門檻、特殊提督與共用 LCG；
- 生還／死亡的精確記錄寫回；
- 人類座位報告過濾；
- 左、右鍵逐則退場與 callback；
- 四段原始 Big5 組字資料存在；
- 完整 `internal/rules`、`internal/screen`、`internal/game` 回歸通過。

本切片達到 READY 實作與 remake 內部自動驗收，但因原版畫面收據不可用，
SPEC 048 不升為 CONFORMED。
