# 防拷問答不驗答案

執行器 dosgolem-logh3 `20e4583`。

## 收據 `cbwrong`：打 `000` 照樣進遊戲

```
-tick 20000 -steps 200000000
-clicks 300,170,40000000,3000000;300,170,95000000,3000000
-press 0,0,0,enter -press-at 130000000 -press-every 2000000
-rclicks 300,200,165000000,3000000;300,200,180000000,3000000
```

`0`、`0`、`0`、Enter 送進防拷問答之後，最後的 VGA framebuffer 是

```
screen.bin SHA-256: 8233320898cdbae7e1137bc9b4ab10ac3429af70ce0dc845668f1cd01c2282ae
```

而這個雜湊與**答對**那條路徑的收據完全相同——`main1`、`rcback`、
`dsFull`、`dsNames`、`fleetEmp`、`main1-rerun`、`rgnEmp` 全都是同一個值，
它們送的是 `-press 1,3,8,enter`。兩條路徑的差別只有那三個按鍵，
結果卻是逐位元組相同的戰略地圖。

`report.txt` 也一致：`視訊模式 12h`、模式切換三次（`10h → 03h → 12h`）、
開過 31 個檔（`STRMAP.BIT`、`STRMAP.RGN`、`FACE.AMG` 都在裡面），
就是本體已經載入戰略地圖的樣子。

結論（`proven`）：**問答只要求「有打東西」，不比對數值**。

## 為什麼這件事會被誤判

`docs/internals/data/screen-copy-protection.json` 記的照片是萊因哈特，
而收據送的 `138` 既不是他的占領（126）也對不上畫面問的欄位——全 22 位
提督裡只有第 76 號的占領是 138。這個對不上原本被當成「還沒解出題目
是怎麼選的」，實際上是**答案根本不參與判斷**，所以對不上也照樣過。

題目怎麼選仍然未解（`docs/re/74` 的開放項），但它不再擋著 remake：
remake 只要照原版放行就好。

## remake 端

`internal/game/game.go` 的 `Key()` 現在只檢查 `len(g.typed) > 0`；
`Answer()` 留著給文件與測試看，不參與判斷。原本那個「答錯次數」計數器
（`tries`）是照著錯誤假設寫的，已經移除。

測試 `TestAnyAnswerEntersGame` 釘住兩件事：空白按 Enter 留在問答畫面、
打 `000` 進遊戲本體。
