package shell

import (
	"errors"
	"fmt"
	"github/JustGopher/Gotaxy/internal/tunnel/serverCore/global"
	"io"
	"log"
	"strconv"
	"strings"

	"github.com/chzyer/readline"
)

type Shell struct {
	Rl       *readline.Instance
	commands map[string]func(args []string)
}

func New() *Shell {
	return &Shell{
		commands: make(map[string]func(args []string)),
	}
}

func (s *Shell) Register(cmd string, handler func(args []string)) {
	s.commands[cmd] = handler
}

func (s *Shell) Run() {
	completer := s.buildCompleter()
	rl, err := readline.NewEx(&readline.Config{
		Prompt:              "\033[31m»\033[0m ",
		HistoryFile:         "/tmp/readline.tmp",
		AutoComplete:        completer,
		InterruptPrompt:     "^C",
		EOFPrompt:           "exit",
		HistorySearchFold:   true,
		FuncFilterInputRune: filterInput,
	})
	if err != nil {
		panic(err)
	}
	defer func(rl *readline.Instance) {
		err := rl.Close()
		if err != nil {
			return
		}
	}(rl)
	s.Rl = rl
	rl.CaptureExitSignal()
	log.SetOutput(rl.Stderr())

	setPasswordCfg := rl.GenPasswordConfig()
	setPasswordCfg.SetListener(func(line []rune, pos int, key rune) (newLine []rune, newPos int, ok bool) {
		rl.SetPrompt(fmt.Sprintf("Enter password(%v): ", len(line)))
		rl.Refresh()
		return nil, 0, false
	})

	for {
		line, err := rl.Readline()
		if errors.Is(err, readline.ErrInterrupt) {
			if len(line) == 0 {
				break
			}
			continue
		} else if err == io.EOF {
			break
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		isExit := false

		// 固定命令
		switch {
		case strings.HasPrefix(line, "mode "):
			switch line[5:] {
			case "vi":
				rl.SetVimMode(true)
			case "emacs":
				rl.SetVimMode(false)
			default:
				fmt.Println("invalid mode:", line[5:])
			}
			continue
		case line == "mode":
			if rl.IsVimMode() {
				fmt.Println("current mode: vim")
			} else {
				fmt.Println("current mode: emacs")
			}
			continue
		case line == "help":
			s.usage(rl.Stderr())
			continue
		case line == "exit":
			global.Cancel()
			isExit = true
		}
		if isExit {
			break
		}
		// 自定义命令分发
		parts := strings.Fields(line)
		cmd, args := parts[0], parts[1:]

		if handler, ok := s.commands[cmd]; ok {
			handler(args)
		} else {
			log.Println("Unknown command:", strconv.Quote(line))
		}
	}
}

func (s *Shell) buildCompleter() *readline.PrefixCompleter {
	items := []readline.PrefixCompleterInterface{
		readline.PcItem("mode", readline.PcItem("vi"), readline.PcItem("emacs")),
		readline.PcItem("exit"),
		readline.PcItem("help"),
	}
	for name := range s.commands {
		items = append(items, readline.PcItem(name))
	}
	return readline.NewPrefixCompleter(items...)
}

func (s *Shell) usage(w io.Writer) {
	_, err := io.WriteString(w, "commands:\n")
	if err != nil {
		return
	}
	_, err = io.WriteString(w, s.buildCompleter().Tree("    "))
	if err != nil {
		return
	}
}

func filterInput(r rune) (rune, bool) {
	switch r {
	case readline.CharCtrlZ:
		return r, false
	}
	return r, true
}
