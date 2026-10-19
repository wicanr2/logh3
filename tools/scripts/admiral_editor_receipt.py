"""149收據：綁定正式二進位、正常GUI、標準測試及核准草稿，不把草稿當oracle。"""
import argparse
import ast
import hashlib
import json
import os
from pathlib import Path
import shutil
import subprocess

from PIL import Image

ROOT=Path(__file__).resolve().parents[2]
DEST=ROOT/'docs/images/admiral-editor.json'
GUI=ROOT/'workplace/admiral-editor-gui-r6'
BINARY=ROOT/'workplace/admiral-editor-gin3-r6'

def sha(path):return hashlib.sha256(path.read_bytes()).hexdigest()

def verify_gui():
    result=json.loads((GUI/'RESULT.json').read_text())
    assert result['passed'] and result['normal_entry'] and not result['state_injection'] and not result['script_flag'] and not result['seed_injection']
    assert result['binary_sha256']==sha(BINARY)
    assert len(result['launches'])==2 and all(x['binary_sha256']==sha(BINARY) for x in result['launches'])
    captures=[]
    for row in result['events']:
        if 'capture' not in row:continue
        path=GUI/row['file'];assert sha(path)==row['sha256']
        with Image.open(path) as im:assert im.size==(960,720)
        captures.append(dict(name=row['capture'],sha256=sha(path),size=[960,720],mode=row['observed']['graphics']))
    assert len(captures)==11
    return result,captures

def source_files():
    tracked=subprocess.check_output(['git','diff','--name-only'],cwd=ROOT,text=True).splitlines()
    new=subprocess.check_output(['git','ls-files','--others','--exclude-standard'],cwd=ROOT,text=True).splitlines()
    selected=[n for n in set(tracked+new) if n.endswith('.go') or n in ('tools/scripts/settings_font.py','internal/font/settings_font.SOURCE.json','internal/font/settings_font.json','internal/font/settings_font_12.png','internal/font/settings_font_36.png')]
    return {n:sha(ROOT/n) for n in sorted(selected)}

def main():
    p=argparse.ArgumentParser(description=__doc__);p.add_argument('--check',action='store_true');args=p.parse_args()
    assert os.getuid()==1000 and DEST.parent.stat().st_uid==1000
    result,captures=verify_gui()
    log=(ROOT/'workplace/admiral-editor-tests.log').read_text()
    passed=[s for s in log.splitlines() if s.startswith('ok ')]
    empty=[s for s in log.splitlines() if s.startswith('?') and '[no test files]' in s]
    assert len(passed)==15 and len(empty)==4 and not any(s.startswith('FAIL') for s in log.splitlines()),log[-1500:]
    for n in ['admiral_editor_gui.py','admiral_editor_preview.py','admiral_editor_receipt.py','settings_font.py']:
        ast.parse((ROOT/'tools/scripts'/n).read_text())
    approved={str(p.relative_to(ROOT)):sha(p) for p in sorted((ROOT/'docs/images').glob('admiral-editor-approved-*.png'))}
    assert len(approved)==5
    exported={'admiral-editor-empire.png':'0-special.png','admiral-editor-alliance.png':'1-special.png','admiral-editor-internal.png':'0-hidden.png','admiral-editor-kemp.png':'0-kemp-fighter.png','admiral-editor-poplan.png':'1-poplan-fighter.png'}
    if args.check:
        record=json.loads(DEST.read_text());assert record['product_sources']==source_files()
        assert record['gui_result_sha256']==sha(GUI/'RESULT.json')
        assert record['tests']['sha256']==sha(ROOT/'workplace/admiral-editor-tests.log')
        assert record['approved']==approved
        for name,digest in record['original_inputs'].items():assert sha(ROOT/name)==digest
        for name,digest in record['references'].items():assert sha(ROOT/name)==digest
        for name,digest in record['generators'].items():assert sha(ROOT/name)==digest
        for name,source in exported.items():assert sha(DEST.parent/name)==sha(GUI/source)
    else:
        for name,source in exported.items():
            target=DEST.parent/name;assert not target.exists() or target.stat().st_uid==1000
            shutil.copyfile(GUI/source,target)
        record=dict(spec='149',kind='remake-feature',comparison='layout-only against approved prototype; no DOS parity claim',
                    binary_sha256=sha(BINARY),gui_result=str((GUI/'RESULT.json').relative_to(ROOT)),gui_result_sha256=sha(GUI/'RESULT.json'),
                    normal_entry=True,script=False,state_injection=False,seed_injection=False,captures=captures,approved=approved,
                    exported={name:sha(DEST.parent/name) for name in exported},
                    product_sources=source_files(),original_inputs={str(p.relative_to(ROOT)):sha(p) for p in sorted((ROOT/'org_game/extracted').glob('ADMIRAL.DT*'))},
                    references={name:sha(ROOT/name) for name in ['docs/internals/data/ship-kind-table.json','docs/internals/data/keireki.json','assets/hd/interface-themes.json']},
                    generators={name:sha(ROOT/name) for name in ['tools/scripts/admiral_editor_gui.py','tools/scripts/admiral_editor_receipt.py','tools/scripts/settings_font.py']},
                    tests=dict(command="LOGH3_TIMEOUT=20m tools/gui.sh xrun sh -c 'go test ./internal/... ./cmd/... > workplace/admiral-editor-tests.log 2>&1; test_status=$?; cat workplace/admiral-editor-tests.log; exit \"$test_status\"'",passed=15,no_tests=4,sha256=sha(ROOT/'workplace/admiral-editor-tests.log')),
                    limits=['原版欄名仍未知','未新增DOS同狀態對拍','未新增Windows／macOS真機或發行包'],
                    tool_image='logh3-hd-preview:20261011-r1')
        DEST.write_text(json.dumps(record,ensure_ascii=False,indent=2)+'\n')
    print('PASS 149來源、兩次正式啟動、11張完整截圖、標準19套件與核准草稿索引')

if __name__=='__main__':main()
