# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

-

## [0.1.0] - 2026-11-13

### Added

- Canonical machine manifest: `asbuilt manifest` prints the JSON description of
  the machine (host, platform and firmware, CPU, NUMA, memory, PCIe, OS, power,
  virtualization) with a schema version, the list of reachable data sources and
  explicit notes about missing evidence.
- Rule catalog of 33 rules across CPU, memory, NUMA, kernel settings, PCIe,
  power and platform profile, each with a severity, the evidence behind it and,
  where the arithmetic allows, a measured performance loss.
- PCIe link verdict taken from world-readable sysfs attributes, so the
  degraded-link check works without root and inside containers where PCI config
  space is not reachable.
- SMBIOS memory population decoding: memory arrays, devices, channels, sockets
  and the parts and speeds installed in them.
- Expected-profile conformance: `asbuilt check --profile FILE` judges the
  machine against a written profile and reports every mismatch as a finding.
- Manifest diff: `asbuilt diff A.json B.json` compares two manifests and
  reports what changed, with an exit code of 1 when anything differs.
- JSON output for every command via `--json`.
- CI-gate exit codes: `--fail-on critical|warn|info` turns findings of that
  severity or worse into a non-zero exit for use in a pipeline.
- `asbuilt rules` and `asbuilt explain RULE_ID` to browse the catalog.
- `--root DIR` to read a recorded fixture tree instead of the live machine.
- Installers for Linux and macOS (`install.sh`) and Windows (`install.ps1`),
  plus signed-per-target release binaries with a published `checksums.txt`.
- Brand assets rendered from source by `tools/brandgen`.
