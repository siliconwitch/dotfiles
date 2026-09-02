return {
  {
    "neovim/nvim-lspconfig",
    dependencies = { "hrsh7th/cmp-nvim-lsp" },
    event = { "BufReadPre", "BufNewFile" },
    config = function()
      vim.lsp.config("*", {
        capabilities = require("cmp_nvim_lsp").default_capabilities(),
      })

      vim.lsp.config("lua_ls", {
        settings = {
          Lua = {
            runtime = { version = "LuaJIT" },
            workspace = { library = { vim.env.VIMRUNTIME } },
            diagnostics = { globals = { "vim" } },
          },
        },
      })

      vim.lsp.enable({ "clangd", "gopls", "lua_ls", "nil_ls" })

      vim.diagnostic.config({
        virtual_text = true,
        severity_sort = true,
      })

      vim.api.nvim_create_autocmd("LspAttach", {
        callback = function(args)
          local buf = args.buf
          local client = vim.lsp.get_client_by_id(args.data.client_id)

          vim.keymap.set("n", "gD", vim.lsp.buf.declaration, { buffer = buf, desc = "Goto declaration" })
          vim.keymap.set("n", "<leader>cf", function()
            vim.lsp.buf.format()
          end, { buffer = buf, desc = "Format buffer" })
          vim.keymap.set("n", "<leader>th", function()
            vim.lsp.inlay_hint.enable(not vim.lsp.inlay_hint.is_enabled({ bufnr = buf }), { bufnr = buf })
          end, { buffer = buf, desc = "Toggle inlay hints" })

          if client and client.name == "clangd" then
            vim.keymap.set("n", "<leader>ch", function()
              local params = vim.lsp.util.make_text_document_params(buf)
              client:request("textDocument/switchSourceHeader", params, function(err, result)
                if err or not result then
                  return
                end

                vim.cmd.edit(vim.uri_to_fname(result))
              end, buf)
            end, { buffer = buf, desc = "Switch source/header" })
          end
        end,
      })
    end,
  },
}
