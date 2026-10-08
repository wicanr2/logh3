# GIN3PS.EXE 的 C runtime 是 Datalight C 3.10（Large）

日期：2026-10-08。起因是評估 matching decompilation，需要先知道用哪一支編譯器。

## 結論

| 項目 | 結論 | 等級 |
|---|---|---|
| `GIN3PS.EXE` 的資料段開頭（`ds:0000`）是 Datalight C 的版本橫幅 | 橫幅文字是 `Datalight C 3.10 L, written by Walter Bright` | 已證實 |
| 這份橫幅來自連結進本體的 C runtime（啟動碼與 NULL 指標檢查） | 啟動碼在入口把 `DS` 設到這個段 | 已證實 |
| 遊戲自己的 C 程式也由 Datalight C 3.10 編譯 | 同一個工具鏈的產物，沒有看到第二套 runtime 的字串 | 強推論 |
| 橫幅的 `L` 是 Large memory model | 與多個 code segment（`3CC9`、`3870`、`39B5`）一致，尚未用編譯驗證 | 強推論 |
| 遊戲碼的精確編譯器版本與旗標 | 只有逐函式編譯比對 bytes 才能定 | 未知 |

先前文件把 `GIN3PS.EXE` 寫成 Borland C 編出來的。那個說法的依據只有 long 輔助函式的呼叫形狀，沒有任何 Borland 字串。
傾印裡沒有 `Borland`、`Turbo-C`、`Abnormal program termination`、`MS Run-Time`、`Microsoft Corp` 這些字樣。

## 輸入

| 項目 | 值 |
|---|---|
| `org_game/extracted/GIN3PS.EXE` SHA-256 | `1efa4bd95e73415efba440af63c6a480665d60a67d8876b5f230f5c391aa2381` |
| 傾印 `~/cht/dosgolem-logh3/workplace/pzx296_199M.bin`（1,048,576 bytes）SHA-256 | `4873ec1171fad68c96a869b0c780de5cb016aecf6bd28ea6903ab9914c19d186` |
| 傾印 `~/cht/dosgolem-logh3/workplace/gunfire.bin`（1,048,575 bytes）SHA-256 | `72e9b0b8adab6c821a684747ebab3ebdf67920d08b665297d44e6cd7526b9def` |
| dosgolem 工作副本 HEAD | `f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c`，工作樹乾淨 |
| 工具 | Docker `golang:1.24-bookworm` 內的 `strings`、`od`，`--network none`，傾印唯讀掛載 |

傾印是 dosgolem 執行中的 1 MB 線性記憶體。GIN3PS 解開後的映像從線性 `0x3490`（載入段 `0349`）開始，這是 dosgolem 的載入位置，真機的段值不同。
另有 `t260_post.bin`、`refit4_post.bin` 兩份傾印，同樣位移出現相同的四個字串（未記雜湊，不作為證據）。

## 證據

資料段起點 `0x54B30` ＝ 段值 `54B3` × 16。

```
0x54b30  "ERROR: NULL pointer\0"
0x54b44  "Datalight C 3.10 L, written by Walter Bright\0"
0x54c10  "Stack Overflow\r\n$"
0x54c71  "Only 32 args allowed"
```

入口 `0349:0079`（線性 `0x3509`）的前六個 bytes：

```
fa  b8 b3 54  8e d8      cli ; mov ax,54B3h ; mov ds,ax
```

同一段啟動碼接著用 `int 21h AH=30h` 取 DOS 版本，再解析命令列。資料段裡跟著 `Bad stack size parameter`、`DOS 1.xx not supported` 這些訊息，屬於啟動碼，不是遊戲文字。

`docs/re/334` 提到的「`ds:0000` 的 Borland 字串」位於 `ds:0000`，就是這個橫幅所在的位置。該文件記錄的是橫幅被執行器造成的越界寫入改了 bit 7，該段的歸屬（Borland）當時沒有字串證據，現在由橫幅文字改正為 Datalight。

## 對既有文件的影響

下列文件把本體寫成 Borland C，或把 long 輔助函式稱為 Borland 的。觀察到的呼叫形狀與位址仍然成立，只是歸屬要改：

- `docs/internals/51-borland-long-helpers.md`、`docs/internals/02-executables-and-credits.md`：已改正文。檔名保持不變，因為有連結指向它。
- `docs/re/146`、`145`、`147`、`163`、`144`、`272`、`287`、`288`、`334`、`345`、`349`、`367`、`docs/internals/33`、`docs/spec/028`、`079`：歷史紀錄，不改寫。`146`、`334` 頂端加了指標。
- `docs/internals/00-README.md` 索引行仍寫 Borland。該檔有未提交的使用者修改，這次沒有動。

連帶要重新檢查的推論：

- `docs/re/288`、`287` 的「Borland 讀 `unsigned char` 固定編成某形狀」是 codegen 慣用法。換成 Datalight 之後形狀是否相同未驗證，用它找欄位存取的掃描器可能有洞。
- `CLAUDE.md` §3a-1 第二條的 `83 C3 NN` 形狀同理。
- `3CC9` 的亂數 `0x41C64E6D`＋`0x3039` 是 runtime 內的 `rand()`，歸屬改為 Datalight runtime。常數與演算法的證據（`docs/re/87`）不受影響。

## 對 matching decompilation 的影響

byte-matching 需要同版編譯器。已查到的來源（2026-10-08 查詢；Zortech C++ 3.1 與 2.06 後續已取得，見下節）：

| 來源 | 內容 | 與 3.10 的關係 |
|---|---|---|
| Internet Archive `zortech-cpp-31` | Zortech C++ 3.1 | 後代，約晚三年以上，codegen 預期有差 |
| Internet Archive `msdos_shareware_fb_ZORTECH` | Frostbyte 收藏的 ZORTECH，內容未檢視 | 未知 |
| Internet Archive `dmcpp` | Digital Mars C++ 8.57 Portable | 遠親，codegen 差異大 |
| GitHub `davidly/dos_compilers`，`Zortech C++ v206` | Zortech C++ 2.06 | 後代，`~/cht/kol` 已用這個收藏庫 |
| Internet Archive 全文檢索 `datalight`、`optimum-c` | 沒有編譯器本體，只有 `Datalight` 圖片、`gx-5000-out` ROM 與手冊類條目 | 找不到 3.10 |

二手資訊（未取得原文）：Datalight Optimum-C 有 3.08（1987）、3.11（1988）、3.12（1988）的紀錄，未見 3.10。

本機 `~/cht/borland/toolchains/` 有 Borland C++ 2.0、Turbo C 2.0、MSVC 1.0／1.52、MASM 6.11，沒有 Datalight／Zortech 系。Borland 與 MSVC 對本作沒有用處。

重新開啟條件：取得 Datalight C 3.10 本體，或證明某個後代版本編同一支已確認函式能得到相同 bytes。

## 候選編譯器探測（2026-10-08）

取得的候選，全部只放本機 `workplace/toolchains/`（已被 gitignore，不進 Git，不散布）：

| 候選 | 來源與雜湊 |
|---|---|
| Zortech C++ 3.1（1992-12-07，10 張 1.44 MB 磁片，內容為 LZH） | Internet Archive `zortech-cpp-31`，各檔 md5 與 IA 中繼資料一致，SHA-256 見 `workplace/toolchains/SHA256SUMS.zortech-ia` |
| ZORTECH.ZIP（195,211 bytes） | Internet Archive `msdos_shareware_fb_ZORTECH`，md5 `d1f30cfecbee33f72b1b6ae9b0926b03`，內容尚未檢視 |
| Zortech C++ 2.06 | GitHub `davidly/dos_compilers` commit `da3f15ce4912893a0dd9f8349dab29116a53ca59`。`BIN/ZTC1.EXE` `0fffc13f…1dabc`、`BIN/ZTC2.EXE` `2dd0d98e…a698e`。該收藏庫自述不保證與官方發行版一致 |

比對 Zortech C++ 2.06 與 GIN3PS runtime：

| 項目 | 結果 |
|---|---|
| 啟動碼訊息字串 `Stack Overflow`、`ERROR: NULL pointer`、`Bad stack size parameter`、`DOS 1.xx not supported`、`Not enough memory`、`written by Walter Bright` | 都在 2.06 的 `CT.OBJ` 與 `ZLC／ZLL／ZLM／ZLS.LIB`。2.06 的橫幅是 `Zortech C 2.00 library, Copyright (C) 1988,1989 L, written by Walter Bright`，格式與 GIN3PS 的橫幅相同。已證實血統相同 |
| `Only 32 args allowed` | 2.06 的 LIB 沒有。該字串屬於不同世代的啟動碼 |
| `3CC9:0009` 的 32 位元乘法（12 bytes 簽章 `8b f2 f7 e6 8b f8 8b c3 f7 e1 03 f8`）與 `3CC9:0020` 的無號除法開頭 | 2.06 的 LIB 與 OBJ 都沒有。2.06 以 `__LMUL@`、`__LDIV@`、`__ULDIV@`、`__LCMP@` 命名，位元組不同 |

推論（強推論）：Datalight C 3.10 產生的 long 運算呼叫慣例是 `AX:BX` 與 `CX:DX`，Zortech 2.06 已不同。因此任何含 long 運算的函式，用 2.06 或 3.1 編譯都不會得到相同 bytes。無 long 運算的葉函式能否相同，尚未試編譯，因為 prologue、暫存器配置與最佳化也可能已改。

byte-matching 的現況：沒有可用的同版編譯器。重新開啟條件不變。目前採「功能等價」路徑：寫 C，對原版函式做隨機輸入差異測試，另記錄 bytes 不同。

## 序言風格鑑別（2026-10-09）

GIN3PS 的函式序言是 `push bp; push si; push di; sub sp,N; mov bp,sp`：先配框架再設 `bp`，區域變數在 `[bp+0..]`，
參數在 `[bp+N+…]`。以下用 `grep -c -P` 對 8 byte 簽章計數：

| 對象 | `55 56 57 83 EC xx 89 E5` | `55 56 57 81 EC xx xx 89 E5` | `55 8B EC` | `55 89 E5` |
|---|---:|---:|---:|---:|
| GIN3PS 傾印（371,594 bytes，以行計） | 374 | 24 | 52 | 未計 |
| Zortech C++ 2.06 的 `LIB/*.LIB` 與 `*.OBJ`（以命中次數計） | 0 | 0 | 755 | 2103 |

兩邊的計數單位不同（行對命中次數），只能用來看「有無」。2.06 的函式庫完全沒有 GIN3PS 這種序言，使用標準框架。
所以代碼產生器在 Datalight C 3.x 與 Zortech 2.x 之間已經改寫，Zortech 2.06 與 3.1 都不可能編出 GIN3PS 的遊戲碼，
不只是 long 運算輔助函式的差異。

IA 其餘候選的檢視結果：`msdos_shareware_fb_ZORTECH/ZORTECH.ZIP` 是 ARC 自解壓的 Zortech 圖形示範（`FGDEMO.EXE`、`RDEMO2.EXE`、`DEMO.DBD`），沒有編譯器。
IA、WinWorld、Vetusware 的檢索都沒有 Datalight C 或 Optimum-C 本體。WinWorld 論壇與 Vetusware 列有 Zortech C++ 1.06、2.0、3.04，
尚未下載（授權範圍只含 Internet Archive）。Zortech 1.06 是 1988 年的版本，與 Datalight 3.x 同期，序言風格是否相同未知。

byte-matching 的條件：取得 Datalight C 3.x（含 `L` 模型函式庫、啟動碼與連結器設定），或證明某個 Zortech 1.x 版本的代碼產生器與它相同。

## 擴大來源後的檢索結果（2026-10-09，使用者授權 Vetusware 與 WinWorld）

| 來源 | 結果 |
|---|---|
| Vetusware 的 Zortech 廠商頁（`/manufacturer/Zortech/?author=25`） | 只有兩筆 Zortech C++ 3.0r4。WinWorld 論壇貼文說 Vetusware 有 1.06，沒有在廠商頁找到 |
| Vetusware 的 Datalight 廠商頁 | 只有 ROM-DOS 7.10 |
| WinWorld 的 Zortech C++ 產品頁 | 2.x 只有 `Zortech CPP 2.06 (1989) (5.25-360k)`；3.x 沒有下載項；沒有 1.x、Zortech C、Zorland C、Datalight C |
| Walter Bright 個人網站 | 只在履歷列出 Datalight C，沒有下載 |
| Digital Mars 新聞組貼文（`c++.announce`，2002-01） | Walter Bright 公開收購 Datalight C 等舊版原件，回覆者只提供 Zortech 與 Symantec C++，沒有人有 Datalight C |

Zortech C++ 版本日期（EDM2）：1.05 為 1988-10、1.06 為 1988-11、1.07 為 1989-06、2.0 為 1990-03、2.06 為 1990-04。EDM2 說首版 DOS 產品基於更早的 Zortech C。
1.0x 與 Datalight 3.x 同期，代碼產生器可能相同，但目前找不到可下載的 1.x。沒有下載任何新檔案。

結論：Datalight C 3.x 與 Zortech 1.x 都無法從授權範圍內的來源取得。重新開啟條件不變。
