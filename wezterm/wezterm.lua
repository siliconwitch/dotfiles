local wezterm = require("wezterm")

local config = wezterm.config_builder()

-- General window styling
config.background = {
    {
        source = {
            File = os.getenv("HOME") .. "/.dotfiles/wezterm/wallpaper.jpg",
        },
        vertical_align = "Middle",
        horizontal_align = "Center",
    },
    {
        source = {
            Color = "#000000",
        },
        height = "100%",
        width = "100%",
        opacity = 0.75,
    }
}
config.colors = require("cyberdream")
config.default_cursor_style = "BlinkingBar"
config.font = wezterm.font("RobotoMono Nerd Font")
config.font_size = 13
config.window_decorations = "RESIZE"
config.window_padding = {
    left = 50,
    right = 50,
    top = 50,
    bottom = 30,
}

config.underline_thickness = 2
config.hide_tab_bar_if_only_one_tab = false
config.tab_bar_at_bottom = true
config.use_fancy_tab_bar = false
config.show_new_tab_button_in_tab_bar = false

config.leader = { key = "q", mods = "ALT", timeout_milliseconds = 2000 }
config.keys = {
    {
        mods = "LEADER",
        key = "c",
        action = wezterm.action.SpawnTab "CurrentPaneDomain",
    },
    {
        mods = "LEADER",
        key = "x",
        action = wezterm.action.CloseCurrentPane { confirm = true }
    },
    {
        mods = "LEADER",
        key = "b",
        action = wezterm.action.ActivateTabRelative(-1)
    },
    {
        mods = "LEADER",
        key = "n",
        action = wezterm.action.ActivateTabRelative(1)
    },
    {
        mods = "LEADER",
        key = "|",
        action = wezterm.action.SplitHorizontal { domain = "CurrentPaneDomain" }
    },
    {
        mods = "LEADER",
        key = "-",
        action = wezterm.action.SplitVertical { domain = "CurrentPaneDomain" }
    },
    {
        mods = "LEADER",
        key = "h",
        action = wezterm.action.ActivatePaneDirection "Left"
    },
    {
        mods = "LEADER",
        key = "j",
        action = wezterm.action.ActivatePaneDirection "Down"
    },
    {
        mods = "LEADER",
        key = "k",
        action = wezterm.action.ActivatePaneDirection "Up"
    },
    {
        mods = "LEADER",
        key = "l",
        action = wezterm.action.ActivatePaneDirection "Right"
    },
    {
        mods = "LEADER",
        key = "LeftArrow",
        action = wezterm.action.AdjustPaneSize { "Left", 5 }
    },
    {
        mods = "LEADER",
        key = "RightArrow",
        action = wezterm.action.AdjustPaneSize { "Right", 5 }
    },
    {
        mods = "LEADER",
        key = "DownArrow",
        action = wezterm.action.AdjustPaneSize { "Down", 5 }
    },
    {
        mods = "LEADER",
        key = "UpArrow",
        action = wezterm.action.AdjustPaneSize { "Up", 5 }
    },
}

config.inactive_pane_hsb = {
    saturation = 0.5,
    brightness = 0.8,
}

return config
