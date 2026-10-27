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
| ios `xcode`, android `studio` | no way to open native projects | M3 |
| `frontend:vendor:puppertino` | no equivalent | D5 |
| windows `install:msix:tools`, `--use-msix-tool` | missing; `tool msix` unregistered | D5 |
| migration info diagnostics | stock aux tasks dropped silently; VITE_PORT blocks | WP1b |
| linux `package` default formats | wake default appimage only (Taskfile: +deb,rpm,aur) | WP1c (`linux.formats`) |
| `DEV=true` builds | no `wails3 build` equivalent | D4b (`build --dev`) |
| `setup:docker` | `setup` children registered at root (pre-existing clir bug) | D1c |
| linux `create:appimage` | stock desktop template had a whitespace-only line; migration fed the deb desktop file to AppImage | D2b |
| windows `create:nsis:installer` | makensis fails under a non-UTF-8 locale on POSIX hosts | D5b |

All gaps above are closed on `feat/wake-taskfile-parity`.

## Stretch: modified Taskfiles

`TestCustomTaskfileCorpus`: 16/16 customisation cases migrate as expected
(root vars, removed mobile, extra tags, static ldflags, env, custom pre-build
and post-package steps to generated hook scripts, frontend build/dev commands,
unrelated tasks, darwin min version, NSIS scope, obfuscation). Dynamic build
flags, unknown template vars and rewritten build commands block with precise
diagnostics. Limitation: generated hook scripts are POSIX-only (the hook schema
has one script per phase); Windows hosts get an info diagnostic.

## Verification log (2026-10-28)

- Unit/integration: `go test ./internal/wake/... ./internal/commands/
  ./internal/dev/... ./internal/doctor/... ./internal/setupwizard/...
  ./internal/templates/... ./cmd/wails3/` all ok; vet clean.
- Stock migration: `TestStockTemplatesMigrateWithoutBlockers` 9/9 templates
  migrate complete and activate; manual check on react and vue projects.
- macOS (arm64, native): 34/34 operation checks on a migrated vue project:
  build, sealed ad-hoc bundle, minos 12.0, cached no-op (0 ran), universal
  lipo, DMG, obfuscated, tags, windows amd64/arm64 cross with GUI subsystem,
  server build + cross, `build --dev` flags, android aab/apk (debug-signed,
  apksigner verified)/universal, puppertino, android devices/studio, setup
  docker, ios devices, run/dev plans, clean. Live: `wails3 dev` hot rebuild
  4 s with clean teardown, `wails3 run`, `wails3 run --server` served HTTP 200,
  `wails3 android run` on the wails-test AVD (Greet round-trip "Hello" from Go).
- Linux (Fedora 44 container, arm64): build, ELF, cached no-op, package
  appimage+deb+rpm+archlinux (dpkg-deb/rpm metadata valid, AppImage desktop
  Categories=Development), windows cross, server, `build --dev`, run/dev plans.
- Windows (win-node1, amd64, native): migrate, activate, build, cached, PE
  subsystem GUI, `build --dev`, arm64, server exe, run/dev plans.
- Not verifiable here: iOS on-device/simulator flows (no Xcode.app on any
  reachable Mac), Authenticode/notarization signing (no identities), MSIX
  (needs Windows SDK tools), Linux live GUI dev session (container has no
  display).
