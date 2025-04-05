{ pkgs ? import <nixpkgs> {} }:

pkgs.mkShell {
  buildInputs = with pkgs; [
    # Go compiler and tools
    go
    gopls
    gotools
    delve
  ];

  # Environment setup using go env
  shellHook = ''
    # Set up Go environment using go env
    go env -w GO111MODULE=on
    
    go env -w GOPATH=$(pwd)/.go
    mkdir -p $(go env GOPATH)/bin $(go env GOPATH)/src
    export PATH=$(go env GOPATH)/bin:$PATH
  '';
}
