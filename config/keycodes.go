package config

import "github.com/BurntSushi/xgb/xproto"

var AppCodes = map[xproto.Keycode]string{
	68: "gnome-terminal",
	69: "obsidian",

	75: "vivaldi-stable",
	76: "emacs",
}

func GetAppCodesValues() []string {
	values := make([]string, 0, len(AppCodes))
	for _, value := range AppCodes {
		values = append(values, value)
	}
	return values
}
