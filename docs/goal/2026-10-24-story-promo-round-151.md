# 第151輪：國徽故事片頭與完整推廣片重剪

使用者要求v1.6.0片頭以帝國與同盟國徽作底圖，沿先前拍攝手法加入故事鋪陳，將既有內容串成完整推廣片。這是本機影片編輯，三平台程式與完整包保持。

## 目前狀態表

| 項目 | 狀態 | 證據與閘門 |
|---|---|---|
| 開工 | 完成 | 路由載入推廣片、文件職責與dist-all；Issue102張，282條verify零失效，#48 OPEN |
| 舊片查證 | 完成 | 126雙旗故事、140名將影片規格與143實際分鏡、CONTACT；已目視既有雙國徽BACKGROUND及v1.5成片總覽 |
| 剪輯契約 | READY | 下列已授權敘事；不改遊戲規則，不公開發布 |
| 素材與舊版保存 | 完成 | 重剪前備份92秒成片、收據及校驗，共10檔；previous/MAPPING.json保存原路徑與SHA-256 |
| 新片 | 完成 | 35段123.566667秒1080p60；雙國徽故事、主角動畫、新版操作與完整原速實戰已串連；18336447 bytes |
| GitHub README | 完成 | 使用者追加明確授權，提交44a8cae1只含README、三張海報實錄及三圖索引，線上五檔下載SHA-256相同；不公開完整版或影片 |
| 驗收與交付 | 通過 | 7414格、單曲連續、40處音效、34剪接點、SNR36.30dB、零黑段、82文字安全矩形、來源與目視通過；現行總索引只更新影片 |

片頭直接沿`workplace/hd127/promo-final/BACKGROUND.png`雙國徽底圖，不重繪或替換國徽。故事沿README既有來源及[140](../spec/140-hero-tactics-and-release.md)：帝國與同盟對峙百餘年，萊因哈特與楊威利立場不同，遠征與補給困境帶出亞姆力札，將下一道命令交給玩家。

1988動畫與正常戰術／要塞／三武器交火沿已驗來源，逐段綁定原PLAN及SOURCE-PLAYBACK；不冒稱全片由v1.6.0新拍。海報、五語、提督編輯及駐地改選、領船、編成、出發沿150正式ZIP新實錄。完整演出不加速，改選提督到下達出發保持連續原速。全片控制在同首BGM_0F長度內，音效另依既有played觀測重混。

工作區沿`workplace/poster151/`，唯一現行影片交付仍在`dist-all/v.1.6.0-20261024/promo/`。前版保存在本輪previous映射，不移動公開tag／Release。製片入口為[poster_story_promo.py](../../tools/scripts/poster_story_promo.py)，驗收追加[133](../playtest/133-poster-five-language-delivery.md)，現行總索引沿[交付索引](../internals/data/v160-delivery.json)。

GitHub入口：[README海報預覽](https://github.com/wicanr2/logh3#海報寬螢幕預覽)、[本次五檔提交](https://github.com/wicanr2/logh3/commit/44a8cae1875a9af39ff9f23e645d7b8f96eeb5ff)、[三圖索引](../images/readme-v160-posters.json)。工作區`GITHUB-README.json`保存實際HTTP下載雜湊；主工作樹只同步提交基準與索引，既有未提交產品及README全文修改保持。
