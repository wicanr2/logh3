"""固定第90輪核心探針，加入完整原池核對；只生成可丟棄診斷。"""
from pathlib import Path
import hashlib

p = Path("tools/scripts/battle_finish_attack_probe_test.go")
b = p.read_bytes()
assert hashlib.sha256(b).hexdigest() == "fe602860ee58c2aceb5d6aa1f6dbe1d427fbb4b592591db860912ef4283cc85b"
s = b.decode().replace("第90輪", "第91輪").replace("finish79-r90-probe-v2-", "finish79-r91-raw-")
s = s.replace('"testing"', '"testing"\n"github.com/wicanr2/logh3/internal/asset"', 1)
anchor = '\traw, e := os.ReadFile(nativePath)'
assert s.count(anchor) == 1
s = s.replace(anchor, '\tif os.Getenv("WINNER_PROBE_TERRAIN")=="1" { outputPath=outputPath[:len(outputPath)-5]+"-terrain.json" }\n'+anchor, 1)
anchor = '\t\tplace := &StrategicPlacement{Slots: slots}'
assert s.count(anchor) == 1
s = s.replace(anchor, anchor + '''
		terrainCalls, blockedCalls := 0,0
		blockedPoints := [][2]int{}
		mapSHA := ""
		if os.Getenv("WINNER_PROBE_TERRAIN")=="1" {
		  terrain,err:=LoadBattleTerrain("/src/docs/internals/data/battle-terrain.json"); if err!=nil {t.Fatal(err)}
		  input,err:=os.ReadFile(fmt.Sprintf("/orig/MAP%02d.CMP",region)); if err!=nil {t.Fatal(err)}
		  cells,err:=asset.DecodeCMP(input,asset.BattleMapCols,asset.BattleMapRows); if err!=nil {t.Fatal(err)}
		  mapSHA=fmt.Sprintf("%x",sha256.Sum256(input))
		  place.Passable=func(x,y int)bool { terrainCalls++; ok:=terrain.Passable(cells,x,y); if !ok {blockedCalls++;blockedPoints=append(blockedPoints,[2]int{x,y})};return ok }
		}
''', 1)
anchor = "\t\tres := probeWinnerAssault(planets, sites, winner, sfs, r, place, row.Round)"
assert s.count(anchor) == 1
s = s.replace(anchor, '''		kinds, err := LoadShipKindTable("/src/docs/internals/data/ship-kind-table.json")
		if err != nil { t.Fatal(err) }
		rawModel := configureFinishRaw(slot, fleets, before, kinds, slots)
''' + anchor, 1)
anchor = '\t\toutput = append(output, map[string]any{"name": row.Name,'
assert s.count(anchor) == 1
s = s.replace(anchor, '''		rawModel["planet"] = pall
		for a,v := range res.Merit { o:=a*28+2; binary.LittleEndian.PutUint32(rawModel["admiral"][o:],binary.LittleEndian.Uint32(rawModel["admiral"][o:])+uint32(v)) }
		for faction:=0;faction<2;faction++ { for _,o:=range []int{0xD0,0xD8} { at:=o+4*faction; binary.LittleEndian.PutUint32(rawModel["work"][at:],binary.LittleEndian.Uint32(rawModel["work"][at:])+uint32(res.Score[faction])) } }
		rawDiff := map[string][]string{}
		rawHex := map[string]string{}
		for k,got := range rawModel { rawHex[k]=hex.EncodeToString(got); rawDiff[k]=[]string{}; for i,v:=range got { if v!=after[k][i] { rawDiff[k]=append(rawDiff[k],fmt.Sprintf("+%X %02X/%02X",i,v,after[k][i])) } } }
''' + anchor.replace('"name": row.Name,', '"name": row.Name, "raw_differences":rawDiff,"raw_hex":rawHex,"terrain_calls":terrainCalls,"blocked_calls":blockedCalls,"blocked_points":blockedPoints,"map_sha256":mapSHA,'), 1)
dest = Path("workplace/finish79-r91-model-test.go")
assert dest.parent.stat().st_uid == 1000
with dest.open("x") as f:
    f.write(s)
print(dest)
