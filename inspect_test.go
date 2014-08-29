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

func runInspector(t *testing.T, infile, outfile string, visitor Visitor) {
	var err error
	var inputSrc []byte
	var expectedOut []byte

	inputSrc, err = ioutil.ReadFile(infile)
	assert.NoError(t, err, "Error reading input")
	expectedOut, err = ioutil.ReadFile(outfile)
	assert.NoError(t, err, "Error reading expected output")

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, infile, inputSrc, parserFlags)
	assert.NoError(t, err, "Error parsing input")

	inspector := NewInspector(visitor)
	result := inspector.Inspect(f)

	actualOutBuf := new(bytes.Buffer)
	err = format.Node(actualOutBuf, fset, result)
	assert.NoError(t, err, "Error formatting output AST")
	assert.Equal(t, string(expectedOut), actualOutBuf.String(), "Expected output doesn't match actual output")
}

func TestPrependingFuncName(t *testing.T) {
	visitor := func(i Inspector, n ast.Node) bool {
		if n, ok := n.(*ast.FuncDecl); ok {
			n.Name = ast.NewIdent(fmt.Sprintf("Foo%s", n.Name.String()))
			i.Replace(n)
			return false
		}

		return true
	}

	runInspector(
		t,
		"test-samples/prepending-func-name.go.in",
		"test-samples/prepending-func-name.go.out",
		visitor)
}

func TestChangePointerToInterface(t *testing.T) {
	visitor := func(i Inspector, n ast.Node) bool {
		if _, ok := n.(*ast.StarExpr); ok {
			newNode := &ast.SelectorExpr{
				X:   ast.NewIdent("pkg"),
				Sel: ast.NewIdent("InterfaceName"),
			}
			i.Replace(newNode)
			return false
		}

		return true
	}

	runInspector(
		t,
		"test-samples/pointer-to-interface.go.in",
		"test-samples/pointer-to-interface.go.out",
		visitor)
}
