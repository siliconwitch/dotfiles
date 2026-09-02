local parsers = {
  "bash", "c", "cpp", "go", "gomod", "gosum", "gowork", "json", "lua",
  "make", "markdown", "markdown_inline", "nix", "regex", "toml", "yaml",
}

return {
  {
    "nvim-treesitter/nvim-treesitter",
    branch = "main",
    build = ":TSUpdate",
    lazy = false,
    config = function()
      require("nvim-treesitter").install(parsers)

      vim.api.nvim_create_autocmd("FileType", {
        callback = function(args)
          local started = pcall(vim.treesitter.start, args.buf)

          if started then
            vim.bo[args.buf].indentexpr = "v:lua.require'nvim-treesitter'.indentexpr()"
          end
        end,
      })
    end,
  },
}
