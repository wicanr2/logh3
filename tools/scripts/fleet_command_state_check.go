//go:build ignore

// 核對第77輪快照。線上欄位沿dosgolem f7523f9 internal/machine/state.go。
// gob的map順序不穩定；完整機器欄位逐項核對，DOS段另逐byte核對。
package main

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/binary"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
)

type machineState struct {
	Magic                             string
	Version                           int
	Mem                               []uint8
	A20                               bool
	HMA                               []uint8
	R                                 [8]uint16
	Seg                               [4]uint16
	IP, Flags                         uint16
	Halted                            bool
	Model                             int
	Steps, Ticks, PortTicks, NextIRQ0 uint64
	IRQ0Pending                       bool
	IRQ0Every, IRQ0Base               uint64
	IRQ0Pinned                        bool
	CPUHz                             uint64
	CycleClock                        bool
	Cycles, NextIRQ0Cyc               uint64
	PITDiv                            uint32
	PITAccess, PITPhase, PITLo        uint8
	Ports                             map[uint16]uint8
	PortsIn                           map[uint16]uint64
	DAC                               []uint8
	DACIndex, DACPhase                uint8
	Planes                            []uint8
	Latch                             [4]uint8
	SeqIdx                            uint8
	Seq                               [8]uint8
	GCIdx                             uint8
	GC                                [16]uint8
	AC                                [32]uint8
	ACIdx                             uint8
	ACFlip, PlanarOn                  bool
	ProgramPath                       string
	FreeSeg                           uint16
	ImageBase                         uint32
	ImageLen                          int
}

func read(path string) (machineState, []byte, error) {
	var m machineState
	f, err := os.Open(path)
	if err != nil {
		return m, nil, err
	}
	defer f.Close()
	z, err := gzip.NewReader(f)
	if err != nil {
		return m, nil, err
	}
	defer z.Close()
	b, err := io.ReadAll(z)
	if err != nil || len(b) < 8 {
		return m, nil, fmt.Errorf("無效快照：%v", err)
	}
	n := binary.LittleEndian.Uint64(b)
	if n > uint64(len(b)-8) {
		return m, nil, fmt.Errorf("段長越界")
	}
	if err = gob.NewDecoder(bytes.NewReader(b[8 : 8+n])).Decode(&m); err != nil {
		return m, nil, err
	}
	if m.Magic != "DOSGOLEM-M" || m.Version != 3 || len(m.Mem) != 1048576 {
		return m, nil, fmt.Errorf("快照版本不符")
	}
	return m, b[8+n:], nil
}

func main() {
	if len(os.Args) != 2 {
		panic("指定專案根目錄")
	}
	root := os.Args[1]
	var rows []map[string]any
	for _, c := range []string{"wait", "hover1", "hover2", "hover3", "popup1", "popup2", "popup3"} {
		var states [2]machineState
		var dos [2][]byte
		for i, s := range []string{"a", "b"} {
			var err error
			states[i], dos[i], err = read(filepath.Join(root, "workplace/oracle/combat79-r77-"+c+"-"+s, "end.state"))
			if err != nil {
				panic(err)
			}
		}
		if !reflect.DeepEqual(states[0], states[1]) || !bytes.Equal(dos[0], dos[1]) {
			panic(c + "機器狀態不同")
		}
		b, err := json.Marshal(states[0])
		if err != nil {
			panic(err)
		}
		rows = append(rows, map[string]any{"case": c, "all_machine_fields_equal": true, "dos_section_bytes_equal": true,
			"machine_fields": reflect.TypeOf(states[0]).NumField(), "canonical_machine_sha256": fmt.Sprintf("%x", sha256.Sum256(b)),
			"dos_section_sha256": fmt.Sprintf("%x", sha256.Sum256(dos[0]))})
	}
	data, err := json.MarshalIndent(rows, "", "  ")
	if err != nil {
		panic(err)
	}
	p := filepath.Join(root, "workplace/round77-state-comparison.json")
	f, err := os.OpenFile(p, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	if _, err = f.Write(append(data, '\n')); err != nil {
		panic(err)
	}
	if err = f.Close(); err != nil {
		panic(err)
	}
	fmt.Println("七組完整機器欄位與DOS段相同")
}
