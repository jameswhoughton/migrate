{ pkgs ? import <nixpkgs> { } }:

let unstable = import <unstable> {};

in 
pkgs.mkShell {
  buildInputs = [
    unstable.go
    unstable.go-tools
    unstable.gotools
    unstable.gopls
    unstable.go-outline
    unstable.gocode
    unstable.gopkgs
    unstable.gocode-gomod
    unstable.godef
    unstable.golint
    unstable.delve
    unstable.sqlite

  ];
}
