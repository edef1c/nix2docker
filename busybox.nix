with import <nixpkgs> {};
let nix2docker = callPackage ./. {}; in
nix2docker {
  repository = "edef/busybox";
  contents = [ busybox ];
  dockerConfig.Entrypoint = [ "/bin/ash" ];
}
