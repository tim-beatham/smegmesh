package lib

import "regexp"

func MatchCaptureGroup(pattern, payload string) map[string]string {
	patterns := make(map[string]string)

	expr := regexp.MustCompile(pattern)

	match := expr.FindStringSubmatch(payload)

	for i, name := range expr.SubexpNames() {
		if i != 0 && name != "" {
			patterns[name] = match[i]
		}
	}

	return patterns
}
