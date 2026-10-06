//go:build ignore

// 正常主選單、公開S3載入、實行、首戰與一回合。無局面或CPU注入。
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/wicanr2/logh3/internal/bootstrap"
	"github.com/wicanr2/logh3/internal/game"
)

func main() {
	out:=flag.String("out","","新輸出");flag.Parse()
	if *out=="" {panic("缺輸出")}
	if err:=os.Mkdir(*out,0755);err!=nil {panic(err)}
	r,err:=bootstrap.Load("/orig","docs/internals/data");if err!=nil {panic(err)}
	g,err:=game.New(r);if err!=nil {panic(err)}
	inputs:=[]map[string]any{}
	click:=func(x,y int){inputs=append(inputs,map[string]any{"kind":"click","x":x,"y":y});g.MoveMouse(x,y);g.Click(x,y);g.Frame()}
	click(300,270);if g.State()==game.StateMainMenu {click(300,270)}
	for _,k:=range []byte("138\r") {g.Key(k);inputs=append(inputs,map[string]any{"kind":"key","key":int(k)})}
	if g.State()!=game.StateLoadGame {panic("未到正常取進度")}
	click(300,322);click(318,208)
	if g.State()!=game.StateCommandPhase {panic("未到正常命令")}
	g.NormalTurn117StrategySeed(0xD3DC167E)
	inputs=append(inputs,map[string]any{"kind":"strategy-seed","seed":uint32(0xD3DC167E)})
	click(544,125);click(560,125)
	for n:=0;n<15000&&g.State()!=game.StateBattle;n++ {
		switch g.State() {
		case game.StateTurnThinking,game.StateTurnWindow,game.StateTurnReport,game.StateEngagement:
			g.Tick(time.Second/60)
		case game.StateOccupationReport,game.StateRebellionReport,game.StateRegionFlipReport,game.StateAdmiralFateReport,game.StateShipDelivery,game.StatePhezzanReport:
			click(240,240)
		default:panic(fmt.Sprintf("正常實行停在%v",g.State()))
		}
		g.Frame()
	}
	if g.State()!=game.StateBattle {panic("未到正常首戰")}
	g.NormalTurn117Seed(0x12345678)
	inputs=append(inputs,map[string]any{"kind":"battle-seed","seed":uint32(0x12345678)})
	g.NormalTurn117Record("entry")
	saveFrame:=func(name string){
		if err:=g.Frame().WriteIndex(filepath.Join(*out,name+".bin"));err!=nil {panic(err)}
		if err:=g.Frame().WritePNG(filepath.Join(*out,name+".png"),g.Palette());err!=nil {panic(err)}
	}
	g.MoveMouse(240,240);saveFrame("entry")
	click(100,100);click(100,110)
	ticks:=0
	for ;ticks<20000&&!g.NormalTurn117Done();ticks++ {
		if g.State()==game.StateTurnReport {click(240,240)} else {g.Tick(time.Second/60)}
		g.Frame()
	}
	complete:=g.NormalTurn117Done()
	g.NormalTurn117Record("next-input");saveFrame("next-input")
	if err:=g.SaveSlot(*out,0);err!=nil {panic(err)}
	d:=map[string]any{"scope":"normal bootstrap and public S3 entry, first turn, seed-only overlay","complete":complete,"ticks":ticks,"inputs":inputs,"rows":g.NormalTurn117Rows()}
	b,err:=json.MarshalIndent(d,"","  ");if err!=nil {panic(err)}
	if err=os.WriteFile(filepath.Join(*out,"observations.json"),append(b,'\n'),0644);err!=nil {panic(err)}
	fmt.Println("正常Game首回合",complete,"觀測",len(g.NormalTurn117Rows()),"ticks",ticks)
}
