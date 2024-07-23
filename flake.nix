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
            in
            pkgs.nixosTest {
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

                  environment = {
                    systemPackages = with pkgs; [
                      glxinfo # grab information about GPU
                      go
                      kitty
                      nordzy-cursor-theme # used in SetCursor() test
                    ];
                    variables = {
                      "WLR_RENDERER_ALLOW_SOFTWARE" = 1;
                    };
                  };

                  services.getty.autologinUser = user;

                  virtualisation.qemu = {
                    # package = lib.mkForce pkgs.qemu_full;
                    options = [
                      "-smp 2"
                      "-m 4G"
                      "-vga none"
                      "-device virtio-gpu-pci"
                      # needs qemu_full:
                      # "-device virtio-vga-gl"
                      # "-display egl-headless,gl=core"
                      # "-display gtk,gl=on"
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
                            glxinfo -B > "$HOME/glxinfo"
                            cd ${./.}
                            go test -v 2>&1 | tee -a "$HOME/test.log"
                            echo $? > "$HOME/test-finished"
                            hyprctl dispatch exit
                          '';
                      hyprlandConf =
                        pkgs.writeText "hyprland.conf"
                          # hyprlang
                          ''
                            bind = SUPER, Q, exec, kitty # Bind() test need at least one bind
                            exec-once = kitty sh -c ${testScript}
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
                  machine.wait_for_file("${home}/test-finished")

                  print(machine.succeed("cat ${home}/glxinfo || true"))
                  print(machine.succeed("cat ${home}/test.log"))
                  print(machine.succeed("test $(cat ${home}/test-finished) -eq 0"))
                '';
            };
        }
      );
    };
}
