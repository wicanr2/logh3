# 072：索敵燃料、回收與保存

2026-10-04，第83輪，#79、[107 READY](../spec/107-scout-fuel.md)。
父提交e57f18293421d99797f0ef0f743ab7ff8dc0f4d9；原版固定f7523f9、乾淨工作樹。
[原版來源](../internals/data/scout-fuel.json)SHA-256為
9934ca4f7832739eaea663e851c82c4684d7d53b20432267ff2f76f329d35ef0。
[產品索引](../internals/data/scout-fuel-game.json)保存所有來源、命令結果、完整圖片與六檔保存的雜湊。
產品61來源SHA-256為67f8e1bdb9ac5d44cb4be0ad1f189b60b339908dc7c18e12093a4b238423afe5。
標準18套件13通過／5無測試，四組目標回歸零失敗／跳過。

實際工具映像身份：

- Go：golang:1.24-bookworm，SHA-256 1a6d4452c65dea36aac2e2d606b01b4a029ec90cc1ae53890540ce6173ea77ac。
- GUI：logh3-ebiten:2，SHA-256 6ccf04cfd896019d83b92641b9c072f65e1403cc4ef41e0afdb09b5ee1bab8f8。
- IDA：ida-pro-9.4-idapython:locked-v1，SHA-256 6f6d59af49d0008c4109a5295b5f374bdc007e2d1ab28cb9de08779584de2780。

| 驗證 | 結果 | 聲明範圍 |
|---|---|---|
| 原版雙跑 | 281案JSON逐byte相同，48報告畫面與輸入步數相同 | 明示Mode／燃料／位置／座位注入，沿完整保存自然呼叫，不是正常GUI |
| 產品規則 | 281案各驗有／無可見typed索敵隊，完整兩張原池、seed及48報告相符 | 包含byte255回繞、冷卻、空清單、先報告後回收 |
| 報告圖像 | 提督12／40，各58240像素及有效16色零差，無遮罩 | 指定報告矩形same-state；原版與remake底圖不同 |
| 完整診斷底圖 | 各248681像素不同，完整差分與bbox保留 | 不宣稱整張戰場parity |
| 多則等待 | 指定兩艦隊順序，每次Click只回收目前一支，最後續接一次 | 原清單順序由IDA核；雙艦隊產品回歸不冒充新原版多艦隊動態收據 |
| 正常操作 | 主選單、138、S3帝國、實行、首戰旗艦索敵及一次正常實行通過 | 無位置／燃料注入，seed在首戰操作前固定12345678 |
| 回收與保存 | 後續三次第8步診斷到期，冷卻三次，再派出重用原池；六檔重讀與完整FLEET／SHIP相同 | 後續僅呼叫第8步，不能當正常連續四回合到期；LoadSlot仍回命令階段 |
| GUI回歸 | 正常X11主選單到首戰，一倍／兩倍邏輯640×480相同 | 使用-intro=false，未新增GUI燃料到期或全原生X11像素驗收 |

重生入口：先沿既有`tools/battle_ownership_oracle.sh entry-a`取得父保存；
在GUI容器執行`python3 tools/scripts/scout_fuel_native_prepare.py`，
再用`bash tools/scout_fuel_oracle.sh a`及`b`重生原版，最後執行凍結器。
每次輸出拒絕覆寫。舊原版收據全部保留，執行器未改動。

產品目標命令：
`LOGH3_SCOUT_OUTPUT=/src/workplace/scout79-r83-game-final bash tools/go.sh test ./internal/game -run '^TestScoutFuel' -count=1 -v`。
標準測試：`bash tools/gui.sh test ./cmd/... ./internal/... -count=1`。
GUI沿`tools/turn_counter_gui.sh`，參數需以容器內`env`傳入，不能只設在主機包裝器外。

初次對拍缺陣營名稱，原因是測試fixture沒有載入正式FleetWindow格式；補齊資源後兩圖零差。
首個正常實行測試點在選單外框，改依正式Highlight幾何後實行成功。
近處索敵目的地會自然歸隊；改查完整地形後，目前視野最遠通路仍只有距離平方76240，
不足以驗到期。保留失敗，縮清驗證聲明，不更改玩家規則、原版期望或放寬門檻。
GUI首批環境參數未傳進容器，因既有輸出拒絕覆寫而退出；正確傳入後乾淨重跑。

107保持READY，完整062 DRAFT。正常GUI到期、正常連續四回合同態、8-4／8-5與完整八步仍缺。
