#!/usr/bin/env python3
# -*- coding: utf-8 -*-

import json
import re
import subprocess

import psutil

def varnishstat():
    result = subprocess.run(
        ['/usr/bin/varnishstat', '-1', '-j'],
        capture_output=True,
        text=True)
    assert result.returncode == 0, f'Command failed with exit code {result.returncode}'
    return json.loads(result.stdout)

def main():
    # Get 'varnishstat' metrics.
    metrics = varnishstat()

    # Filter out some metrics we prefer to ignore.
    blacklist = re.compile(r'^(?:LCK|MEMPOOL)[.]')
    metrics = {k: v for k, v in metrics.items() if not blacklist.match(k)}

    # Add some additional ad-hoc metrics following the 'varnishstat' output
    # format.
    for k, v in psutil.cpu_times(percpu=False)._asdict().items():
        metrics[f'CPU.time.{k}'] = {
            'description': f"psutil.cpu_times(percpu=False)._asdict()['{k}'] * 1000",
            'flag': 'c',
            'format': 'i',
            'value': int(v * 1000),
        }
    for k, v in psutil.virtual_memory()._asdict().items():
        metrics[f'MEMORY.{k}'] = {
            'description': f"psutil.virtual_memory()._asdict()['{k}']",
            'flag': 'g',
            'format': 'i' if k in ['percent'] else 'B',
            'value': int(v),
        }
    for k, v in psutil.swap_memory()._asdict().items():
        metrics[f'SWAP.{k}'] = {
            'description': f"psutil.swap_memory()._asdict()['{k}']",
            'flag': 'g' if k in ['total', 'used', 'free', 'percent'] else 'c',
            'format': 'i' if k in ['percent'] else 'B',
            'value': int(v),
        }
    for nic, counters in psutil.net_io_counters(pernic=True).items():
        for k, v in counters._asdict().items():
            metrics[f'NET.{nic}.{k}'] = {
                'description': f"psutil.net_io_counters(pernic=True)['{nic}']._asdict()['{k}']",
                'flag': 'c',
                'format': 'B' if 'bytes' in k else 'i',
                'value': int(v),
            }

    # Print metrics as JSON.
    print(json.dumps(metrics))

if __name__ == '__main__':
    main()
