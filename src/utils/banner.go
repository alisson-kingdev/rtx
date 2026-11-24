package utils

// Banner returns the RTX ASCII art banner with version info
func Banner() string {
	version := GetVersion()
	lines := []string{
		"    ____        _                    _        ",
		"   / __ \\____ _   _ __ ___  _ __ ___ | | __    ",
		"  / / / / _ \\ | | | '_ ` _ \\| '_ ` _ \\| |/ /    ",
		" / /_/  (_)  |_| | | | | | | | | | |   <     ",
		" \\___\\___/\\__,_\\__,_| |_| |_| |_| |_| |_|\\_\\    ",
		"                                            ",
		"         RTX - IP ROTATOR v" + version + "           ",
		"         Author: MRX                       ",
	}

	result := "\n"
	for i, line := range lines {
		switch {
		case i < 5:
			result += cyan(line) + "\n"
		case i == 5:
			result += magenta(line) + "\n"
		case i == 6:
			result += bold(yellow(line)) + "\n"
		default:
			result += blue(line) + "\n"
		}
	}
	return result
}
