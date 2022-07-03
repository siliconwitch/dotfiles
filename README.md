# My Personal Dev Setup

Configuration for **oh-my-zsh** and **yabai** running on MacOS

## Installing

1. Ensure these tools are installed

    - [git](https://github.com/git-guides/install-git)
    - [oh-my-zsh](https://ohmyz.sh/#install)
    - [brew](https://brew.sh)

1. Clone this repository into a `tools` folder

    ```bash
    git clone --recursive https://github.com/siliconwitch/dev-config.git ~/tools/
    ```

1. Install some standard packages via brew

    ```bash
    brew install fzf iterm2 koekeishiya/formulae/yabai koekeishiya/formulae/skhd go gtkwave icarus-verilog openfpgaloader segger-jlink nordic-nrf-command-line-tools gcc-arm-embedded
    ```
1. Activate the Dracula theme in iTerm2

    1. iTerm2 > Preferences > Profiles > Colors Tab
    1. Open the Color Presets... drop-down in the bottom right corner
    1. Select Import... from the list
    1. Select the Dracula.itermcolors file
    1. Select the Dracula from Color Presets

1. Install OSS FPGA tools via brew
    
    ```bash
    brew tap siliconwitchery/oss-fpga
    brew install --HEAD siliconwitchery/oss-fpga/nextpnr-nexus
    brew install --HEAD siliconwitchery/oss-fpga/nextpnr-ice40
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

1. Install dependencies for Verilog-HDL and configure paths in settings

    ``` bash
    brew install --HEAD universal-ctags/universal-ctags/universal-ctags
    cargo install svls
    ```