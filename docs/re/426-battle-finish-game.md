# 426：戰場結束的正式接線與報告頭像

2026-10-05，第92輪，沿 GitHub #79 與 [110 READY](../spec/110-battle-finish.md)。
父提交 `8b8b2d73a9e634537ad0c4b67daab53b8d024866`。本輪不改 dosgolem。
正式驗證與限制見 [081](../playtest/081-battle-finish-game.md)。

## 報告呼叫參數與實際頭像

第89輪記錄的敗方 caller 頭像參數0及占領 caller 的完整 FLEET+04 word 保持有效。
本輪補核 `3826:0405` 的讀取端，實際頭像依座位選擇。

| 原始定位 | 原始 bytes／運算元 | 已證實語意 |
|---|---|---|
| 3826:040C | `8B 76 30`，`mov si,[bp+30h]` | 讀收件座位 |
| 3826:045E | `3D FF FF`，`cmp ax,0FFFFh` | 區分直接頭像與座位路由 |
| 3826:0465 | `8A 87 4D 8F`，`mov al,[bx-70B3h]` | 非FFFF座位取 DS:8F4D[座位] 情報參謀 |
| 3826:046D | `8A 46 32`，`mov al,[bp+32h]` | FFFF座位只取 caller 頭像低byte |
| 3826:0470 | `30 E4`，`xor ah,ah` | 清頭像高byte |
| 3826:0480 | `9A 15 02 58 39`，`call far ptr 3958h:215h` | 將選出的頭像交給原版繪製端 |

因此，敗方報告用收件方情報參謀；人類勝方占領報告用提督，電腦勝方用對手情報參謀。
這是補充 consumer 語意，不推翻既有 caller 收據。
50張正式 renderer 報告矩形均符合原版，包含頭像、文字及游標。

## 證據身份與位址空間

- 原版 bundle：`org_game/银河英雄传说3SP.zip`，SHA-256 `de905aa0b5a666aa9d31dce36f43a5584f1842950c39d7e1b1f97c72e0d79673`。
- GIN3.COM：SHA-256 `770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918`。
- 原版執行器：dosgolem `f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c`，乾淨工作樹；Go `1.24.13`。
- IDA Pro 9.4既有匯出：`workplace/ida/ownership79-r82-3826.json`，SHA-256 `cfb1671c41cfc0ee40b52485ea325fce812b83bd42c7a3addabb0b2522b5e06f`。
- IDA輸入：`/tmp/ownership79-r82-3826.com`，SHA-256 `bec30e4885bdbd42ac0fdf512cc0e1c271fdab0bc159d09bbcf835bb73d2d2ad`。
- 上表定位為 dosgolem runtime CS:IP。IDA COM基底為 `0x10200`，例如3826:0465對應IDA `0x10665`；dosgolem線性位址另為 `0x386C5`。
- 六個報告格式由固定父RAM擷取，RAM SHA-256 `1575a57e438f8803a0fd0200268bc0e883a64d7210b4bcc42faab1e35f035ecf`。生成器為 [battle_finish_formats.py](../../tools/scripts/battle_finish_formats.py)。

本輪將上述原始指令與已證實語意合併至 [正式來源索引](../internals/data/battle-finish-game.json)。
原有 IDA 匯出中的未分級名稱只作導覽，未改原始函式名、位址或運算元。

## 正式實作範圍

8-4報告返回後呼叫8-5。先固定可戰計數與勝負，再逐則等待敗方報告。
唯一勝方以當前回合、原池、地形、隊形及共用亂數進入占領續接。
占領報告期間已提交Owner、存量及得分；行星上限功績在返回後才增加。
每個等待點同步結構化艦隊、分隊、行星與得分，保存與下一回合均拒絕等待中的狀態。
完成後回既有回合尾巴，雙方可戰繼續，單方或雙方歸零結束。

67案正式規則及Game各自比完整六表、清單順序與seed，另核真正callback前後及六案保存重讀。
兩陣營正常玩家入口與交戰回歸通過；本輪沒有取得自然勝利／占領GUI或完整八步同狀態收據。
110保持READY，完整062保持DRAFT，#79保持OPEN。
