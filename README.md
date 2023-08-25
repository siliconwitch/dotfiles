# My Personal Dev Setup

Customized dev environments that I use on MacOS and Arch Linux. Follow these instructions if you'd like to set up something similar.

## On MacOS

1. Install git, and the base development tools:

    ```sh
    xcode-select --install

    # Set up the git globals
    git config --global user.name "my name"
    git config --global user.email "mymail@gmail.com"
    ```

1. Install the [Homebrew](https://brew.sh) package manager:

    ```sh
    /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
    ```

1. Install all of our applications using `brew`:

    ```sh
    brew install affinity-photo arc blender discord drawio dropbox figma fxfactory fzf gcc-arm-embedded gh go gtkwave homebrew/cask-fonts/font-roboto-mono icarus-verilog kicad koekeishiya/formulae/skhd koekeishiya/formulae/yabai nordic-nrf-command-line-tools obs openfpgaloader python raspberry-pi-imager raycast rekordbox remarkable saleae-logic segger-jlink spotify steam the-unarchiver visual-studio-code vlc vnc-viewer xmind zoom

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

1. Clone this repository:

    ```sh
    git clone --recursive https://github.com/siliconwitch/dev-config.git ~/tools/dev-config
    cd ~/tools/dev-config
    ```

1. Create symlinks to all of our configuration files:

    ```sh
    mkdir -p ~/.config/yabai ~/.config/skhd
    ln -sf `pwd`/zsh/zshrc ~/.zshrc
    ln -sf `pwd`/yabai/yabairc ~/.config/yabai
    ln -sf `pwd`/skhd/skhdrc ~/.config/skhd
    ln -sf `pwd`/vscode/settings.json ~/Library/Application\ Support/Code/User
    ln -sf `pwd`/vscode/keybindings.json ~/Library/Application\ Support/Code/User
    ```

1. Enable Yabai and SKHD with the commands:

    ```sh
    yabai --start-service
    skhd --start-service
    ```

1. Install dependencies for Verilog-HDL and GTKWave:

    ```sh
    cargo install svls
    sudo perl -MCPAN -e 'install Switch'
    ```

## On Arch Linux

1. Start with a full update:

    ```sh
    pacman -Sy archlinux-keyring && pacman -Su
    ```

1. Install all of our required applications:

    ```sh
    pacman -Sy zsh git base-devel tmux github-cli fzf tigervnc micro chromium xorg-server xorg-xinit qtile alacritty ttf-roboto-mono noto-fonts-emoji picom rofi rofi-calc
    ```

1. Set up a new user. **Be sure to change** `username` to your user:

    ```sh
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

    ```sh
    hostnamectl set-hostname remote-lab
    ```

1. Clone this repository:

    ```sh
    git clone --recursive https://github.com/siliconwitch/dev-config.git ~/tools/dev-config
    cd ~/tools/dev-config
    ```

1. Create symlinks to all of our configuration files:

    ```sh
    mkdir -p ~/.config/qtile ~/.config/picom ~/.config/alacritty ~/.vnc 
    ln -sf `pwd`/zsh/zshrc ~/.zshrc
    ln -sf `pwd`/vnc/config ~/.vnc/config
    ln -sf `pwd`/tmux/tmux.conf ~/.tmux.conf
    ln -sf `pwd`/qtile/config.py ~/.config/qtile/config.py
    ln -sf `pwd`/picom/picom.conf ~/.config/picom/picom.conf
    ln -sf `pwd`/alacritty/alacritty.yml ~/.config/alacritty/alacritty.yml
    ```

1. Delete old bash related files:

    ```sh
    rm ~/.bash*
    ```

1. Add your username to the `vncserver.users` file. **Be sure to change** `username` to your user:

	```sh
	sudo sh -c 'echo ":1=username" >> /etc/tigervnc/vncserver.users'
	```

1. Set the VNC password:

    ```sh
    vncpasswd
    ```

1. Start and enable the VNC server:

    ```sh
    sudo systemctl start vncserver@:1
    sudo systemctl enable vncserver@:1
    ```

1. Exit and login again to see the completed environment. **Be sure to change** `username` to your user:

    ```sh
    exit

    # Connect with SSH and pass the the VNC port to localhost
    ssh -L 5901:127.0.0.1:5901 username@<IP_ADDR>
    ```

### Optional stuff

1. Set up for git pushes:

    ```sh
    # Set up the git globals
    git config --global user.name "my name"
    git config --global user.email "mymail@gmail.com"

    # Log into a github account (use the https option)
    gh auth login
    ```

1. Optionally, install the `yay` helper for `pacman`:

    ```sh
    # Clone the repository
    git clone https://aur.archlinux.org/yay.git ~/tools/yay
    
    # go into the cloned directory
    cd ~/tools/yay

    # Make and install yay
    makepkg -si
    ```

1. Optionally, install these hardware related tools:

    - [Efinix Efinity Software](https://www.efinixinc.com/support/efinity.php)

        ```sh
        # You may need to install the cryptx library
        yay -S libxcrypt-compat
        ```

    - [Signalhound Spike](https://signalhound.com/spike/)
    - [VNA/J Software](https://vnaj.dl2sba.com)

        ```sh
        # Java is required
        yay -S jre-openjdk

        # Add the user to the tty group 
        sudo usermod -aG tty lab

        # Run VNA/J jar file from the terminal using
        java -jar vnaJ.3.4.8.jar
        ```

    - [Digilent Waveforms](https://digilent.com/shop/software/digilent-waveforms/download) can be installed through [AUR](https://aur.archlinux.org/packages/digilent.waveforms):

        ```sh
        yay -S aur/digilent.waveforms

        # Okay to remove dependencies after install (press Y)

        # No need to show diff if prompted (press N)
        ```
    
    - [J-Link Tools](https://www.segger.com/downloads/jlink/) can be installed through [AUR](https://aur.archlinux.org/packages/jlink):

        ```sh
        yay -S aur/jlink-software-and-documentation

        # Okay to remove dependencies after install (press Y)

        # No need to show diff if prompted (press N)
        ```
