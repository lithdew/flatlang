package main

import (
	"errors"
	"fmt"
	"github.com/chzyer/readline"
	"github.com/lithdew/flatlang"
	"io"
	"log"
	"strings"
)

func check(err error) {
	if err != nil {
		log.Panic(err)
	}
}

func wrap(fn func() error) { check(fn()) }

func main() {
	l, err := readline.NewEx(&readline.Config{Prompt: ">> "})
	check(err)
	defer wrap(l.Close)

	log.SetOutput(l.Stderr())

	for {
		line, err := l.Readline()
		if err != nil {
			if errors.Is(err, readline.ErrInterrupt) {
				break
			}
			if errors.Is(err, io.EOF) {
				break
			}
		}

		line = strings.TrimSpace(line)
		if len(line) == 0 {
			break
		}

		if !strings.HasSuffix(line, ";") {
			line += ";"
		}

		var (
			lx *flatlang.Lexer
			px *flatlang.Parser
		)

		lx, err = flatlang.Lex([]byte(line), "")
		if err == nil {
			px, err = flatlang.Parse(lx)
			if err == nil {
				fmt.Println(px.Format())
				continue
			}
		}

		log.Println(err)
	}
}
