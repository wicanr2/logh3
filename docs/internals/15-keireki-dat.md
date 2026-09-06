# KEIREKI.DAT — 人物傳

資料：`docs/internals/data/keireki.json`；可讀版本：`docs/lore/character-biographies.md`
產生：`tools/py.sh tools/scripts/keireki_export.py`
證據等級：版面 confirmed；姓名與陣營對應 confirmed。

## 版面

18,046 bytes，開頭是一張 uint16 LE 的位移表。**第一筆位移本身就是表的長度**，
所以筆數 ＝ `offsets[0] / 2` ＝ 83；最後一筆等於檔案長度，當終止哨兵。

```
+0000  A8 00        offsets[0]  = 0x00A8 ＝ 表長 ＝ 第 0 段的起點
+0002  F9 01        offsets[1]  = 0x01F9
...
+00A4  7E 46        offsets[82] = 0x467E
+00A6  7E 46        offsets[83] = 0x467E ＝ 檔案長度
```

`offsets[82] == offsets[83]`，第 82 段長度為 0，所以實際有內文的是 **82 段**，
正好對上 `ADMIRAL.DT*` 的 82 名提督。

每一段是一串 Big5（倚天）文字，以一個 `\0` 收尾，尾端用**全形空白**補到
固定長度——那是原版排版程式用的填充，不是內文。段內沒有換行控制碼，
斷行由遊戲自己算。

## 姓名不在這個檔裡

`KEIREKI.DAT` 只有內文。提督姓名在 `GIN3PS.EXE` 解包後的 64 KB 資料段裡，
`\0` 分隔的字串陣列，第 0 筆「萊因哈特」在 **0x61A8**，共 82 筆，順序與
提督索引一致。取得資料段的方法見 `docs/internals/14-screen-strings.md`。

82 個提督名之後接一張指標表，再往下 **0x6570** 起是 34 個旗艦名
（伯倫希爾、巴爾巴洛沙、人狼、托利斯坦……）。

| 區段 | 位址 | 筆數 |
|---|---|---|
| 提督姓名 | 0x61A8 | 82 |
| 旗艦名 | 0x6570 | 34 |

## 陣營

取自 `ADMIRAL.DTX` 每筆的 `+1`（0 ＝ 帝國、1 ＝ 同盟），把 82 人切成
0–39 與 40–81 兩半，見 `08-admiral-dt.md`。

## 中譯

中文版由微波軟體（Micronet 在台代理）中文化，人物傳是當年的翻譯成果，
逐字保留在 `docs/lore/character-biographies.md`，未經改寫。
