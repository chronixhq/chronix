package sqlsyntax

import (
	"fmt"

	"github.com/antlr4-go/antlr/v4"
)

type collectingErrorListener struct {
	antlr.DefaultErrorListener
	errs []error
}

func newErrorListener() *collectingErrorListener {
	return &collectingErrorListener{errs: make([]error, 0, 2)}
}

func (l *collectingErrorListener) SyntaxError(_ antlr.Recognizer, _ interface{},
	line, column int, msg string, _ antlr.RecognitionException) {
	l.errs = append(l.errs, fmt.Errorf("line %d:%d %s", line, column, msg))
}

func (l *collectingErrorListener) Err() error {
	if len(l.errs) == 0 {
		return nil
	}
	// Return just the first error as a short reason
	return l.errs[0]
}
