{
  lib,
  buildGo122Module,
  fetchFromGitHub,
  git,
  testers,
  makeWrapper,
  versionMeta ? "",
}: let
  currentVersion = (lib.importJSON ./version.json).version;
in
  buildGo122Module rec {
    pname = "githooks";

    version =
      if versionMeta != ""
      then "${currentVersion}+nix.${versionMeta}"
      else currentVersion;

    src = ../../.;
    # In nixpkgs that should be:
    # fetchFromGitHub {
    #   owner = "gabyx";
    #   repo = "githooks";
    #   rev = "v${version}";
    #   hash = "sha256-TD6RiZ4Bq8gU444erYDkuGrKkpDrjMTrSH3qZpBwwqk=";
    # };

    modRoot = "./githooks";
    vendorHash = "sha256-ZcDD4Z/thtyCvXg6GzzKC/FSbh700QEaqXU8FaZaZc4=";
    nativeBuildInputs = [makeWrapper];
    buildInputs = [git];

    ldflags = [
      "-s" # Disable symbole table.
      "-w" # Disable DWARF generation.
    ];

    tags = ["package_manager_enabled"];

    doCheck = false;

    postConfigure = ''
      GH_BUILD_VERSION="${version}" \
        GH_BUILD_TAG="v${version}" \
        go generate -mod=vendor ./...
    '';

    postInstall = ''
      mv "$out/bin/cli" "$out/bin/githooks-cli"
      mv "$out/bin/runner" "$out/bin/githooks-runner"
      mv "$out/bin/dialog" "$out/bin/githooks-dialog"

      wrapProgram "$out/bin/githooks-cli" --prefix PATH : ${lib.makeBinPath [git]}
      wrapProgram "$out/bin/githooks-runner" --prefix PATH : ${lib.makeBinPath [git]}
    '';

    passthru.tests.version = testers.testVersion {
      package = "githooks-cli";
      command = "githooks-cli --version";
      inherit version;
    };

    meta = with lib; {
      description = "Githooks is a Git hooks manager with per-repo and shared Git hooks including version control";
      homepage = "https://github.com/gabyx/Githooks";
      license = licenses.mpl20;
      maintainers = with maintainers; [gabyx];
      mainProgram = "githooks-cli";
    };
  }
