#!/bin/sh
# Claude Code status line. Reads the status JSON on stdin.

jq -r '
  def k: if . >= 1000000 then "\(. / 100000 | floor / 10)M"
    elif . >= 1000 then "\(. / 100 | floor / 10)k"
    else tostring end;
  [ .model.display_name + (.effort.level | if . == null then "" else " (\(.))" end)
      + " \(.context_window.current_usage
          | if . == null then 0
            else .input_tokens + .cache_creation_input_tokens + .cache_read_input_tokens end
          | k)"
      + " (\(.context_window.used_percentage // 0 | floor)%)",
    (.rate_limits.five_hour.used_percentage
      | if . == null then empty else "session \(floor)% used" end)
  ] | join(" · ")'
