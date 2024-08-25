return {
	"nvim-treesitter/nvim-treesitter",
	build = ":TSUpdate",
	config = function()
		require("nvim-treesitter.configs").setup({
			ensure_installed = {
				"c",
				"cmake",
				"cpp",
				"css",
				"csv",
				"dart",
				"diff",
				"go",
				"html",
				"javascript",
				"json",
				"lua",
				"make",
				"python",
				"toml",
				"verilog",
				"vim",
				"vimdoc",
				"yaml",
			},
			sync_install = false,
			highlight = { enable = true },
			indent = { enable = true },
		})

		vim.treesitter.language.register("verilog", "sv")
	end,
}
