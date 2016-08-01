{ go }: with go.stdenv.lib;
{ name ? null, repository, contents, dockerConfig }:
let
  stdenv = go.stdenv;
  hashOf = path: head (splitString "-" (removePrefix "${builtins.storeDir}/" path));
  zip = xs: ys: concatLists (zipListsWith (x: y: [x y]) xs ys);
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
    else (head contents').name;
  contents' = flatten contents;
  closures = genList (n: "closure" + optionalString (n != 0) "-${toString n}") (length contents');
  drv = stdenv.mkDerivation {
    name = "${name'}.tar.gz";
    preferLocalBuild = true;

    buildCommand = nix2docker;
    passAsFile = [ "config" ];
    exportReferencesGraph = zip closures contents';
    config = builtins.toJSON {
      Repository = repository;
      Paths = contents';
      Graphs = closures;
      DockerConfig = dockerConfig;
    };
  };
in drv // {
  imageName = "${repository}:${hashOf drv}";
  imageTag = "${hashOf drv}";
}
