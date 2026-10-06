//go:build ignore

// 第117輪正常原版首回合：完整保存續行，只有測試seed寫入，CPU不改。
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"runtime"

	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/dos"
	"github.com/wicanr2/dosgolem/internal/machine"
	"github.com/wicanr2/dosgolem/internal/state"
)

func must(e error) { if e != nil { panic(e) } }
func digest(b []byte) string { h:=sha256.Sum256(b);return hex.EncodeToString(h[:]) }

func main() {
	exe,e:=os.ReadFile("/orig/GIN3.COM");must(e)
	if digest(exe)!="770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918" {panic("原版身份不符")}
	parent,e:=os.ReadFile("/parent/machine.state");must(e)
	if digest(parent)!="e34be1a1256712a7476a2e0aefdf55d4aee721cac2e97de0a0d0374cccf5c920" {panic("正常父保存不符")}
	m:=machine.New();d:=dos.New(m,"/orig");d.Install();defer d.Close()
	must(state.Load("/parent/machine.state",m,d))
	if m.IRQ0Every!=20000||!m.IRQ0Pinned||m.CycleClock||m.Steps!=435000000 {panic("父時鐘不符")}
	addr:=func(seg,off uint16)uint32{return cpu.Addr(seg,off)}
	bytesAt:=func(a uint32,n int)[]byte{return append([]byte(nil),m.Mem[a:a+uint32(n)]...)}
	farPointer:=func(off uint16)uint32{return addr(m.Read16(addr(0x54B3,off+2)),m.Read16(addr(0x54B3,off)))}
	tables:=func()map[string]string{return map[string]string{
		"admiral":hex.EncodeToString(bytesAt(addr(0x54B3,0xB77C),82*28)),
		"fleet":hex.EncodeToString(bytesAt(addr(0x54B3,0xC074),32*122)),
		"ship":hex.EncodeToString(bytesAt(farPointer(0xCFB4),352*78)),
		"work":hex.EncodeToString(bytesAt(addr(0x54B3,0x8F14),230)),
		"planet":hex.EncodeToString(append(bytesAt(addr(0x54B3,0xCFB8),64),bytesAt(farPointer(0x5336),81*82)...)),
		"stellar":hex.EncodeToString(bytesAt(addr(0x54B3,0xA79C),32*126)),
		"lists":hex.EncodeToString(bytesAt(addr(0x54B3,0x9894),164)),
	}}
	seedAt:=addr(0x54B3,0x94DC)
	seedBefore:=hex.EncodeToString(bytesAt(seedAt,4))
	m.WriteBytes(seedAt,[]byte{0x78,0x56,0x34,0x12})
	inputs:=[]map[string]any{{"kind":"seed","step":m.Steps,"linear":"lin:5E00C","before_hex":seedBefore,"after_hex":"78563412"}}
	rows:=[]map[string]any{}
	save:=func(label string) {
		name:=fmt.Sprintf("%02d-%s",len(rows),label)
		pix:=m.Indexed();if len(pix)!=640*480 {panic("畫布不符")}
		dac:=m.Palette();pal:=make(color.Palette,16);rgb:=[]byte{}
		for i:=range pal {c:=dac[m.VGA.DACIndex(byte(i))];pal[i]=color.RGBA{c[0],c[1],c[2],255};rgb=append(rgb,c[:]...)}
		im:=image.NewPaletted(image.Rect(0,0,640,480),pal);copy(im.Pix,pix)
		f,e:=os.OpenFile(filepath.Join("/out",name+".png"),os.O_CREATE|os.O_EXCL|os.O_WRONLY,0644);must(e);must(png.Encode(f,im));must(f.Close())
		must(os.WriteFile(filepath.Join("/out",name+".bin"),pix,0644));must(os.WriteFile(filepath.Join("/out",name+".pal"),rgb,0644))
		rows=append(rows,map[string]any{"name":name,"step":m.Steps,"cs_ip":fmt.Sprintf("%04X:%04X",m.CPU.Seg[cpu.CS],m.CPU.IP),"tables":tables(),"seed_hex":hex.EncodeToString(bytesAt(seedAt,4)),"screen_sha256":digest(pix),"palette_sha256":digest(rgb)})
	}
	save("entry")
	markers:=map[uint32]string{}
	for _,v:=range []struct{off uint16;name string}{{0x000A,"turn-entry"},{0x003F,"maintenance"},{0x00FA,"movement"},{0x0262,"spotting"},{0x028C,"fleet-combat"},{0x0385,"fortress"},{0x03D0,"planet"},{0x03F0,"final"}} {markers[addr(0x181B,v.off)]=v.name}
	for _,v:=range []struct{off uint16;name string}{{0x001D,"final-1"},{0x02DA,"final-2"},{0x0491,"final-3"},{0x060D,"final-4"},{0x09BF,"final-5"}} {markers[addr(0x185C,v.off)]=v.name}
	seen:=map[uint32]bool{};active,returned,complete,pending:=false,false,false,false
	var release uint64
	for m.Steps<535000000 {
		location:=addr(m.CPU.Seg[cpu.CS],m.CPU.IP)
		if location==addr(0x181B,0x000A) {active=true}
		if label,ok:=markers[location];active&&ok&&!seen[location] {seen[location]=true;save(label)}
		if active&&!returned&&location==addr(0x12AA,0x00E2) {returned=true;save("turn-return")}
		if returned&&location==addr(0x12AA,0x02D4)&&m.Read16(addr(0x54B3,0x8F1A))==1 {save("next-input");complete=true;break}
		if active&&location==addr(0x3765,0x006C) {pending=true}
		if pending&&location==addr(0x3765,0x0091)&&release==0 {
			pending=false;d.MoveMouse(240,200);d.PressMouse(0);release=m.Steps+100000
			inputs=append(inputs,map[string]any{"kind":"report-click","step":m.Steps,"x":240,"y":200})
		}
		for _,click:=range []struct{step uint64;x,y int}{{450000000,100,60},{455000000,100,70}} {
			if m.Steps==click.step {d.MoveMouse(click.x,click.y);d.PressMouse(0);release=m.Steps+3000000;inputs=append(inputs,map[string]any{"kind":"click","step":m.Steps,"x":click.x,"y":click.y})}
		}
		if release!=0&&m.Steps>=release {d.ReleaseMouse(0);inputs=append(inputs,map[string]any{"kind":"release","step":m.Steps});release=0}
		must(m.Step())
	}
	if !complete {save("budget-stop")}
	data:=map[string]any{"schema":"logh3.normal-turn-native.v1","scope":"normal S3 saved-state continuation; seed only; no CPU entry injection","parent_sha256":digest(parent),"program_sha256":digest(exe),"dosgolem_head":"f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c","go_version":runtime.Version(),"uid":os.Getuid(),"tick":20000,"complete":complete,"inputs":inputs,"rows":rows}
	b,e:=json.MarshalIndent(data,"","  ");must(e);must(os.WriteFile("/out/observations.json",append(b,'\n'),0644))
	fmt.Println("正常原版首回合",complete,"觀測",len(rows),"輸入",len(inputs),"末步",m.Steps)
}
