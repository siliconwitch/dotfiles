return {
	"scottmckendry/cyberdream.nvim",
	lazy = false,
	priority = 1000,
	opts = {
		transparent = true,
		terminal_colors = false,
		theme = { variant = "default" },
	},
	config = true,
	init = function()
		vim.cmd("colorscheme cyberdream")
	end,
}
