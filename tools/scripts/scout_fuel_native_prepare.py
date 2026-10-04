"""產生完整machine／DOS保存的觀測程式；執行器來源仍為唯讀固定模組。"""
from pathlib import Path

src = Path("tools/scripts/scout_fuel_probe.go").read_text()
src = src.replace('//go:build ignore\n\n', '', 1)
src = src.replace('"crypto/sha256"', '"crypto/sha256"\n "bytes"\n "image"\n "image/color"\n "image/png"\n "runtime"')
src = src.replace("\t\"github.com/wicanr2/dosgolem/oracle\"", "\t\"github.com/wicanr2/dosgolem/internal/cpu\"\n\t\"github.com/wicanr2/dosgolem/internal/dos\"\n\t\"github.com/wicanr2/dosgolem/internal/machine\"\n\t\"github.com/wicanr2/dosgolem/internal/state\"")
start = src.index('\to, err := oracle.LoadProgram(')
end = src.index('\tvar calls ', start)
src = src[:start] + '\to := &control{callbacks: make(map[addr]func(*control))}\n\to.reload()\n\tdefer o.Close()\n' + src[end:]
src = src.replace('oracle.Addr', 'addr').replace('oracle.Far', 'far').replace('oracle.Oracle', 'control')
src = src.replace('o.Restore(snap)', 'o.reload()')
src = src.replace('far(0x25DB, 0x0714), far(0x25DB, 0x0759)', 'far(0x25DB, 0x0714), far(0x25DB, 0x0759), far(0x3765, 0x006C)')
src = src.replace('"tick": 0', '"tick": 20000')
src = src.replace('logh3.scout-fuel-oracle.v1', 'logh3.scout-fuel-native.v1')
src = src.replace('原版RAM；新LoadProgram堆疊，CallNear F000:0000設DS／ES；每案Restore。無替身、不改指令；訊息速度0、冷VGA，僅核常式分支／資料／文字，不核正常入口或時間。', '每案由固定完整machine／DOS entry.state重載；原8-2自然續到8-3並停在8-4之前。訊息速度明示0，必要時在3765:006C完成頭像後送左鍵並記錄。無替身、不改指令；有艦隊／燃料／位置注入，不稱正常玩家GUI或實體時間。')
src = src.replace('"steps":          o.Steps() - steps, "calls": calls}', '"steps": o.Steps() - steps, "calls": calls, "inputs": o.inputs, "frames": o.frames}')
src = src.replace('"tick": 20000,', '"tick": 20000, "go_version":runtime.Version(), "uid":os.Getuid(), "dosgolem_dirty":false, "parent_state_sha256":"598e10e5f9457842275fa5639219636f322d8bb2e0f8393fe25b93bfbdef0cfe", "command":"bash tools/scout_fuel_oracle.sh <a|b>",')
src = src.replace('必要時在3765:006C完成頭像後送左鍵並記錄', '必要時在3765:0091首次事件讀取返回後送左鍵並記錄')
src += r'''

type addr struct{ Seg, Off uint16 }
func far(seg, off uint16) addr { return addr{seg, off} }
func (a addr) Linear() uint32 { return cpu.Addr(a.Seg, a.Off) }
func (a addr) String() string { return fmt.Sprintf("%04X:%04X", a.Seg, a.Off) }
type regs struct{ SS, SP uint16 }
type control struct {
    m *machine.Machine
    d *dos.DOS
    callbacks map[addr]func(*control)
    inputs []map[string]any
    frames []map[string]any
}
func (o *control) Close() { if o.d != nil { o.d.Close() } }
func (o *control) reload() {
    o.Close()
    o.m = machine.New()
    o.d = dos.New(o.m, "/orig")
    o.d.Install()
    path := "/task/workplace/oracle/ownership79-r82-entry-a/entry.state"
    b, err := os.ReadFile(path); must(err)
    if digest(b) != "598e10e5f9457842275fa5639219636f322d8bb2e0f8393fe25b93bfbdef0cfe" { panic("完整父狀態身份不符") }
    must(state.Load(path, o.m, o.d))
    observed := make([]byte,1<<20)
    for i := range observed { observed[i] = o.m.Read8(uint32(i)) }
    if digest(observed) != "b905eda112fd3a78292bc0d1cdc10068703e0291968640cbaa0d150404404900" {
        raw, err := os.ReadFile("/task/workplace/oracle/ownership79-r82-entry-a/mem-in.bin"); must(err)
        differences := 0
        for i, v := range raw {
            if observed[i] != v {
                if differences < 20 { fmt.Printf("state/dump lin:%X %02X/%02X\n",i,observed[i],v) }
                differences++
            }
        }
        fmt.Println("loaded state bus SHA",digest(observed),"differences",differences)
        panic("完整父狀態RAM不符")
    }
    // 與cmd/probe同用Read8驗1MB傾印；讀VGA會改latch，因此驗後再還原全部狀態。
    must(state.Load(path,o.m,o.d))
    if o.m.IRQ0Every != 20000 || !o.m.IRQ0Pinned || o.m.CycleClock { panic("父狀態時鐘不符") }
    o.inputs = nil
    o.frames = nil
}
func (o *control) Bytes(a addr, n int) []byte { return append([]byte(nil), o.m.Mem[a.Linear():a.Linear()+uint32(n)]...) }
func (o *control) WriteBytes(a addr, b []byte) { o.m.WriteBytes(a.Linear(), b) }
func (o *control) Word(a addr) uint16 { return o.m.Read16(a.Linear()) }
func (o *control) Steps() uint64 { return o.m.Steps }
func (o *control) Regs() regs { return regs{o.m.CPU.Seg[cpu.SS], o.m.CPU.R[cpu.SP]} }
func (o *control) OnCall(a addr, f func(*control)) { o.callbacks[a] = f }
func (o *control) CallBudget(budget uint64, entry addr) (uint32, error) {
    if entry != far(0x185C, 0x0491) { panic("只接受已核對的8-3入口") }
    from, seen := o.Steps(), false
    var release uint64
    pending := false
    for o.Steps()-from < budget {
        a := far(o.m.CPU.Seg[cpu.CS], o.m.CPU.IP)
        if a == entry { seen = true }
        if a == far(0x185C, 0x060D) {
            if !seen { panic("未經8-3就到8-4") }
            if release != 0 {
                o.d.ReleaseMouse(0)
                o.inputs = append(o.inputs, map[string]any{"event":"release", "step":o.Steps()})
            }
            return uint32(o.m.CPU.R[cpu.AX]), nil
        }
        if f := o.callbacks[a]; f != nil { f(o) }
        if a == far(0x3765, 0x006C) {
            pending = true
            pix := o.m.Indexed()
            if len(pix) != 640*480 { panic("原生畫布不是640x480") }
            pal := make(color.Palette,16)
            rgb := make([]byte,0,48)
            dac := o.m.Palette()
            for i := range pal {
                c := dac[o.m.VGA.DACIndex(byte(i))]
                rgb = append(rgb,c[:]...)
                pal[i] = color.RGBA{c[0],c[1],c[2],255}
            }
            pic := image.NewPaletted(image.Rect(0,0,640,480),pal)
            copy(pic.Pix,pix)
            var buf bytes.Buffer
            must(png.Encode(&buf,pic))
            o.frames = append(o.frames, map[string]any{"step":o.Steps(),"screen":pix,"palette_rgb":rgb,"png":buf.Bytes()})
        }
        // 既有fleet_wait_native已核0091為清事件後的輪詢；入口按下會被清掉。
        if pending && a == far(0x3765, 0x0091) {
            pending = false
            o.d.MoveMouse(240, 200)
            o.d.PressMouse(0)
            o.inputs = append(o.inputs, map[string]any{"event":"press", "step":o.Steps(), "x":240, "y":200})
            release = o.Steps()+100000
        }
        if release != 0 && o.Steps() >= release {
            o.d.ReleaseMouse(0)
            o.inputs = append(o.inputs, map[string]any{"event":"release", "step":o.Steps()})
            release = 0
        }
        if err := o.m.Step(); err != nil { return 0, err }
    }
    return 0, fmt.Errorf("固定預算未到8-4：%04X:%04X",o.m.CPU.Seg[cpu.CS],o.m.CPU.IP)
}
'''
dest = Path("workplace/scout79-r83-native-v7.go")
with dest.open('x') as f:
    f.write(src)
print(dest)
