"""只設定隔離Wine的音訊adapter；遊戲的FM／MIDI I/F選擇不變。"""
import hashlib
import json
import os
from pathlib import Path
import subprocess
import sys

assert Path.cwd() == Path('/out') and os.getuid() == 1000
driver = os.environ['LOGH3_WINE_AUDIO_DRIVER']
out = Path(os.environ['LOGH3_WINE_AUDIO_BATCH'])
assert driver in ['pulse', 'alsa'] and (out / 'READY.json').is_file()
record = {'driver': driver, 'wine_prefix': os.environ['WINEPREFIX'],
          'command': sys.argv[1:], 'port_mapping_changed': False}
if driver == 'alsa':
    plugin = Path('/usr/lib/x86_64-linux-gnu/alsa-lib/libasound_module_pcm_pulse.so')
    assert plugin.is_file()
    config = ('</usr/share/alsa/alsa.conf>\n'
              'pcm.!default { type pulse server "' + os.environ['PULSE_SERVER'] + '" }\n'
              'ctl.!default { type pulse server "' + os.environ['PULSE_SERVER'] + '" }\n')
    with (out / 'alsa.conf').open('x') as stream:
        stream.write(config)
    command = ['wine', 'reg', 'add', 'HKCU\\Software\\Wine\\Drivers',
               '/v', 'Audio', '/t', 'REG_SZ', '/d', 'alsa', '/f']
    with (out / 'wine-driver.log').open('x') as log:
        subprocess.run(command, check=True, stdout=log, stderr=subprocess.STDOUT, timeout=45)
    query = subprocess.check_output(['wine', 'reg', 'query',
                                    'HKCU\\Software\\Wine\\Drivers', '/v', 'Audio'],
                                   timeout=10, text=True)
    assert 'alsa' in query
    record.update(alsa_config=config, registry_command=command, registry_query=query,
                  plugin_sha256=hashlib.sha256(plugin.read_bytes()).hexdigest())
with (out / 'DRIVER.json').open('x') as stream:
    json.dump(record, stream, ensure_ascii=False, indent=2)
    stream.write('\n')
os.execvp(sys.argv[1], sys.argv[1:])
