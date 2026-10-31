# 第171輪：三平台完整版與P索敵、行軍補錄

使用者要求打包完整版，推廣影片加入P與行軍畫面。沿170已驗產品來源，main仍`5918a5c0`，169／170未提交來源保留。

## 目前狀態表

| 項目 | 狀態與契約 |
|---|---|
| 版本 | v.1.6.4-20261031；新建，不覆寫既有tag、Release或封包 |
| 共用藝術 | 新stage依158預縮83人物，原始HD及既有素材SHA保持 |
| Linux | PASS，x86_64完整AppImage；既有Type-2 runtime，實包兩陣營、P與行軍／兩回合、前線整軍通過；未測FUSE |
| Windows | PASS，amd64完整ZIP，實際解包及Wine驗證；附閃退日誌CMD，不宣稱169已修或真機通過 |
| macOS | PASS，Intel／Apple Silicon通用ZIP，Go1.24.13／SDK15.5；架構、簽章及素材核對，未真機 |
| 補錄 | PASS，新AppImage正常入口的P範圍、選點與還原、整隊64點實際行軍；不注入局面或seed |
| 完整推廣片 | PASS，保留雙國徽、主角故事與完整原速武器；前線整軍重錄，接P及行軍，37段126.883333秒1080p60，單曲配樂與40音效 |
| 權利／交付 | 沿專案既有dist-all/版本/full與promo，只本機完整交付，不公開原版或動畫 |
| 完成閘門 | PASS，標準21套件16PASS／5無測試、三包manifest／SHA與5條正常入口、音畫／字幕／來源及目視；語法、1695個文件連結及正向索引、擁有權與容器清理通過 |

路由：[版本契約](/home/anr2/.codex/knowledge-base/local/retro-remake-release-versioning.md)、[交付根目錄](/home/anr2/.codex/knowledge-base/local/retro-remake-dist-all-output.md)。子目錄依本專案AGENTS的既有`full`契約，避免新增第二份現行交付。

入口：[本輪交付驗收142](../playtest/142-v164-full-scout-march-promo.md)、[機器收據](../internals/data/v164-full-delivery.json)、[新版截圖](../images/readme-v164.json)、[170功能驗證](../playtest/141-scout-formation-lazy-hd.md)、[166封包與影片](../playtest/139-poster-frame-full-promo.md)、[目前脈絡](../../CONTEXT.md)。本輪流水帳仍寫WORKLOG；169未知Windows閃退、#48及原版全局同態保持原狀態。

使用者於交付後追加授權push。169至171的來源、工具、README截圖與驗收文件提交至main；上述5918a5c0及dirty來源是建置時基準，收據保持，不另重打包。完整版、影片與原始遊戲資料不進Git。
