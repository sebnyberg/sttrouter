{
  description = "sttrouter development environment";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
      in
      {
        devShells.default = pkgs.mkShell {
          buildInputs = [
            pkgs.go
            pkgs.golangci-lint
            pkgs.ffmpeg
            pkgs.git
            pkgs.curl
          ];

          shellHook = ''
            echo "🚀 sttrouter development environment loaded"
            echo "Available tools: go, golangci-lint, ffmpeg, git, curl"
            echo "Run 'make help' for available commands"
          '';
        };
      });
}