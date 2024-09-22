# Configure paths
set -e fish_user_paths

switch (uname)
    case "Darwin"

        set -U fish_user_paths \
            /opt/homebrew/bin \
            /opt/homebrew/opt/openjdk@17/bin \
            /Applications/gtkwave.app/Contents/Resources/bin \
            $HOME/.cargo/bin \
            $HOME/.local/bin \
            $HOME/.nrfutil/bin \
            $HOME/projects/flutter \

    case "Linux"
    
        set -U fish_user_paths \
            /home/linuxbrew/.linuxbrew/bin \      

end

# Abbreviations
abbr --add clr 'clear && tmux clear-history'

abbr --add update 'brew upgrade; brew cleanup; brew autoremove'

abbr --add gs git status
abbr --add gl git log
abbr --add gfp git fetch -p
abbr --add gbd git branch -d
abbr --add gbD git branch -D

abbr --add df hx ~/.config

# Enable FZF and set keybindings
fzf --fish | source

bind \ct fzf-cd-widget
bind -M insert \ct fzf-cd-widget

# Remove the greeting message
set fish_greeting