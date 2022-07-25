package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/chzyer/readline"
	"github.com/wongak/sl/sl"
)

type (
	app struct {
		rl *readline.Instance
	}
)

const prompt = "sl> "

func main() {
	os.Exit(runWithCode())
}

func runWithCode() int {
	app := &app{}
	var err error
	app.rl, err = readline.New(prompt)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error initializing readline: %v\n\n", err)
		return 1
	}
	defer app.rl.Close()

	for {
		line, err := app.rl.Readline()
		if err != nil {
			fmt.Fprintf(app.rl.Config.Stderr, "error on input: %v\n\n", err)
			return 1
		}
		rd := strings.NewReader(line)
		p := sl.NewParser(rd)
		n, err := p.Parse()
		if err != nil {
			fmt.Fprintf(app.rl.Config.Stderr, "%v\n", err)
			continue
		}
		if n != nil {
			fmt.Fprintf(app.rl.Config.Stdout, "%s\n", n)
			continue
		}
	}
	return 0
}
