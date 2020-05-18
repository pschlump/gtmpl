package main

import (
	"fmt"
	"testing"
)

// func PathFind(path, fn string) (full_path, file_type string, err error) {

func Test_FindPath(t *testing.T) {

	fp, ex, err := PathFind("tmpl1;./testdata", "base-table.html")

	if err != nil {
		t.Errorf("Error Got error %s when expecting success", err)
	}
	if ex != ".html" {
		t.Errorf("Expected .html got ->%s<-\n", ex)
	}
	if fp != "testdata/base-table.html" {
		t.Errorf("Expected ->testdata/base-table.html<- got ->%s<-\n", fp)
	}
}

// func RenderTemplate(mdata map[string]interface{}, fns ...string) (tmpl_rendered string, err error) {
func Test_RenderTemlate(t *testing.T) {

	mdata := map[string]interface{}{
		"a": 12,
	}

	tv, err := RenderTemplate(mdata, "base-table.html", "list-of-files/lof.html")

	fmt.Printf("%s %s\n", tv, err)

}
