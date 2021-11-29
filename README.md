# My Personal Dev Setup

Configuration for **VS Code** and **oh-my-zsh** running on MacOS

## Installing

1. Ensure these tools are installed

- [git](https://github.com/git-guides/install-git)
- [oh-my-zsh](https://ohmyz.sh/#install)
- [Dracula Theme](https://draculatheme.com/zsh)
- [zsh-autosuggestions](https://github.com/zsh-users/zsh-autosuggestions)
- [brew](https://brew.sh)
- [fzf](https://github.com/junegunn/fzf)
- [FPGA toolchains](https://github.com/siliconwitchery/homebrew-oss-fpga)

2. Ensure these plugins are installed for VS Code

- **ARM** by dan-c-underwood
- **Beautify** by HookyQR
- **C/C++** by Microsoft
- **Clock in status bar** by Compulim
- **Code Spell Checker** by Street Side Software
- **Cortex Debug** by Marus25
- **Git History** by Don Jayamanne
- **Hex Editor** by Microsoft
- **JSON Tools** by Erik Lynd
- **Live Server** by Ritwick Dey
- **Tabnine** by Tabnine
- **Verilog-HDL** by mshr-h

3. Remove old `.zshrc`, `settings.json` and `keybindings.json`

```bash
rm ~/.zshrc
rm ~/Library/Application\ Support/Code/User/settings.json
rm ~/Library/Application\ Support/Code/User/keybindings.json
```

4. Create symlinks

```bash
ln -s /<full-path>/dev-config/zshrc ~/.zshrc
ln -s /<full-path>/dev-config/settings.json ~/Library/Application\ Support/Code/User/settings.json
ln -s /<full-path>/dev-config/keybindings.json ~/Library/Application\ Support/Code/User/keybindings.json
```

5. Install a couple dependencies for Verilog-HDL and configure paths in settings if needed

``` bash
brew install --HEAD universal-ctags/universal-ctags/universal-ctags
cargo install svls
```