# 艦隊組成多編組對拍收據

日期：2026-09-13  
原版 bundle：`银河英雄传说3SP.zip`，SHA-256
`de905aa0b5a666aa9d31dce36f43a5584f1842950c39d7e1b1f97c72e0d79673`  
`GIN3.COM` SHA-256：`770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918`  
dosgolem HEAD：`e0b6a03c5338ea4aa5676e81c5a19825d847ab51`，工作樹乾淨  
原版顯示模式：VGA `12h`，640×480、16 色

## 原版狀態與輸入

收據 `fmx1` 沿用 `fok` 的正常玩家路徑，走到艦隊組成畫面；完整可重播旗標保存在
本機 `workplace/oracle/fmx1/flags.txt`，`meta.txt` 同時綁定執行器版本。畫面初始化
期間以 `lin:` 位址把艦隊 0 注入為 3／8 組：

| 組序 | 艦艇池索引 | 艦種 | 隊形槽 | 畫面 |
|---:|---:|---:|---:|---|
| 0 | 0 | 0 | 20 | Ｂ、格盤 (4,4)、圖塊 36 |
| 1 | 0 | 2 | 10 | Ｃ、格盤 (2,2)、圖塊 38 |
| 2 | 0 | 4 | 30 | Ｄ、格盤 (6,6)、圖塊 40 |

`report.txt` 記錄每一筆 `lin:` 位址的實際位址與指定值相同。艦隊記錄傾印為
122 bytes，工作表傾印為 230 bytes；工作表的 `ds:984B/984C` 是 `3/5`，且
`ds:981E` 壓實的三筆與艦隊 `+0x10` 起三筆逐位元組相同。

## 雜湊

| 本機證據 | SHA-256 |
|---|---|
| `workplace/oracle/fmx1/screen.bin` | `d49b0d7e02bd4635c41404000052606077308dd1ee501bff82c9d8fca0837c1d` |
| `workplace/oracle/fmx1/pal.bin` | `28b933fe81e79a380123b932302ba2dd8983bd3ac85a8f922e9e3b6ac1bd02e8` |
| `workplace/fmx1_work.bin` | `6ee17025e9e91111a39c2ba2b2d0b17f49270cd7b6597d0e342ba2d8b3911d28` |
| `workplace/fmx1_fleet.bin` | `942c4e20b36d93ef237efc452bd026b1add5dc9b99b3b9bd2aae31f35b72282e` |

## 自動比較與裁決

```sh
tools/go.sh test -count=1 ./internal/game \
  -run 'TestFleetComposition(MultipleGroups)?MatchesOracle$'
```

`TestFleetCompositionMultipleGroupsMatchesOracle` 不硬編第二份 remake fixture：它讀取
上述兩份記憶體傾印，先核對 `3/8`、壓實順序及三個艦種／槽，再用正式字型、
`CompositionPanel` 與 `SPRITE.SPR` renderer 重畫。比較範圍是規則控制的每一格：

- 格盤 `(136,168)`–`(279,311)`：9×9、20,736 像素；
- 編隊列 `(160,342)`–`(287,357)`：八格、2,048 像素。

合計 22,784 像素，未核准差異為 0。部隊數由兩份原版傾印交叉核對為 `3/8`。
游標壓在部隊數附近，該處會隨 `1/8` 變 `3/8` 而出現 88 個游標／文字疊圖差異；
它不在本 issue 的字母與格盤規則範圍，因此沒有把這份區域對拍升格為整張畫面一致。

裁決：#10 要求的第二個多組局面已具可重生原版收據，B／C／D、五個空格與
三個格盤位置逐格相同；證據等級為指定區域的 `same-state`。
