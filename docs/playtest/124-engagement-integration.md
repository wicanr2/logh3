# 交戰演出整合驗收

日期：2026-10-14，第135輪。[118限定CONFORMED](../spec/118-landing-scene.md)；正常GUI與原版完整renderer幀分開驗收。
入口為[計畫135](../goal/2026-10-14-engagement-integration-round-135.md)，原版診斷及失敗分類見[研究448](../re/448-engagement-integration.md)。

獨立正常S3公開操作計畫g／h相同，四類要塞型行星原版雙跑792檔相同，正式renderer重驗普通行星297幀及要塞162幀，共459完整幀、141004800像素全圖／八區／RGB零差。
標準19套件為15 PASS、4無測試。

最終GUI已核對正式片頭、S3同盟、目前CPU政策、24段登陸／反擊／占領、第6回合換手至第7回合、HD／原版顯示、F7、實際音訊與公開六表存讀。兩種顯示各293來源格完整，沒有跳格或倒序。
保存讀回採獨立公開操作的完整六表預期。STELLAR載入重算的五個byte差異另記，不要求不成立的逐檔前後皆相同。
原版／remake整回合同態及同戰場續讀仍不由這份收據代替。

| 正常驗收 | 結果 |
|---|---|
| 入口與條件 | 正式片頭、S3同盟、原版AI，無腳本、局面／seed注入或被動CPU覆寫 |
| 第6回合占領 | 艦隊31攻雷奧米爾，防衛力987→307，再以獨立占領段由Owner0改1 |
| 演出與續行 | 24段登陸／反擊／占領，各293來源格全數呈現，正常返回第7回合 |
| HD與原版顯示 | 各單獨執行；HD演出中F7往返保持資料及流程 |
| 音訊 | 正式PulseAudio輸出，44100 Hz／16-bit／雙聲道、非靜音；各段cue及播放完成閘門通過 |
| 公開保存讀回 | 六表共44560 bytes，完整符合獨立公開操作；五表前後不變，STELLAR重算五個byte |
| 原色圖與HD基底 | 459原版完整幀零差；兩陣營左右基底與無D0安全回退回歸通過，修正前版本明確失敗 |

最終來源為`workplace/round135/gui-hd-k/`與`gui-original-k/`，執行檔為`gin3-final-bases`。
HD297.033秒、原版顯示164.511秒包含載入、鏡頭回授及保存操作，不能當作動畫播放時長比較。
結果文字、量表、攻擊後307防衛力、反擊164艘、占領成功與下方方塊已目視確認；沒有文字溢出或結果殘留。
來源、原圖、差分、完整GUI輸入、格序、音訊與六表雜湊見[凍結來源索引](../internals/data/engagement-integration.json)。

## 重跑入口

所有命令使用專案Docker包裝器。原版輸入唯讀，輸出拒絕覆寫；另取不存在的目錄名稱。原版擷取入口的批次名稱固定，重新擷取須在獨立工作副本執行，不搬走或覆寫本輪原始收據。
正式GUI依序執行，不和其他建置、測試、原版擷取或GUI競爭CPU。

```sh
tools/gui.sh run tools/scripts/landing_player_plan.go workplace/round135/new-plan
tools/gui.sh build -o workplace/round135/new-gin3 ./cmd/gin3
DOSGOLEM_CPUS=2 DOSGOLEM_MEM=3g bash tools/landing_fortress_oracle.sh fortress failure a
DOSGOLEM_CPUS=2 DOSGOLEM_MEM=3g bash tools/landing_fortress_oracle.sh fortress failure b
LOGH3_GUI_IMAGE=logh3-hd-preview:20261011-r1 LOGH3_XVFB_SCRIPT=/src/tools/with-xvfb.sh \
  tools/gui.sh xrun python3 tools/scripts/landing_fortress_prepare.py workplace/round135/new-typed.json
tools/gui.sh run tools/scripts/landing_caller_candidate.go \
  workplace/round135/fortress-typed.json workplace/round135/new-comparison docs/internals/data/engagement-panel.json
LOGH3_GUI_IMAGE=logh3-hd-preview:20261011-r1 LOGH3_XVFB_SCRIPT=/src/tools/with-xvfb.sh \
  LOGH3_CPUS=2 LOGH3_TIMEOUT=25m tools/gui.sh xrun python3 tools/scripts/landing_player_gui.py \
  --plan /src/workplace/round135/new-plan/plan.json --binary /src/workplace/round135/new-gin3 \
  --out /src/workplace/round135/new-gui-hd --music /src/workplace/round132/music
tools/gui.sh test ./cmd/... ./internal/...
LOGH3_EXTRA_RO="$PWD" tools/py.sh tools/scripts/engagement_integration_receipt.py
```

原版另外三類為`fortress capture`、`iserlohn failure`、`iserlohn capture`，各跑a／b。
上方比較工具要求先建立新的空輸出目錄；建立規則依本專案`workplace/`職責，不覆寫收據。
原版顯示GUI另指定`--graphics original`。完整凍結核對工具與來源索引由本頁及[計畫135](../goal/2026-10-14-engagement-integration-round-135.md)索引。

來源索引914075 bytes，SHA-256 `fbb937202fa4a2c7bf5a5cdc18885538e19235b38c7b9a3c8a1a96046b12b10b`。
