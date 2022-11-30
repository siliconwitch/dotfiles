# If running on lab machine
if [ "`uname -n`" = lab ] || [ "`uname -n`" = remote-lab ]; then
 
	# Start tmux on login
	if [ -z "$TMUX" ]; then
		tmux attach -t default || tmux new -s default
	fi

	# Set the history file
	export HISTFILE=~/.histfile
	
	# Add alias for easy updating
	alias update='sudo pacman -Syu'
fi

# If running on MacOS
if [ "`uname -n`" = Rajs-MacBook-Air.local ]; then

	# Set path variables
	export PATH=/opt/homebrew/bin:$PATH
	export PATH=/Applications/gtkwave.app/Contents/Resources/bin:$PATH
	export PATH=/Users/raj/.cargo/bin:$PATH
	export PATH=/Users/raj/.local/bin:$PATH

	# Add alias for easy updating                                                                         
	alias update='brew upgrade; brew cleanup; brew autoremove'  
fi

# Remaining settings are common to all machines

# How many commands zsh will load to memory.
export HISTSIZE=10000

# How maney commands history will save on file.
export SAVEHIST=10000

# History won't save duplicates.
setopt HIST_IGNORE_ALL_DUPS

# History won't show duplicates on search.
setopt HIST_FIND_NO_DUPS

# Add lines to history as they are entered
setopt INC_APPEND_HISTORY

source ~/tools/dev-config/zsh-autosuggestions/zsh-autosuggestions.zsh
source ~/tools/dev-config/zsh-interactive-cd/zsh-interactive-cd.plugin.zsh
source ~/tools/dev-config/zsh-history-substring-search/zsh-history-substring-search.zsh
source ~/tools/dev-config/zsh-dracula-theme/dracula.zsh-theme
source ~/tools/dev-config/zsh-syntax-highlighting/zsh-syntax-highlighting.zsh

# Remove the dracula arrow icon
DRACULA_ARROW_ICON=""

# History search keybindings
bindkey '^[[A' history-substring-search-up
bindkey '^[[B' history-substring-search-down

# Set alias for ls
alias ls='ls --color=auto'
alias la='ls -la --color=auto'
