{
  description = "Juniper Bible deterministic engine VM (NixOS)";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-24.05";
  };

  outputs = { self, nixpkgs }:
  let
    system = "x86_64-linux";
    pkgs = import nixpkgs { inherit system; };
  in {
    # Engine tools package - all tools needed for running reference implementations
    packages.${system}.engine-tools = pkgs.buildEnv {
      name = "capsule-engine-tools";
      paths = [
        pkgs.bash
        pkgs.coreutils
        pkgs.findutils
        pkgs.gnugrep
        pkgs.gnused
        pkgs.gawk
        pkgs.jq
        pkgs.python3
        pkgs.zip
        pkgs.unzip
        pkgs.gnutar
        pkgs.xz
        pkgs.openssl

        # Reference tools for Bible formats
        pkgs.sword
      ];
    };

    # NixOS VM configuration for deterministic execution
    nixosConfigurations.engine-vm = nixpkgs.lib.nixosSystem {
      inherit system;
      modules = [
        ({ pkgs, ... }: {
          # Disable networking for determinism
          networking.networkmanager.enable = false;
          networking.useDHCP = false;
          networking.firewall.enable = true;

          # Deterministic time and locale settings
          time.timeZone = "UTC";
          i18n.defaultLocale = "C.UTF-8";

          # Environment variables for determinism
          environment.variables = {
            TZ = "UTC";
            LC_ALL = "C.UTF-8";
            LANG = "C.UTF-8";
          };

          # Runner user
          users.users.runner = {
            isNormalUser = true;
            extraGroups = [ "wheel" ];
            password = "";
          };
          security.sudo.wheelNeedsPassword = false;

          # Include engine tools
          environment.systemPackages = [
            self.packages.${system}.engine-tools
          ];

          # Mount points for host communication (9p virtio)
          fileSystems."/work/in" = {
            device = "workin";
            fsType = "9p";
            options = [ "trans=virtio" "version=9p2000.L" "msize=104857600" "cache=loose" ];
          };
          fileSystems."/work/out" = {
            device = "workout";
            fsType = "9p";
            options = [ "trans=virtio" "version=9p2000.L" "msize=104857600" "cache=loose" ];
          };

          # Capsule runner service
          systemd.services.capsule-runner = {
            description = "Juniper Bible Runner";
            wantedBy = [ "multi-user.target" ];
            serviceConfig = {
              Type = "oneshot";
              User = "runner";
              WorkingDirectory = "/work/in";
              ExecStart = "/bin/sh /work/in/runner.sh";
              StandardOutput = "file:/work/out/stdout";
              StandardError = "file:/work/out/stderr";
            };
          };

          system.stateVersion = "24.05";
        })
      ];
    };

    # Development shell
    devShells.${system}.default = pkgs.mkShell {
      buildInputs = [
        pkgs.go
        pkgs.gopls
        pkgs.qemu
        self.packages.${system}.engine-tools
      ];

      shellHook = ''
        export TZ=UTC
        export LC_ALL=C.UTF-8
        export LANG=C.UTF-8
        echo "Juniper Bible development environment"
      '';
    };
  };
}
