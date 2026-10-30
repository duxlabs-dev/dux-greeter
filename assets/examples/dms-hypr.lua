-- Hyprland 0.55+ greeter config: dms-greeter --command hyprland -C /etc/greetd/dms-hypr.lua
-- The launcher appends the greeter launch itself.

hl.env("DMS_RUN_GREETER", "1")

hl.config({
	misc = {
		disable_hyprland_logo = true,
	},
})
