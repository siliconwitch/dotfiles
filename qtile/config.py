from libqtile import bar, layout, widget
from libqtile.config import Click, Drag, Group, Key, Match, Screen
from libqtile.lazy import lazy
from libqtile.utils import guess_terminal

# Spawn picom on startup
lazy.spawn("picom")

# Modifer is the command key
mod = "mod4"

# Guess the terminal, or add a specific one here
terminal = guess_terminal()

# Key bindings
keys = [
    # Cmd-Shift-R reload Qtile
    Key([mod, "shift"], "r", lazy.reload_config(), desc = "Reload the config"),
    
    # Cmd-Space starts the Terminal
    Key([mod], "space", lazy.spawn(terminal), desc = "Launch terminal"),

    # Cmd-R runs the prompt widget - TODO change this to Rofi
    Key([mod], "r", lazy.spawncmd(), desc = "Spawn a command using a prompt widget"),
    
    # Cmd-Arrow switches between windows
    Key([mod], "left", lazy.layout.left(), desc = "Move focus to left"),
    Key([mod], "right", lazy.layout.right(), desc = "Move focus to right"),
    Key([mod], "up", lazy.layout.up(), desc = "Move focus up"),
    Key([mod], "down", lazy.layout.down(), desc = "Move focus down"),
    
    # Cmd-Shift-Arrow swaps windows around
    Key([mod, "shift"], "left", lazy.layout.shuffle_left(), desc="Move window to the left"),
    Key([mod, "shift"], "right", lazy.layout.shuffle_right(), desc="Move window to the right"),
    Key([mod, "shift"], "up", lazy.layout.shuffle_up(), desc="Move window up"),
    Key([mod, "shift"], "down", lazy.layout.shuffle_down(), desc="Move window down"),

    # Cmd-Ctrl-Arrow resizes the windows
    Key([mod, "control"], "left", lazy.layout.grow_left(), desc="Resize window left"),
    Key([mod, "control"], "right", lazy.layout.grow_right(), desc="Resize window right"),
    Key([mod, "control"], "up", lazy.layout.grow_up(), desc="Resize window up"),
    Key([mod, "control"], "down", lazy.layout.grow_down(), desc="Resize window down"),
    
    # Cmd-F switches between fullscreen and windowed
    Key([mod], "f", lazy.next_layout(), desc="Switch between fullscreen and windowed layout"),

    # Cmd-Shift-F toggles between floating and tiled layout
    Key([mod, "shift"], "f", lazy.window.toggle_floating()),
    
    # Cmd-B balances the windows if they are resized  
    Key([mod], "b", lazy.layout.normalize(), desc="Balance the window sizes"),

    # Cmd-W kills the selected window
    Key([mod], "w", lazy.window.kill(), desc="Kill focused window"),
]

# Total of 5 desktops. Can be more
groups = [Group(i) for i in "12345"]

# Add some keybindings for the desktops
for i in groups:
    keys.extend(
        [
            # Cmd-Number switches to that desktop
            Key([mod], i.name, lazy.group[i.name].toscreen(),
                desc="Switch to group {}".format(i.name)),
            
            # Cmd-Shift-Number switches to that desktop, and take the current window with it
            Key([mod, "shift"], i.name, lazy.window.togroup(i.name, switch_group=True),
                desc="Switch to & move focused window to group {}".format(i.name),
            ),
        ]
    )

# Configure the layouts
layouts = [
    # Configure the column view
    layout.Columns(

        # No border
        border_width = 0,

        # Margin of 8
        margin = 8,
        ),

    # This is just a full screen view of the current window
    layout.Max(),
]

# Set the widget font and size
widget_defaults = dict(
    font="Roboto Mono",   # font="Source Code Pro",

    fontsize=28,
    padding=6,
)
extension_defaults = widget_defaults.copy()

# Configure screen settings
screens = [
    Screen(
        # Set the wallpaper
        wallpaper = "~/tools/dev-config/qtile/wallpaper.jpg",
        wallpaper_mode = "fill",

        # Configure the bar
        bottom=bar.Bar(
            [
                # The screen selector
                widget.GroupBox(
                    foreground = "FAFAFA",
                    active = "FAFAFA",
                    inactive = "FAFAFA80",
                    highlight_method = "line",
                    highlight_color = ["000000","000000"],
                    this_current_screen_border = "66df48",
                ),

                # Run launcher TODO change this to Rofi
                widget.Prompt(),

                # Gap
                widget.Spacer(),

                # Clock
                widget.Clock(format="%H:%M, %d %b %Y"),
            ],

            # Bar height
            50,

            # Top margin should be 8 to match the other gaps
            margin = [8, 0, 0, 0],

            # Set the background to semi transparent
            background="#00000055"
        ),
        
        right=bar.Gap(8),
        left=bar.Gap(8),
        top=bar.Gap(8)
    ),
]

# Drag floating layouts.
mouse = [
    # Cmd-MouseDrag pulls a window out of the layout and makes it floating
    Drag([mod], "Button1", lazy.window.set_position_floating(), start=lazy.window.get_position()),

    # Cmd-Control-MouseDrag will resize the window
    Drag([mod, "control"], "Button1", lazy.window.set_size_floating(), start=lazy.window.get_size()),
]

dgroups_key_binder = None
dgroups_app_rules = []  # type: list
follow_mouse_focus = True
bring_front_click = True
cursor_warp = False
floating_layout = layout.Floating(
    float_rules=[
        # Run the utility of `xprop` to see the wm class and name of an X client.
        *layout.Floating.default_float_rules,
        Match(wm_class="confirmreset"),  # gitk
        Match(wm_class="makebranch"),  # gitk
        Match(wm_class="maketag"),  # gitk
        Match(wm_class="ssh-askpass"),  # ssh-askpass
        Match(title="branchdialog"),  # gitk
        Match(title="pinentry"),  # GPG key password entry
    ]
)
auto_fullscreen = True
focus_on_window_activation = "smart"
reconfigure_screens = True

# If things like steam games want to auto-minimize themselves when losing
# focus, should we respect this or not?
auto_minimize = True

# When using the Wayland backend, this can be used to configure input devices.
wl_input_rules = None

# XXX: Gasp! We're lying here. In fact, nobody really uses or cares about this
# string besides java UI toolkits; you can see several discussions on the
# mailing lists, GitHub issues, and other WM documentation that suggest setting
# this string if your java app doesn't work correctly. We may as well just lie
# and say that we're a working one by default.
#
# We choose LG3D to maximize irony: it is a 3D non-reparenting WM written in
# java that happens to be on java's whitelist.
wmname = "LG3D"
