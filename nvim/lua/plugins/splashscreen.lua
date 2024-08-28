return {
	"goolord/alpha-nvim",
	dependencies = { "nvim-tree/nvim-web-devicons" },
	config = function()
		local alpha = require("alpha")
		local dashboard = require("alpha.themes.dashboard")

		dashboard.section.header.val = {
			[[      _ _ _                         _ _       _      ]],
			[[     (_) (_)                       (_) |     | |     ]],
			[[  ___ _| |_  ___ ___  _ ____      ___| |_ ___| |__   ]],
			[[ / __| | | |/ __/ _ \| '_ \ \ /\ / / | __/ __| '_ \  ]],
			[[ \__ \ | | | (_| (_) | | | \ V  V /| | || (__| | | | ]],
			[[ |___/_|_|_|\___\___/|_| |_|\_/\_/ |_|\__\___|_| |_| ]],
		}
		alpha.setup(dashboard.opts)
		-- local startify = require("alpha.themes.startify")
		-- startify.file_icons.provider = "devicons"
		-- require("alpha").setup(startify.config)
	end,
}
