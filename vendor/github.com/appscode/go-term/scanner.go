package term

import (
	"fmt"
	"strconv"
	"strings"
)

func Read(prompt string) string {
	for {
		Infoln(prompt)

		var ans string
		fmt.Scanln(&ans)
		ans = strings.TrimSpace(ans)
		if ans != "" {
			return ans
		}
	}
}

func Ask(prompt string, defaultYes bool) bool {
	for {
		if defaultYes {
			Infoln(prompt, "[Yes/no]")

			var ans string
			fmt.Scanln(&ans)
			ans = strings.TrimSpace(ans)
			if strings.EqualFold(ans, "y") || strings.EqualFold(ans, "yes") || ans == "" {
				return true
			}
			if strings.EqualFold(ans, "n") || strings.EqualFold(ans, "no") {
				return false
			}
		} else {
			Infoln(prompt, "[yes/No]")

			var ans string
			fmt.Scanln(&ans)
			ans = strings.TrimSpace(ans)
			if strings.EqualFold(ans, "y") || strings.EqualFold(ans, "yes") {
				return true
			}
			if strings.EqualFold(ans, "n") || strings.EqualFold(ans, "no") || ans == "" {
				return false
			}
		}
	}
}

func List(items []string) (int, string) {
	for i, item := range items {
		fmt.Printf("[%v] %v\n", i+1, item)
	}
	for {
		fmt.Print("Select option: ")
		var ans string
		fmt.Scanln(&ans)
		ans = strings.TrimSpace(ans)
		if i, err := strconv.Atoi(ans); err == nil {
			if i >= 1 && i <= len(items) {
				return i - 1, items[i-1]
			}
		}
	}
}
