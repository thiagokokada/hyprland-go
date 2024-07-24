# hyprland-go

[![Go](https://github.com/thiagokokada/hyprland-go/actions/workflows/go.yml/badge.svg)](https://github.com/thiagokokada/hyprland-go/actions/workflows/go.yml)
[![Test](https://github.com/thiagokokada/hyprland-go/actions/workflows/nix.yaml/badge.svg)](https://github.com/thiagokokada/hyprland-go/actions/workflows/nix.yaml)
[![Hyprland](https://img.shields.io/badge/Made%20for-Hyprland-blue)](https://github.com/hyprwm/Hyprland)
[![stability-alpha](https://img.shields.io/badge/stability-alpha-f4d03f.svg)](https://github.com/mkenney/software-guides/blob/master/STABILITY-BADGES.md#alpha)

An unofficial Go wrapper for Hyprland's IPC.

## Getting started

```
go get -u github.com/thiagokokada/hyprland-go
```

Look at the [`examples`](./examples) directory for examples on how to use the
library.

## What is supported?

- [Dispatchers:](https://wiki.hyprland.org/Configuring/Dispatchers/) for
  calling dispatchers, batch mode supported, e.g.: `c.Dispatch("exec kitty",
  "exec firefox")`
- [Keywords:](https://wiki.hyprland.org/Configuring/Keywords/) for dealing with
  configuration options, e.g.: (`c.SetKeyword("bind SUPER,Q,exec,firefox",
  "general:border_size 1")`)
- [Hyprctl commands:](https://wiki.hyprland.org/Configuring/Using-hyprctl/)
  most commands are supported, e.g.: `c.SetCursor("Adwaita",
  32)`.
  + Commands that returns a JSON in `hyprctl -j` will return a proper struct,
    e.g.: `c.ActiveWorkspace().Monitor`
- [Raw IPC commands:](https://wiki.hyprland.org/IPC/): while not recommended
  for general usage, sending commands directly to the IPC socket of Hyprland is
  supported for i.e.: performance, e.g.: `c.RawRequest("[[BATCH]] dispatch exec
  kitty, keyword general:border_size 1")`

## Planned

- [Events](https://wiki.hyprland.org/Plugins/Development/Event-list/)

## Development

If you are developing inside a Hyprland session, and have Go installed, you can
simply run:

```console
# -short flag is recommended otherwise this will run some possibly dangerous tests, like TestKill()
go test -short -v
```

Keep in mind that this will probably mess your current session. We will reload
your configuration at the end, but any dynamic configuration will be lost.

We also have tests running in CI based in a [NixOS](https://nixos.org/) VM
using [`nixosTests`](https://wiki.nixos.org/wiki/NixOS_VM_tests). Check the
[`flake.nix`](./flake.nix) file. This will automatically start a VM running
Hyprland and run the Go tests inside it.

To run the NixOS tests locally, install [Nix](https://nixos.org/download/) in
any Linux system and run:

```console
nix --experimental-features 'nix-command flakes' flake check -L
```

If you want to debug tests, it is possible to run the VM in interactive mode by
running:

```console
nix --experimental-features 'nix-command flakes' build .#checks.x86-64_linux.testVm.driverInteractive
./result
```

And you can run `start_all()` to start the VM.

## Credits

- [hyprland-ipc-client](https://github.com/labi-le/hyprland-ipc-client) for
inspiration.
