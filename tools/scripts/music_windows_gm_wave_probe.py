"""研究 363：固定來源 GM 與保留官方 SDK 的同輸入波形診斷。

先核對三首匯出 MIDI；bash tools/music_wine.sh python3 /tools/scripts/music_windows_gm_wave_probe.py。
所有採用來源與結果均記入本機收據；沒有原版 MPU／人耳一致聲明。
"""
import array
import hashlib
import json
import math
from pathlib import Path
import subprocess
import shutil
import sys
import wave

root = Path('/out')
assert sys.argv[1:] in [[], ['--utf8']]
utf8 = sys.argv[1:] == ['--utf8']
out = root/('gm-wave-utf8-comparison' if utf8 else 'gm-wave-comparison')
assert not out.exists(), '拒絕覆蓋波形診斷'
out.mkdir()


def sha(p):
    assert p.is_file() and not p.is_symlink()
    with p.open('rb') as f:
        return hashlib.file_digest(f, 'sha256').hexdigest()


def win(p):
    return 'Z:'+str(p).replace('/', '\\')


font = root/'gm-probe/soundfont/FluidR3_GM.sf2'
assert sha(font) == '74594e8f4250680adf590507a306655a299935343583256f3b722c48a1bc1cb0'
sdk = root/'gm-probe/tools/fluidsynth.exe'
package = root/('gm-source-utf8-probe/封裝 測試' if utf8 else 'gm-source-probe/封裝 測試')
current = package/'tools/fluidsynth.exe'
current_font = package/'soundfont/FluidR3_GM.sf2' if utf8 else font
assert sha(current_font) == sha(font)
assert sha(sdk) == '0d6b4a6a747bd0cd5c62a1a42e21f8f1b7b579ce4c8dae105c6920b12ae9d25e'
# SDK 身份另與完整舊收據互核。
gm = json.loads((root/'gm-full/GM-SOURCES.json').read_text())
assert sha(sdk) == gm['tools_sha256'][win(sdk)]
tracks = {t['track']: t for t in gm['tracks']}
commands = []
results = {}
for name in ['BGM_0M', 'END_2M', 'OPEN_M']:
    midi = root/'gm-source-mid'/(name+'.mid')
    assert sha(midi) == tracks[name]['midi_sha256']
    unicode_midi = midi
    if utf8:
        unicode_midi = out/'輸入 曲目'/(name+'.mid')
        unicode_midi.parent.mkdir(exist_ok=True)
        shutil.copyfile(midi, unicode_midi)
        (out/'輸出 波形').mkdir(exist_ok=True)
    records, pcm = {}, {}
    variants = [('sdk', sdk), ('source', current)]
    if name == 'BGM_0M':
        variants += [('source-repeat', current), ('sdk-repeat', sdk)]
    for side, program in variants:
        wav = out/(name+'-'+side+'.wav')
        sf, mid = font, midi
        if utf8 and side.startswith('source'):
            wav = out/'輸出 波形'/wav.name
            sf, mid = current_font, unicode_midi
        args = ['wine', win(program), '-ni', '-F', win(wav), '-r', '44100', '-g', '0.8', win(sf), win(mid)]
        commands.append(args)
        p = subprocess.run(args, stdout=subprocess.PIPE, stderr=subprocess.STDOUT, timeout=900)
        (out/(name+'-'+side+'.log')).write_bytes(p.stdout)
        assert p.returncode == 0, p.stdout.decode(errors='replace')[-3000:]
        with wave.open(str(wav)) as w:
            assert w.getnchannels() == 2 and w.getsampwidth() == 2 and w.getframerate() == 44100
            data = w.readframes(w.getnframes())
            pcm[side] = data
            records[side] = {'path': str(wav.relative_to(root)), 'frames': w.getnframes(), 'wav_sha256': sha(wav),
                'pcm_sha256': hashlib.sha256(data).hexdigest(), 'program_sha256': sha(program)}
        print(name, side, records[side], flush=True)
    a, b = array.array('h'), array.array('h')
    a.frombytes(pcm['sdk'])
    b.frombytes(pcm['source'])
    overlap = min(len(a), len(b))
    count, squares, maximum = 0, 0, 0
    for x, y in zip(a, b):
        diff = abs(x-y)
        count += diff != 0
        squares += diff*diff
        maximum = max(maximum, diff)
    results[name] = {'midi_sha256': sha(midi), 'renders': records,
        'equal_pcm': pcm['sdk'] == pcm['source'], 'compared_samples': overlap,
        'length_difference_samples': len(b)-len(a), 'different_samples': count,
        'max_abs_difference_s16': maximum, 'rmse_s16': math.sqrt(squares/overlap)}
    if name == 'BGM_0M':
        results[name]['source_repeat_equal_pcm'] = pcm['source'] == pcm['source-repeat']
        results[name]['sdk_repeat_equal_pcm'] = pcm['sdk'] == pcm['sdk-repeat']
result = {'scope': '同 MIDI／SF2／2.3.1 float DSP／合成參數的 SDK 與來源試建比較',
    'soundfont_sha256': sha(font), 'commands': commands, 'tracks': results,
    'human_listening_verified': False, 'formal_tool_replacement': False}
(out/'RESULT.json').write_text(json.dumps(result, ensure_ascii=False, indent=2)+'\n')
print(json.dumps(results, ensure_ascii=False, indent=2), flush=True)
