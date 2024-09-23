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

        # Export the DESKTOP variable in order to run graphical apps on a VM
        set -g DISPLAY :1

        # Settings for Radiant
        set -g LIBGL_ALWAYS_SOFTWARE 1
        set -g QTWEBENGINE_DISABLE_SANDBOX 1

end

# Abbreviations
abbr --add clr 'clear && tmux clear-history'
abbr --add clear 'clear && tmux clear-history'

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

# Start tmux when opening a new terminal TODO don't open on remote machine
if [ -z "$TMUX" ] && [ "$TERM_PROGRAM" != "vscode" ]
    tmux attach -t default || tmux new -s default
end