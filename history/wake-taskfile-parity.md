# Wake ⇄ default Taskfile parity tracker

Goal: every operation the default generated Taskfiles provide can be done with
wake (`wails.yaml`, `WAILS_EXP_USE_WAKE`) and no Taskfile; `wails3 migrate`
converts 100% of stock generated projects; stretch: port modified Taskfiles.

Base: `codex/hcl-build-system` @ bb687a997 (PR 6131). Work branch:
`feat/wake-taskfile-parity`.

## Baseline findings (2026-10-28)

- Stock migration: 0/9 templates migrate. Every template hits the same 8
  blockers: 7 × `dynamic-task-reference` inside stock tasks, 1 × missing
  `frontend/dist` on a fresh project. Existing golden test only covers
  `examples/badge` (historical Taskfiles, committed `dist`).
- Migration bindings: `-i` not recognised → TS projects flip to classes;
  `-time-type=Date` dropped.
- Desktop build/package/dev/run on macOS works end-to-end (cached no-op build
  118 ms; dev hot rebuild 1.5 s; clean process reaping).

## Parity matrix (gaps only; everything else COVERED)

| Taskfile task(s) | Gap | WP |
| --- | --- | --- |
| migrate stock projects | dynamic refs, frontend/dist, `-i`, time-type | WP1 |
| windows `build:native` | `-H windowsgui` missing (console window) | D1 |
| windows `build` cross | needed zig/docker; Taskfile used pure-Go CGO=0 | D1 |
| darwin `build:native` | default `-mmacosx-version-min=12.0` not applied | D1 |
| dev builds | `-gcflags=all=-l` dropped | D1 |
| `*:build:docker` | no GOMODCACHE mount, root-owned output, `CGO_CFLAGS=-w` | D1 |
| `setup:docker` | no CLI path to obtain wails-cross; stale hints | D1 |
| darwin `codesign:adhoc` | bundle not ad-hoc sealed | D2 |
| darwin `run` | `Info.dev.plist` (ATS local networking) unused | D2 |
| `sign*` | published as `X.app.signed` / `X.signed.exe` | D2 |
| `build:server`, `run:server` | no server mode | D3 |
| `build:docker`, `run:docker` | no server Docker image | D3 |
| root `VITE_PORT` / `WAILS_VITE_PORT` | dev port only via `--port` | D4 |
| common `go:mod:tidy` | never run | D4 |
| android `package`, `package:fat` | production APK refused | M2 |
| mobile bindings tag | only added for `run` (dead tasks; low) | M2 |
| ios `logs*`, android `logs*` | no standalone log streaming | M1 |
| `ios:device:list` | no iOS device listing | M1 |
| ios `ensure-simulator`, android `ensure-emulator` | no auto-provisioning | M1 |
| ios `deploy-device` | `--device` mandatory | M1 |
| ios `xcode`, android `studio` | no way to open native projects | pending |
| `frontend:vendor:puppertino` | no equivalent | pending |
| windows `install:msix:tools`, `--use-msix-tool` | missing; `tool msix` unregistered | pending |
| migration info diagnostics | stock aux tasks dropped silently; VITE_PORT blocks | pending (after WP1) |
| linux `package` default formats | wake default appimage only (Taskfile: +deb,rpm,aur) | pending (migration) |

## Verification log

- (filled in as work packages land)
