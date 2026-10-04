# 423：8-5可戰計數、結束報告與勝方占領邊界

日期：2026-10-05。Issue：[#79](https://github.com/wicanr2/logh3/issues/79)。
目前實作與規格入口為[110草稿](../spec/110-battle-finish.md)，分級事實保存於
[battle-finish.json](../internals/data/battle-finish.json)。原始名稱、bytes、位址與xref均保留。

## 輸入與工具

| 輸入／工具 | 固定身份 |
|---|---|
| GIN3.COM | SHA-256 `770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918` |
| bundle | SHA-256 `de905aa0b5a666aa9d31dce36f43a5584f1842950c39d7e1b1f97c72e0d79673` |
| 8-1完整父保存 | workplace/oracle/cleanup79-r88-entry-a/entry.state；SHA-256 `a62af47fb7d15a136e2e2ed51760ef0e9dd82538f68d5f2b0011b8` |
| 父1MiB RAM | SHA-256 `1575a57e438f8803a0fd0200268bc0e883a64d7210b4bcc42faab1e35f035ecf` |
| dosgolem | 專用logh3-oracle副本，HEAD `f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c`，乾淨 |
| 逆向工具 | IDA Pro9.4、ida-pro-9.4-idapython:locked-v1；只分析/tmp副本，不改正式.i64 |
| 執行工具 | golang:1.24-bookworm；IRQ0固定20000步，CycleClock關閉；seed在8-5前設12345678 |
| remake父版本 | `a97ccd7c97b3ca2597e41318c3c38d4ab352b945` |

185C與20B4兩個COM由父RAM各取0x2000 bytes，前置256 bytes。
COM雜湊分別為`77b7198ee99167489d486f2e6338929eac233dbd22b94b978bc5cec3d2356ede`、
`2f824b4a8f1fc7f23de1e91ae6bd74995bcef8c711eba2af38b63b9e3f35aa72`。
IDA匯出使用COM基址0x10200；每項另外保存dosgolem CS:IP及線性位址，不把兩者混用。
合計590筆原指令，其中20筆已審為proven；其餘unknown，函式自動名只供導覽。

## 方法與已證實結果

[原版觀測器](../../tools/scripts/battle_finish_native_prepare.py)由第88輪凍結控制器導出。
從185C:001D自然經8-1到8-4，於185C:09BF注入士氣bit0、兩座位操控類型、
回合、seed；六案另注入首顆行星防衛力與回復率。完整注入位址與bytes逐案保存。
停止點是185C:001B，位於8-5返回後，不跳過占領分支。

60案包括四種勝負、四種人類座位、回合5／8／9，另補勝方回合0／1／6與六案回復邊界。
雙跑逐byte相同，原始JSON SHA-256
`d26f893d7a617ed566f6ef7852abfcd1cf8476e746f123f3bad7551d3cc6e477`。

- 可戰計數、返回值、敗方報告順序／文字／座位及勝方占領引數由獨立模型核對。
- 48案沒有開火；完整ADMIRAL、FLEET、SHIP、WORK、PLANET、STELLAR及戰場清單符合獨立模型，seed不變。
- 44則報告保存原生640×480色號、16色及PNG。八次占領callback包含雙CPU報告。
- 有開火的12案、42次守軍返回值0均有完整原始資料；傷害、原池回寫與RND仍未獨立核對。

較早「勝方九回合占領」應精確讀成局部回合由目前戰場回合開始，在步末≥9停止。
第9回合仍執行一次回復，來自do-while控制流。研究339與322保留歷史，不將其
未量測摘要當正式驗收。058的戰略入口從0開始，這個差異不推翻既有戰略收據。

## 探索勘誤

初版探針以為艦隊12／16／17都屬帝國。原始FLEET+01實際為12屬帝國，
16／17／28／29／30／31屬同盟；初版漏測帝國勝利。
探索雙跑保留在workplace/finish79-r89-map-explore-a.json及-b.json，
SHA-256均為`76f798111862a43623b4d33058fdbf46867947f8fcba7507683b24be3bc47591`。
正式探針加入原始陣營斷言，改用mask01／7E。

第二次54案正式雙跑已相同，SHA-256
`da82ed72f67ce01c54c4eae67e603d7ba5a69a9f2e9d976f299e24d4556178d8`，
保留在workplace/finish79-r89-base-a.json及-b.json。
原保存防衛力已滿，不能靠資料不變證明回復算式，因此另加六案後才凍結60案。
獨立核對器最初比較整word陣營引數失敗；20B4:151F只讀低byte。
按consumer訂正核對器後通過，原版輸出未改。

## 限制

本輪沒有修改production path。110與完整062仍DRAFT，#79保持OPEN。
下一閘門是有開火分支的原池／RND、守軍適用條件與可共用規則的READY審查。
原版診斷不是正常玩家勝負，也沒有正式remake報告畫面通過聲明。
