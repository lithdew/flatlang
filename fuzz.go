// +build gofuzz

package flatlang

func Fuzz(data []byte) int {
	parser := NewParser(NewLexer(string(data)))

	_, err := parser.Parse()
	if err != nil {
		return 0
	}

	return 1
}
