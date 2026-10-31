# 索敵選點、行軍隊形與HD按需載入

2026-10-30開始，2026-10-31驗證。來源main `5918a5c0` 加本輪修改；前輪169診斷保留。[規格158](../spec/158-scout-formation-lazy-hd.md)限定remake功能完成，原版選點畫面同態與Windows閃退根因未完成。完整來源、命令、封包、候選、圖像及trace SHA在[機器收據](../internals/data/scout-formation-hd170.json)。

## 已完成

| 項目 | 實際驗證 |
|---|---|
| P索敵 | 正常主選單、讀S1、選同盟、進亞姆力扎、點旗艦與索敵。P、資訊面板及250視野圈可見，小地圖移鏡頭後選點，原視窗精確還原。取消、無效點、再派及原池保存回歸通過 |
| 行軍隊形 | 原版參考、普通強化、名將及撤退改由旗艦接整隊命令；兩陣營在真實槽位表、空地64像素行軍的每一步都維持槽位相對位置。正常同盟實際完成兩回合，敵方走同一移動引擎 |
| 按需圖片 | 啟動只登記211個人物／編輯器／艦型索引，解碼0。首用解碼，人物與艦型共64MiB快取；淘汰／重載與損毀回退通過。字型按頁、每款最多兩頁，未選字型釋放。場景材質首次使用載入；定案海報及名言只解碼目前組別／語系，前一組貼圖淘汰 |
| 正常呈現 | 36張海報路徑，三組、五組雙人名言、五語、300秒回首組、F1／F7、存讀及重啟首組通過；另由工具列實切黑體、宋體、原版強化。全部正常操作共60張截圖，P與字型已目視 |
| 發行縮圖 | 83張人物PNG由109880634降至7392131 bytes，約減少93.3%；沿既有Scale3取樣，83張RGBA逐像素零差。原始HD及其餘素材逐檔SHA保持 |

人物、艦型、字型與海報的雜湊／解碼錯誤在首用時回退，不每幀重試。這次未把圖片讀取改為額外執行緒，不新增競態或遊戲狀態。新縮圖流程只產生發行複本，字型圖集、定案海報與來源綁定艦型圖集保持原尺寸。

P派出補接原版`25DB:0514`的圖示初始化：SHIP+2F為帝國43／同盟73，既有IDA位元組與分級見[研究132追加](../re/132-scout-order.md)。選點中的P前景、視野提示及鏡頭還原依本次使用者契約，不稱原版畫面已對拍。

## 載入量測

相同工具映像、Xvfb、HD圖與v1海報，關閉開場與音效，各一次已暖快取抽樣。舊版用正式AppImage內的同一ELF，排除AppImage解壓成本。候選使用原大小HD，未混入預縮圖的效益。

| 指標 | 舊版 | 按需載入候選 |
|---|---:|---:|
| 程序啟動至主選單擷取 | 11.49秒 | 0.53秒 |
| 容器記憶體峰值 | 847.24MiB | 447.80MiB |
| 主選單人物／艦型解碼 | 啟動全載入 | 0 |

這是Mesa軟體繪圖、GOMAXPROCS2、GOMEMLIMIT1800MiB及GOGC70的單次容器觀測，非Windows真機、實體GPU或一般硬體延遲保證。正常交火抽樣含15個行星攻擊演出識別，沒有把它寫成新的艦隊三武器動畫驗收。

## 重播入口

最終候選：`workplace/scout170/gin3-r2`。Windows交叉編譯為`workplace/scout170/gin3-windows.exe`，未真機執行。海報完整循環沿前一候選`gin3`，兩者僅差正在索敵的P圖示；海報／字型／載入來源相同，二者SHA分別綁收據。

```sh
LOGH3_GUI_IMAGE=logh3-ebiten:20261030-r169 \
LOGH3_XVFB_SCRIPT=/src/tools/with-xvfb.sh tools/gui.sh xrun \
  python3 tools/scripts/scout_hd_gui.py workplace/scout170/driver-new.py \
  --binary /src/workplace/scout170/gin3-r2 \
  --package /src/workplace/poster166/extracted/squashfs-root/usr/bin \
  --plan /src/workplace/scout170/PLAN.json \
  --out /src/workplace/scout170/normal-new --poster v1 --graphics hd --audio on
```

[驅動器](../../tools/scripts/scout_hd_gui.py)沿第169輪固定SHA的正常入口工具，只用X11鍵鼠及只讀trace；沒有狀態、種子、時間或script注入。首次腳本把戰場原點當成(0,40)，誤點行星後逾時，子程序仍存活。修正為讀正式`screen-battle.json`的(16,60)及選單行距後，R2／R3乾淨完成，未因此修改遊戲規則。

## 下一次HD封包

下一次新版本的藝術stage使用[prepare_hd_release.py](../../tools/scripts/prepare_hd_release.py)，不要把大尺寸master直接複製進發行包。來源可用現行本機完整版的`art`，輸出必須不存在，父目錄須由目前UID擁有：

```sh
LOGH3_GUI_IMAGE=logh3-ebiten:20261030-r169 tools/gui.sh xrun \
  python3 tools/scripts/prepare_hd_release.py --orig /orig \
  --art /src/workplace/poster166/extracted/squashfs-root/usr/bin/art \
  --out /src/workplace/scout170/new-release-art
```

流程呼叫[hd-pack](../../cmd/hd-pack/main.go)，預縮人物、逐張核對舊runtime像素、更新藝術清單並保存master來源SHA及`PACKED-HD.json`。其餘圖片與來源文件逐檔保持。新封包以此藝術stage接既有乾淨同源打包；版號與`dist-all`另依正式契約建立，不修改已發布版本。本輪完整試樣在`workplace/scout170/release-art/`，沒有重新打三平台、發布或變更tag。

標準產品21套件16PASS／5無測試，新增字型頁邊界及P原版數值回歸另通過。所有一次性容器退出自刪，未清理其他專案容器或映像。#48與169閃退仍獨立追蹤。狀態入口：[第170輪](../goal/2026-10-30-scout-formation-hd-round-170.md)。
