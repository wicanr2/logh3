# 第四十五輪：正常首戰的AI一次性跳過

工作[#89](https://github.com/wicanr2/logh3/issues/89)，父提交c05be3dd1afa8bfb745a6cda94ce02e46b6a178a。
89張worklist／12張open，269條核實零不成立。沿復古remake、spec閘門及文件職責路由。

第44輪實際WORK+32載入後兩側均1，首戰原版0、remake仍1。
029及研究175已有AI一次性跳過契約，但原版正常觸發與正式接線仍缺。
先用有可回查冷啟動來源的fcmd-s3-base完整狀態，事前seedD3DC167E，
只觀測自然旗標1；另案事前控制旗標0，不能混作正常玩家收據。
來源snapshot、命令、工具版本與全部注入另記完整收據，探索與正式雙跑分開。

核對2D10入口閘門的原始bytes／IDA定位及動態前後表格。
證據充分後先審查088的READY子契約，再接一次性消費；不永久停用原版或增強AI。
正常主選單S3首次實行、下一次AI與保存各抽樣。首戰日期／座位／完整畫面仍有差異時照實保留。
不把新增名單一致替代完整#89或正常索敵same-state，087及#79保持原閘門。

成果續寫[088](../spec/088-normal-battle-player-entry.md)、[研究382](../re/382-normal-battle-player-entry.md)、
[046](../playtest/046-battle-spotting-phase.md)與唯一現況CONTEXT。

## 可重播命令及本輪結果

原版父snapshot身份與冷啟動命令見383；使用既有Go／GUI／IDA image。
四次原版只使用已存在的來源；不覆寫輸出。

```sh
bash tools/normal_ai_gate_oracle.sh normal89-r45-natural-a
bash tools/normal_ai_gate_oracle.sh normal89-r45-natural-b
bash tools/normal_ai_gate_oracle.sh normal89-r45-zero-a
bash tools/normal_ai_gate_oracle.sh normal89-r45-zero-b
tools/gui.sh xrun go run -overlay workplace/round44-normal-overlay.json tools/scripts/battle_normal_probe.go -out workplace/round45-normal-fixed-a
tools/gui.sh xrun go run -overlay workplace/round44-normal-overlay.json tools/scripts/battle_normal_probe.go -out workplace/round45-normal-fixed-b
tools/gui.sh xrun go run -overlay workplace/round44-normal-overlay.json tools/scripts/battle_normal_probe.go -skip-loaded-save -out workplace/round45-normal-nosave
tools/gui.sh xrun go test ./internal/game -run '^Test(NormalAIGate|EnhancedAINormalPlayerPath|BattleSpotting)' -count=1 -v
tools/gui.sh xrun go test ./internal/... ./cmd/... -count=1
tools/gui.sh xrun python3 tools/scripts/normal_ai_gate_receipt.py
```

seed覆蓋若尚不存在，先跑tools/scripts/battle_normal_probe_seed.py；已存在則核對rebellion.go來源相同後沿用。
IDA輸入由自然loaded-mem.bin[0x2D100:0x2D760]前補0x100個零byte寫入新normal89-r45-2D10-check.com。
以tools/ida.sh raw idat -A執行normal_ai_gate_ida.py，輸出/work/normal89-r45-gate.json。
輸入SHA／image／輸出schema見383及JSON；不開正式DB。

一次性旗標已接，S3首戰7艦隊／55筆池與結尾seed相同。38限定PASS、三份正常存讀與18標準套件通過。
088子契約已驗、整體DRAFT；日期、座位及全畫面仍nearby-state，#89／#79保持OPEN。
