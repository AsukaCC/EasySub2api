#!/usr/bin/env python3
"""Verify v0.2.22 adaptations against temporary PostgreSQL and Redis containers."""

import json
import os
from pathlib import Path
import subprocess
import sys


backend = Path(__file__).resolve().parents[1]
subprocess.run(["docker", "info"], check=True, stdout=subprocess.DEVNULL)
package = json.loads(subprocess.check_output(
    ["go", "list", "-json", "-tags=integration", "./internal/repository"], cwd=backend,
))
directory = Path(package["Dir"])
tests = [
    "integration_harness_test.go",
    "fixtures_integration_test.go",
    "release0222_integration_test.go",
    "wallet_points_integration_test.go",
]
files = [str(directory / name) for name in package["GoFiles"] + tests]
environment = os.environ.copy()
environment["CI"] = "true"
result = subprocess.run(
    ["go", "test", "-p=1", "-tags=integration", "-count=1", "-v", "-timeout=10m", *files, *sys.argv[1:]],
    cwd=backend, env=environment,
)
sys.exit(result.returncode)
