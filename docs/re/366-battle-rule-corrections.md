# 366 — 戰場三處差異的原始指令與受控常式驗證

日期：2026-10-02；工作項目：[#53](https://github.com/wicanr2/logh3/issues/53)。
輪次：[第二十四輪](../goal/2026-10-02-battle-rules-round-24.md)；
契約：[規格 078](../spec/078-battle-rule-corrections.md)；
驗收：[對拍 038](../playtest/038-battle-rule-corrections.md)。

## 輸入與工具身份

未修改 `GIN3.COM` SHA-256 `770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918`。
固定 `aspre.bin` 1,048,576 bytes，SHA-256
`3d8d1e7b6a5332253736d31aca9c056c9d2a16255a568fa0498237e5dc954ae6`。
來源為既有 `s3assault`、dosgolem `f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c` 乾淨副本；
該收據完整旗標與 metadata 雜湊收錄於[欄位摘要](../internals/data/battle-rule-corrections.json)。
原版 bundle 雜湊同專案基準 `de905aa0b5a666aa9d31dce36f43a5584f1842950c39d7e1b1f97c72e0d79673`。

本輪重新獨立直接呼叫原版常式兩次，未修改 dosgolem。
先注入上述完整記憶體、設 DS=ES=54B3，停虛擬 IRQ（tick=0）、預置種子
`0x12345678`，每案恢復同一快照。新載入器的堆疊與遠呼叫返回由 `oracle.Call` 管理。
這是明示的同狀態常式診斷，不能代表正常玩家路徑或原版八步順序。
種子位址為 dosgolem 線性 `lin:5E00C`。全部記憶體、艦種旗標、艦隊組成及艦艇
注入 bytes 保存在 `workplace/oracle/battle53-routines/run-c.json` 與 `run-d.json`。

IDA Pro 9.4 使用 `ida-pro-9.4-idapython:locked-v1`，image ID
`sha256:6f6d59af49d0008c4109a5295b5f374bdc007e2d1ab28cb9de08779584de2780`。
原始資料庫未改。三個新 `.COM` 由記憶體段切出，加 256 bytes 零前綴；
工具載入段起點後另加 `100h` 才對到原版 IP。每筆同列保留原始 bytes／運算元／名稱、
IDA .COM 位址、dosgolem runtime CS:IP 與線性位址，不混成同一地址。

| 切段輸入 | SHA-256 |
|---|---|
| re53-203A.com | `48dad4bc7ddb71094453c27e0f46ce00b235edf94be628a35b82d8a9c61bd02f` |
| re53-20B4.com | `3419e8c334c3744eb3d80041c63e9aed1bba9e9815c2ec6a1a74158a51cf4e29` |
| re53-2BE1.com | `c177c8dcd4d5ad1d8d544374d1ec87c857722263f63561e1303f2abde496ba3a` |

`tools/scripts/battle_rules_ida.py` 只能操作 `re53-` 新切段，禁止改名。
依[受審查語意索引](../internals/data/battle-rule-semantics.json)自動合併原始指令、
證據出處及分級，bytes／切段雜湊不符即停止；未列語意顯示 unknown 警示。
查詢範圍只用於導覽，不能拿範圍自行命名當函式邊界的證明。

## 三項已證實結果

- 尾段 `20B4:11C9` 把夾值寫入 `[bp+2]`；`11D2` 卻推入 `[bp+6]`。
  一組 100 艘、量 130 變 0；兩組各 100、量 201 變 0／1，實扣回值 199。
  `1215` 的 `8B 46 08` 返回實扣艘數，百分比先寫資料段 `A768`，雙分隊為 99。
- `203A:0620` 的 `74 09` 與 `0668` 的 `74 0A` 讓種類 5 直接走兩格縮減。
  100 艘扣 30、庫存 1003／301、兩旗標 FF，結果 70 艘、庫存 700／210。
  種類 0 在 FF 不縮，旗標 0 會縮；種類 7 完全不動並回 100。
- `2BE1:01EE` 載入 `593A`、`0202` 比對索引上限 64。
  64 個 word 的原始 128 bytes SHA-256
  `ae8f29f8a84f496742ae96a08b1fbad33520bc3bee5fc3469cc077418df2d614`。
  12 組原版呼叫包含零向量、四象限及世界邊界；`(1,1)=33`、`(4,1)=10` 能區分舊二分法。
  表值須與現有 `atanTable` 全數相同，再讓兩個入口共用。

## 保留勘誤與工具失敗

研究 147「種類 5 不縮」、研究 152「尾段同樣夾」由上述指令與受控輸出否定；
保留原研究，追加本輪勘誤入口。研究 330 §12 的約 8,130 組差異只是當時強推論，
本輪不重述成完整窮舉數字，完成聲明限於固定表與實際驗證範圍。

首次 IDA 探針沒有輸出；加建置／日誌參數後顯示 `idc.get_kernel_version` 不存在，
改用 `idaapi.get_kernel_version` 才取得真實 9.4 與非空指令。這是工具腳本問題。
Go 探針第一次為 `uint32` 線性位址誤放 `[]int` 的編譯失敗，修正後乾淨重跑。
初步回值欄位曾誤寫 `return_percent`；完整返回路徑及雙分隊 199 證明它是實扣艘數，
正式兩輪另存 c／d，a／b 與失敗日誌保留，沒有覆寫歷史。

## 實作後回填

先將規格 078 標 READY 才修改正式三支規則；同一收據下，舊實作十個子案例失敗，
修正後 19 個子案例通過。既有戰場／戰略收據與相關玩家路徑共 54 個通過事件，零跳過；
全部標準產品套件、三份正常 GUI 及獨立正式存讀抽樣通過，規格升為指定常式 CONFORMED。
細節與唯一摘要 SHA-256 見[對拍 038](../playtest/038-battle-rule-corrections.md)。
原版 IRQ 停止的常式診斷與 remake 正常 GUI 分開標範圍，未升格為完整八步同狀態驗證。
