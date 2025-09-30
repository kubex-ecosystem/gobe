package components

import "fmt"

// NewLinkButtons creates formatted link buttons for embeds
func NewLinkButtons(links map[string]string) string {
	if len(links) == 0 {
		return ""
	}

	var buttons string
	for label, url := range links {
		if buttons != "" {
			buttons += " | "
		}
		buttons += fmt.Sprintf("[%s](%s)", label, url)
	}
	return buttons
}
