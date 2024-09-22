-- Map leader key
vim.g.mapleader = " "
vim.g.maplocalleader = "\\"

-- Tab and indent options
vim.o.expandtab = true
vim.o.tabstop = 4
vim.o.shiftwidth = 4
vim.o.breakindent = true

-- Line number and symbol margin options
vim.wo.relativenumber = true
vim.o.updatetime = 250
vim.o.signcolumn = "yes"

-- Seach settings
vim.o.ignorecase = true
vim.o.smartcase = true

-- Keep undo history
vim.o.undofile = true

-- Copy to clipboard by default
vim.o.clipboard = "unnamedplus"

-- Clean up the powerline
vim.o.showmode = false
vim.o.ruler = false
vim.o.laststatus = 3

-- Bootstrap lazy.nvim
local lazypath = vim.fn.stdpath("data") .. "/lazy/lazy.nvim"

if not (vim.uv or vim.loop).fs_stat(lazypath) then
	local lazyrepo = "https://github.com/folke/lazy.nvim.git"
	local out = vim.fn.system({ "git", "clone", "--filter=blob:none", "--branch=stable", lazyrepo, lazypath })
	if vim.v.shell_error ~= 0 then
		vim.api.nvim_echo({
			{ "Failed to clone lazy.nvim:\n", "ErrorMsg" },
			{ out, "WarningMsg" },
			{ "\nPress any key to exit..." },
		}, true, {})
		vim.fn.getchar()
		os.exit(1)
	end
end

vim.opt.rtp:prepend(lazypath)

-- Load plugins
require("lazy").setup("plugins")
