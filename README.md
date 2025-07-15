# hyprland-go

[![Go Reference](https://pkg.go.dev/badge/github.com/thiagokokada/hyprland-go.svg)](https://pkg.go.dev/github.com/thiagokokada/hyprland-go)
[![Go](https://github.com/thiagokokada/hyprland-go/actions/workflows/go.yml/badge.svg)](https://github.com/thiagokokada/hyprland-go/actions/workflows/go.yml)
[![Test](https://github.com/thiagokokada/hyprland-go/actions/workflows/nix.yaml/badge.svg)](https://github.com/thiagokokada/hyprland-go/actions/workflows/nix.yaml)
[![Hyprland](https://img.shields.io/badge/Hyprland-0.47.2-blue)](https://github.com/hyprwm/Hyprland)
[![stability-alpha](https://img.shields.io/badge/stability-alpha-f4d03f.svg)](https://github.com/mkenney/software-guides/blob/master/STABILITY-BADGES.md#alpha)

An unofficial Go wrapper for Hyprland's IPC.

## Getting started

```
go get -u github.com/thiagokokada/hyprland-go
```

Look at the [`examples`](./examples) directory for examples on how to use the
library.

## Why?

- Go: it is a good language to prototype scripts thanks to `go run`, have low
  startup times after compilation and allow creation of static binaries that
  can be run anywhere (e.g.: you can commit the binary to your dotfiles, so you
  don't need to have a working Go compiler!)
- Good developer experience: the API tries to support the most common use cases
  easily while also supporting low-level usage if necessary. Also includes
  proper error handling to make it easier to investigate issues
- Performance: uses tricks like `bufio` and `bytes.Buffer` to archieve high
  performance:
  ```console
  # see examples/hyprtabs
  $ hyperfine -N ./hyprtabs # 1 window in workspace
  Benchmark 1: ./hyprtabs
    Time (mean ± σ):       5.5 ms ±   2.0 ms    [User: 0.8 ms, System: 3.0 ms]
    Range (min … max):     2.2 ms …  11.8 ms    443 runs

  $ hyperfine -N ./hyprtabs # 10 windows in workspace, 122 commands to IPC!
    Benchmark 1: ./hyprtabs
      Time (mean ± σ):      12.0 ms ±   4.7 ms    [User: 0.9 ms, System: 3.4 ms]
      Range (min … max):     4.4 ms …  20.0 ms    490 runs

  $ hyperfine -N ./hyprtabs # 20 windows in workspace, 242 commands to IPC!!
    Benchmark 1: ./hyprtabs
      Time (mean ± σ):      24.0 ms ±  10.9 ms    [User: 0.9 ms, System: 3.3 ms]
      Range (min … max):     9.0 ms …  44.4 ms    77 runs
  ```
  Compare the results above with the original
  [`hyprtabs.sh`](https://gist.github.com/Atrate/b08c5b67172abafa5e7286f4a952ca4d):
  <details>

      $ hyperfine ./hyprtabs.sh # 1 window in workspace
      Benchmark 1: ./hyprtabs.sh
        Time (mean ± σ):     103.0 ms ±   8.1 ms    [User: 51.6 ms, System: 88.1 ms]
        Range (min … max):    92.6 ms … 122.3 ms    30 runs

      $ hyperfine ./hyprtabs.sh # 10 windows in workspace
      Benchmark 1: ./hyprtabs.sh
        Time (mean ± σ):     115.5 ms ±   9.6 ms    [User: 50.2 ms, System: 85.8 ms]
        Range (min … max):    94.8 ms … 136.8 ms    28 runs

      $ hyperfine ./hyprtabs.sh # 20 windows in workspace
      Benchmark 1: ./hyprtabs.sh
        Time (mean ± σ):     121.5 ms ±   5.8 ms    [User: 50.7 ms, System: 82.4 ms]
        Range (min … max):   112.6 ms … 133.6 ms    23 runs

  </details>
- Zero dependencies: smaller binary sizes

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
- [Events:](https://wiki.hyprland.org/Plugins/Development/Event-list/) to
  subscribe and handle Hyprland events, see
  [events](./examples/events/events.go) for an example on how to use it.

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
