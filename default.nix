{ go }: with go.stdenv.lib;
{ name ? null, repository, contents, dockerConfig }:
let
  stdenv = go.stdenv;
  hashOf = path: head (splitString "-" (removePrefix "${builtins.storeDir}/" path));
  nix2docker = stdenv.mkDerivation {
    name = "nix2docker";
    buildInputs = [ go ];
    buildCommand = ''
      mkdir src
      cp -a ${./.} src/nix2docker
      GOPATH=$PWD go build -o $out nix2docker
    '';
  };
  name' = if name != null
    then name
    else contents.name;
  drv = stdenv.mkDerivation {
    name = "${name'}.tar.gz";
    preferLocalBuild = true;

    buildCommand = nix2docker;
    passAsFile = [ "config" ];
    exportReferencesGraph = [ "closure" contents ];
    config = builtins.toJSON {
      Repository = repository;
      Paths = [ contents ];
      Graphs = [ "closure" ];
      DockerConfig = dockerConfig;
    };
  };
in drv // { imageName = "${repository}:${hashOf drv}"; }
