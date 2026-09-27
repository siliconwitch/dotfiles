-- Leader keys must be set before plugins load
vim.g.mapleader = " "
vim.g.maplocalleader = " "

local o = vim.o

-- Use Foot's RGB palette while leaving the editor background to the terminal.
o.termguicolors = true

o.number = true
o.signcolumn = "yes"
o.statusline = "%<%f%( %m%)%( %r%)"
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

vim.keymap.set("n", "<Esc>", "<cmd>nohlsearch<cr>")
vim.keymap.set("n", "<leader>w", "<cmd>set wrap!<cr>", { desc = "Toggle wrap" })

local group = vim.api.nvim_create_augroup("user", { clear = true })

-- Reload files edited outside nvim, e.g. by claude-code or codex
vim.api.nvim_create_autocmd({ "FocusGained", "BufEnter", "CursorHold" }, {
  group = group,
  callback = function()
    if vim.fn.getcmdwintype() ~= "" then
      return
    end

    vim.cmd("checktime")
  end,
})

vim.api.nvim_create_autocmd("FileChangedShellPost", {
  group = group,
  callback = function(args)
    vim.notify("Reloaded from disk: " .. vim.fn.fnamemodify(args.file, ":~:."))
  end,
})

vim.api.nvim_create_autocmd("TextYankPost", {
  group = group,
  callback = function()
    vim.hl.on_yank()
  end,
})

vim.api.nvim_create_autocmd("FileType", {
  group = group,
  pattern = "go",
  callback = function()
    vim.bo.expandtab = false
  end,
})

-- Parsers are built against the plugin revision, so rebuild them after an update
vim.api.nvim_create_autocmd("PackChanged", {
  group = group,
  callback = function(args)
    if args.data.spec.name == "nvim-treesitter" and args.data.kind == "update" then
      vim.cmd.TSUpdate()
    end
  end,
})

vim.pack.add({
  "https://github.com/lewis6991/gitsigns.nvim",
  "https://github.com/neovim/nvim-lspconfig",
  { src = "https://github.com/nvim-treesitter/nvim-treesitter", version = "main" },
  "https://github.com/folke/which-key.nvim",
}, { confirm = false })

require("gitsigns").setup({
  on_attach = function(buf)
    local gitsigns = require("gitsigns")

    vim.keymap.set("n", "]h", function()
      gitsigns.nav_hunk("next")
    end, { buffer = buf, desc = "Next hunk" })
    vim.keymap.set("n", "[h", function()
      gitsigns.nav_hunk("prev")
    end, { buffer = buf, desc = "Previous hunk" })
    vim.keymap.set("n", "<leader>gp", gitsigns.preview_hunk_inline, { buffer = buf, desc = "Expand hunk inline" })
  end,
})

vim.lsp.config("lua_ls", {
  settings = {
    Lua = {
      runtime = { version = "LuaJIT" },
      workspace = { library = { vim.env.VIMRUNTIME } },
      diagnostics = { globals = { "vim" } },
    },
  },
})

vim.lsp.enable({ "clangd", "gopls", "lua_ls", "nil_ls" })

vim.diagnostic.config({
  virtual_text = true,
  severity_sort = true,
})

vim.api.nvim_create_autocmd("LspAttach", {
  group = group,
  callback = function(args)
    vim.keymap.set("n", "<leader>fs", vim.lsp.buf.workspace_symbol, { buffer = args.buf, desc = "Workspace symbols" })
  end,
})

require("nvim-treesitter").install({
  "bash", "c", "cpp", "go", "gomod", "gosum", "gowork", "json", "lua",
  "make", "markdown", "markdown_inline", "nix", "regex", "toml", "yaml",
})

vim.api.nvim_create_autocmd("FileType", {
  group = group,
  callback = function(args)
    local started = pcall(vim.treesitter.start, args.buf)

    if started then
      vim.bo[args.buf].indentexpr = "v:lua.require'nvim-treesitter'.indentexpr()"
    end
  end,
})

require("which-key").setup()

vim.cmd.colorscheme("base16_transparent")
