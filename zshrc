# Path variables
export PATH=/opt/homebrew/bin:$PATH
export PATH=/Applications/gtkwave.app/Contents/Resources/bin:$PATH
export PATH=/Users/raj/.cargo/bin:$PATH
export PATH=/Users/raj/.local/bin:$PATH

# Path to your oh-my-zsh installation
export ZSH="/Users/raj/.oh-my-zsh"

# Theme. More info here https://github.com/ohmyzsh/ohmyzsh/wiki/Themes
ZSH_THEME="dracula"

# Automatically update without prompting
DISABLE_UPDATE_PROMPT="true"

# Load plugins
plugins=(brew fzf git history zsh-interactive-cd zsh-autosuggestions)

source $ZSH/oh-my-zsh.sh