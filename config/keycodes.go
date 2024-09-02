package config

import "github.com/BurntSushi/xgb/xproto"

var AppCodes = map[xproto.Keycode]string{
	75: "vivaldi-stable",
	76: "emacs",
	95: "obsidian",
	96: "",
}

func GetAppCodesValues() []string {
    values := make([]string, 0, len(AppCodes))
    for _, value := range AppCodes {
        values = append(values, value)
    }
    return values
}
