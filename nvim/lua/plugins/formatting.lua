return {
	{
		"stevearc/conform.nvim",
		event = { "BufReadPre", "BufNewFile" },
		config = function()
			local conform = require("conform")

			conform.setup({
				formatters_by_ft = {
					javascript = { "prettier" },
					html = { "prettier" },
					json = { "prettier" },
					css = { "prettier" },
					markdown = { "prettier" },
					lua = { "stylua" },
					python = { "isort", "black" },
					c = { "clang-format" },
					cpp = { "clang-format" },
					go = { "gopls" },
				},
				format_on_save = {
					lsp_fallback = true,
				},
			})

			vim.keymap.set({ "n", "v" }, "<leader>mp", function()
				conform.format({
					lsp_fallback = true,
				})
			end, { desc = "Format file or range (in visual mode)" })
		end,
	},
	{
		"zapling/mason-conform.nvim",
		config = true,
	},
}
