package main

/// see https://geekon.tech/post/lexing-parsing-of-golang-compiler/
import (
	"compress/flate"
	"github.com/scritchley/orc"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"
)

var id int64

func main() {
	arg := os.Args[1]
	println(arg)
	fset := token.NewFileSet()
	f, _ := parser.ParseFile(fset, arg+".go", nil, parser.Mode(0))

	fi, err := os.Create(arg + `.pkg`)
	if err != nil {
		println(err.Error())
	}
	defer fi.Close()

	schema, err := orc.ParseSchema(`struct<id:int,pid:int,etyp:string,name:string,typ:string,f1:string,f2:string>`)
	if err != nil {
		println(err.Error())
	}

	w, err := orc.NewWriter(fi, orc.SetSchema(schema), orc.SetCompression(orc.CompressionZlib{Level: flate.DefaultCompression}))
	if err != nil {
		println(err.Error())
	}

	var typ_id int64
	var etyp string /// f:function, m:method, st:struct-type, it:interface-type, c:constant, t:type

	tmap := make(map[string]int64)

	ast.Inspect(f, func(n ast.Node) bool {

		switch n.(type) {
		case *ast.GenDecl:
			gn, ok := n.(*ast.GenDecl)
			if ok {
				switch gn.Tok {
				case token.IMPORT:
					for _, v := range gn.Specs {
						ip := v.(*ast.ImportSpec)
						println("import: ", ip.Path.Value)
					}
				case token.TYPE:
					for _, v := range gn.Specs {

						name := v.(*ast.TypeSpec).Name.Name
						var typ string

						switch v.(*ast.TypeSpec).Type.(type) {
						case *ast.Ident:
							etyp = "t"
							typ = v.(*ast.TypeSpec).Type.(*ast.Ident).Name
						case *ast.InterfaceType:
							etyp = "it"
							typ = "intrf"
						case *ast.StructType:
							etyp = "st"
							typ = "struct"
						}

						typ_id = genId()

						if etyp != "it" {
							tmap[name] = typ_id
						}

						err := w.Write(typ_id, 0, etyp, name, typ, "", "")
						if err != nil {
							println(err.Error())
						}
					}
				case token.CONST:
					for _, v := range gn.Specs {
						name := v.(*ast.ValueSpec).Names[0].Name
						typ := v.(*ast.ValueSpec).Type.(*ast.Ident).Name

						var val string

						switch v.(*ast.ValueSpec).Values[0].(*ast.CallExpr).Args[0].(type) {
						case *ast.BasicLit:
							val = v.(*ast.ValueSpec).Values[0].(*ast.CallExpr).Args[0].(*ast.BasicLit).Value
						case *ast.Ident:
							val = v.(*ast.ValueSpec).Values[0].(*ast.CallExpr).Args[0].(*ast.Ident).Name
						}

						typ_id = tmap[typ]

						err := w.Write(genId(), typ_id, "c", name, typ, val, "")
						if err != nil {
							println(err.Error())
						}

					}

				}
			}

		case *ast.FuncDecl:
			fn, ok := n.(*ast.FuncDecl)
			var fname, recvr, recvrn, rtyp string
			if ok {
				if fn.Name.IsExported() {
					var cotyp string
					fname = fn.Name.Name

					if fn.Recv != nil {

						switch fn.Recv.List[0].Type.(type) {
						case *ast.StarExpr:
							recvrn = fn.Recv.List[0].Names[0].Name
							rtyp = fn.Recv.List[0].Type.(*ast.StarExpr).X.(*ast.Ident).Name
							recvr = recvrn + " *" + rtyp

						}
					}

					var ityp, otyp []string
					if fn.Type.Params != nil {
						for _, fl := range fn.Type.Params.List {

							_, typ := parse_io(fl)
							ityp = append(ityp, typ)
						}

					}

					if fn.Type.Results != nil {
						var typ string
						for _, fl := range fn.Type.Results.List {

							cotyp, typ = parse_io(fl)
							otyp = append(otyp, typ)
						}

					}

					if rtyp != "" { /// if recvr, is method

						typ_id = tmap[rtyp]

						err := w.Write(genId(), typ_id, "m", fname, recvr, strings.Join(ityp, ", "), strings.Join(otyp, ", "))
						if err != nil {
							println(err.Error())
						}

					} else {

						if cotyp != "" { /// if has output type, func may have 'type' parent

							typ_id = tmap[cotyp]

							err := w.Write(genId(), typ_id, "f", fname, recvr, strings.Join(ityp, ", "), strings.Join(otyp, ", "))
							if err != nil {
								println(err.Error())
							}

						} else { /// bare function

							err := w.Write(genId(), 0, "f", fname, recvr, strings.Join(ityp, ", "), strings.Join(otyp, ", "))
							if err != nil {
								println(err.Error())
							}

						}

					}

				}
			}
		}

		return true
	})

	err = w.Close()
	if err != nil {
		println(err.Error())
	}

}

func genId() int64 {
	id += 1
	return id
}

func parse_io(f *ast.Field) (string, string) {

	var cotyp, typ string

	switch f.Type.(type) {
	case *ast.FuncType:
		var ityp []string
		if f.Type.(*ast.FuncType).Params != nil {
			for _, ft := range f.Type.(*ast.FuncType).Params.List {
				_, t := parse_io(ft)
				ityp = append(ityp, t)
			}
		}

		typ = "func(" + strings.Join(ityp, ", ") + ")"

	case *ast.Ident:
		typ = f.Type.(*ast.Ident).Name

	case *ast.SelectorExpr:
		typ = f.Type.(*ast.SelectorExpr).X.(*ast.Ident).Name + "." + f.Type.(*ast.SelectorExpr).Sel.Name

	case *ast.ArrayType:
		switch f.Type.(*ast.ArrayType).Elt.(type) {
		case *ast.Ident:
			cotyp = f.Type.(*ast.ArrayType).Elt.(*ast.Ident).Name
			typ = "[]" + cotyp
		case *ast.StarExpr:
			switch f.Type.(*ast.ArrayType).Elt.(*ast.StarExpr).X.(type) {
			case *ast.Ident:
				cotyp = f.Type.(*ast.ArrayType).Elt.(*ast.StarExpr).X.(*ast.Ident).Name
				typ = "[]*" + cotyp
			case *ast.SelectorExpr:
				cotyp = f.Type.(*ast.ArrayType).Elt.(*ast.StarExpr).X.(*ast.SelectorExpr).X.(*ast.Ident).Name + "." + f.Type.(*ast.ArrayType).Elt.(*ast.StarExpr).X.(*ast.SelectorExpr).Sel.Name
				typ = "[]*" + cotyp
			}
		}
	case *ast.StarExpr:
		switch f.Type.(*ast.StarExpr).X.(type) {
		case *ast.Ident:
			cotyp = f.Type.(*ast.StarExpr).X.(*ast.Ident).Name
			typ = "*" + cotyp

		case *ast.SelectorExpr:
			cotyp = f.Type.(*ast.StarExpr).X.(*ast.SelectorExpr).X.(*ast.Ident).Name + "." + f.Type.(*ast.StarExpr).X.(*ast.SelectorExpr).Sel.Name
			typ = "*" + cotyp
		}
	}

	if f.Names != nil {
		for _, name := range f.Names {
			typ = name.Name + " " + typ
		}
	}

	return cotyp, typ

}
