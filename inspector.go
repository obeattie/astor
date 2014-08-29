package astor

import (
	"go/ast"
	"sync"
)

// Visitor is called by an Inspector for each node in an AST. If the result is true, each of the children of
// node will be visited, followed by a call with node as nil. This is deliberately similar to ast.Visitor.
type Visitor func(i Inspector, node ast.Node) (recurse bool)

// An Inspector visits each node in an AST, calling a Visitor. The current node may be replaced in the AST with a call
// to Replace().
type Inspector interface {
	// Current returns the node currently being inspected
	Current() ast.Node
	// Replace replaces the node currently being inspected with the passed node
	Replace(ast.Node)
	Visit(ast.Node) Inspector
}

// NewInspector constructs a new Inspector with the passed Visitor.
func NewInspector(v Visitor) Inspector {
	return &inspectorImpl{
		visitorImpl: v,
	}
}

type inspectorImpl struct {
	mtx         sync.Mutex
	node        ast.Node
	visitorImpl Visitor
}

func (i *inspectorImpl) Current() ast.Node {
	return i.node
}

func (i *inspectorImpl) Replace(n ast.Node) {
	i.node = n
}

func (i *inspectorImpl) Visit(n ast.Node) Inspector {
	i.mtx.Lock()
	defer i.mtx.Unlock()

	i.node = n

	if i.visitorImpl(i, n) {
		return i
	} else {
		return nil
	}
}
