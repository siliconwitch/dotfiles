{ lib, stdenvNoCC, fetchurl, jre, makeDesktopItem, makeWrapper }:

stdenvNoCC.mkDerivation {
  pname = "simsmith";
  version = "18.4";

  src = fetchurl {
    url = "https://www.tcpsas.com/sezioneIV/SimSmith/SimSmith.jar";
    hash = "sha256-FgJFPhR3Mka0JEX9Bad0cLOjLex6w7ciLqofKcs9ifY=";
  };
  dontUnpack = true;
  nativeBuildInputs = [ makeWrapper ];
  desktopItem = makeDesktopItem {
    name = "simsmith";
    desktopName = "SimSmith";
    genericName = "Smith chart simulator";
    exec = "simsmith";
    terminal = false;
    categories = [ "Education" "Science" "Electronics" ];
  };

  installPhase = ''
    runHook preInstall

    install -Dm644 "$src" "$out/share/java/SimSmith.jar"
    makeWrapper ${jre}/bin/java "$out/bin/simsmith" \
      --add-flags "-jar $out/share/java/SimSmith.jar"
    install -Dm644 "$desktopItem/share/applications/simsmith.desktop" \
      "$out/share/applications/simsmith.desktop"

    runHook postInstall
  '';

  meta = {
    description = "Java-based Smith chart and RF circuit simulator";
    homepage = "https://www.ae6ty.com/smith_charts/";
    license = lib.licenses.unfree;
    mainProgram = "simsmith";
    platforms = lib.platforms.linux;
  };
}
