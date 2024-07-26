{
  description = "Hyprland's IPC bindings for Go";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  };

  outputs =
    { nixpkgs, ... }:
    let
      supportedSystems = [
        "x86_64-linux"
        "aarch64-linux"
      ];

      # Helper function to generate an attrset '{ x86_64-linux = f "x86_64-linux"; ... }'.
      forAllSystems = nixpkgs.lib.genAttrs supportedSystems;

      # Nixpkgs instantiated for supported system types.
      nixpkgsFor = forAllSystems (system: import nixpkgs { inherit system; });
    in
    {
      checks = forAllSystems (
        system:
        let
          pkgs = nixpkgsFor.${system};
        in
        {
          testVm =
            let
              user = "alice";
              uid = 1000;
              home = "/home/${user}";

              # Testing related file paths
              covHtml = "${home}/hyprland-go.html";
              covOut = "${home}/hyprland-go.out";
              glxinfoOut = "${home}/glxinfo.out";
              testFinished = "${home}/test-finished";
              testLog = "${home}/test.log";
            in
            pkgs.testers.runNixOSTest {
              name = "hyprland-go";

              nodes.machine =
                {
                  config,
                  pkgs,
                  lib,
                  ...
                }:
                {
                  boot.loader.systemd-boot.enable = true;
                  boot.loader.efi.canTouchEfiVariables = true;

                  programs.hyprland.enable = true;

                  users.users.${user} = {
                    inherit home uid;
                    isNormalUser = true;
                  };

                  environment.systemPackages = with pkgs; [
                    glxinfo # grab information about GPU
                    gnome-themes-extra # used in SetCursor() test
                    go
                    kitty
                  ];

                  services.getty.autologinUser = user;

                  virtualisation.qemu = {
                    options = [
                      "-smp 2"
                      "-m 4G"
                      "-vga none"
                      "-device virtio-gpu-pci"
                      # Works only in Wayland, may be slightly faster
                      # "-device virtio-gpu-rutabaga,gfxstream-vulkan=on,cross-domain=on,hostmem=2G"
                    ];
                  };

                  # Start hyprland at login
                  programs.bash.loginShellInit =
                    # bash
                    let
                      testScript =
                        pkgs.writeShellScript "hyprland-go-test"
                          # bash
                          ''
                            set -euo pipefail

                            trap 'echo $? > ${testFinished}' EXIT

                            glxinfo -B > ${glxinfoOut} || true
                            cd ${./.}
                            go test -bench=. -coverprofile ${covOut} -v > ${testLog} 2>&1
                            go tool cover -html=${covOut} -o ${covHtml}
                          '';
                      hyprlandConf =
                        pkgs.writeText "hyprland.conf"
                          # hyprlang
                          ''
                            bind = SUPER, Q, exec, kitty # Bind() test need at least one bind
                            exec-once = kitty sh -c ${testScript}
                            animations {
                              # slow, and nobody is looking anyway
                              enabled = false
                            }
                            decoration {
                              # slow, and nobody is looking anyway
                              blur {
                                enabled = false
                              }
                            }
                            cursor {
                              # improve cursor in VM
                              no_hardware_cursors = true
                            }
                          '';
                    in
                    # bash
                    ''
                      if [ "$(tty)" = "/dev/tty1" ]; then
                        Hyprland --config ${hyprlandConf}
                      fi
                    '';
                };

              testScript = # python
                ''
                  start_all()

                  machine.wait_for_unit("multi-user.target")
                  machine.wait_for_file("${testFinished}")

                  print(machine.succeed("cat ${glxinfoOut} || true"))
                  print(machine.succeed("cat ${testLog}"))
                  print(machine.succeed("exit $(cat ${testFinished})"))

                  machine.copy_from_vm("${covOut}")
                  machine.copy_from_vm("${covHtml}")
                '';
            };
        }
      );

      packages = forAllSystems (
        system:
        let
          pkgs = nixpkgsFor.${system};
        in
        rec {
          default = hyprland-go;
          hyprland-go = pkgs.buildGoModule {
            pname = "hyprland-go";
            version = "0.0.1";

            src = ./.;

            subPackages = [
              "examples/hyprctl"
              "examples/hyprtabs"
            ];

            vendorHash = null;
          };
        }
      );
    };
}
