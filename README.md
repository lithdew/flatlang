# flatlang

[![MIT License](https://img.shields.io/apm/l/atomic-design-ui.svg?)](LICENSE)
[![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=flat-square)](https://pkg.go.dev/github.com/lithdew/flatlang)
[![Discord Chat](https://img.shields.io/discord/697002823123992617)](https://discord.gg/HZEbkeQ)

**flatlang** is an embeddable configuration language that is made for configuring large-scale codegen tools, automation tools, code bases, and applications offline/in realtime.

**flatlang** is heavily inspired on being scriptable and embeddable by [Google Starlark](https://github.com/bazelbuild/starlark), and being suited for code generation and data validation by [CUE](https://cuelang.org/).

** This project is still a heavy WIP! Ping me on Discord if you'd like to contribute or know more.

## Goals

1. **Be human-friendly.** Flatlang discourages having to type the same variable names and method names over and over again while laying out your configuration. This is done by having flatlang represent configuration code as chains of method calls linked together using pipe syntax ('>'). This forces configuration in flatlang to naturally be imperative, such that it naturally reads/can be written from top to bottom.

2. **Be easy to type.** Defining a variable, calling a method with a set of parameters, editing some constants, and refactoring out some code should only require a few key strokes. The goal is to make flatlang a language that is convenient and fun to write in irrespective of your coding environment (terminal, text editor, IDE, etc.). There are zero keywords you have to remember in flatlang.

3. **No recursion.** Recursion burdens the reader into having to follow and pick up base conditions and recursive statements; something which obfuscates configuration code and dilutes readability.

4. **No recursive structures.** Maps and lists in flatlang are only allowed to be composed of literals. Most likely whenever you see the need to recursively compose a map or list, you'd be better off implementing new builtin methods into flatlang to handle your use case, or reconsider the parameters of methods you implement into flatlang.

5. **Types are values.** Options and values may be constrained using sum types, which may very easily be used to configure and represent complicated constraints on values for your programs.

## Design

The lexer and parser design is heavily inspired by both Go's [`text/scanner`](https://golang.org/pkg/text/scanner/) package, and [`BurntSushi/toml`](https://github.com/BurntSushi/toml). The parser is LL(1), and is a recursive-descent parser which uses Pratt parsing for parsing expressions.

## Testing

Both the lexer and parser have manually been fuzz-tested using [go-fuzz](https://github.com/dvyukov/go-fuzz) to check that erroneous inputs are rejected, and that panics won't occur. Both the lexer and parser come accompanied with unit test case suites, and sum up to a total of roughly 1500 lines of code.

You can manually test the lexer/parser for flatlang by running the command below. The command below assumes you have Go installed:

```
$ go run github.com/lithdew/flatlang/cmd/parser
```

## Example

```
// Database URL.

db = 'sqlite://:memory:';

// SQL query handlers.

all_posts = sql db 'select * from posts limit :limit offset :offset' 'posts';
find_post = sql db 'select * from posts where id = :id' 'post';
new_post  = sql db 'insert into posts (author, content) values (:author, :content) returning id' 'id';

// An example expanded-form of a SQL query.

expanded_form = sql {db: db, query: 'select * from posts limit :limit offset :offset', set: 'posts'};

// Pagination helper.

paginate =
    > default {offset: 0, limit: 1024}
    > require {offset: >=0, limit: >=0 & <=1024}
    > set ['limit', 'offset'];

// GET "/posts" returns at most 1024 posts.

get "/posts"
	> paginate
	> find_post
	> encode;

// GET "/posts/:id" returns a post with a specific id = ":id".

get "/posts/:id"
    > find_post
    > encode;

// POST "/posts" creates a new post. Author and content must be provided.

post {path: "/posts", max_body_size: 128}
	> decode
	> new_post
	> encode;

// A simple health check endpoint that returns {status: "ok"}.

get "/health" > set {status: "ok"} > encode;
```

