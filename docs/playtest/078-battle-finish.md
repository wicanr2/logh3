# 078：戰場結束判定原版雙跑與限定獨立驗算

2026-10-05，第89輪，沿[#79](https://github.com/wicanr2/logh3/issues/79)。
本輪屬固定狀態診斷，不宣稱正常玩家或正式remake同態。
[110 DRAFT](../spec/110-battle-finish.md)、完整062 DRAFT。

## 結果

| 核對 | 結果 |
|---|---|
| 原版60案雙跑 | 完整JSON逐byte相同，含所有原始六表、清單、注入、輸入、報告、seed與返回值 |
| 獨立計數／路由 | 60案通過；依實際清單與FLEET+00/+01判定 |
| 完整表獨立比較 | 48案通過；無開火，含六案回復、溢位及零回復率 |
| 有開火分支 | 12案已擷取；占領傷害、原池及RND未獨立驗算 |
| 原版報告 | 44張原生640×480全圖、16色與PNG已保存；未比較正式renderer |
| 守軍 | 42次返回0；不推廣到其他戰場狀態 |
| 產品標準回歸 | `go test ./internal/... ./cmd/... -count=1`：18套件，13PASS、5無測試 |

原版雙跑SHA-256：`d26f893d7a617ed566f6ef7852abfcd1cf8476e746f123f3bad7551d3cc6e477`。
[來源索引](../internals/data/battle-finish.json)SHA-256：
`31e4aa66008895c8ca13417f837001cf8f96f3d4ce901640c38187fef8122f7e`。
索引包含26來源、590原指令／20項proven、逐案七張資料的前後雜湊、完整注入與44張報告雜湊。
原始大檔保留在workplace，未加入Git。

## 重現

沿用既有第88輪父保存與固定工具版本，輸出檔不存在時執行：

```sh
bash tools/gui.sh xrun python3 tools/scripts/battle_finish_native_prepare.py
bash tools/battle_finish_oracle.sh a
bash tools/battle_finish_oracle.sh b
bash tools/gui.sh xrun python3 tools/scripts/battle_finish_review.py
```

各入口拒絕覆寫既有原始JSON。`battle_finish_review.py`固定原始JSON雜湊，
只允許已知相同PNG重用；完整來源與控制器SHA均列在索引。
dosgolem固定HEAD與IRQ0 20000步，原版GIN3.COM唯讀；每案在8-5開始前固定seed12345678。
等待報告時使用原控制器的滑鼠事件，press與release的確切步數逐案保存。

IDA輸入由`battle_finish_ida_prepare.py`產生，兩次原始查詢為：

```sh
idat -A -L/work/finish79-r89-185C.log \
  '-S/work/finish79-r89-ida.py /work/finish79-r89-185C.json 185C 0001:ret 09BF:ret 0A80:ret 0AF9:ret 0B18:ret' \
  /tmp/finish79-r89-ida/finish79-r89-185C.com
idat -A -L/work/finish79-r89-20B4.log \
  '-S/work/finish79-r89-ida.py /work/finish79-r89-20B4.json 20B4 1512:ret 1662:ret 1878:ret' \
  /tmp/finish79-r89-20B4-ida/finish79-r89-20B4.com
```

命令在`tools/ida.sh raw`既有Docker image內，外層timeout90s；COM先複製至/tmp，
沒有寫入正式.i64。輸入雜湊、IDA基址、原始CS:IP、線性位址及xref分欄保存於索引。

本輪檔案UID/GID為1000；收尾容器與完整擁有權檢查記錄於WORKLOG。
沒有重做第88輪已完成清理，沒有以本輪證據升格正式畫面或整場戰鬥。
