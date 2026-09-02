# PATH additions
export PATH=$PATH:$HOME/.cargo/bin
export PATH=$PATH:$HOME/.local/bin
export PATH=$PATH:$HOME/.npm/bin
export PATH=$PATH:$HOME/.nrfutil/bin
export PATH=$PATH:$HOME/go/bin

# Pull the latest passwords in the background
(pass git pull &>/dev/null &)

# Aliases
alias nv='nvim .'

alias l='eza -l'
alias ll='eza -la'
alias lt='eza --tree'
alias cat='bat'
alias rm='trash-put'

alias gs='git status -u'
alias gd='git diff'
alias gl='git log'
alias gc='git add . && git commit -m'
alias gp='git pull'
alias gpp='git push'
alias gbd='git branch -d'
alias gbD='git branch -D'
alias gco='git checkout'
alias gcb='git checkout -b'
alias gt='eza --tree --long --git --git-ignore --no-permissions --no-user --no-time --no-filesize'

alias make='make -j16'
alias clean='make clean'
alias debug='make debug'
alias recover='make recover'
alias rtt='make rtt'
alias release='make release'

alias rebuild-my-nix='sudo nixos-rebuild switch --flake ~/.config#mist'
alias update-my-nix='nix flake update --flake ~/.config && sudo nixos-rebuild switch --flake ~/.config#mist ; sudo fwupdmgr refresh --force ; sudo fwupdmgr update'

if [[ -n ${SSH_CONNECTION:-} ]]; then
  alias sudo='sudo '

  alias reboot='echo "Blocked over SSH: reboot"'
  alias poweroff='echo "Blocked over SSH: poweroff"'
  alias halt='echo "Blocked over SSH: halt"'
  alias shutdown='echo "Blocked over SSH: shutdown"'
  alias suspend='echo "Blocked over SSH: suspend"'
  alias hibernate='echo "Blocked over SSH: hibernate"'

  alias systemctl='echo "Blocked over SSH: systemctl"'
  alias loginctl='echo "Blocked over SSH: loginctl"'
fi

# Clear scrollback + screen
function clear-scrollback-and-screen { printf '\033[3J'; zle clear-screen }
zle -N clear-scrollback-and-screen
bindkey '^o' clear-scrollback-and-screen

# History search keybindings
bindkey '^[[A' history-substring-search-up
bindkey '^[[B' history-substring-search-down
bindkey "$terminfo[kcuu1]" history-substring-search-up
bindkey "$terminfo[kcud1]" history-substring-search-down

# Options
setopt autocd
unsetopt BEEP

# History (kept in the XDG state dir, out of the home root)
mkdir -p ~/.local/state/zsh
export HISTFILE=~/.local/state/zsh/history
export HISTSIZE=1000000
export SAVEHIST=$HISTSIZE
setopt EXTENDED_HISTORY
setopt HIST_EXPIRE_DUPS_FIRST
setopt HIST_FIND_NO_DUPS
setopt HIST_IGNORE_ALL_DUPS
setopt HIST_IGNORE_DUPS
setopt HIST_IGNORE_SPACE
setopt HIST_SAVE_NO_DUPS
setopt HIST_VERIFY
setopt SHARE_HISTORY

# Prompt
source ~/.config/zsh/prompt.zsh

# zsh plugins are installed and loaded by NixOS (programs.zsh in
# configuration.nix). zoxide is a tool, so init it here.
eval "$(zoxide init zsh --cmd cd)"

# Start the terminal multiplexer if connecting over SSH
if [[ -n $SSH_CONNECTION && -z $HERDR_ENV && -z $NO_HERDR ]]; then
  herdr
fi
