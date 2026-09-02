return {
  {
    "akinsho/toggleterm.nvim",
    keys = {
      { [[<C-\>]], desc = "Toggle terminal", mode = { "n", "t" } },
      { "<leader>tf", "<cmd>ToggleTerm direction=float<cr>", desc = "Floating terminal" },
    },
    opts = {
      open_mapping = [[<C-\>]],
      size = 15,
    },
  },
}
