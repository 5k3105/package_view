package main

import (
	"github.com/scritchley/orc"
	"github.com/therecipe/qt/core"
	"github.com/therecipe/qt/gui"
	"github.com/therecipe/qt/widgets"
	"strings"
)

type PkgTree struct {
	*widgets.QTreeView
	Im *gui.QStandardItemModel
}

func New_PkgTree(window *widgets.QMainWindow, pkg string) *PkgTree {
	pt := &PkgTree{
		QTreeView: widgets.NewQTreeView(window),
	}

	model := gui.NewQStandardItemModel(window)
	model.SetColumnCount(1)

	pt.SetModel(model)
	pt.Im = model
	pt.SetFont(font)
	pt.SetStyleSheet(ptstylesheet)

	pt.LoadFile(pkg)
	pt.SetColumnWidth(0, 250)
	pt.Header().SetStretchLastSection(true)
	pt.SetSortingEnabled(true)

	// SortChildren
	root := model.InvisibleRootItem().Index()
	for i := 0; i < model.RowCount(root); i++ {
		idx := model.Index(i, 0, root)
		item := model.ItemFromIndex(idx)
		if model.HasChildren(idx) {
			item.SortChildren(0, core.Qt__AscendingOrder)
		}
	}
	return pt
}

func (pt *PkgTree) LoadFile(pkg string) {
	model := pt.Im
	model.Clear()
	model.SetColumnCount(1)
	filename := pkg + `.pkg`

	fnheader := gui.NewQStandardItem2(pkg)
	fnheader.SetFont(font)
	model.SetHorizontalHeaderItem(0, fnheader)

	r, err := orc.Open(filename)
	if err != nil {
		println(err.Error())
	}

	c := r.Select("id", "pid", "etyp", "name", "typ", "f1", "f2")

	var pid int64
	parent := model.InvisibleRootItem()
	path := make(map[int64]*gui.QStandardItem)
	path[0] = parent
	/// init type folders:
	st := gui.NewQStandardItem2(`Struct Types`)
	child := []*gui.QStandardItem{st}
	parent.AppendRow(child)
	t := gui.NewQStandardItem2(`Types`)
	child = []*gui.QStandardItem{t}
	parent.AppendRow(child)
	bf := gui.NewQStandardItem2(`Bare Functions`)
	child = []*gui.QStandardItem{bf}
	parent.AppendRow(child)
	it := gui.NewQStandardItem2(`Interface Types`)
	child = []*gui.QStandardItem{it}
	parent.AppendRow(child)

	for c.Stripes() {
		for c.Next() {
			id := c.Row()[0].(int64)
			id_parent := c.Row()[1].(int64)
			etyp := c.Row()[2].(string)
			name := c.Row()[3].(string)
			typ := c.Row()[4].(string)
			f1 := c.Row()[5].(string)
			f2 := c.Row()[6].(string)

			var line string
			if id_parent == 0 {
				switch etyp {
				case "t":
					line = name + " " + typ
				case "f":
					if strings.Contains(f2, ",") {
						line = "func " + name + "(" + f1 + ") (" + f2 + ")"
					} else {
						line = "func " + name + "(" + f1 + ") " + f2
					}
				case "m":
					if strings.Contains(f2, ",") {
						line = "func (" + typ + ") " + name + "(" + f1 + ") (" + f2 + ")"
					} else {
						line = "func (" + typ + ") " + name + "(" + f1 + ") " + f2
					}
				case "c":
					line = name + " " + f1
				case "st":
					line = name // + " " + typ --- name is the type, dont use interfaces for parents
				case "it":
					line = name
				}

				fn := gui.NewQStandardItem2(line)
				fn.SetData(core.NewQVariant9(id), 0x0100)

				child := []*gui.QStandardItem{fn}
				switch etyp {
				case "t":
					t.AppendRow(child)
				case "f":
					bf.AppendRow(child)
				case "st":
					st.AppendRow(child)
				case "it":
					it.AppendRow(child)
				}

				path[id] = child[0]

			}
		}
	}

	r, err = orc.Open(filename)
	if err != nil {
		println(err.Error())
	}

	c = r.Select("id", "pid", "etyp", "name", "typ", "f1", "f2")
	for c.Stripes() {
		for c.Next() {
			id := c.Row()[0].(int64)
			id_parent := c.Row()[1].(int64)
			etyp := c.Row()[2].(string)
			name := c.Row()[3].(string)
			typ := c.Row()[4].(string)
			f1 := c.Row()[5].(string)
			f2 := c.Row()[6].(string)

			var line string
			if id_parent != 0 {
				switch etyp {
				case "t":
					line = name + " " + typ
				case "f":
					if strings.Contains(f2, ",") {
						line = "func " + name + "(" + f1 + ") (" + f2 + ")"
					} else {
						line = "func " + name + "(" + f1 + ") " + f2
					}
				case "m":
					if strings.Contains(f2, ",") {
						line = "func (" + typ + ") " + name + "(" + f1 + ") (" + f2 + ")"
					} else {
						line = "func (" + typ + ") " + name + "(" + f1 + ") " + f2
					}
				case "c":
					line = name + " " + f1
				case "st":
					line = name // + " " + typ --- name is the type, dont use interfaces for parents
				case "it":
					line = name
				}

				pid = id_parent
				parent = path[pid]

				fn := gui.NewQStandardItem2(line)
				fn.SetData(core.NewQVariant9(id), 0x0100)

				child := []*gui.QStandardItem{fn}
				parent.AppendRow(child)
				path[id] = child[0]

			}
		}
	}

}

var ptstylesheet string = `
	QTreeView::branch:has-siblings:!adjoins-item {
		 border-image: url(:/tree/vline.png) 0;
	 }

	 QTreeView::branch:has-siblings:adjoins-item {
		 border-image: url(:/tree/branch-more.png) 0;
	 }

	 QTreeView::branch:!has-children:!has-siblings:adjoins-item {
		 border-image: url(:/tree/branch-end.png) 0;
	 }

	 QTreeView::branch:has-children:!has-siblings:closed,
	 QTreeView::branch:closed:has-children:has-siblings {
			 border-image: none;
			 image: url(:/tree/branch-closed.png);
	 }

	 QTreeView::branch:open:has-children:!has-siblings,
	 QTreeView::branch:open:has-children:has-siblings  {
			 border-image: none;
			 image: url(:/tree/branch-open.png);
	 }
	 

	QTreeView {
			selection-background-color: lightGrey;
	}

	QTreeView::item:selected
	{		background-color: lightGrey;
			color: black;		
	} 
`
