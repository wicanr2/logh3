# 084：戰場旗艦方向圖示

狀態：**CONFORMED**，只限本規格的方向圖示。工作項[#79](https://github.com/wicanr2/logh3/issues/79)。
2026-10-02第36輪完成DRAFT及證據審查：IDA控制流、原版圖塊1024 bytes、43邊界344筆可見方向與Facing全部相符。
分級索引見[方向證據](../internals/data/battle-directions.json)，尚未知呈現不納入本規格。
證據[研究376](../re/376-second-battle-render.md)、[S1逐圈紀錄](../internals/data/battle-second-frames.json)。

限定旗艦旁的八向金色三角圖示，不處理同順位重疊物件與格表殘留，不改移動與戰損。
GIN3.COM、RAM、IDA工具與位址空間雜湊由研究376固定。

typed輸入沿用BattleFleet.Facing/X/Y與BattleUnit.Faction/Squadron/Detail/Out。
旗艦存在且未Out，在玩家陣營或Detail==3時畫方向圖示；其餘不畫。
方向由已證實DirectionIndex(Facing)取得，SPRITE.SPR第0..7塊、16×16。
位置為艦隊現在座標減鏡頭、加研究376八筆位移與Field原點。
色號0透明，裁切於既有Field安全矩形，不重畫旗艦內容、不擴張命中範圍。
缺sprites或索引越界安全略過；不以生成圖片補原版素材。

圖塊經既有DecodeTiles解析，正式Game.Frame繪製。方向是執行期呈現，不新增存檔欄位。
S1的43邊界比全畫面差分與方向圖示區域；完整資料原先為零差異。
正常S1入口與移動GUI抽樣、既有S3逐圈與標準測試另驗。
保留原始全部差分，不把本圖示修正外推成完整畫面或八步完成。

第36輪依READY契約接正式Game.Frame，43邊界各440個原版物件方向像素相同，合計18920。
原版位置直接讀每幀物件+6/+8，不在測試複製方向位移公式，也不將其餵入引擎。
既有S3全畫面35邊界、三份正常隨附存檔實行／完整存讀、18個標準套件通過。
正式S1正常入口移動PNG已檢視。全畫面還差434到664像素，重疊順序與203列仍unknown。
完整畫面與080／083仍未完成，本規格不豁免任何未知差分。
