-- cyberdream theme for wezterm
return {
    foreground = "#ffffff",
    background = "#16181a",

    cursor_bg = "#ffffff",
    cursor_fg = "#16181a",
    cursor_border = "#ffffff",

    selection_fg = "#ffffff",
    selection_bg = "#3c4048",

    scrollbar_thumb = "#16181a",
    split = "#7b8496",

    ansi = { "#16181a", "#ff6e5e", "#5eff6c", "#f1ff5e", "#5ea1ff", "#bd5eff", "#5ef1ff", "#ffffff" },
    brights = { "#3c4048", "#ff6e5e", "#5eff6c", "#f1ff5e", "#5ea1ff", "#bd5eff", "#5ef1ff", "#ffffff" },
    indexed = { [16] = "#ffbd5e", [17] = "#ff6e5e" },

    tab_bar = {
        background = "rgba(0,0,0,0)",
        active_tab = {
            bg_color = "rgba(0,0,0,0.0)",
            fg_color = "#ffffff",
            intensity = "Bold",
        },
        inactive_tab = {
            bg_color = "rgba(0,0,0,0.0)",
            fg_color = "#7b8496",
        },
        inactive_tab_hover = {
            bg_color = "rgba(0,0,0,0.0)",
            fg_color = "#7b8496",
            intensity = "Bold",
        }
    }
}
