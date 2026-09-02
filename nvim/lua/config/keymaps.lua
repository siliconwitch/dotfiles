local map = vim.keymap.set

map("n", "<Esc>", "<cmd>nohlsearch<cr>")

map("n", "<C-h>", "<C-w>h", { desc = "Left window" })
map("n", "<C-j>", "<C-w>j", { desc = "Lower window" })
map("n", "<C-k>", "<C-w>k", { desc = "Upper window" })
map("n", "<C-l>", "<C-w>l", { desc = "Right window" })
map("n", "<leader>wv", "<cmd>vsplit<cr>", { desc = "Vertical split" })
map("n", "<leader>ws", "<cmd>split<cr>", { desc = "Horizontal split" })

map("t", "<Esc><Esc>", [[<C-\><C-n>]], { desc = "Exit terminal mode" })

-- Build with makeprg into the quickfix list (autocmds.lua sets `compiler go`)
map("n", "<leader>m", "<cmd>make!<cr><cmd>cwindow<cr>", { desc = "Build" })
map("n", "<leader>q", "<cmd>cwindow<cr>", { desc = "Quickfix window" })
