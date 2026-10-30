package launcher

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/duxlabs-dev/dux-go/log"
	"github.com/duxlabs-dev/dux-greeter/core/internal/config"
)

var niriOverrideFiles = []string{
	"/usr/share/greetd/niri_overrides.kdl",
	"/etc/greetd/niri_overrides.kdl",
}

var hyprlandLuaPattern = regexp.MustCompile(`(^|[^[:alnum:]_])hl\.`)

const hyprlandBaseConfig = `hl.env("DMS_RUN_GREETER", "1")

hl.config({
	misc = {
		disable_hyprland_logo = true,
	},
})
`

const aqueousBaseConfig = `[keybinds]
spawn_terminal = []
toggle_start_menu = []
screenshot = []
lock_screen = []
toggle_overview = []
`

type launchPlan struct {
	logTag string
	argv   []string
	env    []string
}

func buildPlan(compositor, configPath, qsCmd string) (launchPlan, error) {
	switch compositor {
	case "niri":
		return buildNiriPlan(configPath, qsCmd)
	case "hyprland":
		return buildHyprlandPlan(configPath, qsCmd)
	case "sway":
		return buildAppendExecPlan("sway", configPath, fmt.Sprintf("\nexec \"%s; swaymsg exit\"\n", qsCmd), "sway", "--unsupported-gpu", "-c")
	case "scroll":
		return buildAppendExecPlan("scroll", configPath, fmt.Sprintf("\nexec \"%s; scrollmsg exit\"\n", qsCmd), "scroll", "-c")
	case "miracle", "miracle-wm":
		return buildAppendExecPlan("miracle", configPath, fmt.Sprintf("\nexec \"%s; miraclemsg exit\"\n", qsCmd), "miracle-wm", "-c")
	case "labwc":
		return buildLabwcPlan(configPath, qsCmd)
	case "mango", "mangowc":
		return buildMangoPlan(configPath, qsCmd)
	case "aqueous":
		return buildAqueousPlan(configPath, qsCmd)
	default:
		return launchPlan{}, fmt.Errorf("unsupported compositor: %s\nSupported compositors: niri, hyprland, sway, scroll, miracle, mango, labwc, aqueous", compositor)
	}
}

func buildNiriPlan(configPath, qsCmd string) (launchPlan, error) {
	if err := requireCommand("niri"); err != nil {
		return launchPlan{}, err
	}

	base := config.NiriGreeterConfig
	if configPath != "" {
		content, err := os.ReadFile(configPath)
		if err != nil {
			return launchPlan{}, err
		}
		base = string(content)
	}

	var builder strings.Builder
	builder.WriteString(base)
	for _, overrideFile := range niriOverrideFiles {
		if _, err := os.Stat(overrideFile); err != nil {
			continue
		}
		fmt.Fprintf(&builder, "\ninclude %q\n", overrideFile)
	}
	fmt.Fprintf(&builder, "\nspawn-at-startup \"sh\" \"-c\" \"%s; niri msg action quit --skip-confirmation\"\n", qsCmd)

	tempConfig, err := writeTempConfig(builder.String(), "")
	if err != nil {
		return launchPlan{}, err
	}
	return launchPlan{logTag: "niri", argv: []string{"niri", "-c", tempConfig}}, nil
}

func buildHyprlandPlan(configPath, qsCmd string) (launchPlan, error) {
	startHyprland := commandExists("start-hyprland")
	if !startHyprland && !commandExists("Hyprland") {
		return launchPlan{}, fmt.Errorf("neither 'start-hyprland' nor 'Hyprland' was found in PATH")
	}

	lua, err := hyprlandLuaConfig(configPath, qsCmd)
	if err != nil {
		return launchPlan{}, err
	}
	tempConfig, err := writeTempConfig(lua, ".lua")
	if err != nil {
		return launchPlan{}, err
	}

	// Without MANAGERPID, Hyprland starts hyprland-session.target, pulling graphical-session units up as the greeter user.
	env := []string{"HYPRLAND_NO_SD_TARGET=1"}
	if startHyprland {
		return launchPlan{logTag: "hyprland", argv: []string{"start-hyprland", "--", "--config", tempConfig}, env: env}, nil
	}
	return launchPlan{logTag: "hyprland", argv: []string{"Hyprland", "-c", tempConfig}, env: env}, nil
}

func hyprlandLuaConfig(configPath, qsCmd string) (string, error) {
	launch := fmt.Sprintf("\nhl.on(\"hyprland.start\", function()\n\thl.exec_cmd(%s)\nend)\n", luaLongString(fmt.Sprintf(`sh -c "%s; hyprctl dispatch 'hl.dsp.exit()'"`, qsCmd)))
	if configPath == "" {
		return hyprlandBaseConfig + launch, nil
	}

	content, err := os.ReadFile(configPath)
	if err != nil {
		return "", err
	}
	if !IsHyprlandLuaConfig(configPath, content) {
		log.Warnf("Hyprland 0.57 only reads Lua config; ignoring %s and using the built-in config", configPath)
		return hyprlandBaseConfig + launch, nil
	}
	// A top-level `return` in the user's file would make the appended handler a syntax error.
	return "do\n" + string(content) + "\nend\n" + launch, nil
}

func luaLongString(s string) string {
	level := ""
	for strings.Contains(s+"]", "]"+level+"]") {
		level += "="
	}
	return "[" + level + "[" + s + "]" + level + "]"
}

func buildAppendExecPlan(logTag, configPath, execLine, binary string, argvPrefix ...string) (launchPlan, error) {
	if err := requireCommand(binary); err != nil {
		return launchPlan{}, err
	}

	base := ""
	if configPath != "" {
		content, err := os.ReadFile(configPath)
		if err != nil {
			return launchPlan{}, err
		}
		base = string(content)
	}

	tempConfig, err := writeTempConfig(base+execLine, "")
	if err != nil {
		return launchPlan{}, err
	}
	return launchPlan{logTag: logTag, argv: append(append([]string{binary}, argvPrefix...), tempConfig)}, nil
}

func buildLabwcPlan(configPath, qsCmd string) (launchPlan, error) {
	if err := requireCommand("labwc"); err != nil {
		return launchPlan{}, err
	}
	if configPath != "" {
		return launchPlan{logTag: "labwc", argv: []string{"labwc", "--config", configPath, "--session", qsCmd}}, nil
	}
	return launchPlan{logTag: "labwc", argv: []string{"labwc", "--session", qsCmd}}, nil
}

func buildMangoPlan(configPath, qsCmd string) (launchPlan, error) {
	if err := requireCommand("mango"); err != nil {
		return launchPlan{}, err
	}
	session := fmt.Sprintf("%s && mmsg dispatch quit", qsCmd)
	if configPath != "" {
		return launchPlan{logTag: "mango", argv: []string{"mango", "-c", configPath, "-s", session}}, nil
	}
	return launchPlan{logTag: "mango", argv: []string{"mango", "-s", session}}, nil
}

func buildAqueousPlan(configPath, qsCmd string) (launchPlan, error) {
	if err := requireCommand("aqueous"); err != nil {
		return launchPlan{}, err
	}

	var err error
	if configPath == "" {
		configPath, err = writeTempConfig(aqueousBaseConfig, ".toml")
	} else {
		configPath, err = filepath.Abs(configPath)
		if err != nil {
			return launchPlan{}, err
		}
		_, err = os.ReadFile(configPath)
	}
	if err != nil {
		return launchPlan{}, err
	}

	return launchPlan{
		logTag: "aqueous",
		argv:   []string{"aqueous", "-no-xwayland", "-c", fmt.Sprintf(`compositor_pid=$PPID; %s; kill -TERM "$compositor_pid"`, qsCmd)},
		env: []string{
			"AQUEOUS_CONFIG=" + configPath,
			"AQUEOUS_LAYOUT=" + configPath,
			"AQUEOUS_INPUT=" + configPath,
			"AQUEOUS_RULES=" + configPath,
			"AQUEOUS_OUTPUTS=" + configPath,
		},
	}, nil
}

func IsHyprlandLuaConfig(configPath string, content []byte) bool {
	if strings.HasSuffix(configPath, ".lua") {
		return true
	}
	for line := range strings.Lines(string(content)) {
		if strings.HasPrefix(strings.TrimSpace(line), "#") {
			continue
		}
		if hyprlandLuaPattern.MatchString(line) {
			return true
		}
	}
	return false
}

func writeTempConfig(content, suffix string) (string, error) {
	file, err := os.CreateTemp("", "dms-greeter-*"+suffix)
	if err != nil {
		return "", err
	}
	defer file.Close()
	if _, err := file.WriteString(content); err != nil {
		os.Remove(file.Name())
		return "", err
	}
	return file.Name(), nil
}

func requireCommand(name string) error {
	if commandExists(name) {
		return nil
	}
	return fmt.Errorf("required command %q was not found in PATH", name)
}

func commandExists(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}
