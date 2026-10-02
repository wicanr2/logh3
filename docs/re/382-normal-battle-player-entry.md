# 382 正常S3首次戰場入口不同

日期2026-10-03，第四十四輪，Issue [#79](https://github.com/wicanr2/logh3/issues/79)、
新上游閘門[#89](https://github.com/wicanr2/logh3/issues/89)。父提交5751b76bce200b3b229fafcda51fd39a2bbd3907。
路由載入重製／畫面對拍技能及文件職責。正式產品與dosgolem未改；[088](../spec/088-normal-battle-player-entry.md) DRAFT。

## 輸入與工具

- 未修改GIN3.COM SHA-256 `770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918`。
- 原版bundle SHA-256 `de905aa0b5a666aa9d31dce36f43a5584f1842950c39d7e1b1f97c72e0d79673`。
- dosgolem專用logh3-oracle副本HEAD `f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c`，clean；
  Go與GUI image身份、完整argv及輸出雜湊見[收據](../internals/data/battle-normal-entry.json)。
- 原版地址為dosgolem線性RAM：FLEET `lin:60BA4`、SHIP `lin:668B2`、STELLAR `lin:5F2CC`、
  WORK `lin:5DA44`、seed `lin:5E00C`。不混用IDA位址。各列保留原索引、完整bytes及差異位移。
- 原版以主選單取進度／138／S3／帝國／首次戰略實行／敵艦報告確認到達希利。
  只在300M步寫入seed D3DC167E，完整事件見flags；未改局面、艦隊、鏡頭或遊戲檔。
- remake用bootstrap.Load／Game.New，送正式畫面座標輸入。原版滑鼠y+40為畫面y。
  主選單一下進入是007已列的既有差異；原版需兩下。亂數觀測／固定只用測試期Go覆蓋，
  正式rebellion.go沒有改。固定前實讀D3DC167E，固定後仍同值；不重擲。

## proven結果與限制

| 比較點 | 原版 | remake | 結論 |
|---|---|---|---|
| 載入後三張全表 | FLEET3904、SHIP27456、STELLAR4032 | 同長度、零差異 | 原始資料／載入三表相同 |
| 載入後WORK | +0C=00、+0E=01、+10=01 | +0C=02、+0E=00、+10=00 | 座位／控制區已不同；不猜補含義 |
| 首戰艦隊槽 | 12,16,17,28,29,30,31 | 12,28,31 | 7對3，索敵之前已不同 |
| 戰場池含未放出索敵隊 | 55 | 27 | 口徑相同，不混同battleUnits活動分隊數 |
| 首戰戰略回合 | 380 | 381 | 日期相鄰，不能same-state |
| 入口全表不同bytes | FLEET223、SHIP1293、STELLAR66、WORK50 | 全紀錄保存 | 包含未參戰及指標，不能逐項稱玩法bug |
| 全畫面 | 307200像素 | 5863色號及RGB不同 | nearby-state，未遮罩 |
| 戰場視窗 | 161280像素 | 73不同，bbox240,240至255,255 | 游標／滑鼠狀態差異仍保存 |

原版5份輸出獨立雙跑相同；其中兩份為各1MiB RAM，其餘為PNG／色號／DAC。
兩份entry.state全檔不同，保存各自雜湊，不將snapshot全檔列為雙跑通過。
remake17份輸出雙跑相同；另一次省略實行前診斷存檔，六張全表及畫面仍相同。
原版結尾seed B6E4B78A，remake ECDD600B；這只證明入口亂數狀態不同，不要求整段骰序一致。

原版PNG左圖右欄為第13艦隊資料，remake右圖為輸入摘要及7990304日期。
兩圖已目視；完整差異屬狀態不同，其他幾何／色彩原因尚unknown。沒有用裁切或遮罩宣稱通過。
PNG診斷讀原版PNG有效PLTE色盤，不能直接用DAC pal.bin當色號色盤。

## 根因仍unknown

007明示排除讀檔後陣營小選單，現行Game直接進命令階段。這是已知契約缺口，
但本次尚未證明它單獨造成7對3名單差異。也不能只看到結尾seed不同就判AI錯。
下一步先追原版選陣營寫WORK、首次AI命令／移動、回合更新，找最早分歧後審查088 READY。
不要重做已相同的三張parser，不完整翻譯AI executable，不改087的受控golden。

首次探測未加seed觀測，保留在workplace/round44-normal-a；後續測試期覆蓋量到同一初值。
影像分析最初錯假設GUI容器有Pillow，改沿用pngout標準函式庫；seed報告只印地址與長度，
核對改用精確地址／4 bytes及flags值。這兩項是驗證工具問題，不是遊戲缺陷。
完整新收據及source SHA由[核對器](../../tools/scripts/battle_normal_entry_receipt.py)重生；
正式重播與驗收限制見[046](../playtest/046-battle-spotting-phase.md)。
