local wezterm = require("wezterm")
local config = wezterm.config_builder()

config.colors = require("cyberdream")
config.default_cursor_style = "BlinkingBar"
config.enable_tab_bar = false
config.font = wezterm.font("RobotoMono Nerd Font")
config.font_size = 13
config.underline_thickness = 2
config.window_decorations = "RESIZE"

config.background = {
	{
		source = {
			File = os.getenv("HOME") .. "/.config/wezterm/wallpaper.jpg",
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
	},
}

config.window_padding = {
	left = 50,
	right = 50,
	top = 50,
	bottom = 50,
}

return config
