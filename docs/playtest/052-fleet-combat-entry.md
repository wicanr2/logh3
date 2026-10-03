# 052：艦隊戰鬥入口原版雙跑

2026-10-03，第62輪。比較範圍為原版獨立雙跑，固定相同輸入／入口／seed與全部注入。
接敵案例屬受控入口診斷，未建立remake配對，也不作正常玩家命令提交聲明。
[研究400](../re/400-fleet-combat-entry.md)保存來源雜湊及原始定位，
[092](../spec/092-fleet-combat-entry.md)維持DRAFT。
[固定資料](../internals/data/fleet-combat-entry.json) SHA-256為
132e4ae3c6ac38ddaf62075d3bd450c3537dda5c8a4a12a4d376424a3f239121，含338份檔案指紋。

| 收據 | 雙跑結果 | 限制 |
|---|---|---|
| natural a／b | 入口／末態RAM、色號、色盤、PNG相同 | 序列化state不同；只採固定a作父快照 |
| 14個門檻案例 | 每案5份完整檔案相同，參戰名單及初始化符合原指令 | 只驗空命令返回；末態包含後續執行 |
| near-outs | 有敵方候選，但自身Out=FF，無命令／演出／損失 | 明示注入接敵位置與情報 |
| near-human | 到達完整戰鬥命令標頭，雙跑相同 | 尚停提示等待 |
| command a／b | 固定點擊後到人類操作循環，4份檔案相同 | 未選目標／武器，無損失、存讀驗收 |

合計36次執行、89組檔案比較相同。各完整640×480色號307200像素零差，色盤完全相同，
故有效RGB亦零差；無排除區。標頭及操作等待兩張圖已目視。
沒有remake或跨狀態差分圖，不能解讀為原版／remake對拍通過。

重播入口：`tools/fleet_combat_entry_oracle.sh`。
已存在的批次拒絕覆寫，原始輸入唯讀，完整旗標保存每個run的flags.txt／meta.txt。

```sh
bash tools/fleet_combat_entry_oracle.sh combat79-r62-natural-a
bash tools/fleet_combat_entry_oracle.sh combat79-r62-gate-action1-a
bash tools/fleet_combat_entry_oracle.sh combat79-r62-gate-near-human-a
bash tools/fleet_combat_entry_oracle.sh combat79-r62-command-a
bash tools/py.sh /work/tools/scripts/fleet_combat_entry_receipt.py docs/internals/data/fleet-combat-entry.json
```

重建時需要研究400的原版與固定父快照。工具固定dosgolem HEAD且拒絕髒工作樹。
可用門檻案例及完整a／b命令列列在凍結JSON；不覆寫歷史run或已凍結收據。
原始RAM／state／PNG留在workplace，未加入Git。

產品程式未改。本輪18標準產品套件、工具語法、凍結收據重讀及拒覆寫結果由WORKLOG記錄。
新工具／文件入口掛[內部索引](../internals/00-README.md)，#79／#48保持OPEN。
Docker一次性容器收尾清除，輸出UID/GID1000:1000；完整Goal仍未完成。
