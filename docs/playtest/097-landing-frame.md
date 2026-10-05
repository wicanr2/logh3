# 097：登陸完整單幀及名冊分支

2026-10-06，沿 [#48](https://github.com/wicanr2/logh3/issues/48)。
來源：[研究437](../re/437-landing-frame.md)、[完整來源](../internals/data/landing-frame.json)。
118 DRAFT，正式Game未改；此項不關閉#48。

## 比較等級

`same-state`，限定原版明示注入診斷的十八張完整單幀。
原版沿固定完整S3保存，種子只在0486前設12345678一次；各分支注入保存於report。
相同來源、初始狀態、輸入、相機與動畫階段，原版每案a／b逐byte重現。
第107輪GUI仍是當時程式的獨立成果，本輪未驗正常玩家整回合同態。

## 結果

| 案例 | 完整幀 | 比較像素 | 差異 |
|---|---:|---:|---:|
| 同盟登陸失敗 | 3 | 921600 | 0 |
| 同盟占領成功 | 4 | 1228800 | 0 |
| 同盟先反擊 | 3 | 921600 | 0 |
| 帝國登陸失敗 | 3 | 921600 | 0 |
| 雙攻方31→30 | 5 | 1536000 | 0 |
| 合計 | 18 | 5529600 | 0 |

每幀八區為左右圖、左右文字、中間文字、左右量表及完整下方。
區域矩形與全圖307200像素都保留；最大RGB通道差0、差異bbox null。
所有原圖、remake、差分與完整色號在 `workplace/landing48-r109-probe-r4/`，
逐檔大小與SHA-256見來源索引。輸出非公開發行包。

## 重生入口

先確認專用dosgolem版本與父state，包裝器拒絕髒執行器及覆寫收據。
原始遊戲唯讀，容器UID1000、有界CPU／記憶體／PID、無網路。

```sh
bash tools/gui.sh xrun python3 tools/scripts/landing_frame_branch_prepare.py
bash tools/landing_frame_branch_oracle.sh counterfirst a
bash tools/landing_frame_branch_oracle.sh counterfirst b
bash tools/landing_frame_branch_oracle.sh empire a
bash tools/landing_frame_branch_oracle.sh empire b
bash tools/landing_frame_branch_oracle.sh multiple a r2
bash tools/landing_frame_branch_oracle.sh multiple b r2
bash tools/gui.sh xrun python3 tools/scripts/landing_frame_prepare.py
bash tools/gui.sh xrun go run tools/scripts/landing_frame_probe.go \
  workplace/landing48-r109-typed-r4.json workplace/landing48-r109-probe-r4
```

probe-r4輸出目錄須事先存在，且收據重生須用乾淨副本。
未取得原版時不提供替代資料；不能把沒有執行當通過。
完整索引核對器另用 `--freeze`／`--check`；舊118按父提交保存的版本核實，不改寫108索引。

## 驗證與限制

- 正式標準命令 `bash tools/gui.sh test ./internal/... ./cmd/... -count=1`：18套件，13通過、5無測試。
- typed投影逐筆與RAM的SHIP投影、可見性、深度、行星投影與NULL排序核對。
- 每張完整色號及16色RGB核對；原版像素只作期望，沒有借用上方或下方。
- r1／r2／r3的失敗差分及雙名冊r1跳過第二支保留。正式通過只認r4。
- 新來源、入口、UID/GID、root-owned檔、誤建Markdown目錄與Docker容器於收尾核對。

完整反擊段、下一攻方剛換上方時的不同下方快照、正式分段演出、正常GUI／保存仍缺。
聲音只核原版選號與次序，不稱人耳、完整PCM或逐硬體時序通過。
