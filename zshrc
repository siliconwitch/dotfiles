# Path variables
export PATH=/opt/homebrew/bin:$PATH
export PATH=/Applications/gtkwave.app/Contents/Resources/bin:$PATH
export PATH=/Users/raj/.cargo/bin:$PATH
export PATH=/Users/raj/.local/bin:$PATH

# How many commands zsh will load to memory.
export HISTSIZE=10000

# How maney commands history will save on file.
export SAVEHIST=10000

# History won't save duplicates.
setopt HIST_IGNORE_ALL_DUPS

# History won't show duplicates on search.
setopt HIST_FIND_NO_DUPS

source ~/tools/dev-config/zsh-syntax-highlighting/zsh-syntax-highlighting.zsh
source ~/tools/dev-config/zsh-autosuggestions/zsh-autosuggestions.zsh
source ~/tools/dev-config/zsh-interactive-cd/zsh-interactive-cd.plugin.zsh
source ~/tools/dev-config/zsh-history-substring-search/zsh-history-substring-search.zsh
source ~/tools/dev-config/zsh-dracula-theme/dracula.zsh-theme

# Remove the dracula arrow icon
DRACULA_ARROW_ICON=""

# History search keybindings
bindkey '^[[A' history-substring-search-up
bindkey '^[[B' history-substring-search-down