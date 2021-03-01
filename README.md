# My personal dev setup.

Configuration for **VS Code** and **oh-my-zsh**

## Installing

1. Remove old `.zshrc` and `settings.json`

```bash
rm ~/.zshrc
rm ~/Library/Application\ Support/Code/User/settings.json
```

2. Create symlinks

```bash
ln -s /<full-path>/tools/dev-config/zshrc ~/.zshrc
ln -s /<full-path>/vscode-config/settings.json ~/Library/Application\ Support/Code/User/settings.json
```

3. Ensure these tools are installed

- [oh-my-zsh](https://ohmyz.sh/#install)
- [Dracula](https://draculatheme.com/zsh)
- [brew](https://brew.sh)
- [zsh-autosuggestions](https://github.com/zsh-users/zsh-autosuggestions)
- git

4. Ensure these plugins are installed for VS Code

- **ARM** by dan-c-underwood
- **Better C++ Syntax** by Jeff Hykin
- **C/C++** by Microsoft
- **Code Spell Checker** by Street Side Software
- **Cortex Debug** by Marus25
- **Vim** by vscodevim