# ENDING.EXE — 結局文字與職員表

資料：`docs/internals/data/ending-text.json`；可讀版本：`docs/lore/ending-and-credits.md`
產生：`tools/py.sh tools/scripts/ending_export.py`
證據等級：字串區版面 confirmed；`.VO5` 內容未解。

## 檔案

64,677 bytes，**未打包的 MZ**（不像 `GIN3PS.EXE` 有 GLUE 外殼），
所以字串直接搜得到。文字區在 **0x63FE–0x765C**，`\0` 分隔，
Big5（倚天）編碼。

| 區段 | 位址 | 內容 |
|---|---|---|
| 結局旁白 | 0x63FE | `【帝國勝利】`、`【同盟勝利】` 兩段，各自的旁白緊接在標記後 |
| 尾聲對白 | 0x659B | 萊因哈特／吉爾菲艾斯、楊威利／尤里安各一段 |
| 職員表 | 0x666C | 每行 0x4D bytes 的固定寬度，用全形空白排版 |
| 演出名單 | 0x6F00 一帶 | 提督的羅馬拼音，兩行一位（名／姓），順序同 `ADMIRAL.DT*` |
| 腳本檔清單 | 0x75CD | 18 個 `.VO5` 檔名（小寫，不含副檔名）|

程式在 0x63F6 有 `".VO5"`、0x63FB 有 `"rb"`、0x63F1 有 `"%s\n"`——
`sprintf` 出檔名再 `fopen` 的典型寫法。

## 中文化的痕跡

職員表裡有三行是中文版才有的：

```
Ｃhinese Ｖersion Ｄistribution 亨和企業．微波軟體
Ｃhinese Ｍessage  Ｔranslation 邵  程衍
Ｃhinese Ｖersion Ｐrogramming Ａndrew Ｌiao
```

外加 `(C)1994 亨和企業．微波軟體` 一行。這是整包遊戲檔案裡唯一記下
中文化工作者姓名的地方。

日文原版的「ミラージュ」在中文版寫成「ミラㄧジュ」——Big5 沒有片假名
長音符，中文化時拿字形一樣的注音符號「ㄧ」頂替。

## `.VO5` 未解

50 個 `.VO5` 檔（`ENDA10.VO5` 等）是二進位，開頭 `50 C8 00 00`，
不是 `.AMG` 的檔頭（`.AMG` 是 `01 00 00 00`）。內容看起來有
「重複標記」的壓縮結構，但解碼常式還沒定位。**未解**。
