{
  description = "A terminal-based Jellyfin client with mpv playback";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-26.05";

    home-manager = {
      url = "github:nix-community/home-manager/release-26.05";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs =
    {
      self,
      nixpkgs,
      home-manager,
    }:
    let
      systems = [
        "x86_64-linux"
        "aarch64-linux"
        "aarch64-darwin"
      ];
      forAllSystems = nixpkgs.lib.genAttrs systems;
    in
    {
      packages = forAllSystems (
        system:
        let
          pkgs = nixpkgs.legacyPackages.${system};
        in
        rec {
          jfsh = pkgs.callPackage ./package.nix {
            commit = self.shortRev or self.dirtyShortRev or "dirty";
          };
          default = jfsh;
        }
      );

      apps = forAllSystems (system: {
        default = {
          type = "app";
          program = "${self.packages.${system}.default}/bin/jfsh";
          meta.description = "Run jfsh";
        };
      });

      homeManagerModules.default = import ./nix/home-manager.nix {
        jfshPackages = self.packages;
      };

      checks = forAllSystems (
        system:
        let
          pkgs = nixpkgs.legacyPackages.${system};
          home = home-manager.lib.homeManagerConfiguration {
            inherit pkgs;
            modules = [
              self.homeManagerModules.default
              {
                home = {
                  username = "jfsh-test";
                  homeDirectory = if pkgs.stdenv.isDarwin then "/Users/jfsh-test" else "/home/jfsh-test";
                  stateVersion = "26.05";
                };
                programs.jfsh = {
                  enable = true;
                  settings = {
                    host = "https://jellyfin.example.com";
                    username = "test";
                    device = "test-device";
                    skipSegments = [
                      "Intro"
                      "Outro"
                    ];
                  };
                  passwordFile = "/run/secrets/jfsh-password";
                };
              }
            ];
          };
          generatedConfig = home.config.xdg.configFile."jfsh/jfsh.yaml".source;
        in
        {
          package = self.packages.${system}.default;
          home-manager = home.activationPackage;

          # Verify the module translates typed options without embedding a password.
          home-manager-config = pkgs.runCommand "jfsh-home-manager-config" { } ''
            ${pkgs.gnugrep}/bin/grep -F "host: https://jellyfin.example.com" ${generatedConfig}
            ${pkgs.gnugrep}/bin/grep -F "username: test" ${generatedConfig}
            ${pkgs.gnugrep}/bin/grep -F "device: test-device" ${generatedConfig}
            ${pkgs.gnugrep}/bin/grep -F "password_file: /run/secrets/jfsh-password" ${generatedConfig}
            ${pkgs.gnugrep}/bin/grep -F -- "- Intro" ${generatedConfig}
            ${pkgs.gnugrep}/bin/grep -F -- "- Outro" ${generatedConfig}
            if ${pkgs.gnugrep}/bin/grep -q "^password:" ${generatedConfig}; then
              exit 1
            fi
            touch "$out"
          '';
        }
      );
    };
}
