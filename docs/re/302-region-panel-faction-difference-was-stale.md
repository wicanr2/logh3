# 302 — 左下摘要沒有陣營相反規則

日期：2026-09-14

狀態：**confirmed**（目前 dosgolem 重錄、既有收據雜湊與 remake 正常玩家路徑）

## 舊結論

[`278`](278-region-panel-frozen-content-differs-by-faction.md) 用 dosgolem
`6841e68` 量到：同一組表面輸入下，帝國 `cmd1e` 左下摘要有字，同盟
`cmd1eAlly` 卻是空的，因此留下「可能有奇偶性或陣營分支」的假說。

後續 [`282`](282-the-translation-dropped-the-click-pre-move.md) 已找出執行器
漏送「按鍵前先移動游標」事件；[`284`](284-the-left-panel-is-sticky.md) 也在
修復後記錄帝國 `cmd1e` 應是空白，但 #4 與程式註解仍保留舊假說。

## 新證據

以專案固定副本 `~/cht/dosgolem-logh3`、HEAD `e0b6a03`，照 #4 的同盟輸入
重錄 `rgn4AllyCorrect`：

```text
-tick 20000 -press 1,3,8,enter -press-at 105000000 -press-every 2000000
-steps 230000000
-clicks 40000000:300:190,95000000:300:190,180000000:352:85
-rclicks 130000000:300:200,155000000:300:200
-mouse-x 600 -mouse-y 300 -mouse-at 210000000 -click-hold 3000000
```

| 收據 | dosgolem | 左下 `(15,356)`–`(312,399)` 非色 9 像素 | `screen.bin` SHA-256 |
|---|---|---:|---|
| `cmd1e`（帝國）| `e0b6a03` | 88（只有外框）| `e2853bce13482656b6a8ba2d6683b7ec0a118e1566b2dc473211de0d49ecc78d` |
| `cmd1eAlly`（同盟）| 舊檔 `2833ec9` | 88（只有外框）| `0fb736a6b6d25a2f6d210b95113127a6d4590bfa3c118c4d7d45645d9782f4b8` |
| `rgn4AllyCorrect`（同盟重錄）| `e0b6a03` | 88（只有外框）| `0fb736a6b6d25a2f6d210b95113127a6d4590bfa3c118c4d7d45645d9782f4b8` |

同盟新重錄與舊 `cmd1eAlly` 整張逐位元組相同；帝國與同盟的左下區也都
只有外框。差異不是 faction 分支，而是舊帝國收據缺少點擊前移事件。

## remake 驗證

`TestAllianceCommandPhaseScreens` 不再讓十二張共用「先把滑鼠搬到星圖外」
的串接狀態。主畫面、命令選單與子畫面各自從正常玩家路徑重走，輸入順序
與各自收據相同；十二組逐位元組比較全部通過。原始定位與舊觀測完整保留在
`docs/re/278`，本文件只追加勘誤，不重寫其形成原因。
