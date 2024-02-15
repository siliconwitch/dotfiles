# My Personal Dev Setup

Customized dev environment that I use on my various machines. Follow these instructions if you'd like to set up something similar.

## On MacOS

1. Install git and the base development tools:

    ```sh
    xcode-select --install

    # Set up the git globals
    git config --global user.name "my name"
    git config --global user.email "mymail@gmail.com"
    git config --global core.excludesfile ~/.config/git/ignore
    ```

1. Clone this repository:

    ```sh
    git clone --recursive https://github.com/siliconwitch/dotfiles.git ~/.dotfiles
    cd ~/.dotfiles
    ```

1. Create symlinks to all of our configuration files:

    ```sh
    mkdir -p ~/.config/yabai ~/.config/skhd ~/.config/tmux ~/.config/git
    ln -sf `pwd`/zsh/zshrc ~/.zshrc
    ln -sf `pwd`/yabai/yabairc ~/.config/yabai
    ln -sf `pwd`/skhd/skhdrc ~/.config/skhd
    ln -sf `pwd`/tmux/tmux.conf ~/.config/tmux
    ln -sf `pwd`/git/ignore ~/.config/git
    ln -sf `pwd`/vscode/settings.json ~/Library/Application\ Support/Code/User
    ln -sf `pwd`/vscode/keybindings.json ~/Library/Application\ Support/Code/User
    ```

1. Install the [Homebrew](https://brew.sh) package manager:

    ```sh
    /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
    ```

1. Install all of our applications using `brew`:

    ```sh
    brew install affinity-photo arc blender discord drawio dropbox figma fxfactory fzf gcc-arm-embedded gh go gtkwave homebrew/cask-fonts/font-roboto-mono icarus-verilog jq kicad koekeishiya/formulae/skhd koekeishiya/formulae/yabai nordic-nrf-command-line-tools obs openfpgaloader python raspberry-pi-imager raycast rekordbox remarkable saleae-logic segger-jlink steam the-unarchiver visual-studio-code vlc vnc-viewer xmind zoom micro

    brew install --HEAD siliconwitchery/oss-fpga/nextpnr-ice40 siliconwitchery/oss-fpga/nextpnr-nexus
    ```

1. Other manually installed apps:

    - [1Blocker](https://apps.apple.com/se/app/1blocker-ad-blocker/id1365531024?l=en-GB)
    - [Cuprum](https://apps.apple.com/se/app/cuprum/id1088670425?l=en-GB&mt=12)
    - [DaVinci Resolve](https://apps.apple.com/se/app/davinci-resolve/id571213070?l=en-GB&mt=12)
    - [Final Cut Pro](https://apps.apple.com/se/app/final-cut-pro/id424389933?l=en-GB&mt=12)
    - [GIPHY Capture](https://apps.apple.com/se/app/giphy-capture-the-gif-maker/id668208984?l=en-GB&mt=12)
    - [Grammarly for Safari](https://apps.apple.com/se/app/grammarly-writing-app/id1462114288?l=en-GB&mt=12)
    - [Matlab](https://www.mathworks.com)
    - [NanoVNA-Saver](https://github.com/NanoVNA-Saver/nanovna-saver/releases)
    - [SimNEC](http://www.ae6ty.com/smith_charts.html)
    - [Spark](https://apps.apple.com/se/app/spark-email-app-by-readdle/id1176895641?l=en-GB&mt=12)
    - [Things](https://apps.apple.com/se/app/things-3/id904280696?l=en-GB&mt=12)
    - [VNA QT](https://nanorfe.com/nanovna-software.html)
    - [Waveforms](https://digilent.com/shop/software/digilent-waveforms/)
    - [Xcode](https://apps.apple.com/se/app/xcode/id497799835?l=en-GB&mt=12)

1. Install dependencies for Verilog-HDL and GTKWave:

    ```sh
    cargo install svls
    sudo perl -MCPAN -e 'install Switch'
    ```
