autoload -Uz add-zsh-hook
autoload -Uz vcs_info

_git_fetch_async() {
  {
    git rev-parse --is-inside-work-tree &>/dev/null || return

    local remote_url
    remote_url=$(git remote get-url origin 2>/dev/null) || return

    if [[ "$remote_url" == git@* ]]; then
      local host="${remote_url#git@}"
      host="${host%%:*}"
      nc -z -w 1 "$host" 22 &>/dev/null || return
    else
      local host
      host=$(echo "$remote_url" | awk -F/ '{print $3}')
      nc -z -w 1 "$host" 443 &>/dev/null || return
    fi

    git fetch -p --quiet &>/dev/null
    kill -USR1 $$
  } &!
}

TRAPUSR1() {
  vcs_info
  zle && zle reset-prompt
}

add-zsh-hook precmd _git_fetch_async
add-zsh-hook precmd vcs_info

zstyle ':vcs_info:*' enable git
zstyle ':vcs_info:*' formats " %F{cyan}%c%u%b%m%f"
zstyle ':vcs_info:*' actionformats " %F{cyan}%c%u%b%m%f %a"
zstyle ':vcs_info:*' stagedstr "%F{green}"
zstyle ':vcs_info:*' unstagedstr "%F{red}"
zstyle ':vcs_info:*' check-for-changes true

zstyle ':vcs_info:git*+set-message:*' hooks git-untracked git-aheadbehind

+vi-git-untracked() {
  if git --no-optional-locks status --porcelain 2> /dev/null | grep -q "^??"; then
    hook_com[staged]+="%F{red}"
  fi
}

+vi-git-aheadbehind() {
  local ahead behind arrows=""
  ahead=$(git rev-list @{upstream}..HEAD 2>/dev/null | wc -l | tr -d ' ')
  behind=$(git rev-list HEAD..@{upstream} 2>/dev/null | wc -l | tr -d ' ')
  (( ahead > 0 )) && arrows+="↑"
  (( behind > 0 )) && arrows+="↓"
  [[ -n "$arrows" ]] && hook_com[misc]+=" %F{13}${arrows}"
}

setopt PROMPT_SUBST
export PROMPT='%F{13}%1~$vcs_info_msg_0_ '
