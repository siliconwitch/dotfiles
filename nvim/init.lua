-- Leader keys must be set before plugins load
vim.g.mapleader = " "
vim.g.maplocalleader = " "

local o = vim.o

-- Draw with the terminal's 16 ANSI colours so Foot owns the palette and the background.
o.termguicolors = false

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

-- Highlight groups use ANSI slots only and leave Normal unpainted, so Foot owns the palette and transparency.
local groups = {
  -- Editor chrome
  Normal = { ctermfg = 7 },
  NormalFloat = { ctermfg = 15, ctermbg = 0 },
  FloatBorder = { ctermfg = 8, ctermbg = 0 },
  WinSeparator = { ctermfg = 8 },
  LineNr = { ctermfg = 8 },
  CursorLineNr = { ctermfg = 15, bold = true },
  CursorLine = { underline = true },
  SignColumn = { ctermfg = 8 },
  Visual = { ctermbg = 8 },
  Search = { ctermfg = 0, ctermbg = 3 },
  CurSearch = { ctermfg = 0, ctermbg = 11 },
  IncSearch = { ctermfg = 0, ctermbg = 11 },
  MatchParen = { ctermfg = 11, underline = true },
  Pmenu = { ctermfg = 15, ctermbg = 0 },
  PmenuSel = { reverse = true },
  PmenuSbar = { ctermbg = 8 },
  PmenuThumb = { ctermbg = 7 },
  StatusLine = { ctermfg = 15 },
  StatusLineNC = { ctermfg = 8 },
  Folded = { ctermfg = 7, ctermbg = 0 },
  NonText = { ctermfg = 8 },
  Whitespace = { ctermfg = 8 },
  SpecialKey = { ctermfg = 8 },
  Directory = { ctermfg = 12 },
  Title = { ctermfg = 12, bold = true },
  ErrorMsg = { ctermfg = 9 },
  WarningMsg = { ctermfg = 11 },
  ModeMsg = { ctermfg = 15 },
  QuickFixLine = { ctermbg = 8 },
  SpellBad = { undercurl = true },

  -- Syntax
  Comment = { ctermfg = 7 },
  Identifier = { ctermfg = 9 },
  Constant = { ctermfg = 3 },
  Number = { ctermfg = 3 },
  Boolean = { ctermfg = 3 },
  String = { ctermfg = 10 },
  Character = { ctermfg = 10 },
  SpecialChar = { ctermfg = 14 },
  Function = { ctermfg = 12 },
  Special = { ctermfg = 12 },
  Statement = { ctermfg = 13 },
  Keyword = { ctermfg = 13 },
  Operator = { ctermfg = 7 },
  PreProc = { ctermfg = 13 },
  Type = { ctermfg = 11 },
  Delimiter = { ctermfg = 7 },
  Error = { ctermfg = 9 },
  Todo = { ctermfg = 11, bold = true },

  -- Treesitter refinements
  ["@variable"] = { ctermfg = 9 },
  ["@variable.builtin"] = { ctermfg = 9 },
  ["@variable.member"] = { ctermfg = 2 },
  ["@property"] = { ctermfg = 2 },
  ["@attribute"] = { ctermfg = 3 },
  ["@constructor"] = { ctermfg = 12 },
  ["@module"] = { ctermfg = 13 },
  ["@label"] = { ctermfg = 13 },
  ["@string.escape"] = { ctermfg = 14 },
  ["@markup.heading"] = { ctermfg = 12 },
  ["@markup.list"] = { ctermfg = 9 },
  ["@markup.strong"] = { ctermfg = 11, bold = true },
  ["@markup.italic"] = { ctermfg = 13, italic = true },
  ["@markup.strikethrough"] = { strikethrough = true },
  ["@markup.link.url"] = { ctermfg = 3, underline = true },
  ["@markup.link.label"] = { ctermfg = 9 },
  ["@markup.quote"] = { ctermfg = 14 },
  ["@markup.raw"] = { ctermfg = 2 },

  -- Diffs
  Added = { ctermfg = 10 },
  Changed = { ctermfg = 12 },
  Removed = { ctermfg = 9 },
  DiffAdd = { ctermfg = 10, ctermbg = 0 },
  DiffChange = { ctermbg = 0 },
  DiffDelete = { ctermfg = 9, ctermbg = 0 },
  DiffText = { ctermbg = 8 },
  GitSignsAdd = { ctermfg = 10 },
  GitSignsChange = { ctermfg = 12 },
  GitSignsDelete = { ctermfg = 9 },

  -- Diagnostics
  DiagnosticError = { ctermfg = 9 },
  DiagnosticWarn = { ctermfg = 11 },
  DiagnosticInfo = { ctermfg = 12 },
  DiagnosticHint = { ctermfg = 7 },
  DiagnosticUnderlineError = { undercurl = true },
  DiagnosticUnderlineWarn = { undercurl = true },
  DiagnosticUnderlineInfo = { underdotted = true },
  DiagnosticUnderlineHint = { underdashed = true },
}

for group, options in pairs(groups) do
  vim.api.nvim_set_hl(0, group, options)
end
