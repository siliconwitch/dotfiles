{ pkgs, lib, ... }:

let
  # Chromium webapp shortcut → /run/current-system/sw/share/applications
  webapp = display: url: pkgs.makeDesktopItem {
    name        = lib.replaceStrings [ " " ] [ "-" ] display;
    desktopName = display;
    exec        = "${pkgs.chromium}/bin/chromium --app=${url}";
    terminal    = false;
    icon        = "web-browser";
  };
in
{
  environment.systemPackages = [
    (webapp "Discord"               "https://discord.com/app")
    (webapp "Draw.io"               "https://app.diagrams.net")
    (webapp "Figma"                 "https://www.figma.com")
    (webapp "YouTube"               "https://www.youtube.com")
    (webapp "YouTube Music"         "https://music.youtube.com")
  ];
}
