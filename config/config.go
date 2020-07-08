package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
)

var (
	Terminals = map[string]*Terminal{
		"termite":        NewTerminal("termite", "-t", "-e"),
		"alacritty":      NewTerminal("alacritty", "-t", "-e"),
		"gnome-terminal": NewTerminal("gnome-terminal", "-t", "--"),
	}
)

type Terminal struct {
	Command     string
	ExecOpt     string
	ExecCommand string
	TitleOpt    string
	Title       string
}

func NewTerminal(command, titleOpt, execOpt string) *Terminal {
	if execOpt == "" {
		execOpt = "-e"
	}
	if titleOpt == "" {
		titleOpt = "-T"
	}
	term := &Terminal{
		Command:  command,
		TitleOpt: titleOpt,
		ExecOpt:  execOpt,
	}
	return term
}

func (term *Terminal) ToCmd() (*exec.Cmd, error) {
	if term.Title == "" {
		return nil, fmt.Errorf("title is empty")
	}
	if term.ExecCommand == "" {
		return nil, fmt.Errorf("execCmd is empty")
	}
	cmd := exec.Command(term.Command, term.TitleOpt, term.Title, term.ExecOpt, term.ExecCommand)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()
	return cmd, nil
}

func (term *Terminal) String() string {
	return fmt.Sprintf("%s %s %s %s %s", term.Command, term.TitleOpt, term.Title, term.ExecOpt, term.ExecCommand)
}

type Config struct {
	Term  string  `json:"term"`
	Ratio float32 `json:"ratio"`
	// Pos defines terminal pop out position, support 'top|bottom'
	Pos string `json:"pos"`
}

func ParseConfig(content []byte) (*Config, error) {
	var conf *Config
	if content != nil {
		conf = new(Config)
		err := json.Unmarshal(content, conf)
		if err != nil {
			return nil, fmt.Errorf("unmarshal config json: %v", err)
		}
	}

	// set default config
	if conf.Term == "" {
		conf.Term = "termite"
	}
	if conf.Pos == "" {
		conf.Pos = "top"
	}
	if conf.Ratio == 0 {
		conf.Ratio = 0.35
	}

	// validate config
	_, ok := Terminals[conf.Term]
	if !ok {
		return nil, fmt.Errorf("not support term: %s", conf.Term)
	}
	if conf.Pos != "top" && conf.Pos != "bottom" {
		return nil, fmt.Errorf("not support pos: %s", conf.Pos)
	}
	if conf.Ratio <= 0 {
		return nil, fmt.Errorf("ratio %f <= 0", conf.Ratio)
	}

	return conf, nil
}

func getConfigFile() string {
	homeDir, _ := os.UserHomeDir()
	return path.Join(homeDir, ".config", "i3-go-quickterm", "config.json")
}

func getConfigData() ([]byte, error) {
	configFile := getConfigFile()
	return ioutil.ReadFile(configFile)
}

func GetConfig() (*Config, error) {
	content, err := getConfigData()
	if err != nil {
		log.Printf("read user config: %v", err)
	}
	return ParseConfig(content)
}