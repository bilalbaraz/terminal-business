package components

import "fmt"

func ErrorScreen(title, msg string) string {
	return fmt.Sprintf("%s\n\n%s\n\nPress b to return to menu or q to quit.", title, msg)
}
