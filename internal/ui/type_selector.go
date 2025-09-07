package ui

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/Hangell/gommit/platform"
)

const (
	// Emojis Unicode
	IconWIP      = "🚧"
	IconPrune    = "🔥"
	IconFeat     = "💡"
	IconFix      = "🐛"
	IconDocs     = "📝"
	IconStyle    = "💅"
	IconRefactor = "🎨"
	IconPerf     = "⚡"
	IconTest     = "✅"
	IconBuild    = "🔧"
	IconCI       = "🤖"
	IconChore    = "📦"
	IconRevert   = "⏪"
)

const (
	// Fallback ASCII para terminais sem suporte Unicode
	FAWIP      = "[WIP]"
	FAPrune    = "[-]"
	FAFeat     = "[+]"
	FAFix      = "[fix]"
	FADocs     = "[doc]"
	FAStyle    = "[fmt]"
	FARefactor = "[ref]"
	FAPerf     = "[perf]"
	FATest     = "[test]"
	FABuild    = "[build]"
	FACI       = "[ci]"
	FAChore    = "[chore]"
	FARevert   = "[revert]"
)

// ANSI (cores simples) — respeita NO_COLOR
var (
	useColor = func() bool {
		return os.Getenv("NO_COLOR") == "" // qualquer valor desativa
	}

	clrDim = func(s string) string {
		if !useColor() {
			return s
		}
		return "\x1b[2m" + s + "\x1b[0m"
	}
	clrType = func(s string) string {
		if !useColor() {
			return s
		}
		return "\x1b[36m" + s + "\x1b[0m"
	} // cyan
	clrIcon = func(s string) string {
		if !useColor() {
			return s
		}
		return "\x1b[33m" + s + "\x1b[0m"
	} // yellow
	clrError = func(s string) string {
		if !useColor() {
			return s
		}
		return "\x1b[31m" + s + "\x1b[0m"
	} // red
)

type CommitType struct {
	Key         string
	Icon        string
	Description string
}

var commitTypes = []CommitType{
	{"WIP", icon(IconWIP, FAWIP), "Work in progress"},
	{"feat", icon(IconFeat, FAFeat), "A new feature"},
	{"fix", icon(IconFix, FAFix), "Fixing a bug"},
	{"chore", icon(IconChore, FAChore), "Updating Dependencies, Deployments, Configuration files"},
	{"refactor", icon(IconRefactor, FARefactor), "Improving structure / format of the code"},
	{"prune", icon(IconPrune, FAPrune), "Removing code or files"},
	{"docs", icon(IconDocs, FADocs), "Writing documentation"},
	{"perf", icon(IconPerf, FAPerf), "Improving performance"},
	{"test", icon(IconTest, FATest), "Adding tests"},
	{"build", icon(IconBuild, FABuild), "Changes to build system or dependencies"},
	{"ci", icon(IconCI, FACI), "Changes to CI/CD configuration"},
	{"style", icon(IconStyle, FAStyle), "Changes that do not affect the meaning of the code"},
	{"revert", icon(IconRevert, FARevert), "Revert to a commit"},
}

func SelectCommitType() (CommitType, error) {
	// Se stdin não é TTY (pipe), ainda vamos tentar ler uma linha.
	reader := bufio.NewReader(os.Stdin)

	for {
		clearScreen()
		displayCommitTypes()

		fmt.Print("\n? Select the type of change you're committing: ")
		input, err := reader.ReadString('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				// EOF: sai com erro “cancelado”
				return CommitType{}, fmt.Errorf("selection aborted (EOF)")
			}
			return CommitType{}, fmt.Errorf("error reading input: %w", err)
		}

		input = strings.TrimSpace(input)

		// Comandos de saída
		switch strings.ToLower(input) {
		case "q", "quit", "exit":
			return CommitType{}, fmt.Errorf("selection aborted by user")
		}

		// Seleção por número
		if num, err := strconv.Atoi(input); err == nil && num > 0 && num <= len(commitTypes) {
			return commitTypes[num-1], nil
		}

		// Seleção por nome exato
		for _, ct := range commitTypes {
			if strings.EqualFold(input, ct.Key) {
				return ct, nil
			}
		}

		// Busca parcial
		var matches []CommitType
		inputLower := strings.ToLower(input)
		for _, ct := range commitTypes {
			if strings.Contains(strings.ToLower(ct.Key), inputLower) ||
				strings.Contains(strings.ToLower(ct.Description), inputLower) {
				matches = append(matches, ct)
			}
		}

		if len(matches) == 1 {
			return matches[0], nil
		}

		if len(matches) > 1 {
			fmt.Printf("\n%s\n", clrDim("Multiple matches found. Please be more specific:"))
			// Mostrar com tabwriter pra manter alinhado
			w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			for i, m := range matches {
				fmt.Fprintf(w, "  %2d)\t%s\t%s\t- %s\n", i+1, clrIcon(m.Icon), clrType(m.Key), m.Description)
			}
			w.Flush()
			fmt.Print(clrDim("Press Enter to continue..."))
			reader.ReadString('\n')
			continue
		}

		fmt.Printf("\n%s '%s'. %s\n", clrError("Invalid selection:"), input, clrDim("Please try again."))
		fmt.Print(clrDim("Press Enter to continue..."))
		reader.ReadString('\n')
	}
}

func displayCommitTypes() {
	title := "Select the type of change that you're committing: (Use number, type name, or search)"
	fmt.Println(clrDim(title))
	fmt.Println()

	w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
	for i, ct := range commitTypes {
		fmt.Fprintf(w, "  %2d)\t%s\t%-10s\t%s\n", i+1, clrIcon(ct.Icon), clrType(ct.Key), ct.Description)
	}
	w.Flush()

	fmt.Println()
	fmt.Println(clrDim("(You can type: number, commit type name, or search term — or 'q' to quit)"))
}

func clearScreen() {
	// ANSI VT100 (já habilitado no Windows pelo pacote platform)
	fmt.Print("\033[H\033[2J")
}

func icon(emoji, fallback string) string {
	if platform.SupportsUnicode() {
		return emoji
	}
	return fallback
}
