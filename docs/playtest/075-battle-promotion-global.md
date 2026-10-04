# 075：升級全局回寫的原版診斷驗證

比較範圍事先限定為固定8-4入口的資料診斷，沒有remake正常玩家或畫面完成聲明。
完整輸入、版本、雜湊與勘誤見[研究420](../re/420-battle-promotion-global.md)。

## 重播

從父提交710d2d13的專案與固定dosgolem副本執行下列命令。
來源位於`workplace/`，缺原版或父保存時不能以代用品通過。
工具拒絕覆寫輸出；第二次重生需使用乾淨工作樹及獨立輸出目錄。

```sh
bash tools/gui.sh xrun python3 tools/scripts/battle_promotion_global_prepare.py
bash tools/battle_promotion_global_oracle.sh a
bash tools/battle_promotion_global_oracle.sh b
bash tools/go.sh run ./tools/scripts/battle_promotion_global_probe.go
bash tools/gui.sh xrun python3 tools/scripts/battle_promotion_global_receipt.py
bash tools/go.sh run ./tools/scripts/battle_promotion_global_probe.go -output workplace/promotion79-r86-global-probe-check.json
```

原版自然續行至185C:060D後才注入，09BF停止。
66案每案seed12345678在執行前固定，IRQ0間隔20000步，原始JSON保存完整命令、注入、報告按下／放開及步數。
原始JSON欄位`paths`是EMS frame觀測，不代表入口當下已映射路徑。
模型從重算前歸屬自行建表，只與返回後EMS作比較。

## 結果

| 項目 | 結果 |
|---|---|
| 原版66案雙跑 | 完整JSON24871318 bytes相同 |
| 成長與三表 | 呼叫、473個RND輸出、53次重算及三表沿第85輪獨立模型相同 |
| 全局53次 | STELLAR213696、WORK12190 bytes逐byte相同；完整四路徑矩陣相同 |
| 最終66案 | 完整STELLAR／WORK及6706-byte行星相同，零升級案亦核 |
| 模型重跑 | 76521 bytes完全相同，沒有遮罩或忽略欄位 |
| 產品與標準測試 | 本輪未改產品，標準套件及正常首戰X11沿第83輪最近通過結果 |
| 擁有權 | 新工具、觀測器、JSON及文件UID／GID1000；收尾檢查見WORKLOG |

收據[全局來源](../internals/data/battle-promotion-global.json)為21來源，
SHA-256 `38db9c02ae4f4883f13d4cb5fc5316620b23b4e179eb5ecbfb4332cbb13d81b6`。
沒有本輪畫面對拍、產品圖或差分圖。
正式8-4、typed窄回寫、12級軍階UI、逐則通知與玩家存讀仍待完成，108 DRAFT，#79 OPEN。
