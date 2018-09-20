package main

import (
	"github.com/therecipe/qt/gui"
	"github.com/therecipe/qt/widgets"
	"os"
)

var font *gui.QFont

func init() {
	font = gui.NewQFont2("corbel", 12, 1, false)
}

func main() {
	widgets.NewQApplication(len(os.Args), os.Args)
	window := widgets.NewQMainWindow(nil, 0)

	pkg := os.Args[1]
	window.SetWindowTitle("Pkg: " + pkg)
	pkgtree := New_PkgTree(window, pkg)

	window.SetCentralWidget(pkgtree)
	widgets.QApplication_SetStyle2("fusion")
	window.ShowMaximized()
	widgets.QApplication_Exec()
}
