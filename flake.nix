{
  description = "Githooks Dev";

  inputs = {
    # Nixpkgs
    nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";

    # Format the repo with nix-treefmt.
    treefmt-nix = {
      url = "github:numtide/treefmt-nix";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs =
    {
      nixpkgs,
      flake-utils,
      ...
    }@inputs:
    flake-utils.lib.eachDefaultSystem
      # Creates an attribute map `{ devShells.<system>.default = ...}`
      # by calling this function:
      (
        system:
        let
          overlays = [ ];

          # Import nixpkgs and load it into pkgs.
          pkgs = import nixpkgs {
            inherit system overlays;
          };

          # Configure formatter.
          treefmtEval = inputs.treefmt-nix.lib.evalModule pkgs {
            projectRootFile = "README.md";

            # Markdown, JSON, YAML, etc.
            programs.prettier.enable = true;

            # Go
            programs.gofmt.enable = true;

            # Shell.
            programs.shfmt = {
              enable = true;
              indent_size = 4;
            };

            programs.shellcheck.enable = true;
            settings.formatter.shellcheck = {
              options = [
                "-e"
                "SC1091"
              ];
            };

            # Nix.
            programs.nixfmt.enable = true;
          };

          treefmt = treefmtEval.config.build.wrapper;

          # Things needed only at compile-time.
          packages = with pkgs; [
            coreutils
            findutils
            gitFull
            gnugrep
            bash
            jq
            curl
            just

            go_1_24
            golines
            gotools
            golangci-lint
            golangci-lint-langserver

            treefmt
          ];
        in
        with pkgs;
        {
          devShells.default = mkShell {
            # To make CGO and the debugger delve work.
            # https://nixos.wiki/wiki/Go#Using_cgo_on_NixOS
            hardeningDisable = [ "fortify" ];
            inherit packages;
          };
        }
      );
}
