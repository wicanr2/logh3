# FACE.AMG — 人物頭像（.AMG 圖庫格式）

證據等級：**confirmed**（反組譯原版解碼常式 ＋ 與原版畫面逐像素相同）
資料：`docs/internals/data/face-amg.json`
實作：`internal/asset/amg.go`、測試 `internal/asset/amg_test.go`
素材：`workplace/assets/face/face-NN.png`（83 張，一次性產物，不進 Git；
以 `tools/py.sh tools/scripts/amg_export.py` 從自備原版重建）

## 檔案結構

```
0x0000  索引表：N 筆 × 6 bytes     N ＝ 第一筆的位移 ÷ 6
        +0 u16  該張的位元組長度
        +2 u32  檔案內位移
0x01F2  影像資料（第一筆的位移剛好等於索引表長度，兩者前後相接）
```

`FACE.AMG` 有 83 筆，每張影像 80 × 100 像素、4 位元色號。

## 每張影像

```
+0   u32  固定 1
+4   u32  size − 6
+8   u16  直行數 w（每直行 8 像素）      FACE.AMG 全部是 10 → 80 像素寬
+10  u16  每直行的 word 數 h（每 word 2 列） 全部是 50 → 100 列
+12  u8   0 ＝ 一般；0xFF ＝ 空白（第 18、31、33–35、37–39、82 張）
+13  …    壓縮的 planar 資料
```

## 壓縮：直行掃描 ＋ 相等即重複

原版的解碼常式在 `GIN3PS.EXE` `348A:02F9`（定位方法見 §「怎麼找到的」）：

```
02F9  AD           lodsw              ; 取兩個 byte
02FA  38 E0        cmp al,ah
02FC  74 13        je  0311           ; 相等 ⇒ 重複組
02FE  26 88 05     mov es:[di],al     ; 同一個 x 的第一列
0301  03 FB        add di,bx          ; bx ＝ 80（mode 12h 每列位元組數）
0303  26 88 25     mov es:[di],ah     ; 相鄰的第二列
0306  03 FB        add di,bx
0308  FE CA        dec dl             ; 這一直行還剩幾個 word
030A  75 ED        jnz 02F9
030C  E8 1F 00     call 032E          ; 走完一直行 → x+1
030F  EB E8        jmp 02F9
0311  8A 24        mov ah,[si]        ; 重複組：值對 ＝（al, 下一個 byte）
0314  8A 0C        mov cl,[si]        ; 次數
0319  26 88 05     mov es:[di],al     ; 同樣兩列一組，重複 cx 次
```

換直行與換 plane（`032E`）：

```
032E  mov di,cs:[02B7] / inc di       ; 下一個直行
0339  mov dl,cs:[02B4]                ; 重設每直行的 word 數
033E  dec dh / jnz ret                ; 還有直行 → 繼續
0343  mov di,cs:[02B5]                ; 直行用完 → di 回到最初起點
034D  mov dh,cs:[02B3]                ; 重設直行數
0354  mov al,cs:[02B9] / shl al,1     ; map mask 左移 ＝ 下一個 plane
0361  out dx,al                       ; 03C5h
0362  test al,10h / jz ret            ; 還沒到第 4 個 plane → 繼續
0368  inc sp / inc sp / ret           ; 四個 plane 做完 → 丟掉返回位址收工
```

**四個 plane 共用同一個狀態機**：重複組的次數可以跨過 plane 邊界，
原版在那裡只是換 map mask 再回到同一個迴圈。把每個 plane 當成獨立一段解，
跨界的那幾個 word 會落在錯的位置——症狀是人臉看得出來、輪廓也對，
但有 1355 個像素的 **bit0** 不同（其他三個位元全對），
看起來很像色盤或執行器的 planar 寫入有問題。

## 驗證

| 判準 | 結果 |
|---|---|
| 每張是否剛好用完自己那段位元組 | 83 / 83 全部剛好用完 |
| 尺寸 | 83 張全是 80 × 100 |
| 色號 | 全部落在 0–15 |
| 與原版畫面逐像素比對（第 0 張 vs 收據 `face-fresh`，對齊 144,192） | **0 / 8000 不同** |

前三項在 `internal/asset/amg_test.go`；第四項是同一支測試的
`TestFaceAMGMatchesOracle`，沒有收據就 skip。

## 編號就是提督索引（strong inference）

畫面上出現的頭像，編號與 `ADMIRAL.DT*` 的提督索引對得上：

| 畫面 | 頭像 | 署名 | `WORK` +0x37 的提督索引 |
|---|---:|---|---:|
| 戰略情報報告（`info1e`）| 13 | 情報專任參謀克斯拉准將 | 13（職位碼 `3A`）|
| 月底船艦接收報告（`turnmonth2`）| 6 | 軍事專任參謀梅克林格准將 | 6 |
| 月初會計報告（`key1`，帝國）| 16 | 內政專任參謀奧貝斯坦上校 | 16 |
| 月初會計報告（`keyAlly`，同盟）| 60 | 內政專任參謀邱吾權上校 | 60 |

四個獨立資料點；前兩張的頭像用 `tools/scripts/face_find.py` 比到
**0 個像素不同**，後兩張是 remake 照提督索引取圖之後整張畫面逐位元組
相同（`docs/spec/011`）。同盟那一筆是第一個**索引 ≥ 40** 的資料點——
在它之前這條規則只在帝國那半邊驗過。83 張比 82 名提督多一張，
多的那一張是誰未確認。

（`WORK` 的 `+0x37` 是三個職位各一對提督索引，帝國在前、同盟在後，
見 `docs/re/61`；先前寫成「索引 ＋ 職務碼」。）

## 怎麼找到的

在 dosgolem 開 `-vram-sites`（統計「誰在寫視訊記憶體」），跑兩次不同步數
再相減：頭像畫出來那一段新增的寫入者只有 `348A:02FE/0303/0319/031E`
四處，合計 4000 次寫入 ＝ 4 plane × 10 × 100。接著 `-peek` 取出那一段
位元組直接讀。

先前在壓縮位元組上猜格式（PackBits、三種位元組 RLE、96 組 LZSS／LZ77 參數、
nibble RLE）全部落空——**判準都用「83 張要解出同一個長度」，沒有一組通過**。
差別在於：猜格式是在搜尋空間裡亂走，找繪圖常式是直接問「誰畫的」。
