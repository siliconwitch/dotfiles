# My Personal Dev Setup

Configuration for **oh-my-zsh** and **yabai** running on MacOS

## Installing

1. Ensure these tools are installed

    - [git](https://github.com/git-guides/install-git)
    - [oh-my-zsh](https://ohmyz.sh/#install)
    - [iTerm2](https://iterm2.com/downloads.html)
    - [brew](https://brew.sh)
    - [VSCode](https://code.visualstudio.com/download)

1. Clone this repository into a `tools` folder

    ```bash
    git clone --recursive https://github.com/siliconwitch/dev-config.git ~/tools/dev-config
    ```

1. Remove old `.zshrc` and create symlinks to our settings

    ```bash
    rm ~/.zshrc
    ln -s ~/tools/dev-config/zshrc ~/.zshrc
    ln -s ~/tools/dev-config/zsh-autosuggestions ~/.oh-my-zsh/custom/plugins
    ln -s ~/tools/dev-config/zsh-dracula-theme/dracula.zsh-theme ~/.oh-my-zsh/custom/themes/dracula.zsh-theme
    mkdir ~/.config/yabai ~/.config/skhd
    ln -s ~/tools/dev-config/yabairc ~/.config/yabai
    ln -s ~/tools/dev-config/skhdrc ~/.config/skhd
    ```

1. Remove default VSCode settings, and create symlinks to our settings

    ```bash
    rm ~/Library/Application\ Support/Code/User/settings.json
    rm ~/Library/Application\ Support/Code/User/keybindings.json
    ln -s ~/tools/dev-config/settings.json ~/Library/Application\ Support/Code/User
    ln -s ~/tools/dev-config/keybindings.json ~/Library/Application\ Support/Code/User
    ```

1. Install dev tools using brew

    ```bash
    brew install fzf
    brew install koekeishiya/formulae/yabai
    brew install koekeishiya/formulae/skhd
    brew install go
    brew install gtkwave
    brew install icarus-verilog
    brew install openfpgaloader
    brew install homebrew/cask-drivers/segger-jlink
    brew install homebrew/cask-drivers/nordic-nrf-command-line-tools
    brew install kicad
    brew install ~/tools/dev-config/gcc-arm-embedded.rb # Installs an older version of GCC ARM suitable for Apple silicon
    brew install --HEAD siliconwitchery/oss-fpga/nextpnr-nexus
    brew install --HEAD siliconwitchery/oss-fpga/nextpnr-ice40
    brew install --HEAD universal-ctags/universal-ctags/universal-ctags
    brew install rust
    ```
    
1. Enable Yabai and SKHD with the command

    ```brew
    brew services start yabai
    brew services start skhd
    ```

1. Activate the Dracula theme in iTerm2

    1. iTerm2 > Preferences > Profiles > Colors Tab
    1. Open the Color Presets... drop-down in the bottom right corner
    1. Select Import... from the list
    1. Select the Dracula.itermcolors file
    1. Select the Dracula from Color Presets

1. Install dependencies for Verilog-HDL and configure paths in settings

    ``` bash
    cargo install svls
    ```

1. Cleanup

    ```bash
    brew autoremove
    ```