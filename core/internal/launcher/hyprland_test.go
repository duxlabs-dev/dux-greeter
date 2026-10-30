package launcher

import (
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestHyprlandPlanAlwaysLaunchesLua(t *testing.T) {
	const qsCmd = `/opt/q"s\bin]]/qs -p /greeter`
	const launch = `hl.exec_cmd([=[sh -c "/opt/q"s\bin]]/qs -p /greeter; hyprctl dispatch 'hl.dsp.exit()'"]=])`
	const userLua = "hl.config({ input = { kb_layout = \"de\" } })\n"

	tests := []struct {
		name       string
		file       string
		content    string
		wantPrefix string
	}{
		{name: "hyprlang config is replaced", file: "hyprland.conf", content: "input {\n    kb_layout = de\n}\nexec-once = waybar\n", wantPrefix: hyprlandBaseConfig},
		{name: "lua config is appended to", file: "greeter.lua", content: userLua, wantPrefix: "do\n" + userLua + "\nend\n"},
		{name: "lua content without suffix is appended to", file: "greeter.conf", content: userLua, wantPrefix: "do\n" + userLua + "\nend\n"},
		{name: "lua config ending in return still loads", file: "greeter.lua", content: userLua + "return {}\n", wantPrefix: "do\n" + userLua + "return {}\n\nend\n"},
		{name: "hyprlang mentioning hl. in a comment is replaced", file: "hyprland.conf", content: "# see hl.dsp docs\ninput {\n    kb_layout = de\n}\n", wantPrefix: hyprlandBaseConfig},
		{name: "no config", wantPrefix: hyprlandBaseConfig},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			t.Setenv("TMPDIR", dir)
			t.Setenv("PATH", dir)
			if err := os.WriteFile(filepath.Join(dir, "Hyprland"), []byte("#!/bin/sh\n"), 0o700); err != nil {
				t.Fatal(err)
			}
			configPath := ""
			if tt.file != "" {
				configPath = filepath.Join(dir, tt.file)
				if err := os.WriteFile(configPath, []byte(tt.content), 0o600); err != nil {
					t.Fatal(err)
				}
			}

			plan, err := buildPlan("hyprland", configPath, qsCmd)
			if err != nil {
				t.Fatal(err)
			}
			if !slices.Contains(plan.env, "HYPRLAND_NO_SD_TARGET=1") {
				t.Fatalf("env = %q, want HYPRLAND_NO_SD_TARGET=1", plan.env)
			}
			generated := plan.argv[len(plan.argv)-1]
			if !strings.HasSuffix(generated, ".lua") {
				t.Fatalf("generated config %q is not a .lua file", generated)
			}
			got, err := os.ReadFile(generated)
			if err != nil {
				t.Fatal(err)
			}
			if strings.Contains(string(got), "exec-once") {
				t.Fatalf("generated config contains hyprlang exec-once:\n%s", got)
			}
			if !strings.HasPrefix(string(got), tt.wantPrefix) {
				t.Fatalf("generated config does not start with %q:\n%s", tt.wantPrefix, got)
			}
			if !strings.Contains(string(got), launch) {
				t.Fatalf("generated config lacks %s:\n%s", launch, got)
			}
			if luac, err := exec.LookPath("luac"); err == nil {
				if out, err := exec.Command(luac, "-p", generated).CombinedOutput(); err != nil {
					t.Fatalf("generated config is not valid Lua: %s", out)
				}
			}
		})
	}
}
