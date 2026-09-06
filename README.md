# logh3 — 銀河英雄傳說III SP DOS remake／研究專案

《銀河英雄傳說III SP》（Legend of the Galactic Heroes III SP）PC DOS 中文版的
重製與保存專案。原版 1993 年由 BOTHTEC／QUEST 與 Micro Vision 製作，
1994 年由亨和企業．微波軟體發行中文版。

專案有兩個並行的目的：把遊戲重做成可維護的版本，以及保存這款 1994 年台灣
中文化作品的技術細節——原始碼與開發資料都不存在，只剩執行檔可以問。

## 現況

主選單已達成**逐位元組一致**的同狀態對拍：remake 畫出來的 640×480 色號陣列
與調色盤，跟原版在無頭執行器裡跑出來的收據 SHA-256 相同。

完成度看 [`VERIFICATION-MATRIX.md`](VERIFICATION-MATRIX.md)。

## 怎麼跑

需要自備原版遊戲檔（本儲存庫不含原版檔與素材），放在 `org_game/extracted/`。

```
tools/go.sh test ./...                                   # 含對拍回歸測試
tools/go.sh run ./cmd/logh3 -font /orig/GRAPH.IMG -out workplace/remake
tools/go.sh run ./cmd/oradiff \
  -oracle workplace/re06/screen.bin -remake workplace/remake/screen.bin
```

原版側的畫面由 [dosgolem](https://github.com/wicanr2/dosgolem)（無頭、決定性的
DOS 執行器）產生，它是本專案唯一的 oracle。

## 文件

| 檔 | 內容 |
|---|---|
| [`AGENTS.md`](AGENTS.md) | 協作規範與對拍契約 |
| [`CONTEXT.md`](CONTEXT.md) | 目前真相與唯一剩餘工作 |
| [`VERIFICATION-MATRIX.md`](VERIFICATION-MATRIX.md) | 完成度的唯一數字來源 |
| [`docs/internals/`](docs/internals/) | 逆向出來的事實（格式、結構、演算法），含機器可讀的 `data/*.json` |
| [`docs/re/`](docs/re/) | 逐次調查紀錄 |
| [`docs/playtest/`](docs/playtest/) | 對拍收據 |
| [`WORKLOG.md`](WORKLOG.md) | 流水帳 |

## 授權

未定。原版遊戲檔與素材不屬於本專案，也不隨本專案散布。
