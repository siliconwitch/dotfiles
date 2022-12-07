# My Personal Dev Setup

Customized dev environments that I use on MacOS and Arch Linux. Follow these instructions if you'd like to set up something similar.

## On MacOS

1. Install git, and the base development tools:

    ```bash
    xcode-select --install

    # Set up the git globals
    git config --global user.name "my name"
    git config --global user.email "mymail@gmail.com"
    ```

1. Install the [Homebrew](https://brew.sh) package manager.

    ```bash
    /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
    ```

1. Install all of our applications using `brew`:

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
    brew tap homebrew/cask-fonts
    brew install font-roboto-mono
    ```
    
1. Clone this repository:

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

1. Start with a full update:

    ```bash
    pacman -Sy archlinux-keyring && pacman -Su
    ```

1. Install all of our required applications:

    ```bash
    pacman -Sy zsh git base-devel tmux github-cli fzf tigervnc micro chromium xorg-server xorg-xinit qtile alacritty ttf-roboto-mono noto-fonts-emoji picom rofi rofi-calc
    ```

1. Set up a new user. **Be sure to change** `username` to your user:

    ```bash
    useradd -m -G wheel -s /bin/zsh username
    passwd username

    # Edit the sudoers file
    nano /etc/sudoers
    #
    #   Uncomment the line
    #   %wheel   ALL=(ALL)   ALL
    #

    # Switch to the newly created user account
    su username
    ```

1. Change the hostname to `lab` or `remote-lab`:

    ```bash
    hostnamectl set-hostname remote-lab
    ```

1. Clone this repository:

    ```bash
    git clone --recursive https://github.com/siliconwitch/dev-config.git ~/tools/dev-config
    cd ~/tools/dev-config
    ```

1. Create symlinks to all of our configuration files:

    ```bash
    mkdir -p ~/.config/qtile ~/.config/picom ~/.config/alacritty ~/.vnc 
    ln -sf `pwd`/zsh/zshrc ~/.zshrc
    ln -sf `pwd`/vnc/config ~/.vnc/config
    ln -sf `pwd`/tmux/tmux.conf ~/.tmux.conf
    ln -sf `pwd`/qtile/config.py ~/.config/qtile/config.py
    ln -sf `pwd`/picom/picom.conf ~/.config/picom/picom.conf
    ln -sf `pwd`/alacritty/alacritty.yml ~/.config/alacritty/alacritty.yml
    ```

1. Delete old bash related files:

    ```bash
    rm ~/.bash*
    ```

1. Add your username to the `vncserver.users` file. **Be sure to change** `username` to your user:

	```bash
	sudo sh -c 'echo ":1=username" >> /etc/tigervnc/vncserver.users'
	```

1. Set the VNC password:

    ```bash
    vncpasswd
    ```

1. Start and enable the VNC server:

    ```bash
    sudo systemctl start vncserver@:1
    sudo systemctl enable vncserver@:1
    ```

1. Exit and login again to see the completed environment. **Be sure to change** `username` to your user:

    ```bash
    exit

    # Connect with SSH and pass the the VNC port to localhost
    ssh -L 5901:127.0.0.1:5901 username@<IP_ADDR>
    ```

### Optional stuff

1. Set up for git pushes:

    ```bash
    # Set up the git globals
    git config --global user.name "my name"
    git config --global user.email "mymail@gmail.com"

    # Log into a github account (use the https option)
    gh auth login
    ```

1. Optionally, install the `yay` helper for `pacman`:

    ``` bash
    # Clone the repository
    git clone https://aur.archlinux.org/yay.git ~/tools/yay
    
    # go into the cloned directory
    cd ~/tools/yay

    # Make and install yay
    makepkg -si
    ```

1. Optionally, install these hardware related tools:

    - [Efinix Efinity Software](https://www.efinixinc.com/support/efinity.php)

        ```bash
        # You may need to install the cryptx library
        yay -S libxcrypt-compat
        ```

    - [Signalhound Spike](https://signalhound.com/spike/)
    - [VNA/J Software](https://vnaj.dl2sba.com)

        ```bash
        # Java is required
        yay -S jre-openjdk

        # Add the user to the tty group 
        sudo usermod -aG tty lab

        # Run VNA/J jar file from the terminal using
        java -jar vnaJ.3.4.8.jar
        ```

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
