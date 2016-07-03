{ go }:
{ name ? null, repository, contents, dockerConfig }:
let
  nix2docker = go.stdenv.mkDerivation {
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
in
go.stdenv.mkDerivation {
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
}
