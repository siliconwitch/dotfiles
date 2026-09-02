return {
  {
    "lewis6991/gitsigns.nvim",
    event = { "BufReadPre", "BufNewFile" },
    opts = {
      on_attach = function(buf)
        local gitsigns = require("gitsigns")

        vim.keymap.set("n", "]h", function()
          gitsigns.nav_hunk("next")
        end, { buffer = buf, desc = "Next hunk" })
        vim.keymap.set("n", "[h", function()
          gitsigns.nav_hunk("prev")
        end, { buffer = buf, desc = "Previous hunk" })
        vim.keymap.set("n", "<leader>gp", gitsigns.preview_hunk, { buffer = buf, desc = "Preview hunk" })
        vim.keymap.set("n", "<leader>gs", gitsigns.stage_hunk, { buffer = buf, desc = "Stage hunk" })
        vim.keymap.set("n", "<leader>gr", gitsigns.reset_hunk, { buffer = buf, desc = "Reset hunk" })
        vim.keymap.set("n", "<leader>gb", gitsigns.blame_line, { buffer = buf, desc = "Blame line" })
      end,
    },
  },
  {
    "sindrets/diffview.nvim",
    cmd = { "DiffviewOpen", "DiffviewFileHistory", "DiffviewClose" },
    keys = {
      { "<leader>gd", "<cmd>DiffviewOpen<cr>", desc = "Diff working tree" },
      { "<leader>gh", "<cmd>DiffviewFileHistory %<cr>", desc = "File history" },
      { "<leader>gq", "<cmd>DiffviewClose<cr>", desc = "Close diff view" },
    },
    opts = {},
  },
}
