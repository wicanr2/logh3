# 規格

`AGENTS.md` 的流程：

```
RE 證據 → DRAFT spec → 證據審查 → READY → implementation → 同狀態驗證 → CONFORMED
```

只有標 `READY` 的規格可以動 production path；標 `CONFORMED` 表示實作完成
而且有同狀態收據撐著。每一份規格的頂端都要寫現在是哪一個狀態。

`docs/internals/` 是逆向出來的事實，`docs/re/` 是怎麼查到的，
這裡是**要照著做什麼**——把事實翻成 remake 的行為與驗收條件。

| # | 規格 | 狀態 |
|---|---|---|
| [`001`](001-region-summary-panel.md) | 星域摘要面板（左下角）| `CONFORMED` |
| [`002`](002-capital-panel.md) | 右側面板（日期、錢、自己的首都）| `CONFORMED` |
