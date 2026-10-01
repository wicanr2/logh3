# 038 — 戰場三項規則修正驗收

日期：2026-10-02；工作項目：[#53](https://github.com/wicanr2/logh3/issues/53)。
契約：[規格 078](../spec/078-battle-rule-corrections.md)；
證據：[研究 366](../re/366-battle-rule-corrections.md)。
已證實（proven）：指定常式輸出與修正後 remake 相同；既有戰鬥回歸無跳過，
正常戰場入口／實行及獨立正式存讀抽樣通過。

唯一原版欄位摘要：[battle-rule-corrections.json](../internals/data/battle-rule-corrections.json)。
SHA-256 `59bfe878dbdcc7431075e36f0c662eb06ad44c98a9cbc175b0a8a45b25608b0b`。
兩輪完整原版收據共同 SHA-256 `529062efcbfcb3a55fcddcb233dd0e56868ae5afe5f9a4a75c456e1a581659dd`。
分級索引 SHA-256 `49107efdd314279160252f02552eb63fb638b6795fd8137a5bcc81057604060d`。
兩輪固定原版呼叫、初始狀態、完整注入與返回值以其中身份及本機收據為準。
常式層可以比較指定欄位，不宣稱整場戰鬥、畫面、所有副作用或八步時機一致。

## 重生入口

確認所有掛載路徑存在、形態及 UID/GID 正確，並核對專用 dosgolem 的固定 HEAD／乾淨工作樹。
原版記憶體雜湊不符即停止。來源先由 `tools/py.sh` 複製到已存在的 workplace，
與版控原檔逐位元組核對。新收據檔禁止覆寫，重跑使用新批次並更新下列收據入口。

```sh
DOSGOLEM_ORIG=/home/anr2/cht/logh3/org_game/extracted \
DOSGOLEM_EXTRA_MOUNT=/home/anr2/cht/logh3/workplace:/out \
bash /home/anr2/cht/dosgolem-logh3/tools/go.sh run /out/battle_rules_probe.go \
  -output /out/oracle/battle53-routines/run-c.json
# 獨立再跑一次，輸出 run-d.json。
```

`tools/scripts/battle_rules_receipt.py` 讀兩輪相同收據，生成解出的數值與經審查指令索引。
輸入身份由主機 Git／Docker 控制面提供，分析與寫檔只在 `tools/py.sh` 容器執行：

```sh
LOGH3_EXTRA_RO=/home/anr2/cht/logh3 bash tools/py.sh \
  /work/tools/scripts/battle_rules_receipt.py \
  --golem-head "$(git -C /home/anr2/cht/dosgolem-logh3 rev-parse HEAD)" \
  --golem-dirty "$(git -C /home/anr2/cht/dosgolem-logh3 status --porcelain)" \
  --remake-head "$(git rev-parse HEAD)" --remake-dirty "$(git status --porcelain)" \
  --go-image-id "$(docker image inspect golang:1.24-bookworm --format '{{.Id}}')"
```

語意索引複本放 `workplace/ida/re53-semantics.json`，正式 IDA 匯出必須加
`--index=/work/re53-semantics.json`；其他原始旗標與切段身份見研究 366。

基準 `aspre.bin` 來自既有 s3assault。若本機輸入遺失，以下為其完整原版重生旗標；
本輪沒有重生完整玩家序列，而是重生上列固定記憶體的受控常式。
輸出會在專用 dosgolem 工作副本；由 Docker 複製到本專案之前必須核對
1,048,576 bytes 與研究 366 的固定 SHA-256，不符就停止，不更改期望值。

```sh
bash tools/oracle.sh battle53-checkpoint -tick 20000 \
  -press 1,3,8,enter -press-at 105000000 -press-every 2000000 \
  -steps 1950000000 -click-hold 3000000 \
  -clicks '40000000:300:230,95000000:300:230,190000000:300:282,240000000:318:168,300000000:544:85,330000000:560:85,405000000:544:85,435000000:560:85,510000000:544:85,540000000:560:85,615000000:544:85,645000000:560:85,720000000:544:85,750000000:560:85,825000000:544:85,855000000:560:85,920000000:240:200,935000000:240:200,1000000000:240:200,1015000000:240:200,1080000000:240:200,1095000000:240:200,1110000000:240:200,1125000000:240:200,1140000000:240:200,1155000000:240:200,1170000000:240:200,1185000000:240:200,1200000000:240:200,1215000000:240:200,1230000000:240:200,1245000000:240:200,1260000000:240:200,1275000000:240:200,1290000000:240:200,1305000000:240:200,1400000000:240:200,1420000000:240:200' \
  -dump-mem-at '1496151955:lin:0:1048576:/src/workplace/battle53-checkpoint-pre.bin;1895887711:lin:0:1048576:/src/workplace/battle53-checkpoint-post.bin'
```

## 修正前後與實際驗收

`TestBattleRulesMatchesOracle` 核對原版 EXE、兩輪完整收據雜湊及 64 格原表，
再用解出的欄位驗 12 個方位、5 個扣損條件及 2 個分攤局面。
缺原版／原版收據會明示跳過，本輪零跳過。舊實作在十個子案例失敗，
新版 19 個子案例全過；兩側資料相同，沒有重擲或挑過關結果。

| 條件 | 舊 remake | 原版／新 remake |
|---|---|---|
| 方位 `(1,1)` | 32 | 33 |
| 方位 `(4,1)` | 11 | 10 |
| 種類 5、旗標 FF、100 艘／庫存 1003／301 扣 30 | 70 艘，庫存不動 | 70 艘，庫存 700／210 |
| 一組 100 艘、分攤量 130 | 留 1 艘，實扣 99 | 0 艘，實扣 100 |
| 兩組各 100 艘、分攤量 201 | 1／1 艘，實扣 198 | 0／1 艘，實扣 199 |

標準命令 `bash tools/gui.sh test -count=1 ./internal/... ./cmd/...`：18 套件完整回報成功或無測試。
另記錄以下固定測試事件，20 個主測試、含子測試共 54 個通過事件，零失敗、零跳過：

```sh
bash tools/gui.sh test -count=1 -json ./internal/rules ./internal/game ./internal/screen \
  -run 'Test(BattleRulesMatchesOracle|PlanetAssaultMatchesOracle|FortressGunMatchesOracle|StrategicAssaultMatchesOracle|BattleMoveMatchesOracle|BattleFacingMatchesOriginal|FormationBaseMatchesOriginal|BattleScrollMatchesOracle|BattleStartsOnContact|BattleTurnsDeduct|BattleEndsAfterNineTurns|BattleResultPersists|BattleDoubleClickAdvances|BattleMatchesOracle|FortressRangeMatchesOracle|BattleUnitsMatchOracle|BattleDataMatchesOracle|BattleMenuMatchesOracle|FortressAimStoresBearing|FortressGunFiresOnTurn)$'
```

射程候選與扣量範圍、行星反擊、戰略占領、移動、地形、戰鬥結果回寫、現有戰場畫面及
命令／資料面板均沿用既有範圍。舊測試允許的誤差未用來取代新原版常式的精確欄位比較。
每份本機日誌、實際 Go 版本與三份 IDA 匯出雜湊都保存於唯一摘要的 validation／ida。
IDA 匯出逐指令自動附上七筆已審查語意，其他位置保持 unknown 警示。

`tools/scripts/battle_rules_gui_check.py` 使用正式 `cmd/gin3`／renderer，未用 direct-entry 或記憶體注入：
主選單 → 防拷 → 正式取進度 → S3 → 帝國 → 系統實行 → 敵艦報告 → 戰場第一回合，
再從戰術選單實行一次到第二回合。兩張 640×480 PNG、307200-byte 色號及 StateBattle 收據齊全；
已目視確認回合 1／2。這是 remake 正常路徑抽樣，未新增原版同狀態畫面聲明。

獨立正常新遊戲 → 編隊 → 回合 → 存／讀／再存，帝國 1／同盟 3 支，
FLEET 3904／SHIP 27456 bytes 全表零差異，行星完整保留。
只證明正式存讀回歸，沒有中途戰場保存聲明。
GUI 收據 SHA-256 `6e8bd2257968defc9feaca8c9299e7cdb93584c90765fff5962d3347ee7e1655`；
完整正常輸入、命令、輸出與來源身份亦存於唯一摘要。
最後一處函式註解訂正後，於 `gui-final/` 獨立重跑三例並核對最終來源身份；
前一批收據保留。入口為 `tools/gui.sh xrun python3 tools/scripts/battle_rules_gui_check.py`
加 `--output /src/workplace/oracle/battle53-routines/gui-final` 及原有三個身份參數。

本輪父提交 `1b4b08012f447e230ad9c42e80114ecf7c8b1883`，驗證時明示工作樹含本輪修改。
執行器維持乾淨 f7523f9，沒有因新執行器而沿用舊收據。
寫入前後 UID/GID 1000:1000，收尾檢查專案容器與 root-owned 殘留。
所有新增入口已掛入 internals 索引；#53 可依本輪判準關閉。
完整八步 #79、實機／跨平台音樂與其他未完成項不因本輪通過而關閉。
