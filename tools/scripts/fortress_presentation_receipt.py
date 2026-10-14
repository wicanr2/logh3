"""核對144正式要塞演出：原版雙跑、423完整幀、正常GUI及六表存讀。

LOGH3_EXTRA_RO="$PWD" tools/py.sh tools/scripts/fortress_presentation_receipt.py --freeze
不加--freeze只重核，不覆寫凍結索引。
"""
import argparse
import hashlib
import json
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
SOURCE = Path('/ref') if Path('/ref/internal/game/fortressengagement.go').exists() else ROOT
OUT = ROOT / 'docs/internals/data/fortress-presentation.json'


def sha(p):
    return hashlib.sha256(p.read_bytes()).hexdigest()


def load(p):
    return json.loads(p.read_text())


def collect():
    parent = ROOT / 'docs/internals/data/fortress-scene.json'
    assert sha(parent) == '22160d0c3c08d8c72fba9208328de2e2da11818d305955f67a117b1a137c04b7'
    old = load(parent)
    # 舊索引的測試來源保留在原提交，不能拿本輪擴充版冒充舊版。
    for name, expected in old['sources'].items():
        p = ROOT / 'workplace/round134/fortress_scene_oracle_test-r133.go' if name == 'internal/rules/fortress_scene_oracle_test.go' else SOURCE / name
        assert sha(p) == expected, name
    result = dict(schema='logh3.fortress-presentation.v1', date='2026-10-14', issue=48,
        parent_commit=(ROOT/'workplace/round134/parent-head.txt').read_text().strip(), source_dirty=True,
        parent_receipt_sha256=sha(parent), native={}, renderer={}, gui={}, sources={},
        images={'go': 'golang:1.24-bookworm@sha256:1a6d4452c65dea36aac2e2d606b01b4a029ec90cc1ae53890540ce6173ea77ac',
                'build': 'logh3-ebiten:2@sha256:d7f4d3e04c75fc9e9b2b3606d9f015d8f1bcde806c25d6780780450ad40827ab',
                'gui': 'logh3-hd-preview:20261011-r1@sha256:932b5ad00e356b4252945c35938d9b30b939109abfaaca8b6673f1a17f532aff'})
    stems = ['fortress48-r133-r3-'+k for k in ('hit','empty','off')] + ['fortress48-r134-r3-'+k for k in ('iserlohn','empire')]
    total = 0
    for stem in stems:
        a,b=[ROOT/'workplace/oracle'/(stem+'-'+side) for side in ('a','b')]
        names=sorted(p.name for p in a.iterdir() if p.name!='report.json')
        assert names==sorted(p.name for p in b.iterdir() if p.name!='report.json')
        for n in names:assert (a/n).read_bytes()==(b/n).read_bytes(),(stem,n)
        r,rr=load(a/'report.json'),load(b/'report.json')
        command_a,command_b=r.pop('command'),rr.pop('command')
        assert r==rr and r['dosgolem_head']=='f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c' and not r['dosgolem_dirty']
        assert r['uid']==1000 and r['go_version']=='go1.24.13' and r['seed_once']
        total+=len(names)
        result['native'][stem]=dict(commands=[command_a,command_b],report=r,
            reports_sha256=[sha(a/'report.json'),sha(b/'report.json')],
            equal_files={n:sha(a/n) for n in names})
    assert total==1956
    result['native_equal_files']=total
    pal=load(ROOT/'docs/internals/data/fleet-presentation-spec.json')['palette']
    rgb=[[(v<<2)|(v>>4) for v in c] for c in pal]
    count=pixels=0
    for kind,expected in [('hit',128),('off',42),('iserlohn',125),('empire',128)]:
        typed_path=ROOT/f'workplace/round134/{kind}-typed-final.json'
        typed=load(typed_path)
        folder=ROOT/f'workplace/round134/{kind}-production-final'
        stats=load(folder/'comparison.json')
        assert len(stats)==len(typed['Cases'])==expected
        for c,s in zip(typed['Cases'],stats,strict=True):
            assert c['Name']==s['case'] and s['pixels']==307200 and s['different']==s['max_channel_difference']==0
            assert s['difference_bbox'] is None and len(s['regions'])==8 and all(r['different']==0 for r in s['regions'])
            assert 'production renderer' in s['classification']
            native=ROOT/c['NativePath']
            assert sha(native)==c['NativeSHA256']
            assert native.read_bytes()==(folder/(c['Name']+'-remake.bin')).read_bytes()
            assert load(native.with_name(native.name.replace('-screen.bin','-palette.json')))[:16]==rgb
        count+=len(stats);pixels+=sum(s['pixels'] for s in stats)
        result['renderer'][kind]=dict(typed_sha256=sha(typed_path),typed=typed,comparisons=stats,
            outputs={str(p.relative_to(ROOT)):sha(p) for p in sorted(folder.iterdir())})
    assert (count,pixels)==(423,129945600)
    result['frames'],result['pixels']=count,pixels
    binary=ROOT/'workplace/round134/gin3-verified'
    result['binary_sha256']=sha(binary)
    for kind in ('hd','off'):
        folder=ROOT/f'workplace/round134/gui-{kind}-accepted'
        r=load(folder/'RESULT.json')
        assert r['passed'] and r['normal_entry'] and r['ai']==-1
        assert not any(r[k] for k in ('state_injection','seed_injection','script_flag','passive_override'))
        assert r['binary_sha256']==sha(binary) and r['fortress_units']==[7,8]
        assert sum(h['lost'] for h in r['hits'])==4932 and all(r['public_save_load']['equal'].values())
        assert r['planet']['gun_state']==2 and r['planet']['gun_aim']==161 and r['battle_resume_not_claimed']
        rows=[json.loads(line) for line in (folder/'FRAMES.jsonl').read_text().splitlines()]
        groups={}
        for row in rows:
            p=row.get('playback')
            if p and p['kind']=='fortress':groups.setdefault(p['id'],[]).append(p)
        assert len(groups)==2
        for group in groups.values():
            assert any(p['result'] for p in group)
            if kind=='off':assert all(not p['pictures'] for p in group)
            else:
                expected={(p['segment_hex'],p['frame'],p['side']) for p in group[-1]['pictures']}
                seen=set()
                for p in group:
                    assert p['interval']==8
                    seen.update((x['segment_hex'],x['frame'],x['side']) for x in p['pictures'][:p['applied']])
                assert expected==seen,'正常GUI遺漏原圖格'
        if kind=='hd':
            assert r['f7_during_animation'] and r['audio_format']==[2,2,44100] and 32<r['audio_peak']<32768
            assert sha(folder/'audio.wav')==r['audio_sha256']
            assert any(p['sound_playing'] for g in groups.values() for p in g)
            assert any(row.get('music_track','').startswith(('BGM_6','BGM_7')) for row in rows if row.get('playback',{}).get('kind')=='fortress')
        result['gui'][kind]=dict(result=r,trace_sha256=sha(folder/'FRAMES.jsonl'),result_sha256=sha(folder/'RESULT.json'),
            screenshots={p.name:sha(p) for p in sorted(folder.glob('*.png'))},groups=groups)
    standard=ROOT/'workplace/round134/standard-final.log'
    text=standard.read_text()
    assert text.count('ok  \t')==15 and text.count('[no test files]')==4 and 'FAIL' not in text
    result['standard']=dict(sha256=sha(standard),passed=15,no_tests=4)
    originals=Path('/orig') if Path('/orig/GIN3.COM').exists() else SOURCE/'org_game/extracted'
    result['originals']={n:sha(originals/n) for n in ('GIN3.COM','GRAPH.IMG','COMANI.BMG','ANI_PNL1.A98','3DBLOCK.SPR','COM_FLAG.SPR','ADMIRAL.DTA')}
    names=['internal/game/fortressengagement.go','internal/game/fortressengagement_test.go','internal/game/engagement.go',
        'internal/game/game.go','internal/game/battle.go','internal/game/battlecmd.go','internal/game/intro.go',
        'internal/game/passivebattle.go','internal/game/playbackobservation.go','internal/rules/fortressgun.go',
        'internal/rules/fortress_scene_oracle_test.go','internal/screen/fortress_frame.go','internal/screen/engagement.go',
        'cmd/gin3/telemetry.go','docs/spec/144-fortress-gun-presentation.md','docs/internals/data/engagement-panel.json',
        'tools/fortress_presentation_oracle.sh','tools/fortress_presentation_native_run.sh',
        'tools/scripts/fortress_presentation_native.go','tools/scripts/fortress_frame_prepare.py',
        'tools/scripts/fortress_frame_probe.go','tools/scripts/fortress_presentation_gui.py','tools/scripts/fortress_presentation_receipt.py']
    result['sources']={n:sha(SOURCE/n) for n in names}
    result['limits']=['same-state只限423張完整typed renderer幀與通知指定區域；正常GUI是獨立remake路徑',
        '正常試射命中友軍；無狀態／seed／被動CPU覆寫；不宣稱原版整場同態',
        '公開存讀六表相同；既有088不承諾同戰場續讀，本輪額外續戰探針回星域8，保持未完成',
        '未新增完整登陸／占領正常GUI、Windows／macOS真機、人耳或硬體逐週期驗收']
    return result


parser=argparse.ArgumentParser()
parser.add_argument('--freeze',action='store_true')
args=parser.parse_args()
result=collect()
if args.freeze:
    with OUT.open('x') as f:json.dump(result,f,ensure_ascii=False,indent=2);f.write('\n')
else:assert result==load(OUT),'來源或驗收與凍結索引不符'
print('PASS: 1956雙跑檔；423完整幀129945600像素；正常GUI／音訊／F7／六表存讀；標準15 PASS／4無測試')
print('index',OUT.stat().st_size,sha(OUT))
