return {
  {
    "nvim-telescope/telescope.nvim",
    dependencies = {
      "nvim-lua/plenary.nvim",
      "nvim-tree/nvim-web-devicons",
      { "nvim-telescope/telescope-fzf-native.nvim", build = "make" },
    },
    cmd = "Telescope",
    keys = {
      { "<leader><leader>", "<cmd>Telescope buffers sort_mru=true ignore_current_buffer=true<cr>", desc = "Switch buffer" },
      { "<leader>ff", "<cmd>Telescope find_files<cr>", desc = "Find files" },
      { "<leader>fg", "<cmd>Telescope live_grep<cr>", desc = "Grep project" },
      { "<leader>fw", "<cmd>Telescope grep_string<cr>", desc = "Grep word under cursor" },
      { "<leader>fr", "<cmd>Telescope oldfiles<cr>", desc = "Recent files" },
      { "<leader>fh", "<cmd>Telescope help_tags<cr>", desc = "Help" },
      { "<leader>fd", "<cmd>Telescope diagnostics<cr>", desc = "Diagnostics" },
      { "<leader>fs", "<cmd>Telescope lsp_dynamic_workspace_symbols<cr>", desc = "Workspace symbols" },
      { "gd", "<cmd>Telescope lsp_definitions<cr>", desc = "Goto definition" },
      { "grr", "<cmd>Telescope lsp_references<cr>", desc = "References" },
      { "gri", "<cmd>Telescope lsp_implementations<cr>", desc = "Implementations" },
    },
    config = function()
      local telescope = require("telescope")

      telescope.setup({})
      telescope.load_extension("fzf")
    end,
  },
}
