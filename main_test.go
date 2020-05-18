package main

import "testing"

// func PathFind(path, fn string) (full_path, file_type string, err error) {

func Test_GzipServer(t *testing.T) {

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
