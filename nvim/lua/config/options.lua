local o = vim.o

-- Render with the terminal's 16-color palette and no painted background,
-- so foot's colors and alpha show through
o.termguicolors = false

o.number = true
o.signcolumn = "yes"
o.cursorline = true
o.scrolloff = 4
o.wrap = false
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
