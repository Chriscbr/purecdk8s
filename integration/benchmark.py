#!/usr/bin/env python3
"""Compare fresh, prebuilt integration apps (requires Go, Node.js, npm).

Reuses the integration fixtures, migration script, and Dockerfile's importer
versions. Setup, imports, tests, compilation, and output checks are not timed.
Each sample includes process startup, construct creation, and App.Synth() to
disk, including upstream's JSII/Node startup. This is not a warm-process
microbenchmark of App.Synth() alone. Run on an otherwise idle machine.
"""

import argparse
import json
import os
from pathlib import Path
import platform
import re
import shutil
import statistics
import subprocess
import tempfile
import time


ROOT = Path(__file__).resolve().parent.parent
DEFAULT_EXAMPLES = [
    "getting-started", "cncf-demo", "plus-http-echo", "plus-web-cache-db",
]


def run(command, cwd, env=None):
    result = subprocess.run(
        [str(arg) for arg in command], cwd=cwd, env=env,
        stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True,
    )
    if result.returncode:
        raise RuntimeError(
            f"{command} failed in {cwd}:\n{result.stdout}{result.stderr}"
        )
    return result.stdout.strip()


def output_tree(directory):
    return {
        str(path.relative_to(directory)): path.read_bytes()
        for path in sorted(directory.rglob("*")) if path.is_file()
    }


def positive_int(value):
    number = int(value)
    if number < 1:
        raise argparse.ArgumentTypeError("must be at least 1")
    return number


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("examples", nargs="*", help="fixture names (default: four core/plus examples)")
    parser.add_argument("--runs", type=positive_int, default=15)
    parser.add_argument("--warmups", type=positive_int, default=2)
    parser.add_argument("--json", type=Path, help="save environment and raw samples")
    args = parser.parse_args()
    examples = args.examples or DEFAULT_EXAMPLES
    for name in examples:
        if Path(name).name != name or not (ROOT / "integration/examples" / name / "go.mod").is_file():
            parser.error(f"unknown integration example: {name}")
    for tool in ("go", "node", "npm"):
        if shutil.which(tool) is None:
            parser.error(f"{tool} must be installed")

    dockerfile = (ROOT / "integration/Dockerfile").read_text()
    cli_version = re.search(r"ARG CDK8S_CLI_VERSION=(\S+)", dockerfile).group(1)
    constructs_version = re.search(r'"constructs@([^"]+)"', dockerfile).group(1)
    report = {
        "platform": platform.platform(),
        "cpu": run(["sysctl", "-n", "machdep.cpu.brand_string"], ROOT)
        if platform.system() == "Darwin" else platform.processor(),
        "go": run(["go", "version"], ROOT),
        "node": run(["node", "--version"], ROOT),
        "cdk8s_cli": cli_version,
        "commit": run(["git", "rev-parse", "HEAD"], ROOT),
        "runs": args.runs,
        "warmups": args.warmups,
        "scope": "fresh prebuilt process through manifest writes; warm filesystem caches",
        "results": [],
    }
    print(f"{report['cpu']}; {report['platform']}; {report['go']}; Node {report['node']}", flush=True)
    with tempfile.TemporaryDirectory(prefix="purecdk8s-benchmark-") as temporary:
        scratch = Path(temporary)
        print("Preparing importers and integration applications (not timed)...", flush=True)
        run(["npm", "install", "--prefix", scratch / "npm", "--no-audit", "--no-fund",
             f"cdk8s-cli@{cli_version}", f"constructs@{constructs_version}"], scratch)
        pure_cli = scratch / "purecdk8s"
        run(["go", "build", "-o", pure_cli, "./cmd/purecdk8s"], ROOT)
        projects = {}
        modules = {}
        for name in examples:
            for implementation in ("upstream", "pure"):
                print(f"  {name}: {implementation}", flush=True)
                project = scratch / implementation / name
                shutil.copytree(ROOT / "integration/examples" / name, project,
                                ignore=shutil.ignore_patterns("imports", "dist"))
                cli = scratch / "npm/node_modules/.bin/cdk8s"
                if implementation == "pure":
                    run([ROOT / "integration/migrate.sh", project, ROOT], ROOT)
                    cli = pure_cli
                run([cli, "import", "--no-save"], project)
                run(["go", "mod", "tidy"], project)
                run(["go", "test", "./..."], project)
                modules[name, implementation] = run(["go", "list", "-m", "all"], project)
                if implementation == "pure" and re.search(
                    r"github.com/(aws/(constructs-go|jsii-runtime-go)|cdk8s-team/)",
                    modules[name, implementation],
                ):
                    raise RuntimeError(f"{name} retained an upstream runtime dependency")
                run(["go", "build", "-o", project / "benchmark-app", "."], project)
                projects[name, implementation] = project

        print("\nMedian wall time; byte-for-byte output checked on every run:")
        print(f"{'Example':<24} {'Resources':>9} {'Upstream ms':>12} {'Pure ms':>10} {'Speedup':>9}", flush=True)
        for name in examples:
            samples = {"upstream": [], "pure": []}
            expected = None
            for iteration in range(args.warmups + args.runs):
                # Alternate which implementation runs first to reduce ordering bias.
                order = ("upstream", "pure") if iteration % 2 == 0 else ("pure", "upstream")
                for implementation in order:
                    project = projects[name, implementation]
                    output = project / "dist"
                    if output.exists():
                        shutil.rmtree(output)
                    env = dict(os.environ, CDK8S_OUTDIR=str(output))
                    for variable in ("ENVIRONMENT", "CDK8S_DISABLE_SORT", "CDK8S_RECORD_CONSTRUCT_METADATA"):
                        env.pop(variable, None)
                    start = time.perf_counter_ns()
                    run([project / "benchmark-app"], project, env)
                    elapsed_ms = (time.perf_counter_ns() - start) / 1_000_000
                    actual = output_tree(output)
                    if not actual:
                        raise RuntimeError(f"{name}/{implementation} produced no output")
                    if expected is None:
                        expected = actual
                    if actual != expected:
                        raise RuntimeError(f"{name}/{implementation} output differs from upstream warmup")
                    if iteration >= args.warmups:
                        samples[implementation].append(elapsed_ms)
            medians = {key: statistics.median(values) for key, values in samples.items()}
            speedup = medians["upstream"] / medians["pure"]
            resources = sum(len(re.findall(rb"^kind:", contents, re.MULTILINE))
                            for filename, contents in expected.items() if filename.endswith(".yaml"))
            print(f"{name:<24} {resources:>9} {medians['upstream']:>12.2f} {medians['pure']:>10.2f} {speedup:>8.1f}x", flush=True)
            report["results"].append({
                "example": name, "resources": resources,
                "samples_ms": samples, "median_ms": medians, "speedup": speedup,
                "modules": {key: modules[name, key].splitlines() for key in samples},
            })
    if args.json:
        args.json.write_text(json.dumps(report, indent=2) + "\n")
        print(f"Raw samples saved to {args.json}")


if __name__ == "__main__":
    main()
