package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"time"

	"go.i3wm.org/i3/v4"
)

const (
	MarkQuickTermPattern = "quick_term_go.*"
	MarkQuickTerm        = "quick_term_go_%s"
)

var (
	InPlace *bool = flag.Bool("i", false, "Run in place")
)

type Terminal struct {
	Command     string
	ExecOpt     string
	ExecCommand string
	TitleOpt    string
	Title       string
}

func NewTerminal(command, titleOpt, title, execOpt, execCmd string) (*Terminal, error) {
	if execOpt == "" {
		execOpt = "-e"
	}
	if titleOpt == "" {
		titleOpt = "-T"
	}
	if title == "" {
		return nil, fmt.Errorf("title is empty")
	}
	if execCmd == "" {
		return nil, fmt.Errorf("execCmd is empty")
	}
	term := &Terminal{
		Command:     command,
		TitleOpt:    titleOpt,
		Title:       title,
		ExecOpt:     execOpt,
		ExecCommand: execCmd,
	}
	return term, nil
}

func (term *Terminal) ToCmd() *exec.Cmd {
	cmd := exec.Command(term.Command, term.TitleOpt, term.Title, term.ExecOpt, term.ExecCommand)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()
	return cmd
}

func (term *Terminal) String() string {
	return fmt.Sprintf("%s %s %s %s %s", term.Command, term.TitleOpt, term.Title, term.ExecOpt, term.ExecCommand)
}

func initSway() {
	// ref: https://github.com/i3/go-i3/pull/5
	i3.SocketPathHook = func() (string, error) {
		out, err := exec.Command("sway", "--get-socketpath").CombinedOutput()
		if err != nil {
			return "", fmt.Errorf("getting sway socketpath: %v (output: %s)", err, out)
		}
		return string(out), nil
	}

	i3.IsRunningHook = func() bool {
		out, err := exec.Command("pgrep", "-c", "sway\\$").CombinedOutput()
		if err != nil {
			log.Printf("sway running: %v (output: %s)", err, out)
		}
		return bytes.Compare(out, []byte("1")) == 0
	}
}

func FindMarkedTerm(shell string) *i3.Node {
	tree, err := i3.GetTree()
	if err != nil {
		log.Fatal(err)
	}
	shellMark := regexp.MustCompile(fmt.Sprintf(MarkQuickTerm, shell))
	term := tree.Root.FindChild(func(node *i3.Node) bool {
		for _, mark := range node.Marks {
			if shellMark.Match([]byte(mark)) {
				return true
			}
		}
		return false
	})
	return term
}

func RunTerm(shell string) error {
	log.Printf("Run new terminal")
	term, err := NewTerminal(
		"termite",
		"-t", "shell - sway-quickterm",
		"-e", fmt.Sprintf("%s -i", os.Args[0]))
	if err != nil {
		return fmt.Errorf("new terminal %v", err)
	}
	cmd := term.ToCmd()
	return cmd.Run()
}

func RestoreTerm(term *i3.Node, shell string) error {
	log.Printf("Restore terminal %d", term.ID)
	ws, err := GetCurrentWorkspace()
	if err != nil {
		return fmt.Errorf("get current focused workspace: %v", err)
	}
	nodeWs, err := FindNodeWorkspace(term)
	if err != nil {
		return fmt.Errorf("find terminal workspace: %v", err)
	}
	if nodeWs.Name != ws.Name {
		log.Printf("Pop hide terminal, exists: %s != current: %s", nodeWs.Name, ws.Name)
		shellMark := fmt.Sprintf(MarkQuickTerm, shell)
		if err := PopIt(shellMark, "top", 0.25); err != nil {
			return fmt.Errorf("try retore term: %v", err)
		}
	} else {
		if err := MoveBack(fmt.Sprintf("[con_id=%d]", term.ID)); err != nil {
			return fmt.Errorf("move term back: %v", err)
		}
	}
	return nil
}

func ToggleTerm(shell string) error {
	term := FindMarkedTerm(shell)
	if term == nil {
		return RunTerm(shell)
	}
	return RestoreTerm(term, shell)
}

func LaunchInplace(shell string) error {
	shellMark := fmt.Sprintf(MarkQuickTerm, shell)
	i3.RunCommand(fmt.Sprintf("mark %s", shellMark))
	if err := MoveBack(fmt.Sprintf("[con_mark=%s]", shellMark)); err != nil {
		return fmt.Errorf("moveback: %v", err)
	}
	if err := PopIt(shellMark, "top", 0.25); err != nil {
		return fmt.Errorf("popit %v", err)
	}
	cmd := exec.Command("zsh")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()

	return cmd.Run()
}

func main() {
	flag.Parse()
	initSway()
	if *InPlace {
		time.Sleep(time.Second)
		if err := LaunchInplace("shell"); err != nil {
			ExitErr(fmt.Errorf("launch in place: %v", err))
		}
		return
	}
	if err := ToggleTerm("shell"); err != nil {
		ExitErr(fmt.Errorf("toggle term: %v", err))
		return
	}
}
