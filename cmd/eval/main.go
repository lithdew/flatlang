package main

import (
	"errors"
	"github.com/chzyer/readline"
	"github.com/davecgh/go-spew/spew"
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
			lx  *flatlang.Lexer
			px  *flatlang.Parser
			res interface{}
		)

		lx, err = flatlang.Lex([]byte(line), "")
		if err == nil {
			px, err = flatlang.Parse(lx)
			if err == nil {
				res, err = flatlang.Eval(lx, px.Result)
				if err == nil {
					if len(res.([]interface{})) > 0 {
						spew.Dump(res)
					}
					continue
				}
			}
		}
		log.Println(err)
	}
}
