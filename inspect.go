package astor

import (
	"fmt"
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
	// Inspect walks the AST for the node passed, calling the Visitor, and returning the modified tree
	Inspect(node ast.Node) ast.Node
	// Visit calls the Visitor for the node, returning its replacement, and optionally an Inspector to be called for its
	// children
	Visit(node ast.Node) (replacement ast.Node, i Inspector)
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

func (i *inspectorImpl) Visit(n ast.Node) (ast.Node, Inspector) {
	i.mtx.Lock()
	i.node = n
	result := i.visitorImpl(i, n)
	replacement := i.node
	i.node = nil
	i.mtx.Unlock()

	if result {
		return replacement, i
	} else {
		return replacement, nil
	}
}

func (i *inspectorImpl) Inspect(node ast.Node) ast.Node {
	var ii Inspector
	if node, ii = i.Visit(node); ii == nil {
		return node
	}

	// inspect children
	// (the order of the cases matches the order
	// of the corresponding node types in ast.go)
	switch n := node.(type) {
	// Comments and fields
	case *ast.Comment:
		// nothing to do

	case *ast.CommentGroup:
		for l, c := range n.List {
			n.List[l] = ii.Inspect(c).(*ast.Comment)
		}

	case *ast.Field:
		if n.Doc != nil {
			n.Doc = ii.Inspect(n.Doc).(*ast.CommentGroup)
		}
		n.Names = inspectIdentList(ii, n.Names)
		n.Type = ii.Inspect(n.Type).(ast.Expr)
		if n.Tag != nil {
			n.Tag = ii.Inspect(n.Tag).(*ast.BasicLit)
		}
		if n.Comment != nil {
			n.Comment = ii.Inspect(n.Comment).(*ast.CommentGroup)
		}

	case *ast.FieldList:
		for l, f := range n.List {
			n.List[l] = ii.Inspect(f).(*ast.Field)
		}

	// Expressions
	case *ast.BadExpr, *ast.Ident, *ast.BasicLit:
		// nothing to do

	case *ast.Ellipsis:
		if n.Elt != nil {
			n.Elt = ii.Inspect(n.Elt).(ast.Expr)
		}

	case *ast.FuncLit:
		n.Type = ii.Inspect(n.Type).(*ast.FuncType)
		n.Body = ii.Inspect(n.Body).(*ast.BlockStmt)

	case *ast.CompositeLit:
		if n.Type != nil {
			n.Type = ii.Inspect(n.Type).(ast.Expr)
		}
		n.Elts = inspectExprList(i, n.Elts)

	case *ast.ParenExpr:
		n.X = ii.Inspect(n.X).(ast.Expr)

	case *ast.SelectorExpr:
		n.X = ii.Inspect(n.X).(ast.Expr)
		n.Sel = ii.Inspect(n.Sel).(*ast.Ident)

	case *ast.IndexExpr:
		n.X = ii.Inspect(n.X).(ast.Expr)
		n.Index = ii.Inspect(n.Index).(ast.Expr)

	case *ast.SliceExpr:
		n.X = ii.Inspect(n.X).(ast.Expr)
		if n.Low != nil {
			n.Low = ii.Inspect(n.Low).(ast.Expr)
		}
		if n.High != nil {
			n.High = ii.Inspect(n.High).(ast.Expr)
		}
		if n.Max != nil {
			n.Max = ii.Inspect(n.Max).(ast.Expr)
		}

	case *ast.TypeAssertExpr:
		n.X = ii.Inspect(n.X).(ast.Expr)
		if n.Type != nil {
			n.Type = ii.Inspect(n.Type).(ast.Expr)
		}

	case *ast.CallExpr:
		n.Fun = ii.Inspect(n.Fun).(ast.Expr)
		n.Args = inspectExprList(i, n.Args)

	case *ast.StarExpr:
		n.X = ii.Inspect(n.X).(ast.Expr)

	case *ast.UnaryExpr:
		n.X = ii.Inspect(n.X).(ast.Expr)

	case *ast.BinaryExpr:
		n.X = ii.Inspect(n.X).(ast.Expr)
		n.Y = ii.Inspect(n.Y).(ast.Expr)

	case *ast.KeyValueExpr:
		n.Key = ii.Inspect(n.Key).(ast.Expr)
		n.Value = ii.Inspect(n.Value).(ast.Expr)

	// Types
	case *ast.ArrayType:
		if n.Len != nil {
			n.Len = ii.Inspect(n.Len).(ast.Expr)
		}
		n.Elt = ii.Inspect(n.Elt).(ast.Expr)

	case *ast.StructType:
		n.Fields = ii.Inspect(n.Fields).(*ast.FieldList)

	case *ast.FuncType:
		if n.Params != nil {
			n.Params = ii.Inspect(n.Params).(*ast.FieldList)
		}
		if n.Results != nil {
			n.Results = ii.Inspect(n.Results).(*ast.FieldList)
		}

	case *ast.InterfaceType:
		n.Methods = ii.Inspect(n.Methods).(*ast.FieldList)

	case *ast.MapType:
		n.Key = ii.Inspect(n.Key).(ast.Expr)
		n.Value = ii.Inspect(n.Value).(ast.Expr)

	case *ast.ChanType:
		n.Value = ii.Inspect(n.Value).(ast.Expr)

	// Statements
	case *ast.BadStmt:
		// nothing to do

	case *ast.DeclStmt:
		n.Decl = ii.Inspect(n.Decl).(ast.Decl)

	case *ast.EmptyStmt:
		// nothing to do

	case *ast.LabeledStmt:
		n.Label = ii.Inspect(n.Label).(*ast.Ident)
		n.Stmt = ii.Inspect(n.Stmt).(ast.Stmt)

	case *ast.ExprStmt:
		n.X = ii.Inspect(n.X).(ast.Expr)

	case *ast.SendStmt:
		n.Chan = ii.Inspect(n.Chan).(ast.Expr)
		n.Value = ii.Inspect(n.Value).(ast.Expr)

	case *ast.IncDecStmt:
		n.X = ii.Inspect(n.X).(ast.Expr)

	case *ast.AssignStmt:
		n.Lhs = inspectExprList(i, n.Lhs)
		n.Rhs = inspectExprList(i, n.Rhs)

	case *ast.GoStmt:
		n.Call = ii.Inspect(n.Call).(*ast.CallExpr)

	case *ast.DeferStmt:
		n.Call = ii.Inspect(n.Call).(*ast.CallExpr)

	case *ast.ReturnStmt:
		n.Results = inspectExprList(i, n.Results)

	case *ast.BranchStmt:
		if n.Label != nil {
			n.Label = ii.Inspect(n.Label).(*ast.Ident)
		}

	case *ast.BlockStmt:
		n.List = inspectStmtList(i, n.List)

	case *ast.IfStmt:
		if n.Init != nil {
			n.Init = ii.Inspect(n.Init).(ast.Stmt)
		}
		n.Cond = ii.Inspect(n.Cond).(ast.Expr)
		n.Body = ii.Inspect(n.Body).(*ast.BlockStmt)
		if n.Else != nil {
			n.Else = ii.Inspect(n.Else).(ast.Stmt)
		}

	case *ast.CaseClause:
		n.List = inspectExprList(i, n.List)
		n.Body = inspectStmtList(i, n.Body)

	case *ast.SwitchStmt:
		if n.Init != nil {
			n.Init = ii.Inspect(n.Init).(ast.Stmt)
		}
		if n.Tag != nil {
			n.Tag = ii.Inspect(n.Tag).(ast.Expr)
		}
		n.Body = ii.Inspect(n.Body).(*ast.BlockStmt)

	case *ast.TypeSwitchStmt:
		if n.Init != nil {
			n.Init = ii.Inspect(n.Init).(ast.Stmt)
		}
		n.Assign = ii.Inspect(n.Assign).(ast.Stmt)
		n.Body = ii.Inspect(n.Body).(*ast.BlockStmt)

	case *ast.CommClause:
		if n.Comm != nil {
			n.Comm = ii.Inspect(n.Comm).(ast.Stmt)
		}
		n.Body = inspectStmtList(i, n.Body)

	case *ast.SelectStmt:
		n.Body = ii.Inspect(n.Body).(*ast.BlockStmt)

	case *ast.ForStmt:
		if n.Init != nil {
			n.Init = ii.Inspect(n.Init).(ast.Stmt)
		}
		if n.Cond != nil {
			n.Cond = ii.Inspect(n.Cond).(ast.Expr)
		}
		if n.Post != nil {
			n.Post = ii.Inspect(n.Post).(ast.Stmt)
		}
		n.Body = ii.Inspect(n.Body).(*ast.BlockStmt)

	case *ast.RangeStmt:
		if n.Key != nil {
			n.Key = ii.Inspect(n.Key).(ast.Expr)
		}
		if n.Value != nil {
			n.Value = ii.Inspect(n.Value).(ast.Expr)
		}
		n.X = ii.Inspect(n.X).(ast.Expr)
		n.Body = ii.Inspect(n.Body).(*ast.BlockStmt)

	// Declarations
	case *ast.ImportSpec:
		if n.Doc != nil {
			n.Doc = ii.Inspect(n.Doc).(*ast.CommentGroup)
		}
		if n.Name != nil {
			n.Name = ii.Inspect(n.Name).(*ast.Ident)
		}
		n.Path = ii.Inspect(n.Path).(*ast.BasicLit)
		if n.Comment != nil {
			n.Comment = ii.Inspect(n.Comment).(*ast.CommentGroup)
		}

	case *ast.ValueSpec:
		if n.Doc != nil {
			n.Doc = ii.Inspect(n.Doc).(*ast.CommentGroup)
		}
		n.Names = inspectIdentList(ii, n.Names)
		if n.Type != nil {
			n.Type = ii.Inspect(n.Type).(ast.Expr)
		}
		n.Values = inspectExprList(i, n.Values)
		if n.Comment != nil {
			n.Comment = ii.Inspect(n.Comment).(*ast.CommentGroup)
		}

	case *ast.TypeSpec:
		if n.Doc != nil {
			n.Doc = ii.Inspect(n.Doc).(*ast.CommentGroup)
		}
		n.Name = ii.Inspect(n.Name).(*ast.Ident)
		n.Type = ii.Inspect(n.Type).(ast.Expr)
		if n.Comment != nil {
			n.Comment = ii.Inspect(n.Comment).(*ast.CommentGroup)
		}

	case *ast.BadDecl:
		// nothing to do

	case *ast.GenDecl:
		if n.Doc != nil {
			n.Doc = ii.Inspect(n.Doc).(*ast.CommentGroup)
		}
		for l, s := range n.Specs {
			n.Specs[l] = ii.Inspect(s).(ast.Spec)
		}

	case *ast.FuncDecl:
		if n.Doc != nil {
			n.Doc = ii.Inspect(n.Doc).(*ast.CommentGroup)
		}
		if n.Recv != nil {
			n.Recv = ii.Inspect(n.Recv).(*ast.FieldList)
		}
		n.Name = ii.Inspect(n.Name).(*ast.Ident)
		n.Type = ii.Inspect(n.Type).(*ast.FuncType)
		if n.Body != nil {
			n.Body = ii.Inspect(n.Body).(*ast.BlockStmt)
		}

	// Files and packages
	case *ast.File:
		if n.Doc != nil {
			n.Doc = ii.Inspect(n.Doc).(*ast.CommentGroup)
		}
		n.Name = ii.Inspect(n.Name).(*ast.Ident)
		n.Decls = inspectDeclList(i, n.Decls)
		// don't inspect n.Comments - they have been
		// visited already through the individual
		// nodes

	case *ast.Package:
		for l, f := range n.Files {
			n.Files[l] = ii.Inspect(f).(*ast.File)
		}

	default:
		fmt.Printf("astor.Inspect: unexpected node type %T", n)
		panic("astor.Inspect")
	}

	ii.Visit(nil)
	return node
}

// Helper functions for common node lists. They may be empty. Copied/adapted shamelessly from go/ast.

func inspectIdentList(i Inspector, list []*ast.Ident) []*ast.Ident {
	newList := make([]*ast.Ident, len(list))
	for l, x := range list {
		newList[l] = i.Inspect(x).(*ast.Ident)
	}
	return newList
}

func inspectExprList(i Inspector, list []ast.Expr) []ast.Expr {
	newList := make([]ast.Expr, len(list))
	for l, x := range list {
		newList[l] = i.Inspect(x).(ast.Expr)
	}
	return newList
}

func inspectStmtList(i Inspector, list []ast.Stmt) []ast.Stmt {
	newList := make([]ast.Stmt, len(list))
	for l, x := range list {
		newList[l] = i.Inspect(x).(ast.Stmt)
	}
	return newList
}

func inspectDeclList(i Inspector, list []ast.Decl) []ast.Decl {
	newList := make([]ast.Decl, len(list))
	for l, x := range list {
		newList[l] = i.Inspect(x).(ast.Decl)
	}
	return newList
}
