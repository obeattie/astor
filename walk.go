package astor

import (
	"fmt"
	"go/ast"
)

// Helper functions for common node lists. They may be empty. Copied/adapted shamelessly from go/ast.

func inspectIdentList(i Inspector, list []*ast.Ident) []*ast.Ident {
	newList := make([]*ast.Ident, len(list))
	for l, x := range list {
		newList[l] = Inspect(i, x).(*ast.Ident)
	}
	return newList
}

func inspectExprList(i Inspector, list []ast.Expr) []ast.Expr {
	newList := make([]ast.Expr, len(list))
	for l, x := range list {
		newList[l] = Inspect(i, x).(ast.Expr)
	}
	return newList
}

func inspectStmtList(i Inspector, list []ast.Stmt) []ast.Stmt {
	newList := make([]ast.Stmt, len(list))
	for l, x := range list {
		newList[l] = Inspect(i, x).(ast.Stmt)
	}
	return newList
}

func inspectDeclList(i Inspector, list []ast.Decl) []ast.Decl {
	newList := make([]ast.Decl, len(list))
	for l, x := range list {
		newList[l] = Inspect(i, x).(ast.Decl)
	}
	return newList
}

func Inspect(i Inspector, node ast.Node) ast.Node {
	if i = i.Visit(node); i == nil {
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
			n.List[l] = Inspect(i, c).(*ast.Comment)
		}

	case *ast.Field:
		if n.Doc != nil {
			n.Doc = Inspect(i, n.Doc).(*ast.CommentGroup)
		}
		n.Names = inspectIdentList(i, n.Names)
		n.Type = Inspect(i, n.Type).(ast.Expr)
		if n.Tag != nil {
			n.Tag = Inspect(i, n.Tag).(*ast.BasicLit)
		}
		if n.Comment != nil {
			n.Comment = Inspect(i, n.Comment).(*ast.CommentGroup)
		}

	case *ast.FieldList:
		for l, f := range n.List {
			n.List[l] = Inspect(i, f).(*ast.Field)
		}

	// Expressions
	case *ast.BadExpr, *ast.Ident, *ast.BasicLit:
		// nothing to do

	case *ast.Ellipsis:
		if n.Elt != nil {
			n.Elt = Inspect(i, n.Elt).(ast.Expr)
		}

	case *ast.FuncLit:
		n.Type = Inspect(i, n.Type).(*ast.FuncType)
		n.Body = Inspect(i, n.Body).(*ast.BlockStmt)

	case *ast.CompositeLit:
		if n.Type != nil {
			n.Type = Inspect(i, n.Type).(ast.Expr)
		}
		n.Elts = inspectExprList(i, n.Elts)

	case *ast.ParenExpr:
		n.X = Inspect(i, n.X).(ast.Expr)

	case *ast.SelectorExpr:
		n.X = Inspect(i, n.X).(ast.Expr)
		n.Sel = Inspect(i, n.Sel).(*ast.Ident)

	case *ast.IndexExpr:
		n.X = Inspect(i, n.X).(ast.Expr)
		n.Index = Inspect(i, n.Index).(ast.Expr)

	case *ast.SliceExpr:
		n.X = Inspect(i, n.X).(ast.Expr)
		if n.Low != nil {
			n.Low = Inspect(i, n.Low).(ast.Expr)
		}
		if n.High != nil {
			n.High = Inspect(i, n.High).(ast.Expr)
		}
		if n.Max != nil {
			n.Max = Inspect(i, n.Max).(ast.Expr)
		}

	case *ast.TypeAssertExpr:
		n.X = Inspect(i, n.X).(ast.Expr)
		if n.Type != nil {
			n.Type = Inspect(i, n.Type).(ast.Expr)
		}

	case *ast.CallExpr:
		n.Fun = Inspect(i, n.Fun).(ast.Expr)
		n.Args = inspectExprList(i, n.Args)

	case *ast.StarExpr:
		n.X = Inspect(i, n.X).(ast.Expr)

	case *ast.UnaryExpr:
		n.X = Inspect(i, n.X).(ast.Expr)

	case *ast.BinaryExpr:
		n.X = Inspect(i, n.X).(ast.Expr)
		n.Y = Inspect(i, n.Y).(ast.Expr)

	case *ast.KeyValueExpr:
		n.Key = Inspect(i, n.Key).(ast.Expr)
		n.Value = Inspect(i, n.Value).(ast.Expr)

	// Types
	case *ast.ArrayType:
		if n.Len != nil {
			n.Len = Inspect(i, n.Len).(ast.Expr)
		}
		n.Elt = Inspect(i, n.Elt).(ast.Expr)

	case *ast.StructType:
		n.Fields = Inspect(i, n.Fields).(*ast.FieldList)

	case *ast.FuncType:
		if n.Params != nil {
			n.Params = Inspect(i, n.Params).(*ast.FieldList)
		}
		if n.Results != nil {
			n.Results = Inspect(i, n.Results).(*ast.FieldList)
		}

	case *ast.InterfaceType:
		n.Methods = Inspect(i, n.Methods).(*ast.FieldList)

	case *ast.MapType:
		n.Key = Inspect(i, n.Key).(ast.Expr)
		n.Value = Inspect(i, n.Value).(ast.Expr)

	case *ast.ChanType:
		n.Value = Inspect(i, n.Value).(ast.Expr)

	// Statements
	case *ast.BadStmt:
		// nothing to do

	case *ast.DeclStmt:
		n.Decl = Inspect(i, n.Decl).(ast.Decl)

	case *ast.EmptyStmt:
		// nothing to do

	case *ast.LabeledStmt:
		n.Label = Inspect(i, n.Label).(*ast.Ident)
		n.Stmt = Inspect(i, n.Stmt).(ast.Stmt)

	case *ast.ExprStmt:
		n.X = Inspect(i, n.X).(ast.Expr)

	case *ast.SendStmt:
		n.Chan = Inspect(i, n.Chan).(ast.Expr)
		n.Value = Inspect(i, n.Value).(ast.Expr)

	case *ast.IncDecStmt:
		n.X = Inspect(i, n.X).(ast.Expr)

	case *ast.AssignStmt:
		n.Lhs = inspectExprList(i, n.Lhs)
		n.Rhs = inspectExprList(i, n.Rhs)

	case *ast.GoStmt:
		n.Call = Inspect(i, n.Call).(*ast.CallExpr)

	case *ast.DeferStmt:
		n.Call = Inspect(i, n.Call).(*ast.CallExpr)

	case *ast.ReturnStmt:
		n.Results = inspectExprList(i, n.Results)

	case *ast.BranchStmt:
		if n.Label != nil {
			n.Label = Inspect(i, n.Label).(*ast.Ident)
		}

	case *ast.BlockStmt:
		n.List = inspectStmtList(i, n.List)

	case *ast.IfStmt:
		if n.Init != nil {
			n.Init = Inspect(i, n.Init).(ast.Stmt)
		}
		n.Cond = Inspect(i, n.Cond).(ast.Expr)
		n.Body = Inspect(i, n.Body).(*ast.BlockStmt)
		if n.Else != nil {
			n.Else = Inspect(i, n.Else).(ast.Stmt)
		}

	case *ast.CaseClause:
		n.List = inspectExprList(i, n.List)
		n.Body = inspectStmtList(i, n.Body)

	case *ast.SwitchStmt:
		if n.Init != nil {
			n.Init = Inspect(i, n.Init).(ast.Stmt)
		}
		if n.Tag != nil {
			n.Tag = Inspect(i, n.Tag).(ast.Expr)
		}
		n.Body = Inspect(i, n.Body).(*ast.BlockStmt)

	case *ast.TypeSwitchStmt:
		if n.Init != nil {
			n.Init = Inspect(i, n.Init).(ast.Stmt)
		}
		n.Assign = Inspect(i, n.Assign).(ast.Stmt)
		n.Body = Inspect(i, n.Body).(*ast.BlockStmt)

	case *ast.CommClause:
		if n.Comm != nil {
			n.Comm = Inspect(i, n.Comm).(ast.Stmt)
		}
		n.Body = inspectStmtList(i, n.Body)

	case *ast.SelectStmt:
		n.Body = Inspect(i, n.Body).(*ast.BlockStmt)

	case *ast.ForStmt:
		if n.Init != nil {
			n.Init = Inspect(i, n.Init).(ast.Stmt)
		}
		if n.Cond != nil {
			n.Cond = Inspect(i, n.Cond).(ast.Expr)
		}
		if n.Post != nil {
			n.Post = Inspect(i, n.Post).(ast.Stmt)
		}
		n.Body = Inspect(i, n.Body).(*ast.BlockStmt)

	case *ast.RangeStmt:
		if n.Key != nil {
			n.Key = Inspect(i, n.Key).(ast.Expr)
		}
		if n.Value != nil {
			n.Value = Inspect(i, n.Value).(ast.Expr)
		}
		n.X = Inspect(i, n.X).(ast.Expr)
		n.Body = Inspect(i, n.Body).(*ast.BlockStmt)

	// Declarations
	case *ast.ImportSpec:
		if n.Doc != nil {
			n.Doc = Inspect(i, n.Doc).(*ast.CommentGroup)
		}
		if n.Name != nil {
			n.Name = Inspect(i, n.Name).(*ast.Ident)
		}
		n.Path = Inspect(i, n.Path).(*ast.BasicLit)
		if n.Comment != nil {
			n.Comment = Inspect(i, n.Comment).(*ast.CommentGroup)
		}

	case *ast.ValueSpec:
		if n.Doc != nil {
			n.Doc = Inspect(i, n.Doc).(*ast.CommentGroup)
		}
		n.Names = inspectIdentList(i, n.Names)
		if n.Type != nil {
			n.Type = Inspect(i, n.Type).(ast.Expr)
		}
		n.Values = inspectExprList(i, n.Values)
		if n.Comment != nil {
			n.Comment = Inspect(i, n.Comment).(*ast.CommentGroup)
		}

	case *ast.TypeSpec:
		if n.Doc != nil {
			n.Doc = Inspect(i, n.Doc).(*ast.CommentGroup)
		}
		n.Name = Inspect(i, n.Name).(*ast.Ident)
		n.Type = Inspect(i, n.Type).(ast.Expr)
		if n.Comment != nil {
			n.Comment = Inspect(i, n.Comment).(*ast.CommentGroup)
		}

	case *ast.BadDecl:
		// nothing to do

	case *ast.GenDecl:
		if n.Doc != nil {
			n.Doc = Inspect(i, n.Doc).(*ast.CommentGroup)
		}
		for l, s := range n.Specs {
			n.Specs[l] = Inspect(i, s).(ast.Spec)
		}

	case *ast.FuncDecl:
		if n.Doc != nil {
			n.Doc = Inspect(i, n.Doc).(*ast.CommentGroup)
		}
		if n.Recv != nil {
			n.Recv = Inspect(i, n.Recv).(*ast.FieldList)
		}
		n.Name = Inspect(i, n.Name).(*ast.Ident)
		n.Type = Inspect(i, n.Type).(*ast.FuncType)
		if n.Body != nil {
			n.Body = Inspect(i, n.Body).(*ast.BlockStmt)
		}

	// Files and packages
	case *ast.File:
		if n.Doc != nil {
			n.Doc = Inspect(i, n.Doc).(*ast.CommentGroup)
		}
		n.Name = Inspect(i, n.Name).(*ast.Ident)
		n.Decls = inspectDeclList(i, n.Decls)
		// don't inspect n.Comments - they have been
		// visited already through the individual
		// nodes

	case *ast.Package:
		for l, f := range n.Files {
			n.Files[l] = Inspect(i, f).(*ast.File)
		}

	default:
		fmt.Printf("astor.Inspect: unexpected node type %T", n)
		panic("astor.Inspect")
	}

	i.Visit(nil)
	return node
}
