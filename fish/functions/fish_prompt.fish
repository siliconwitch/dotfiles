function fish_prompt

    # If running on remote print "user@hostname:"
    if test -n "$SSH_TTY"
        echo -n (set_color purple)$USER@(prompt_hostname)(set_color normal):
    end

    # Then print current working directory
    echo -n (set_color -o purple)(basename (prompt_pwd))(set_color normal)' '

end