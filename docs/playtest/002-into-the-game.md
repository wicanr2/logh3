# 收據 002 — 開機到遊戲本體（same-state，可重現）

比較等級：**oracle 擷取**（尚無 remake 對應畫面，所以不是對拍收據；
這是「原版狀態可重現」的證據）。

## 固定序列

```
tools/oracle.sh key1 -steps 140000000 -tick 20000 \
  -clicks "300,170,40000000,3000000;300,170,95000000,3000000" \
  -press "1,3,8,enter" -press-at 105000000 -press-every 2000000
```

| 項目 | 值 |
|---|---|
| 執行器 | dosgolem-logh3 `6c569a4`（工作樹乾淨） |
| `GIN3.COM` SHA-256 | `770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918` |
| `screen.bin` SHA-256 | `6f41692755f983d0c2d1a2a5358c0a8124730cfc5455f6fe428c1b8040770da0` |
| `pal.bin` SHA-256 | `28b933fe81e79a380123b932302ba2dd8983bd3ac85a8f922e9e3b6ac1bd02e8` |
| 重現 | 獨立第二次跑（`key1-rerun`）兩份雜湊完全相同 |

## 內容

640×480 mode 12h：左側星域地圖、中央 796 年 1 月會計報告、右側狀態面板。
產物在 `workplace/oracle/key1/`（一次性證據，gitignore）。

## 為什麼這組數字長這樣

三個閘門的通過條件與步數見 `docs/re/08-boot-path-to-game-copy-protection.md`。
步數是硬等出來的上界，不是最小值；換執行器版本要重做（收據綁執行器版本）。
