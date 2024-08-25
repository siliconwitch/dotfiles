return {
	{
		"williamboman/mason.nvim",
		config = function()
			require("mason").setup()
		end,
	},
	{
		"williamboman/mason-lspconfig.nvim",
		config = function()
			require("mason-lspconfig").setup({
				ensure_installed = {
					"clangd",
					"gopls",
					"lua_ls",
					"svls",
					"tsserver",
				},
			})
		end,
	},
	{
		"neovim/nvim-lspconfig",
		config = function()
			local lspconfig = require("lspconfig")
			lspconfig.clangd.setup({})
			lspconfig.gopls.setup({})
			lspconfig.svls.setup({})
			lspconfig.lua_ls.setup({})
			lspconfig.tsserver.setup({})
			vim.keymap.set("n", "K", vim.lsp.buf.hover, {})
		end,
	},
}
