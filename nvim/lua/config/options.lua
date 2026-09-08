local o = vim.o

-- Use Foot's RGB palette while leaving the editor background to the terminal.
o.termguicolors = true

o.number = true
o.signcolumn = "yes"
o.scrolloff = 4
o.wrap = true
o.splitright = true
o.splitbelow = true

o.ignorecase = true
o.smartcase = true

o.expandtab = true
o.tabstop = 4
o.shiftwidth = 4
o.undofile = true
o.confirm = true

-- Pick up edits made outside nvim quickly (checktime autocmd runs on CursorHold)
o.autoread = true
o.updatetime = 300

-- Yank and paste through the Wayland clipboard (wl-clipboard)
o.clipboard = "unnamedplus"
