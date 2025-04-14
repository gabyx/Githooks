{
  description = "Githooks Dev";

  nixConfig = {
    extra-substituters = [
      # Nix community's cache server
      "https://cache.nixos.org/"
      "https://nix-community.cachix.org"
    ];
    extra-trusted-public-keys = [
      "nix-community.cachix.org-1:mB9FSh9qf2dCimDSUo8Zy7bkq5CX+/rkCWyvRCYg3Fs="
    ];
  };

  inputs = {
    # Nixpkgs
    nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs =
    {
      self,
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

          # Things needed only at compile-time.
          packages = with pkgs; [
            go_1_24
            golines
            gotools
            golangci-lint
            golangci-lint-langserver
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
