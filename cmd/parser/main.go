package main

import (
	"errors"
	"github.com/chzyer/readline"
	"github.com/lithdew/flatlang"
	"github.com/picmonkey/go-spew/spew"
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

		parser := flatlang.NewParser(flatlang.NewLexer(line))

		program, err := parser.Parse()
		if err == nil {
			spew.Dump(program)
			continue
		}

		log.Println(err)
	}
}
