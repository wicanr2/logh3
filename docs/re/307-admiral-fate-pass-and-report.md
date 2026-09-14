# 307 — 失蹤提督的生死判定、狀態寫回與報告

## 結論

原版 `056A:0219` 每回合依提督索引 0–81 掃描 `ds:B77C`；只有
`+0x00 == 0x80` 的失蹤提督進判定。判定成功後，死亡走 `28E0:057F`，
生還由呼叫端直接寫回，最後以 `056A:0311` 顯示本人姓名、頭像及結果。
控制流、門檻、狀態寫回與組字均為**已證實（proven）**。

輸入 `mrx056A.com` SHA-256 為
`5bafdb75a75879472ef475819acc033b80d8c077d1ac8c905e0759d6d336ef6d`；
`28E0` overlay `code28E0.com` SHA-256 為
`dc08ae1bbb530e41e56b38b38870ab56b34b423b7a00711af97bfe333aa3b6c3`。
工具為 IDA Pro 9.4；位址均為原版 16-bit DOS segment:offset。正式 `.i64`
唯讀，查詢只在可丟棄副本執行。

## 判定

`056A:0219` 的原始分支化簡如下：

1. 提督 0 與 52：`rand() < 0x2000` 時確定生還，不進死亡抽籤。
2. 其他提督：先以 `rand() < 0x0444` 決定本回合是否查明下落。
3. 查明後再抽 `rand() % admiral[+0x0B]`：人類座位 `<= 45` 死亡，
   電腦座位 `<= 30` 死亡；其餘生還。
4. 未通過第一道門檻者保持 `+0x00 == 0x80`，本回合沒有報告。

IDA 匯出 `report056A-admiral-disasm.txt` SHA-256
`9d03c605b8b28cf9bdf37fb29381de03ecc18ccc78d007148e8d9bd02565d1e0`。
重要原始定位與 bytes：

```text
056A:023A  26 80 3f 80        cmp byte ptr es:[bx],80h
056A:024F  9a 0c 00 d5 3c     call far 3CD5:000C
056A:0254  3d 00 20           cmp ax,2000h
056A:0261  9a 0c 00 d5 3c     call far 3CD5:000C
056A:0266  3d 44 04           cmp ax,0444h
056A:0292  f7 f1              div cx
056A:0294  83 fa 2d           cmp dx,45
056A:02B1  f7 f1              div cx
056A:02B3  83 fa 1e           cmp dx,30
```

## 狀態寫回

死亡時呼叫 `28E0:057F(admiral, 1)`。IDA 匯出
`report28E0-cleanup.txt` SHA-256
`487edafa0528ab820b0cd6f525d0c6c196a305f562c54b392ac8d37cc722ec07`：

```text
28E0:05A1  or  byte ptr es:[si+9],80h
28E0:05B5  mov es:[si],al
28E0:05BA  mov word ptr es:[si+2],0
28E0:05C0  mov word ptr es:[si+4],0
```

第二參數非零時 `AL=0`。生還分支 `056A:02D6` 把 `+0x00` 設 `0xFF`，
再把 `+0x0A` 設為該陣營首都星域。兩條路最後都呼叫
`056A:0311(admiral, dead)`。

## 報告組字與版面

`056A:0311` 先以 `3982:022D(admiral[+1])` 跳過電腦座位，再串接：

- `ds:080A`：`下落不明的提督　\n`
- `ds:081D`：`%s已經\n`
- 死亡 `ds:0825`：`確認死亡。　　　　　`
- 生還 `ds:083A`：`脫離成央B確定生還。`

最後呼叫 `3826:0307(buffer, -1, admiral)`；`3826:0356` 在第二參數為
`-1` 時以第三參數的提督索引畫本人。`report056A-admiral-report.txt` 與
`report3826-compositor.txt` SHA-256 分別為
`1e2a14f32f9f935eaab99c381945c07440f3aa89d0876d110711064a2b9083e2`、
`8f5beea5c67ddc5f72fbfddbccd9af8ceda92b38b1ce9ef8c746897d6d73cb2d`。

## 決定性執行期收據與限制

`workplace/oracle/admiral-fate-survive/` 綁 dosgolem-logh3 `e0b6a03`，從
新遊戲正常路徑進第一次「實行」。在 `#210000000` 將提督 0 的
`lin:602AC` 設為 `0x80`；於 `056A:0219` 入口 `#215832564` 把
`ds:94DC`（`lin:5E00C`）設為 0。第一個 `rand()` 回 0，
`056A:0311` 於 `#215832643` 以 `DI=0` 命中，證實特殊提督生還分支與報告
呼叫。完整旗標與實際線性位址在該批次 `flags.txt`／`report.txt`。

dosgolem 在這條 `3826:0307` off-screen compositor 路徑停於滑鼠等待時，
最終 framebuffer 的文字／頭像沒有形成可用的逐像素 oracle。因此這份收據
只裁決控制流、RNG 與參數，不裁決報告像素；畫面目前只能依同一 compositor
的既有已證實幾何標為**強推論（strong inference）**，不可宣稱 same-state。
