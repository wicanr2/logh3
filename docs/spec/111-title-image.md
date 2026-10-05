# 111：SPTIT.350 標題圖解碼

狀態：**CONFORMED**。2026-10-05，GitHub #68／#70；限定圖形解碼與保存。
範圍只限來源圖、palette及保存匯出。動畫末端等待、淡出／淡入與正常標題輸入另待完整時序證據；本規格不授權改玩家轉場。

## 原始證據與審查

- 固定DOS bundle SHA-256 `de905aa0b5a666aa9d31dce36f43a5584f1842950c39d7e1b1f97c72e0d79673`。
- `GIN3.COM` SHA-256 `770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918`。
- `SPTIT.350` 33346 bytes，SHA-256 `c38203aa19c9cab462b9e4e7806225e668c7593ecbdbf30d03b883bdea0ba747`。
- dosgolem專用副本HEAD `f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c`，乾淨；Go image `golang:1.24-bookworm`。
- `tools/oracle.sh title68-r96-a -tick 20000 -steps 950000000 -regs-at 041F:024B,041A:0035 -regs-max 2000`。
- 獨立b批次同啟動與950000000步，沒有額外regs旗標。兩次640×350色號SHA均 `addde5e0929eb5f99a5430a714b19041a2cbe5530a4720a3641abebbfa3bf8ae`，256×3 palette SHA均 `d7382d63e411bd3d804231e2c666f763fd7e051556c909516492562223fe8b21`。
- 可丟棄解碼器四個plane剛好消耗全檔，224000色號逐byte符合原版。227列候選於run超界退出。

上述是完整圖與palette的proven證據，不依靠檔案尺寸或副檔名判斷。
原始執行位址僅供回查既有動畫讀取端，沒有新增靜態函式語意。
完整來源入口：[研究430](../re/430-title-image.md)及[保存素材收據](../internals/data/remaining-assets.json)。

## 解碼契約

`DecodeTitle350([]byte) (*Image,error)` 回640×350、每像素一byte色號，Used必須等於輸入長度。
四個plane各28000 bytes，初始全零，一欄一欄向下排列；解出第i個byte的位置為 `(i%350)*80+i/350`。

| token | 消耗與結果 |
|---|---|
| XX YY，兩byte不同 | 只消耗XX，將XX XOR進一格 |
| XX XX n，n有bit7 | 消耗3 bytes，重複n&7F格 |
| XX XX n1 n2，n1無bit7 | 消耗4 bytes，重複 `(n1<<8)\|n2` 格 |
| XX為0 | 只推進，不改plane |
| plane末端 | 四次各有一個01分隔byte，之後開始下一plane |

四個分隔byte的檔案位移分別為9684、18023、26614、33345，四次值均01。
合成時plane0..3分別對應色號bit0..3，每byte的最高位在左。
截斷、零長run、超過28000格、缺／錯分隔byte或尾端多資料回error，不回部分圖。

`SPTIT.RGB` 48 bytes＝16色；每通道0..15先乘4成六位元DAC，再以 `(v<<2)|(v>>4)` 展成PNG八位元色。
來源48通道與原版有效16色逐byte相符；此規格不推論TITSPT.RGB的時序用途。

## 驗收及權利

- 真實來源完整解碼、Used、色號SHA及原版雙跑符合。
- 全部截斷邊界及錯分隔、超長run／尾端多資料均回error。
- PNG重讀的尺寸、索引與palette符合typed圖；素材清單記輸入及輸出SHA。
- PNG可存private repo的 `assets/original/title/`，由 `tools/scripts/` 產生，不手改；不進一般發行包。
- 本規格不改存檔格式，不把圖解碼完成稱為#68玩家轉場完成。

## 限定驗收

完整來源、四個分隔、全部33346種截斷長度及錯誤run／尾端資料通過。
原版兩批標題圖與palette一致，794張素材獨立雙跑及PNG重讀相同。
`TestRemainingAssetExports`實際核全部保存PNG身份、尺寸及色號SHA通過。
標準18套件13PASS／5無測試在既有logh3-ebiten:2與Xvfb通過。
圖形子範圍從READY升CONFORMED；精確轉場等待及正式玩家接線保持未完成。
