{ jfshPackages }:

{
  config,
  lib,
  pkgs,
  ...
}:

let
  cfg = config.programs.jfsh;
  yaml = pkgs.formats.yaml { };
in
{
  options.programs.jfsh = {
    enable = lib.mkEnableOption "jfsh, a terminal-based Jellyfin client";

    package = lib.mkOption {
      type = lib.types.package;
      default = jfshPackages.${pkgs.stdenv.hostPlatform.system}.default;
      defaultText = lib.literalExpression "inputs.jfsh.packages.\${pkgs.system}.default";
      description = "The jfsh package to install.";
    };

    settings = {
      host = lib.mkOption {
        type = lib.types.nonEmptyStr;
        description = "URL of the Jellyfin server, including the HTTP or HTTPS scheme.";
        example = "https://jellyfin.example.com";
      };

      username = lib.mkOption {
        type = lib.types.nonEmptyStr;
        description = "Jellyfin username.";
        example = "sammar";
      };

      device = lib.mkOption {
        type = lib.types.nullOr lib.types.nonEmptyStr;
        default = null;
        description = "Device name reported to Jellyfin. The hostname is used when unset.";
        example = "living-room";
      };

      skipSegments = lib.mkOption {
        type = lib.types.listOf (
          lib.types.enum [
            "Unknown"
            "Commercial"
            "Preview"
            "Recap"
            "Outro"
            "Intro"
          ]
        );
        default = [ ];
        description = "Jellyfin media segment types to skip automatically.";
        example = [
          "Intro"
          "Outro"
        ];
      };
    };

    passwordFile = lib.mkOption {
      type = lib.types.nullOr lib.types.str;
      default = null;
      description = ''
        Absolute path to a runtime file containing the Jellyfin password.
        Use a string rather than a Nix path so the secret is not copied into the Nix store.
      '';
      example = "/run/secrets/jfsh-password";
    };
  };

  config = lib.mkIf cfg.enable {
    assertions = [
      {
        assertion = lib.hasPrefix "http://" cfg.settings.host || lib.hasPrefix "https://" cfg.settings.host;
        message = "programs.jfsh.settings.host must start with http:// or https://";
      }
      {
        assertion = cfg.passwordFile == null || lib.hasPrefix "/" cfg.passwordFile;
        message = "programs.jfsh.passwordFile must be an absolute runtime path";
      }
    ];

    home.packages = [ cfg.package ];

    # Keep generated configuration immutable; jfsh stores tokens and IDs under XDG_STATE_HOME.
    xdg.configFile."jfsh/jfsh.yaml".source = yaml.generate "jfsh.yaml" (
      {
        host = cfg.settings.host;
        username = cfg.settings.username;
        skip_segments = cfg.settings.skipSegments;
      }
      // lib.optionalAttrs (cfg.settings.device != null) {
        device = cfg.settings.device;
      }
      // lib.optionalAttrs (cfg.passwordFile != null) {
        password_file = cfg.passwordFile;
      }
    );
  };
}
