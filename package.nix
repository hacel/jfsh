{
  lib,
  buildGoModule,
  makeBinaryWrapper,
  mpv,
  version ? "0.1.17",
  commit ? "",
}:

buildGoModule {
  pname = "jfsh";
  inherit version;

  src = lib.fileset.toSource {
    root = ./.;
    fileset = lib.fileset.unions [
      ./go.mod
      ./go.sum
      ./keys.go
      ./main.go
      ./model.go
      ./update.go
      ./view.go
      ./internal
    ];
  };
  vendorHash = "sha256-rWUdgbTkfdjm+fAQG4OUmYs3cBVDAa+cSZE7T2t1Low=";

  env.CGO_ENABLED = 0;
  ldflags = [
    "-s"
    "-w"
    "-X main.version=${version}"
  ]
  ++ lib.optional (commit != "") "-X main.commit=${commit}";

  nativeBuildInputs = [ makeBinaryWrapper ];

  # Prefer a user-selected mpv while guaranteeing one is available as a fallback.
  postInstall = ''
    wrapProgram "$out/bin/jfsh" \
      --suffix PATH : ${lib.makeBinPath [ mpv ]}
  '';

  meta = {
    description = "Terminal-based Jellyfin client with mpv playback";
    homepage = "https://github.com/hacel/jfsh";
    license = lib.licenses.unlicense;
    mainProgram = "jfsh";
    platforms = [
      "x86_64-linux"
      "aarch64-linux"
      "aarch64-darwin"
    ];
  };
}
