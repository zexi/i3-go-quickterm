package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"time"

	"go.i3wm.org/i3/v4"

	"github.com/zexi/i3-go-quickterm/config"
)

const (
	MarkQuickTermPattern = "quick_term_go.*"
	MarkQuickTerm        = "quick_term_go_%s"
)

var (
	InPlace *bool = flag.Bool("i", false, "Run in place")
)

func isSwayRunning() bool {
	_, err := exec.Command("pgrep", "-c", "sway$").CombinedOutput()
	if err != nil {
		return false
	}
	return true
}

func initEnv() {
	initSway := func() {
		// ref: https://github.com/i3/go-i3/pull/5
		i3.SocketPathHook = func() (string, error) {
			out, err := exec.Command("sway", "--get-socketpath").CombinedOutput()
			if err != nil {
				return "", fmt.Errorf("getting sway socketpath: %v (output: %s)", err, out)
			}
			return string(out), nil
		}

		i3.IsRunningHook = func() bool {
			return isSwayRunning()
		}
	}
	if isSwayRunning() {
		initSway()
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

func RunTerm(conf *config.Config, shell string) error {
	log.Printf("Run new terminal")
	term, ok := config.Terminals[conf.Term.Command]
	if !ok {
		return fmt.Errorf("not support config term: %s", conf.Term.Command)
	}
	term.Title = "shell - i3-go-quickterm"
	term.ExecCommand = fmt.Sprintf("%s -i", os.Args[0])
	cmd, err := term.ToCmd(conf.Term.ExtraArgs...)
	if err != nil {
		return fmt.Errorf("to terminal command: %v", err)
	}
	return cmd.Run()
}

func RestoreTerm(conf *config.Config, term *i3.Node, shell string) error {
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
		if err := PopIt(shellMark, conf.Pos, conf.Ratio); err != nil {
			return fmt.Errorf("try retore term: %v", err)
		}
	} else {
		if err := MoveBack(fmt.Sprintf("[con_id=%d]", term.ID)); err != nil {
			return fmt.Errorf("move term back: %v", err)
		}
	}
	return nil
}

func ToggleTerm(conf *config.Config, shell string) error {
	term := FindMarkedTerm(shell)
	if term == nil {
		return RunTerm(conf, shell)
	}
	return RestoreTerm(conf, term, shell)
}

func LaunchInplace(conf *config.Config, shell string) error {
	shellMark := fmt.Sprintf(MarkQuickTerm, shell)
	i3.RunCommand(fmt.Sprintf("mark %s", shellMark))
	if err := MoveBack(fmt.Sprintf("[con_mark=%s]", shellMark)); err != nil {
		return fmt.Errorf("moveback: %v", err)
	}
	if err := PopIt(shellMark, conf.Pos, conf.Ratio); err != nil {
		return fmt.Errorf("popit: %v", err)
	}

	// clear i3 sdk warning message
	clearCmd := exec.Command("clear")
	clearCmd.Stdout = os.Stdout
	clearCmd.Run()

	loginSh := os.Getenv("SHELL")
	cmd := exec.Command(loginSh)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()

	return cmd.Run()
}

func main() {
	flag.Parse()
	initEnv()

	conf, err := config.GetConfig()
	if err != nil {
		ExitErr(fmt.Errorf("get config: %v", err))
	}

	if *InPlace {
		// TODO: fix this sleep
		time.Sleep(time.Second)
		if err := LaunchInplace(conf, "shell"); err != nil {
			ExitErr(fmt.Errorf("launch in place: %v", err))
		}
		return
	}
	if err := ToggleTerm(conf, "shell"); err != nil {
		ExitErr(fmt.Errorf("toggle term: %v", err))
		return
	}
}
