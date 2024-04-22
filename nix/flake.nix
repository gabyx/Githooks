{
  description = "Githooks is a Git hooks manager with per-repo and shared Git hooks including version control";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = inputs @ {self, ...}:
    inputs.flake-utils.lib.eachDefaultSystem (system: let
      overlays = [];
      pkgs = import (inputs.nixpkgs) {inherit system overlays;};
    in {
      packages.default =
        pkgs.callPackage ./pkgs/default.nix {};
    });
}
