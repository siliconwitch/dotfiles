-- Uses Foot's palette without painting Normal, so Foot controls transparency.
vim.cmd("highlight clear")
vim.g.colors_name = "base16_transparent"

local terminal_colors = {
  [0] = "#1c1b19",
  [1] = "#ef2f27",
  [2] = "#519f50",
  [3] = "#fbb829",
  [4] = "#2c78bf",
  [5] = "#e02c6d",
  [6] = "#0aaeb3",
  [7] = "#baa67f",
  [8] = "#918175",
  [9] = "#f75341",
  [10] = "#98bc37",
  [11] = "#fed06e",
  [12] = "#68a8e4",
  [13] = "#ff5c8f",
  [14] = "#2be4d0",
  [15] = "#fce8c3",
}

for index = 0, 15 do
  vim.g["terminal_color_" .. index] = terminal_colors[index]
end

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

  -- Syntax (helix scope -> ANSI color, per base16_transparent.toml)
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
  local resolved = vim.tbl_extend("force", {}, options, {
    fg = terminal_colors[options.ctermfg],
    bg = terminal_colors[options.ctermbg],
  })

  vim.api.nvim_set_hl(0, group, resolved)
end
