package lipgloss

import "strings"

type Color string

type Style struct{}

type Border struct{}

type Position int

const (
	Top Position = iota
	Left
	Center
)

func NewStyle() Style                                { return Style{} }
func (Style) Bold(bool) Style                        { return Style{} }
func (Style) Foreground(Color) Style                 { return Style{} }
func (Style) Border(Border) Style                    { return Style{} }
func (Style) BorderForeground(Color) Style           { return Style{} }
func (Style) Padding(int, int) Style                 { return Style{} }
func (Style) Width(int) Style                        { return Style{} }
func (Style) Render(v string) string                 { return v }
func RoundedBorder() Border                          { return Border{} }
func ThickBorder() Border                            { return Border{} }
func Place(_, _ int, _, _ Position, s string) string { return s }
func JoinHorizontal(_ Position, items ...string) string {
	return strings.Join(items, "")
}
func JoinVertical(_ Position, items ...string) string {
	return strings.Join(items, "\n")
}
