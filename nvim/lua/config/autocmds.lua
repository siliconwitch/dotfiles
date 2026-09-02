local group = vim.api.nvim_create_augroup("user", { clear = true })

-- Reload files edited outside nvim, e.g. by claude-code or codex
vim.api.nvim_create_autocmd({ "FocusGained", "BufEnter", "CursorHold" }, {
  group = group,
  callback = function()
    if vim.fn.getcmdwintype() ~= "" then
      return
    end

    vim.cmd("checktime")
  end,
})

vim.api.nvim_create_autocmd("FileChangedShellPost", {
  group = group,
  callback = function(args)
    vim.notify("Reloaded from disk: " .. vim.fn.fnamemodify(args.file, ":~:."))
  end,
})

vim.api.nvim_create_autocmd("TextYankPost", {
  group = group,
  callback = function()
    vim.hl.on_yank()
  end,
})

vim.api.nvim_create_autocmd("FileType", {
  group = group,
  pattern = "go",
  callback = function()
    vim.bo.expandtab = false
    vim.cmd.compiler("go")
  end,
})
