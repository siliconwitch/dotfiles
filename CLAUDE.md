# NixOS configuration

`flake.nix` builds the `mist` system from `configuration.nix`,
`hardware-configuration.nix`, `apparmor.nix` and `webapps.nix`. A broken config
means a broken boot or lost data, so work through this before every change and
every recommendation.

Everything else in the repo is live dotfiles, read from `~/.config` directly:
no rebuild, no Home Manager. `agents/AGENTS.md` is linked into
`~/.claude/CLAUDE.md` and `~/.codex/AGENTS.md`. `agents/skills/` is linked into
`~/.claude/skills` and `~/.agents/skills`. Editing either shared source is a
commit here.

## Before changing anything

- Read `flake.nix` and every module it lists.
- Verify each option and package attribute at https://search.nixos.org.
- Confirm the change with me before writing.

## After changing anything

- Run `nixos-rebuild dry-build --flake ~/.config#mist` and fix what it reports.
- Ask me to run `sudo nixos-rebuild switch --flake ~/.config#mist`.
- Offer `sudo nixos-rebuild test --flake ~/.config#mist` to trial a risky
  change without making it the boot default, and say that the previous
  generation is in the systemd-boot menu.
- Say when a change needs a reboot. Confirm the live system matches the config
  once it is applied.

## Constraints

- `hardware-configuration.nix` is machine generated and machine specific. Never
  edit it by hand or copy it between machines.
- Declarative only. No imperative one-off changes through the shell.
- Flat, readable `configuration.nix` entries beat nested modules. Split a
  section out only once its size justifies it.
- No usernames or machine-specific values in shared modules.
- The repo is public, and `.gitignore` denies by default. Tracking a new file
  needs an allow line.

## Documentation

Add a link when a new tool arrives, and remove one when a tool goes.

- NixOS manual: https://nixos.org/manual/nixos/stable/
- Options: https://search.nixos.org/options
- Packages: https://search.nixos.org/packages
- Wiki: https://wiki.nixos.org
- Niri: https://niri-wm.github.io/niri/
- Neovim: https://neovim.io/doc/user/
- lazy.nvim: https://lazy.folke.io
