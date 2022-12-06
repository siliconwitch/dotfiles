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
    mkdir -p ~/.config/yabai ~/.config/skhd ~/.config/alacritty
    ln -sf `pwd`/zsh/zshrc ~/.zshrc
    ln -sf `pwd`/yabai/yabairc ~/.config/yabai
    ln -sf `pwd`/skhd/skhdrc ~/.config/skhd
    ln -sf `pwd`/alacritty/alacritty.yml ~/.config/alacritty/alacritty.yml
    ln -sf `pwd`/vscode/settings.json ~/Library/Application\ Support/Code/User
    ln -sf `pwd`/vscode/keybindings.json ~/Library/Application\ Support/Code/User
    ```

1. Download and install the [Roboto Mono](https://fonts.google.com/specimen/Roboto+Mono?query=Roboto+mono) font.

1. Install [Homebrew](https://brew.sh).

1. Then install these tools using the brew command:

    ```bash
    brew install --cask visual-studio-code
    brew install --cask alacritty
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

1. Update `makepkg.conf` to match the number of cores:

    ```bash
    # Check how many threads are available on the CPU
    nproc

    # Edit the line in the makepkg.conf file
    nano /etc/makepkg.conf
    #
    #   # Uncomment the line, and set the number to
    #   # your number of threads. E.g 16
    #   MAKEFLAGS="-j16"
    #
    ```
1. Install the `yay` helper for `pacman`. Be sure to install git first:

    ```bash
    sudo pacman -Sy git
    git clone https://aur.archlinux.org/yay.git
    cd yay
    makepkg -si
    ```

1. Update the mirrors and system:

    ```bash
    yay
    ```

1. Install these tools using `yay`

    ```bash
     yay -S tmux base-devel github-cli fzf tigervnc micro chromium xorg-server xorg-xinit qtile alacritty ttf-roboto-mono picom rofi rofi-calc
    ```

1. Clone this repository into a folder named `tools`:

    ```bash
    git clone --recursive https://github.com/siliconwitch/dev-config.git ~/tools/dev-config
    cd ~/tools/dev-config
    ```

1. Create symlinks to all of our configuration files:

    ```bash
    mkdir -p ~/.config/qtile ~/.config/alacritty ~/.vnc 
    ln -sf `pwd`/zsh/zshrc ~/.zshrc
    ln -sf `pwd`/vnc/config ~/.vnc/config
    ln -sf `pwd`/tmux/tmux.conf ~/.tmux.conf
    ln -sf `pwd`/qtile/config.py ~/.config/qtile/config.py
    ln -sf `pwd`/alacritty/alacritty.yml ~/.config/alacritty/alacritty.yml
    ```

1. Edit the file `/etc/tigervnc/vncserver.users` and add the line:

	```bash
	:1=lab
	```

1. Start and enable the VNC server

    ```bash
    sudo systemctl start vncserver@:1
    sudo systemctl enable vncserver@:1
    ```

1. Optionally, install these hardware related tools:

    - [Efinix Efinity Software](https://www.efinixinc.com/support/efinity.php)
    - [Signalhound Spike](https://signalhound.com/spike/)
    - [Digilent Waveforms](https://digilent.com/shop/software/digilent-waveforms/download) can be installed through [AUR](https://aur.archlinux.org/packages/digilent.waveforms):

        ```bash
        yay -S aur/digilent.waveforms

        # Okay to remove dependencies after install (press Y)

        # No need to show diff if prompted (press N)
        ```
    
    - [J-Link Tools](https://www.segger.com/downloads/jlink/) can be installed through [AUR](https://aur.archlinux.org/packages/jlink):

        ```bash
        yay -S aur/jlink-software-and-documentation

        # Okay to remove dependencies after install (press Y)

        # No need to show diff if prompted (press N)
        ```
