# 432：設定子選單與 BUG! 可達性勘誤

2026-10-05，第98輪，[#66](https://github.com/wicanr2/logh3/issues/66)。

## 來源

GIN3.COM SHA-256 `770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918`；bundle SHA-256 `de905aa0b5a666aa9d31dce36f43a5584f1842950c39d7e1b1f97c72e0d79673`。
乾淨dosgolem `f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c`，golang:1.24-bookworm，tick20000。從GIN3.COM正常啟動、帝國、138、兩次右鍵、系統指令、設定；沒有RAM注入，原檔唯讀。
`tools/oracle.sh`保存完整flags及meta。基準`workplace/oracle/settings66-r98-base-c/`與base-b、既有settings1e的screen SHA-256均為`4b2ee8cc8d071d58fe466c2ca95b72aad079cfaf96a04fb5edeee2cf369622d6`。
260000000步記憶體1048576 bytes，SHA-256 `24f40493c0e464034256f185fa5131290667378d19ab3b179e1acd6544b5aba5`。本畫面虛擬滑鼠y為0至399，繪製y等於輸入y加40。

IDA9.4從當輪記憶體取runtime 042F段4352 bytes，SHA-256 `e087ec02a6aa9b4a3483dbf7f2754ee3319c4232001b30d92da6e3505f100537`。一次性COM資料庫v3匯出269條指令，分列檔案位移、runtime CS:IP及IDA EA，不混用位址；未審查語意標unknown，不改正式i64。

## BUG! 正常入口不可選

**proven，限定本版正常出貨入口。** runtime DGROUP基底54B30，熱區03D8第20筆位移04C8為`0e 00 02 01 00 14 00 00 ca 03 b3 54`，+4可選旗標是0。
runtime `042F:01B1`讀DS:0210，`01B4`寫DS:04CC，原始bytes `A0 10 02 A2 CC 04`。正常基準兩者均0。
261000000步點(410,25)、按住200000指令，到290000001停止，bug-b／bug-c的screen SHA-256均為`48fff4b4c0bcc070710e76b6d36c9d7118b2683e6ae5c3dfe2c0c43805feeb5d`。游標到繪製(410,65)，仍是設定；`3A9C:0000`零命中。
正常快照與點擊後完整WORK230 bytes相同。原版雙跑、正式bootstrap及正常GUI的限定同態比較見[087](../playtest/087-settings-interface.md)。

這推翻286§六「旗標1，知道位置就能點到」。除錯callee及七項寫端的靜態證據保持，不能當正常入口。DS:0210特殊啟用來源本輪未追；不改座位、AI或WORK格式。

## MIDI I/F

**proven。** 動作22命中runtime `042F:08F8`。`0923`建3667下拉選單，DS:02B0七項，原點由94F4／94F5加9／7，寬9格。`0940`取列號、`094F`關閉、`0957`比FF取消，取消跳0995，不寫設定。
`0962..0971`列號乘2、查DS:0564，經DS:F138遠指標寫INI+4。`0975..098B`保存並重初始化音源，`0990`重畫設定。

| 列 | 標籤 | INI+4 |
|---:|---|---|
| 0 | 無 | FFFF |
| 1 | Roland : 300H | 0300 |
| 2 | Roland : 320H | 0320 |
| 3 | Roland : 330H | 0330 |
| 4 | Roland : 332H | 0332 |
| 5 | Roland : 334H | 0334 |
| 6 | Roland : 336H | 0336 |

261000000步點(350,325)，到290000001停止，midi-b／midi-c的screen SHA-256均為`a69a087efb7cd74800eaf613f99dbb4dccd5ba6623079018ec7032dce05e4470`。彈窗完整(304,180)至(447,359)，七列文字從(320,202)開始，每列20像素。
選無／330H均返回設定，INI記憶體分別`00 00 01 FF FF FF`／`00 00 01 FF 30 03`；934E裝置遮罩2，Device仍FM。這只證明本執行環境沒有MPU；現代軟體MIDI映射仍待使用者裁決。
兩個選擇均補第二跑，原始screen／palette／整份RAM相同。右鍵取消亦雙跑相同，沒有開檔，0971寫端零命中，WORK230 bytes不變；其他五個埠的對應以原始表與查表指令證實，未逐一宣稱實際硬體偵測通過。
生命週期勘誤：F138不能在任何時點都當有效INI內容。開窗／取消後同一指標的bytes為`00 00 05 04 00 00`，不等於實檔；取消跳過讀檔及寫端。凍結器初次誤把該指標當INI而失敗，改以零開檔／零寫端／WORK不變核取消，保留這項限制。
dosgolem預設保護原檔，本輪只驗RAM與寫入意圖，不宣稱原始MUSIC.INI已寫回。

## 工具失敗

第一次dump-mem-at省略步數被拒絕。第二次快照排程等於停止步數，主迴圈先退出，沒有輸出；停止到下一道指令後成功。
初版IDA匯出器漏遞增游標，外層90秒退出後容器留存，已明確停止。修正游標並加容器內70秒逾時；v2受到自動資料項目影響，text為db。v3只在一次性資料庫清理該範圍自動項目，269條原始指令成功。這些是工具問題，未改產品遷就。

完整本輪來源由[設定收據索引](../internals/data/settings-interface.json)保存。MIDI軟體映射仍待決，不由原版畫面升格正式接線完成。
