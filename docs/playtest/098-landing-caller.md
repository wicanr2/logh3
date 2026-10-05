# 098：登陸caller完整中間幀

2026-10-06，沿[#48](https://github.com/wicanr2/logh3/issues/48)。
父提交246da42，原版固定dosgolem f7523f9、Go1.24.13；既有六案a／b原始收據保持。
GUI工具logh3-ebiten:2，IDA工具ida-pro-9.4-idapython:locked-v1，全部Docker、UID1000。
對拍等級：same-state完整typed診斷。沒有正常玩家路徑或全回合同態完成聲明。

依原始RAM、文字及量表寫入者事件，各自保存快照；COMANI從正式解碼器重建。
297張640×480、91238400像素、八固定區域及16色RGB全部零差。
包含counter-only、登陸失敗、成功、先反擊、帝國、31→30雙攻方。
無裁切、縮放、遮罩或原版入口像素；投影座標、可見旗標、深度及NULL順序一併核對。

首批r1把計時等待返回誤認成格提交，實際差分全部保留。
原始348A:0273／027A等待後，0276／027D才進格寫入；r2按原始時點修正，詳[438](../re/438-landing-caller.md)。
第109輪3021來源在改正式規格前完整核對，索引保持SHA-256
`a6f8f5fc53543285f8cbd9c24442a6730788875c0541d821ec2dcdde706d2ff8`。

重播：

```sh
bash tools/gui.sh xrun python3 tools/scripts/landing_caller_prepare.py workplace/landing48-r110-typed-r2.json
bash tools/gui.sh run tools/scripts/landing_caller_probe.go workplace/landing48-r110-typed-r2.json workplace/landing48-r110-probe-r2
```

需新輸出名稱與已存在且UID1000的輸出目錄；既有r1／r2輸出不能覆寫。
完整原始命令、來源及SHA-256合併進[本輪唯一來源](../internals/data/landing-caller.json)。

候選PlanetFrameView、ProjectPlanetScene及RenderPlanetFrame另以同一組297幀完整驗證，r3全部零差。
實際投影座標、可見旗標、深度及NULL順序亦符合原池。
候選JSON只新增planet_frame欄位，既有欄位相同；完整5007 bytes、SHA-256
`1b647169e5a010eabb8f3cdefdac8c3ae8899798ba263476eb3426d0ddbd6580`。
未寫入正式資料，也未接Game。原演出仍合併Hit，118不標CONFORMED。

候選重播：

```sh
bash tools/gui.sh xrun python3 tools/scripts/landing_caller_formats.py --output workplace/<new-panel>.json
bash tools/gui.sh xrun python3 tools/scripts/landing_caller_prepare.py workplace/<new-typed>.json
bash tools/gui.sh run tools/scripts/landing_caller_candidate.go workplace/<new-typed>.json workplace/<existing-new-output> workplace/<new-panel>.json
```

標準18套件13PASS／5無測試。Game139.302秒，gin3 76.917秒。
來源Python AST、候選探針編譯及原版297幀／16色通過；沒有新增正常GUI／公開存讀或音訊驗收。
自動審核兩次拒絕正式資料生成，明示要求使用者對READY確認；確認題仍待答。

另沿新開[#91](https://github.com/wicanr2/logh3/issues/91)比對固定父RAM的七支鎖定向量。
a／b完整相同，只有31不同：Action0、原版target0／bits2，現行target-1／bits0。
原池SHA-256及原控制流見[438](../re/438-landing-caller.md)，未重新執行0091或改規則。
