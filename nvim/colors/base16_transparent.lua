-- Port of helix's base16_transparent theme: 16 ANSI terminal colors, no
-- painted background, so foot's palette and alpha define the look
vim.cmd("highlight clear")
vim.g.colors_name = "base16_transparent"

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

for group, opts in pairs(groups) do
  vim.api.nvim_set_hl(0, group, opts)
end
