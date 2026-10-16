{
  description = "Mist - A minimal NixOS based workstation";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    # Same branch as nixpkgs, locked separately so big apps update on their own
    nixpkgs-apps.url = "github:NixOS/nixpkgs/nixos-unstable";
    nixpkgs-master.url = "github:NixOS/nixpkgs/master";
    battui = {
      url = "github:siliconwitch/battui";
      inputs.nixpkgs.follows = "nixpkgs";
    };
    launtui = {
      url = "github:siliconwitch/launtui";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs = { nixpkgs, nixpkgs-apps, nixpkgs-master, battui, launtui, ... }: {
    nixosConfigurations.mist = nixpkgs.lib.nixosSystem {
      specialArgs = { username = "raj"; inherit nixpkgs-apps nixpkgs-master; };
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
