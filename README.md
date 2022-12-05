# My Personal Dev Setup

Customized dev environments that I use on MacOS and Arch Linux. Follow these instructions if you'd like to set up something similar.

## On MacOS

1. Ensure that the MacOS command line tools are installed:

    ```bash
    xcode-select --install
    ```

1. Set the basic git globals:

    ```bash
    git config --global user.name "my name"
    git config --global user.email "mymail@gmail.com"
    ```
1. Clone this repository into a folder named `tools`:

    ```bash
    git clone --recursive https://github.com/siliconwitch/dev-config.git ~/tools/dev-config
    cd ~/tools/dev-config
    ```

1. Create symlinks to all of our configuration files:

    ```bash
    mkdir -p ~/.config/yabai ~/.config/skhd
    ln -sf `pwd`/zsh/zshrc ~/.zshrc
    ln -sf `pwd`/yabai/yabairc ~/.config/yabai
    ln -sf `pwd`/skhd/skhdrc ~/.config/skhd
    ln -sf `pwd`/vscode/settings.json ~/Library/Application\ Support/Code/User
    ln -sf `pwd`/vscode/keybindings.json ~/Library/Application\ Support/Code/User
    ```

1. Install [Homebrew](https://brew.sh).

1. Then install these tools using the brew command:

    ```bash
    brew install --cask visual-studio-code
    brew install --cask iterm2
    brew install --cask gcc-arm-embedded
    brew install kicad
    brew install gh
    brew install fzf
    brew install karabiner-elements
    brew install koekeishiya/formulae/yabai
    brew install koekeishiya/formulae/skhd
    brew install wget
    brew install go
    brew install rust
    brew install gtkwave
    brew install icarus-verilog
    brew install openfpgaloader
    brew install homebrew/cask-drivers/segger-jlink
    brew install homebrew/cask-drivers/nordic-nrf-command-line-tools
    brew install --HEAD siliconwitchery/oss-fpga/nextpnr-nexus
    brew install --HEAD siliconwitchery/oss-fpga/nextpnr-ice40
    brew install --HEAD universal-ctags/universal-ctags/universal-ctags
    ```
    
1. You will need to [enable the Hyper key](https://holmberg.io/hyper-key/) using karabiner-elements.

1. Enable Yabai and SKHD with the commands:

    ```brew
    brew services start yabai
    brew services start skhd
    ```

1. Activate the Dracula theme in iTerm2:

    1. iTerm2 > Preferences > Profiles
    1. Click on Other Actions > Import JSON Profiles
    1. Select ~/tools/dev-config/iterm-config.json

1. Install dependencies for Verilog-HDL and GTKWave

    ``` bash
    cargo install svls
    sudo perl -MCPAN -e 'install Switch'
    ```

1. Cleanup

    ```bash
    brew cleanup
    brew autoremove
    ```

## On Arch Linux

1. Update the mirrors and system:

    ```bash
    sudo pacman -Syu
    ```

1. Install these tools using pacman

    ```bash
    sudo pacman -S tmux git base-devel github-cli fzf tigervnc micro chromium xorg-server xorg-xinit qtile alacritty picom
    ```

1. Clone this repository into a folder named `tools`:

    ```bash
    git clone --recursive https://github.com/siliconwitch/dev-config.git ~/tools/dev-config
    cd ~/tools/dev-config
    ```

1. Create symlinks to all of our configuration files:

    ```bash
    mkdir -p ~/.config/qtile ~/.vnc
    ln -sf `pwd`/zsh/zshrc ~/.zshrc
    ln -sf `pwd`/vnc/config ~/.vnc/config
    ln -sf `pwd`/tmux/tmux.conf ~/.tmux.conf
    ln -sf `pwd`/qtile/config.py ~/.config/qtile/config.py
    ```

1. Start and enable the VNC server

    ```bash
    systemctl start vncserver@:1
    systemctl enable vncserver@:1
    ```

1. Optionally, install these hardware related tools:

    - [Efinix Efinity Software](https://www.efinixinc.com/support/efinity.php)
    - [Signalhound Spike](https://signalhound.com/spike/)
    - [Digilent Waveforms](https://digilent.com/shop/software/digilent-waveforms/download) can be installed through [AUR](https://aur.archlinux.org/packages/digilent.waveforms):

        ```bash
        ```
    
    - [J-Link Tools](https://www.segger.com/downloads/jlink/) can be installed through [AUR](https://aur.archlinux.org/packages/jlink):

        ```bash
        ```