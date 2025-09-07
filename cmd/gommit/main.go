// SPDX-License-Identifier: GPL-3.0-only

package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/Hangell/gommit/internal/commit"
	"github.com/Hangell/gommit/internal/git"
	"github.com/Hangell/gommit/internal/install"
	"github.com/Hangell/gommit/internal/ui"
)

var version = "dev"

func main() {
	fs := flag.NewFlagSet("gommit", flag.ContinueOnError)
	showVersion := fs.Bool("version", false, "print version and exit")
	doInstall := fs.Bool("install", false, "install or update gommit into your user bin and exit")
	dryRun := fs.Bool("dry-run", false, "print the message and exit (do not run git commit)")

	typeFlag := fs.String("type", "", "commit type (feat, fix, docs, style, refactor, perf, test, build, ci, chore, revert; extras: WIP, prune)")
	scopeFlag := fs.String("scope", "", "optional scope (e.g. ui, api, editor)")
	subjectFlag := fs.String("subject", "", "commit subject (imperative, <=72 chars)")
	bodyFlag := fs.String("body", "", "commit body (use \\n for new lines if provided by flag)")
	footerFlag := fs.String("footer", "", "commit footer (e.g. Closes #123; BREAKING CHANGE: ...)")
	asEditor := fs.Bool("as-editor", false, "run as Git editor (writes COMMIT_EDITMSG)")

	allowEmpty := fs.Bool("allow-empty", false, "allow an empty commit")
	amend := fs.Bool("amend", false, "amend the previous commit")
	noVerify := fs.Bool("no-verify", false, "bypass pre-commit and commit-msg hooks")
	signoff := fs.Bool("signoff", false, "add Signed-off-by trailer")
	autoStage := fs.Bool("auto-stage", true, "auto stage all changes (git add -A) when nothing staged")
	showStatus := fs.Bool("show-status", true, "print staged changes summary before committing")

	fs.SetOutput(os.Stderr)
	if err := fs.Parse(os.Args[1:]); err != nil {
		os.Exit(2)
	}

	if *showVersion {
		fmt.Println("gommit version", version)
		return
	}

	// ===== INSTALAÇÃO =====
	if *doInstall {
		res, err := install.InstallSelf(version)
		if err != nil {
			log.Fatalf("install failed: %v", err)
		}
		fmt.Println(res.Message)
		return
	}

	// ===== editor mode =====
	if *asEditor {
		if fs.NArg() < 1 {
			log.Fatal("editor mode: missing COMMIT_EDITMSG path")
		}
		path := fs.Arg(0)
		msg := buildOrPromptMessage(*typeFlag, *scopeFlag, *subjectFlag, *bodyFlag, *footerFlag, true)
		if err := git.WriteCommitEditMsg(path, msg); err != nil {
			log.Fatalf("failed to write commit message: %v", err)
		}
		return
	}

	// ===== precisa ser repo git =====
	if !git.InRepo() {
		log.Fatal("not a git repository (or any of the parent directories)")
	}

	// ===== staged / auto-stage / amend =====
	if *amend {
		if subj, err := git.LastCommitSubject(); err == nil && subj != "" {
			fmt.Printf("Amend mode: last commit → %s\n", subj)
		} else {
			fmt.Println("Amend mode: last commit will be updated.")
		}
	}
	if !*allowEmpty && !*amend {
		staged, err := git.HasStagedChanges()
		if err != nil {
			log.Fatalf("failed to check staged changes: %v", err)
		}
		if !staged {
			dirty, err := git.WorkingTreeDirty()
			if err != nil {
				log.Fatalf("failed to check working tree: %v", err)
			}
			if dirty {
				if *autoStage {
					fmt.Println("No staged changes detected. Running: git add -A")
					if err := git.StageAll(); err != nil {
						log.Fatalf("git add -A failed: %v", err)
					}
					if *showStatus {
						if sum, err := git.StagedSummary(); err == nil && strings.TrimSpace(sum) != "" {
							fmt.Println("\nStaged changes:\n" + strings.TrimRight(sum, "\n"))
						}
					}
				} else {
					log.Fatal("no staged changes. Run 'git add .' or pass --auto-stage")
				}
			} else {
				log.Fatal("nothing to commit. Working tree clean (use --allow-empty or --amend)")
			}
		} else if *showStatus {
			if sum, err := git.StagedSummary(); err == nil && strings.TrimSpace(sum) != "" {
				fmt.Println("Staged changes:\n" + strings.TrimRight(sum, "\n"))
			}
		}
	}

	// ===== wizard / mensagem =====
	msg := buildOrPromptMessage(*typeFlag, *scopeFlag, *subjectFlag, *bodyFlag, *footerFlag, false)

	if *dryRun {
		fmt.Println()
		fmt.Println("────────────────────────────────────────────────")
		fmt.Println("Commit message preview:")
		fmt.Println("────────────────────────────────────────────────")
		fmt.Println(msg)
		return
	}

	// ===== commit =====
	if err := git.CommitWithMessage(msg, git.Options{
		AllowEmpty: *allowEmpty,
		Amend:      *amend,
		NoVerify:   *noVerify,
		Signoff:    *signoff,
	}); err != nil {
		log.Fatalf("git commit failed: %v", err)
	}
}

// --- wizard / montagem ---------------------------------------------------

func buildOrPromptMessage(typeFlag, scopeFlag, subjectFlag, bodyFlag, footerFlag string, quiet bool) string {
	// TYPE
	var selected ui.CommitType
	if typeFlag == "" {
		var err error
		selected, err = ui.SelectCommitType()
		if err != nil {
			log.Fatalf("error selecting commit type: %v", err)
		}
	} else {
		if t, ok := ui.FindType(typeFlag); ok {
			selected = t
		} else {
			log.Fatalf("invalid --type '%s'", typeFlag)
		}
	}

	// SCOPE
	scope := scopeFlag
	if scope == "" && !quiet {
		scope = promptLine("Scope (optional, Enter to skip): ")
	}

	// SUBJECT
	subject := subjectFlag
	if subject == "" {
		subject = promptLine("Subject (required, imperative, <=72 chars): ")
	}
	subject = strings.TrimSpace(subject)
	if subject == "" {
		log.Fatal("subject is required")
	}
	if len([]rune(subject)) > 72 {
		log.Fatal("subject must be <= 72 chars")
	}

	// BODY
	body := unescapeNewlines(bodyFlag)
	if body == "" && !quiet {
		fmt.Println("Body (optional, multiline). Finish with a single '.' on a new line, or press Enter twice:")
		body = readMultiline()
	}

	// Breaking changes?
	breaking := false
	breakingDesc := ""
	if !quiet {
		breaking = promptYesNo("Are there any breaking changes?", false)
		if breaking {
			breakingDesc = promptLine("Describe the breaking change: ")
		}
	}

	// Issues (Closes / Refs)?
	var closes, refs []string
	if !quiet && promptYesNo("Does this change affect any open issues?", false) {
		c := promptLine("Issues to close (comma, ex: 12,45) — Enter to skip: ")
		r := promptLine("Issues to reference (comma, ex: 7,9) — Enter to skip: ")
		closes = splitCSVNums(c)
		refs = splitCSVNums(r)
	}
	issueLines := buildIssueFooter(closes, refs)

	// FOOTER manual (se passado por flag)
	footer := strings.TrimSpace(unescapeNewlines(footerFlag))

	// Montagem final
	headerType := selected.Key
	if breaking && !strings.HasSuffix(headerType, "!") {
		headerType += "!"
	}

	pre := headerType
	if scope != "" {
		pre += "(" + strings.TrimSpace(scope) + ")"
	}
	if ico := ui.EmojiFor(selected.Key); ico != "" {
		pre += ": " + ico
	}
	header := pre + " " + subject

	var b strings.Builder
	b.WriteString(header)

	if body != "" {
		b.WriteString("\n\n")
		b.WriteString(strings.TrimRight(body, "\n"))
	}

	if breaking && strings.TrimSpace(breakingDesc) != "" {
		b.WriteString("\n\n")
		b.WriteString("BREAKING CHANGE: ")
		b.WriteString(strings.TrimSpace(breakingDesc))
	}

	if footer != "" {
		b.WriteString("\n\n")
		b.WriteString(footer)
	}

	if issueLines != "" {
		b.WriteString("\n\n")
		b.WriteString(issueLines)
	}

	// valida
	typeForValidation := selected.Key
	if err := (commit.Message{
		Type:    typeForValidation,
		Scope:   scope,
		Subject: subject,
		Body:    body,
		Footer:  strings.TrimSpace(strings.Join([]string{footer, issueLines}, "\n\n")),
	}).Validate(); err != nil {
		log.Fatalf("invalid commit message: %v", err)
	}

	return b.String()
}

func promptLine(label string) string {
	fmt.Print(label)
	r := bufio.NewReader(os.Stdin)
	s, _ := r.ReadString('\n')
	return strings.TrimSpace(s)
}

func readMultiline() string {
	r := bufio.NewScanner(os.Stdin)
	r.Buffer(make([]byte, 0, 64*1024), 1_000_000)

	var lines []string
	emptyStreak := 0

	for {
		if !r.Scan() {
			break // EOF
		}
		line := r.Text()
		trim := strings.TrimSpace(line)

		if trim == "." {
			break // ponto sozinho
		}
		if trim == "" {
			emptyStreak++
			if emptyStreak >= 2 {
				break // enter duas vezes
			}
		} else {
			emptyStreak = 0
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

func unescapeNewlines(s string) string {
	return strings.ReplaceAll(s, `\n`, "\n")
}

// helpers cz-like

func promptYesNo(label string, defYes bool) bool {
	suf := "y/N"
	if defYes {
		suf = "Y/n"
	}
	fmt.Printf("%s (%s): ", label, suf)
	r := bufio.NewReader(os.Stdin)
	in, _ := r.ReadString('\n')
	in = strings.TrimSpace(strings.ToLower(in))
	if in == "" {
		return defYes
	}
	return in == "y" || in == "yes" || in == "s"
}

func splitCSVNums(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	var out []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if !strings.HasPrefix(p, "#") {
			p = "#" + p
		}
		out = append(out, p)
	}
	return out
}

func buildIssueFooter(closes, refs []string) string {
	var lines []string
	for _, id := range closes {
		lines = append(lines, "Closes "+id)
	}
	for _, id := range refs {
		lines = append(lines, "Refs "+id)
	}
	return strings.Join(lines, "\n")
}
