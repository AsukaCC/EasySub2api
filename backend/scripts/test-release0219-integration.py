#!/usr/bin/env python3
"""Run release and wallet DB regressions without unrelated legacy ID fixtures.

The repository's full integration suite still contains pre-UUID test fixtures.
Use Go's source-file mode to select current production files and these tests;
the normal integration harness still creates containers and applies migrations.
"""

import json
import os
from pathlib import Path
import subprocess
import sys


backend = Path(__file__).resolve().parents[1]
subprocess.run(["docker", "info"], check=True, stdout=subprocess.DEVNULL)
package = json.loads(subprocess.check_output(
    ["go", "list", "-json", "-tags=integration", "./internal/repository"],
    cwd=backend,
))
directory = Path(package["Dir"])
tests = [
    "integration_harness_test.go",
    "fixtures_integration_test.go",
    "wallet_points_integration_test.go",
    "release0219_integration_test.go",
]
files = [str(directory / name) for name in package["GoFiles"] + tests]
environment = os.environ.copy()
environment["CI"] = "true"  # Docker unavailable must never silently pass.
result = subprocess.run(
    ["go", "test", "-tags=integration", "-count=1", "-v", "-timeout=10m", *files, *sys.argv[1:]],
    cwd=backend, env=environment,
)
sys.exit(result.returncode)
