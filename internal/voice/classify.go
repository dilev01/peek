package voice

import (
	"regexp"
	"strconv"
	"strings"
)

type CommandType int

const (
	CmdAnnotation CommandType = iota
	CmdGotoLine
	CmdNextPage
	CmdPrevPage
	CmdGotoTop
	CmdGotoBottom
	CmdSearch
	CmdNextHeading
	CmdPrevHeading
)

type Command struct {
	Type CommandType
	Line int
	Text string
}

var patterns = []struct {
	re      *regexp.Regexp
	cmdType CommandType
}{
	{regexp.MustCompile(`(?i)^go to line (\d+)$`), CmdGotoLine},
	{regexp.MustCompile(`(?i)^line (\d+)$`), CmdGotoLine},
	{regexp.MustCompile(`(?i)^next( page)?$`), CmdNextPage},
	{regexp.MustCompile(`(?i)^back$`), CmdPrevPage},
	{regexp.MustCompile(`(?i)^previous( page)?$`), CmdPrevPage},
	{regexp.MustCompile(`(?i)^top$`), CmdGotoTop},
	{regexp.MustCompile(`(?i)^bottom$`), CmdGotoBottom},
	{regexp.MustCompile(`(?i)^find (.+)$`), CmdSearch},
	{regexp.MustCompile(`(?i)^next heading$`), CmdNextHeading},
	{regexp.MustCompile(`(?i)^previous heading$`), CmdPrevHeading},
}

func Classify(text string) Command {
	text = strings.TrimSpace(text)
	text = strings.TrimSuffix(text, ".")

	for _, p := range patterns {
		matches := p.re.FindStringSubmatch(text)
		if matches != nil {
			cmd := Command{Type: p.cmdType, Text: text}
			switch p.cmdType {
			case CmdGotoLine:
				if len(matches) > 1 {
					cmd.Line, _ = strconv.Atoi(matches[1])
				}
			case CmdSearch:
				if len(matches) > 1 {
					cmd.Text = matches[1]
				}
			}
			return cmd
		}
	}

	return Command{Type: CmdAnnotation, Text: text}
}
