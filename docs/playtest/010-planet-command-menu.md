# 010 — 行星命令選單的正常路徑收據

日期：2026-09-14

比較等級：`same-state`（只裁決不透明的選單矩形）

原版 bundle：`银河英雄传说3SP.zip`，SHA-256

`de905aa0b5a666aa9d31dce36f43a5584f1842950c39d7e1b1f97c72e0d79673`

原版執行檔：`GIN3.COM`，SHA-256

`770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918`

dosgolem：`e0b6a03c5338ea4aa5676e81c5a19825d847ab51`，乾淨工作樹

remake：`TestCommandMenuMatchesOracle`

## 狀態與輸入

沿用 `s3plinfo` 的正常玩家路徑：載入 S3，六次「實行」進入星域 26 的
戰術畫面。該星域第一顆行星原本被分隊壓住，因此在第 1,394,000,000 道指令
把行星 63 的 `+0x46/+0x48` 寫成 `(1100,640)`：

```
-poke lin:545A6@1394000000=4c 04 80 02
```

`report.txt` 的實際寫入位址為線性 `0x545A6`。第 1,395,000,000 道指令送
左鍵 `(60,60)`、按住 3,000,000 道指令；在選「情報」前停於
1,410,000,000。完整旗標保存在每份收據的 `flags.txt`，SHA-256
`ed24c5c816878c5d46fc0ed3fb3587eda381fdacfb090bd008f45ee9e5c07bd5`。

重跑方式：

```sh
mapfile -t receipt_args < workplace/oracle/s3plmenu/flags.txt
tools/oracle.sh s3plmenu-rerun "${receipt_args[@]}"
```

## 兩次原版重現

| 批次 | framebuffer SHA-256 | palette SHA-256 |
|---|---|---|
| `s3plmenu` | `9089c04d14781368fca3b6f60bae67c85dcb931d582deee0f114d1ed5939d11e` | `28b933fe81e79a380123b932302ba2dd8983bd3ac85a8f922e9e3b6ac1bd02e8` |
| `s3plmenu-rerun` | `9089c04d14781368fca3b6f60bae67c85dcb931d582deee0f114d1ed5939d11e` | `28b933fe81e79a380123b932302ba2dd8983bd3ac85a8f922e9e3b6ac1bd02e8` |

兩份 framebuffer 與 palette 均逐位元組相同。畫面是自有、無要塞行星的
三項選單「情報／補給／補充」，滑鼠在第一項。

## remake 比較

`TestCommandMenuMatchesOracle/s3plmenu` 使用相同滑鼠座標，按原版規則先加
`ui.MouseDY = 40`，得到框左上角 `(33,81)`、大小 `62×98`、反白第 0 列。
比較整個不透明選單矩形 6,076 個像素，差異為 **0**。

`TestPlanetMenuOpensAndEnables` 另從正常 `Game` 戰鬥狀態點到畫面內的行星，
確認選單種類、可選旗標、右鍵關閉與 `Game.Frame()` 實際畫出選單。這條測試
抓到並修正一個產品缺陷：原本繪製層只接受 `battleCmdUnit >= 0`，行星選單
雖已開啟狀態卻不會出現在畫面上。

裁決只涵蓋選單矩形；背景戰場、動態分隊與右側狀態不是本收據的聲明範圍。
