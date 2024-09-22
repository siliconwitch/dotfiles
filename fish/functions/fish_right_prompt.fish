function fish_right_prompt

    # Print last status if it was not zero
    set -l cmd_status $status
    if test $cmd_status -ne 0
        echo -n (set_color red)"✘ $cmd_status"
    end

    # Print the current branch status
    set -l branch (command git symbolic-ref --short HEAD 2>/dev/null)

    if not test -z (command git status -z --porcelain 2>/dev/null)
        echo -n (set_color red)' '$branch
    else
        echo -n (set_color green)' '$branch
    end
     
    # Print the commit difference between local and remote
    command git rev-list --count --left-right 'HEAD...@{upstream}' 2>/dev/null \
        | read -d \t -l status_ahead status_behind
    if test $status -ne 0
        set status_ahead 0
        set status_behind 0
    end

    if test $status_ahead -ne 0
        echo -n ' '(set_color purple)''
    end
    if test $status_behind -ne 0
        echo -n ' '(set_color purple)''
    end

    set_color normal

end
