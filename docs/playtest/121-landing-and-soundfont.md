# 121：交戰分段與音色選擇

日期：2026-10-14。音色選擇已驗收；#48完成正式分段接線，整合驗收仍未齊。入口為[第132輪](../goal/2026-10-14-landing-soundfont-round-132.md)，實作依[118](../spec/118-landing-scene.md)與[114](../spec/114-midi-interface.md)。

#66、#69與#16已關閉；#48保持開啟，剩餘條件見下方。

本輪可重生核對器為[收據工具](../../tools/scripts/landing_soundfont_receipt.py)，原版雙跑、完整幀及兩陣營音色實錄的雜湊集中於[來源索引](../internals/data/landing-soundfont.json)。

## 使用者驗收

- 使用者明確回報現行音樂已聆聽、PASS並要求#69結案，已關閉。
- 使用者回報目前remake人工檢視與原版一致，要求#16關閉，已關閉。原版`+20`／`+22`欄名未知仍依[提督資料](../internals/08-admiral-dt.md)保存。
- 使用者指定#66改為Roland MT-32／GeneralUser GS選擇，並確認Gsound為GeneralUser GS.sf2。MT-32目錄實際是ROM；以Munt離線合成、OGG播放。

這些是本輪使用者回報，不改寫舊機器收據，也不新增Windows／macOS真機或DOS整場同態聲明。

## 音色選擇已通過

| 抽樣 | 結果 |
|---|---|
| 兩陣營正常片頭、防拷、月初及設定 | 帝國`gui-f0-final`、同盟`gui-f1-r3`均PASS，未用-script、狀態或seed注入 |
| GeneralUser GS／Roland MT-32 | 實際播放器曲庫路徑符合選擇，各錄兩秒44100 Hz雙聲道PCM，四份均非靜音 |
| 取消與靜音 | 右鍵／Escape只關音色選單；靜音中換兩種音色保持靜音，HD按鈕可解除靜音 |
| 偏好與退出 | `preferences.json`原子保存MT-32；正常退出、重開片頭後恢復，兩次設定「結束」及公開退出均正常 |
| 顯示 | F7雙向切換保持所選音色；兩種選項完整可讀 |
| 失敗保持 | 缺曲庫、損壞偏好及寫入失敗不覆蓋原偏好；測試`TestSoundFontPlayerPreferences`通過 |
| MT-32曲庫 | 11首完整解碼、非靜音、峰值低於-0.2 dBFS；本機合成與實際GUI播放均通過 |

MT-32使用Munt 2.8.2官方提交`3b05ec276f9e605af86b0eaef7f5eda43477a31f`，來源封存`workplace/munt-2.8.2-source.tar.gz`，SHA-256 `209e1042a3e4392f0c807467e9afcc56cc5b14fd3bbb49655a4689e8febf51ec`。控制ROM為1.07，SHA-256 `a73a06c23ed38370e58a11fb1b86f7ea4c547061a60d7aa62bae446235fd2dff`；PCM為`d9164063f293410cf33f2f64cdcea6893b44723fe41938e07ec9aef58b406238`。GeneralUser GS.sf2為`9575028c7a1f589f5770fccc8cff2734566af40cd26ed836944e9a5152688cfe`，沿143既有11首曲庫。

新MT-32曲庫不是SF2，使用使用者提供的ROM離線模擬。原MIDI通道排序後映射到MT-32的1..8，其餘事件位元組保持，不猜補GM音色對應。這是明示的音色近似；使用者先前的聆聽PASS不外推成新MT-32音訊的人耳驗收。

## #48已完成的範圍

正式規則在每次登陸、反擊與占領結果保存獨立前後副本，含擁有者、分隊、攻擊力、完整度與seed。原`PlanetAssault`運算順序、公式、戰損、得分與seed保持。Game改用這些副本排程，避免從最終Hit或換手後行星反推前段。上方與下方保留各自快照，換攻方時不提前重畫下方。

正式renderer從空畫布生成底圖、文字、量表、COMANI與3DBLOCK下方；追加HD繪製紀錄不改原色像素。每段結果獨立等待，一次點擊只消費一次；動畫關閉仍保留文字、量表、下方及結果。音效採既有播放完成閘門，修正行星演出每次Tick重複扣音效剩餘時間，最快速度的靜態下方仍先於cue出現。

| 驗證 | 結果與範圍 |
|---|---|
| 原版六案雙跑 | baseline、failure、capture、counterfirst、empire、multiple，1041份RAM／色號／palette逐檔相同 |
| 完整幀 | 297幀、91238400色號像素、每幀八區及RGB最大差皆0，無遮罩、無借原版像素 |
| 規則觀測 | 原反擊收據的結算／seed保持；快照不受後續改值影響 |
| Game診斷 | 兩陣營占領前擁有者、登陸／占領獨立等待、動畫關閉與音效名義／實際後端閘門通過 |
| 標準測試 | `tools/gui.sh test ./cmd/... ./internal/...`，19套件，15PASS／4無測試 |

原版輸入bundle SHA-256為`de905aa0b5a666aa9d31dce36f43a5584f1842950c39d7e1b1f97c72e0d79673`。專用dosgolem副本HEAD為`f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c`，工作樹乾淨，Go1.24.13。每案完整命令、父state雜湊、線性注入、seed、IRQ0旗標與步數保存於來源索引及原report.json。原版RAM、色號與PNG留在`workplace/oracle/`，差分在`workplace/round132/landing-comparison/`。本輪沒有改執行器。

| Docker工具 | 本輪映像身份 |
|---|---|
| `golang:1.24-bookworm`，原版收據 | `sha256:1a6d4452c65dea36aac2e2d606b01b4a029ec90cc1ae53890540ce6173ea77ac` |
| `logh3-ebiten:2`，建置／測試／renderer | `sha256:d7f4d3e04c75fc9e9b2b3606d9f015d8f1bcde806c25d6780780450ad40827ab` |
| `logh3-hd-preview:20261011-r1`，正常X11／音訊 | `sha256:932b5ad00e356b4252945c35938d9b30b939109abfaaca8b6673f1a17f532aff` |
| `logh3-audio:3-munt2.8.2`，Munt／編碼 | `sha256:bb5fd867a03942eeab1a087fc2a4f5a4044c9687828961faf86d7f179576643f` |

上述297幀是同狀態完整renderer診斷。Game結算到畫面的正常整合不能由它替代，118保持READY，#48保持OPEN。剩餘項目為要塞砲動態演出、要塞最小渲染／音訊抽樣、目前CPU政策下的正常GUI／公開存讀，以及DOS原版整段同狀態。

第107輪占領工具在本輪可重現登陸／占領報告與保存，但它明確呼叫`SetPassiveBattleComputer(true)`。目前bootstrap採現行CPU政策，舊輸入不能直接比較；獨立重播停在StateRegionInfo，不列為通過，也不據此宣稱產品缺陷。本輪早期S3簡易X11驅動未能推進戰術回合，亦未列為#48正常流程證據。

## 本機使用及重生

開發執行檔為`workplace/round132/gin3`，兩套曲庫已放在`workplace/round132/music/{hd,mt32}/`。在Docker內可用以下命令啟動，或從設定選擇。已發布v.1.3.0及其封包未更動。

```sh
tools/gui.sh xrun /src/workplace/round132/gin3 -orig /orig -data /src/docs/internals/data -art /src/assets/hd -graphics hd -music /src/workplace/round132/music -soundfont mt32
```

MT-32合成與正常GUI命令如下。輸出目錄必須不存在，重跑改用新的名稱。原版、ROM及音樂來源唯讀；資源、網路與逾時由包裝器限制，UID/GID1000:1000。

```sh
LOGH3_FONT_DIR=/home/anr2/cht/mt32 bash tools/audio.sh python3 /src/tools/scripts/mt32_prepare.py --midi /src/workplace/modal131/music-source --out /work/round132/mt32
LOGH3_GUI_IMAGE=logh3-hd-preview:20261011-r1 LOGH3_XVFB_SCRIPT=/src/tools/with-xvfb.sh LOGH3_CPUS=2 tools/gui.sh xrun python3 tools/scripts/landing_soundfont_gui.py --binary /src/workplace/round132/gin3 --music /src/workplace/round132/music --out /src/workplace/round132/gui-f0-final --faction 0
LOGH3_GUI_IMAGE=logh3-hd-preview:20261011-r1 LOGH3_XVFB_SCRIPT=/src/tools/with-xvfb.sh LOGH3_CPUS=2 tools/gui.sh xrun python3 tools/scripts/landing_soundfont_gui.py --binary /src/workplace/round132/gin3 --music /src/workplace/round132/music --out /src/workplace/round132/gui-f1-r3 --faction 1
```

原版雙跑入口為`bash tools/landing_scene_oracle.sh <baseline|failure|capture> <a|b>`與`bash tools/landing_frame_branch_oracle.sh <counterfirst|empire|multiple> <a|b>`；multiple另傳`r2`。收據固定HEAD與父state，既有目錄拒絕覆寫。297幀重生用以下命令，先建立使用者擁有的新輸出目錄。

```sh
tools/gui.sh run tools/scripts/landing_caller_candidate.go workplace/landing48-r110-typed-r3.json workplace/round132/landing-comparison docs/internals/data/engagement-panel.json
tools/gui.sh xrun python3 tools/scripts/landing_soundfont_receipt.py --out workplace/round132/new-index.json
```

本輪索引326263 bytes，SHA-256 `7cb6b951e551b66078862871aa32b4c1c89cc6c04dd3aee4343fd8df5cb0f8a0`。原版檔案、ROM、合成音訊及執行檔不加入Git；版控保存程式、方法、來源及輸出雜湊。
