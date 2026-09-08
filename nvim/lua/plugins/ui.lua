return {
  {
    "folke/snacks.nvim",
    priority = 1000,
    lazy = false,
    opts = {
      dashboard = {
        enabled = true,
        sections = {
          {
            section = "terminal",
            cmd = "chafa "
              .. vim.fn.shellescape(vim.fn.stdpath("config") .. "/assets/moss-boulder-garden.png")
              .. " --format symbols --colors full --symbols block+border --size 51x14 --stretch --probe off --threads 2 --work 4",
            height = 14,
            padding = 1,
          },
          {
            pane = 2,
            { section = "keys", gap = 1, padding = 1 },
            { section = "startup" },
          },
        },
      },
    },
  },
  {
    "nvim-lualine/lualine.nvim",
    dependencies = { "nvim-tree/nvim-web-devicons" },
    event = "VeryLazy",
    config = function()
      local colors = {
        blue = "#68a8e4",
        bright_magenta = "#ff5c8f",
        bright_red = "#f75341",
        cyan = "#0aaeb3",
        foreground = "#fce8c3",
        green = "#519f50",
        magenta = "#e02c6d",
        muted = "#918175",
        red = "#ef2f27",
      }

      -- Branch colors follow zsh/prompt.zsh: red for unstaged or untracked,
      -- green for staged, cyan when clean, magenta ahead/behind arrows
      local function refresh_git_status(buf)
        if not vim.api.nvim_buf_is_valid(buf) or vim.bo[buf].buftype ~= "" then
          return
        end

        local file = vim.api.nvim_buf_get_name(buf)
        if file == "" then
          return
        end

        local root = vim.fs.root(file, ".git")
        if not root then
          vim.b[buf].git_status = nil
          return
        end

        vim.system({ "git", "-C", root, "status", "--porcelain=v2", "--branch" }, { text = true }, function(result)
          if result.code ~= 0 then
            return
          end

          local branch, oid
          local ahead, behind = 0, 0
          local staged, dirty = false, false
          for line in result.stdout:gmatch("[^\n]+") do
            local head = line:match("^# branch%.head (.+)")
            if head then
              branch = head
            end

            local sha = line:match("^# branch%.oid (%w+)")
            if sha then
              oid = sha
            end

            local plus, minus = line:match("^# branch%.ab %+(%d+) %-(%d+)")
            if plus then
              ahead, behind = tonumber(plus), tonumber(minus)
            end

            local index, worktree = line:match("^[12u] (.)(.)")
            if index then
              staged = staged or index ~= "."
              dirty = dirty or worktree ~= "."
            end

            dirty = dirty or line:sub(1, 1) == "?"
          end

          if branch == "(detached)" and oid then
            branch = oid:sub(1, 7)
          end

          vim.schedule(function()
            if not vim.api.nvim_buf_is_valid(buf) then
              return
            end

            vim.b[buf].git_status = {
              root = root,
              branch = branch,
              arrows = (ahead > 0 and "↑" or "") .. (behind > 0 and "↓" or ""),
              color = (dirty and colors.red) or (staged and colors.green) or colors.cyan,
            }
            require("lualine").refresh()
          end)
        end)
      end

      local group = vim.api.nvim_create_augroup("user_statusline_git", { clear = true })
      vim.api.nvim_create_autocmd(
        { "BufReadPost", "BufNewFile", "BufEnter", "BufWritePost", "FocusGained", "DirChanged" },
        {
          group = group,
          callback = function(args)
            refresh_git_status(args.buf)
          end,
        }
      )
      vim.api.nvim_create_autocmd("User", {
        group = group,
        pattern = "GitSignsUpdate",
        callback = function(args)
          refresh_git_status(args.buf)
        end,
      })

      local function git_branch()
        local status = vim.b.git_status
        if not status then
          return ""
        end

        return "(" .. status.branch .. ")"
      end

      local function git_arrows()
        local status = vim.b.git_status
        return status and status.arrows or ""
      end

      local function file_path()
        local name = vim.api.nvim_buf_get_name(0)
        if name == "" then
          return "[No Name]"
        end

        local status = vim.b.git_status
        local path
        if status and name:sub(1, #status.root + 1) == status.root .. "/" then
          path = name:sub(#status.root + 2)
        else
          path = vim.fn.fnamemodify(name, ":~:.")
        end

        if vim.bo.modified then
          path = path .. " [+]"
        end
        if vim.bo.readonly then
          path = path .. " [RO]"
        end

        return path
      end

      require("lualine").setup({
        options = {
          theme = {
            normal = {
              a = { fg = colors.blue, bg = "none", gui = "bold" },
              b = { fg = colors.foreground, bg = "none" },
              c = { fg = colors.bright_magenta, bg = "none" },
            },
            insert = { a = { fg = colors.green, bg = "none", gui = "bold" } },
            visual = { a = { fg = colors.magenta, bg = "none", gui = "bold" } },
            replace = { a = { fg = colors.bright_red, bg = "none", gui = "bold" } },
            inactive = {
              a = { fg = colors.muted, bg = "none" },
              b = { fg = colors.muted, bg = "none" },
              c = { fg = colors.muted, bg = "none" },
            },
          },
          section_separators = "",
          component_separators = "",
        },
        sections = {
          lualine_a = {
            {
              "mode",
              fmt = function(mode)
                return mode:sub(1, 3)
              end,
            },
          },
          lualine_b = {},
          lualine_c = {
            {
              git_branch,
              color = function()
                local status = vim.b.git_status
                return { fg = status and status.color or colors.bright_magenta }
              end,
              padding = { left = 1, right = 0 },
            },
            { git_arrows, color = { fg = colors.bright_magenta }, padding = { left = 1, right = 0 } },
            { file_path },
          },
          lualine_x = {},
          lualine_y = {},
          lualine_z = {},
        },
        inactive_sections = {
          lualine_a = {},
          lualine_b = {},
          lualine_c = { { file_path } },
          lualine_x = {},
          lualine_y = {},
          lualine_z = {},
        },
      })

      refresh_git_status(vim.api.nvim_get_current_buf())
    end,
  },
  {
    "folke/noice.nvim",
    dependencies = { "MunifTanjim/nui.nvim" },
    event = "VeryLazy",
    opts = {
      views = {
        cmdline_popup = {
          position = { row = "100%", col = "100%" },
          size = { width = "auto", height = "auto", min_width = 20 },
          border = { style = "none", padding = { 0, 1 } },
        },
      },
    },
  },
  {
    "folke/which-key.nvim",
    event = "VeryLazy",
    opts = {},
  },
  {
    "folke/trouble.nvim",
    cmd = "Trouble",
    keys = {
      { "<leader>xx", "<cmd>Trouble diagnostics toggle<cr>", desc = "Diagnostics list" },
    },
    opts = {},
  },
  {
    "mbbill/undotree",
    cmd = "UndotreeToggle",
    keys = {
      { "<leader>u", "<cmd>UndotreeToggle<cr>", desc = "Undo tree" },
    },
  },
}
