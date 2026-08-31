{ pkgs, lib, username, nixpkgs-master, ... }:

{
  # Nix base settings
  nixpkgs.config.allowUnfree = true;
  nix.settings.experimental-features = [ "nix-command" "flakes" ];
  nix.gc = {
    automatic = true;
    dates = "daily";
    options = "--delete-older-than 7d";
  };

  # Boot & kernel
  boot.loader.systemd-boot.enable = true;
  boot.loader.systemd-boot.configurationLimit = 5;
  boot.loader.efi.canTouchEfiVariables = true;
  boot.kernelParams = [ "i8042.dumbkbd=1" ]; # Lenovo keyboard quirk
  boot.kernelPackages = pkgs.linuxPackages_testing; # Once 7.2 stable lands in nixpkgs: switch to linuxPackages_latest

  # Hardware & firmware
  hardware.enableAllFirmware = true;
  hardware.graphics.enable = true;
  hardware.bluetooth.enable = true;
  services.fwupd.enable = true;
  services.hardware.bolt.enable = true;
  nixpkgs.config.segger-jlink.acceptLicense = true;
  nixpkgs.config.permittedInsecurePackages = [ "segger-jlink-qt4-952" ];
  services.udev.packages = with pkgs; [ saleae-logic-2 segger-jlink ];

  # Remap Lenovo Copilot key
  services.keyd = {
    enable = true;
    keyboards.default = {
      ids = [ "*" ];
      settings = {
        global.chord_timeout = 10;
        main."leftmeta+leftshift+f23" = "leftmeta";
      };
    };
  };

  # Power buttons (niri handles sleep, along with the script below)
  services.logind.settings.Login = {
    HandleLidSwitch = "ignore";
    HandleLidSwitchExternalPower = "ignore";
    HandleLidSwitchDocked = "ignore";
    HandlePowerKey = "ignore";
    HandleSuspendKey = "ignore";
  };

  # Power notifications and tasks
  services.udev.extraRules = ''
    SUBSYSTEM=="power_supply", KERNEL=="ADP0", ENV{POWER_SUPPLY_ONLINE}=="0", ACTION=="change", TAG+="systemd", ENV{SYSTEMD_WANTS}+="ac-unplug-suspend.service"
    SUBSYSTEM=="power_supply", KERNEL=="ADP0", ACTION=="change", TAG+="systemd", ENV{SYSTEMD_WANTS}+="ac-notifications.service"
  '';

  systemd.services.ac-unplug-suspend = {
    serviceConfig.Type = "oneshot";
    script = ''
      # Grace period so unplugging to open the lid right away doesn't suspend
      sleep 30

      grep -q closed /proc/acpi/button/lid/LID0/state || exit 0
      [ "$(cat /sys/class/power_supply/ADP0/online)" = 0 ] || exit 0

      systemctl suspend
    '';
  };

  systemd.services.ac-notifications = {
    serviceConfig.Type = "oneshot";
    path = with pkgs; [ util-linux libnotify ];
    script = ''
      rd=/run/user/$(id -u ${username})
      [ -S "$rd/bus" ] || exit 0

      if [ "$(cat /sys/class/power_supply/ADP0/online)" = 1 ]; then
        message="Plugged in"
      else
        message="Unplugged"
      fi
      runuser -u ${username} -- env XDG_RUNTIME_DIR="$rd" DBUS_SESSION_BUS_ADDRESS="unix:path=$rd/bus" notify-send "Power" "$message"
    '';
  };

  systemd.user.services.battery-notifications = {
    wantedBy = [ "default.target" ];
    path = with pkgs; [ coreutils libnotify ];
    script = ''
      notified=0; suspended=0
      while true; do
        level=$(cat /sys/class/power_supply/BAT0/capacity)
        if [ "$(cat /sys/class/power_supply/BAT0/status)" = "Discharging" ]; then
          if [ $level -le 5 ] && [ $suspended -eq 0 ]; then
            suspended=1; systemctl suspend
          elif [ $level -le 10 ] && [ $notified -lt 2 ]; then
            notified=2; notify-send -t 5000 "Battery Low" "$level% remaining"
          elif [ $level -le 20 ] && [ $notified -lt 1 ]; then
            notified=1; notify-send -t 3000 "Battery Low" "$level% remaining"
          fi
        else
          notified=0; suspended=0
        fi
        sleep 60
      done
    '';
  };

  services.tlp = {
    enable = true;
    settings = {
      CPU_SCALING_GOVERNOR_ON_AC = "performance";
      CPU_SCALING_GOVERNOR_ON_BAT = "powersave";
      STOP_CHARGE_THRESH_BAT0 = 1;
    };
  };

  services.battui.enable = true;

  # Audio
  security.rtkit.enable = true;
  services.pipewire = {
    enable = true;
    alsa.enable = true;
    pulse.enable = true;
    extraConfig.pipewire."50-raop-latency"."context.modules" = [{
      name = "libpipewire-module-raop-discover";
      args."stream.rules" = [{
        # only the HiFiBerry. Blocks stray AirPlay receivers (e.g. a MacBook) from stealing playback
        matches = [{ "raop.hostname" = "~hifiberry"; }];
        actions."create-stream"."stream.props"."sess.latency.msec" = 2000;
      }];
    }];
  };
  systemd.services.airplay-rescan = {
    description = "Re-scan AirPlay/RAOP sinks after resume";
    after    = [ "suspend.target" ];
    wantedBy = [ "suspend.target" ];
    serviceConfig = {
      Type = "oneshot";
      ExecStart = "${pkgs.systemd}/bin/systemctl --user --machine=${username}@.host restart pipewire.service";
    };
  };

  # Networking
  networking.hostName = "mist";
  networking.wireless.iwd.enable = true;

  # Swap (compressed RAM, 32 GB machine, no hibernation, nothing on disk)
  zramSwap.enable = true;

  # GPG agent (passphrase caching for pass)
  programs.gnupg.agent = {
    enable = true;
    pinentryPackage = pkgs.pinentry-curses;
  };

  # mDNS (*.local hostname resolution)
  services.avahi = {
    enable = true;
    nssmdns4 = true;
    openFirewall = true;
    # PipeWire RAOP discovery keys tunnels by mDNS name only, so the HiFiBerry's
    # link-local IPv6 record shadows its IPv4 one and the AirPlay sink never appears
    ipv6 = false;
  };

  # Printers (auto-discovers via avahi)
  services.printing.enable = true;

  # Local file sharing (opens 53317 TCP/UDP for discovery and transfer)
  programs.localsend.enable = true;

  # Removable drives (udiskie automounts them to /run/media)
  services.udisks2.enable = true;
  systemd.user.services.udiskie = {
    wantedBy = [ "graphical-session.target" ];
    partOf = [ "graphical-session.target" ];
    after = [ "graphical-session.target" ];
    serviceConfig.ExecStart = "${pkgs.udiskie}/bin/udiskie --automount --notify --no-tray";
  };

  # SSH
  services.openssh = {
    enable = true;
    ports = [ 3439 ];
    openFirewall = true;
    settings = {
      PermitRootLogin = "no";
      PasswordAuthentication = true;
      MaxAuthTries = 3;
      ClientAliveInterval = 15;
      ClientAliveCountMax = 3;
    };
  };

  # Localisation
  time.timeZone = "Europe/Stockholm";
  i18n.defaultLocale = "en_US.UTF-8";
  console.keyMap = "us";

  # Fonts
  fonts.packages = with pkgs; [
    nerd-fonts.anonymice
    nerd-fonts.roboto-mono
    noto-fonts-color-emoji
  ];

  # Desktop & lockscreen
  programs.niri.enable = true;
  services.gnome.gnome-keyring.enable = false;
  programs.hyprlock.enable = true;
  # The hyprlock module enables hypridle for idle tracking, which we do not use.
  services.hypridle.enable = lib.mkForce false;

  # Locks on `loginctl lock-session` and before every suspend, whatever asked for it
  services.systemd-lock-handler.enable = true;
  systemd.user.services.hyprlock = {
    description = "Lock the screen";
    path = with pkgs; [ coreutils systemd ];
    onSuccess = [ "unlock.target" ];
    partOf    = [ "lock.target" ];
    before    = [ "lock.target" ];
    wantedBy  = [ "lock.target" ];
    serviceConfig = {
      ExecStart = "${pkgs.hyprlock}/bin/hyprlock --immediate-render --no-fade-in";
      # hyprlock cannot fork once locked, so hold the job until niri reports the lock.
      # Stays under logind's 5 s inhibitor delay, so a failed lock still lets the suspend run
      ExecStartPost = pkgs.writeShellScript "wait-until-locked" ''
        for _ in $(seq 30); do
          [ "$(loginctl show-session auto --property=LockedHint --value)" = yes ] && exit 0
          sleep 0.1
        done
      '';
      Restart = "on-failure";
      RestartSec = 0;
    };
  };

  programs.dconf = {
    enable = true;
    profiles.user.databases = [{
      settings."org/gnome/desktop/interface".color-scheme = "prefer-dark";
    }];
  };
  services.greetd = {
    enable = true;
    settings = rec {
      initial_session = {
        command = "${pkgs.niri}/bin/niri-session";
        user = username;
      };
      default_session = initial_session;
    };
  };
  xdg.mime.defaultApplications = {
    "text/html" = "firefox.desktop";
    "application/xhtml+xml" = "firefox.desktop";
    "x-scheme-handler/http" = "firefox.desktop";
    "x-scheme-handler/https" = "firefox.desktop";
    "application/pdf" = "firefox.desktop";
    "image/png" = "firefox.desktop";
    "image/jpeg" = "firefox.desktop";
    "image/gif" = "firefox.desktop";
    "image/webp" = "firefox.desktop";
    "image/svg+xml" = "firefox.desktop";
    "image/bmp" = "firefox.desktop";
    "image/avif" = "firefox.desktop";
  };

  # The LG UltraFine 5K sometimes wakes announcing both of its internal
  # DisplayPort tiles. niri can't combine tiled displays (niri#1921); it strips
  # the duplicate tile's identity and enables it as a phantom monitor that the
  # mouse can wander into. Identity-less outputs only match by connector name
  # in niri's config, and connector names shift between hotplugs, so instead
  # watch the event stream and turn off any identity-less output while the LG
  # is connected.
  systemd.user.services.ultrafine-tile-off = {
    wantedBy = [ "graphical-session.target" ];
    partOf = [ "graphical-session.target" ];
    after = [ "graphical-session.target" ];
    path = with pkgs; [ niri jq libnotify ];
    script = ''
      drop_tile() {
        niri msg --json outputs | jq -r '
          [.[]] as $outs
          | if ($outs | any(.make == "LG Electronics" and .model == "LG UltraFine"))
            then $outs[] | select(.make == "Unknown" and .model == "Unknown" and .current_mode != null) | .name
            else empty
            end' |
        while read -r conn; do
          niri msg output "$conn" off
          notify-send "Display" "Disabled 5K duplicate tile ($conn)"
        done
      }
      drop_tile
      niri msg --json event-stream | jq --unbuffered -r 'select(.WorkspacesChanged) | "event"' |
      while read -r _; do drop_tile; done
    '';
  };

  # Shell
  programs.zsh = {
    enable = true;
    enableCompletion = true;
    autosuggestions.enable = true;
    autosuggestions.strategy = [ "history" "completion" ];
    syntaxHighlighting.enable = true;
    histFile = "$HOME/.local/state/zsh/history";
    # Plugins with no NixOS module of their own. mkAfter so they load after
    # syntax-highlighting (history-substring-search must come after it).
    interactiveShellInit = lib.mkAfter ''
      source /run/current-system/sw/share/zsh/plugins/zsh-history-substring-search/zsh-history-substring-search.zsh
      source /run/current-system/sw/share/zsh/plugins/you-should-use/you-should-use.plugin.zsh
    '';
  };
  environment.sessionVariables.ZDOTDIR = "$HOME/.config/zsh";

  # User
  users.users.${username} = {
    isNormalUser = true;
    extraGroups = [ "wheel" "dialout" ];
    shell = pkgs.zsh;
    initialPassword = "changeme";
    openssh.authorizedKeys.keys = [
      "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIPrOnZOe0GfROlpDHe72esqJdsMXZ+XkWkdTQYKHDGOs raj@ipad"
    ];
  };

  # Packages
  environment.systemPackages = with pkgs; [
    # Desktop related
    brightnessctl      # internal display brightness
    cliphist
    ddcutil            # external monitor brightness
    libnotify          # notify-send
    mako               # notifications
    playerctl          # media keys
    pulseaudio         # pactl (talks to pipewire-pulse)
    swaybg             # wallpaper
    udiskie            # drive mounting (udiskie-umount)
    wl-clipboard
    xwayland-satellite # X11 app support

    # Shell plugins
    zsh-completions
    zsh-history-substring-search
    zsh-you-should-use

    # Terminal apps & tools
    bat                # cat alias
    bluetui
    btop
    claude-code
    cloc               # lines of code
    codex              # OpenAI Codex CLI
    nur.repos.charmbracelet.crush # https://github.com/charmbracelet/crush
    csvlens
    delta              # git pager
    eza                # ls/tree
    fastfetch          # system info
    fd                 # find alternative
    ffmpeg
    foot               # terminal
    fzf                # fuzzy finder
    gh                 # Github CLI
    git
    gnupg              # gpg
    helix              # hx editor
    herdr              # terminal multiplexer
    impala             # Wi-Fi TUI
    jq                 # JSON processor
    libqalculate       # qalc
    netcat-openbsd     # nc (zsh prompt)
    pandoc             # document converter
    pass               # password manager
    poppler-utils      # PDF utils
    ripgrep            # rg
    segger-jlink       # J-Link tools (unfree)
    trash-cli          # rm alias
    unzip
    vhs                # Generates terminal GIFs
    wiremix            # PipeWire TUI mixer
    yazi               # file manager
    zoxide             # cd
    zip

    # Languages & LSPs
    clang                        # C/C++ toolchain
    clang-tools                  # clangd, clang-format
    gcc-arm-embedded             # Toolchain for ARM microcontrollers
    gnumake                      # make
    go
    gopls
    lua
    lua-language-server
    markdown-oxide               # markdown LSP
    marp-cli                     # markdown slides
    nil                          # nix LSP
    nodejs                       # node + npm
    (nrfutil.withExtensions [ "nrfutil-device" "nrfutil-91" ]) # Nordic Semi CLI
    python3
    python3Packages.weasyprint   # HTML → PDF CLI
    ruff                         # python linter/formatter
    typescript                   # tsc
    typescript-language-server
    vscode-langservers-extracted # html/css/json/eslint LSPs

    # GUI apps
    chromium
    davinci-resolve # video editor (unfree)
    firefox
    freecad
    kicad           # EDA
    nanovna-qt
    (makeDesktopItem {
      name = "nanovna-qt";
      desktopName = "NanoVNA-QT";
      exec = "${nanovna-qt}/bin/vna_qt";
      terminal = false;
      categories = [ "Science" ];
    })
    prismlauncher   # Minecraft launcher
    roomeqwizard
    saleae-logic-2  # logic analyzer (unfree)
    (callPackage ./simsmith.nix { })
    vlc             # media player
  ];

  # Specific to projects & apps
  programs.direnv.enable = true; # Enable per project dev shells
  programs.nix-ld.enable = true; # Required for prebuilt linux binaries (eg. nrfutil)
  programs.nix-ld.libraries = with pkgs; [
    stdenv.cc.cc
    zlib
    libidn2
    libunistring
    gnutls
    gmp
    libkrb5
    libxcrypt-legacy
    e2fsprogs
    ncurses
  ];
  nixpkgs.overlays = [
    (final: prev:
    let
      # Imported once and shared by the agents below, so nixpkgs master is
      # only evaluated a single time.
      master = import nixpkgs-master {
        inherit (prev.stdenv.hostPlatform) system;
        config.allowUnfree = true;
      };
    in
    {
      # Prevent chrome bugging us for passwords
      chromium = prev.chromium.override {
        commandLineArgs = "--password-store=basic";
      };
      # Always pull the latest coding agents
      inherit (master) claude-code codex;
    })
  ];

  system.stateVersion = "25.11";
}
