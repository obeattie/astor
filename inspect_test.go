package astor

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

const parserFlags = parser.ParseComments | parser.AllErrors

func TestPrependingFuncName(t *testing.T) {
	var err error
	var inputSrc []byte
	var expectedOut []byte

	inputSrc, err = ioutil.ReadFile("test-samples/prepending-func-name.in.go")
	assert.NoError(t, err, "Error reading input")
	expectedOut, err = ioutil.ReadFile("test-samples/prepending-func-name.out.go")
	assert.NoError(t, err, "Error reading expected output")

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "prepending-func-name.go", inputSrc, parserFlags)
	assert.NoError(t, err, "Error parsing input")

	visitor := func(i Inspector, n ast.Node) bool {
		if n, ok := n.(*ast.FuncDecl); ok {
			n.Name = ast.NewIdent(fmt.Sprintf("Foo%s", n.Name.String()))
			i.Replace(n)
		}

		return true
	}

	inspector := NewInspector(visitor)
	result := inspector.Inspect(f)

	actualOutBuf := new(bytes.Buffer)
	err = format.Node(actualOutBuf, fset, result)
	assert.NoError(t, err, "Error formatting output AST")
	assert.Equal(t, string(expectedOut), actualOutBuf.String(), "Expected output doesn't match actual output")
}
