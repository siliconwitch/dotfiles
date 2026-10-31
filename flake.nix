{
  description = "Mist - A minimal NixOS based workstation";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    # Same branch as nixpkgs, locked separately so big apps update on their own
    nixpkgs-apps.url = "github:NixOS/nixpkgs/nixos-unstable";
    nixpkgs-master.url = "github:NixOS/nixpkgs/master";
    # Temporary fix: patchelf 0.19.2 makes J-Link programs crash at startup, so stay on a revision
    # with 0.15.2. Drop this once nixos-unstable has patchelf 0.15.2 or 0.19.3 (nixpkgs#577387)
    nixpkgs-jlink.url = "github:NixOS/nixpkgs/6510408d147d8c5e6d76043e899e9251756f6a75";
    battui = {
      url = "github:siliconwitch/battui";
      inputs.nixpkgs.follows = "nixpkgs";
    };
    launtui = {
      url = "github:siliconwitch/launtui";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs = { nixpkgs, nixpkgs-apps, nixpkgs-master, nixpkgs-jlink, battui, launtui, ... }: {
    nixosConfigurations.mist = nixpkgs.lib.nixosSystem {
      specialArgs = { username = "raj"; inherit nixpkgs-apps nixpkgs-master nixpkgs-jlink; };
      modules = [
        ./configuration.nix
        ./hardware-configuration.nix
        ./apparmor.nix
        battui.nixosModules.default
        launtui.nixosModules.default
      ];
    };
  };
}
