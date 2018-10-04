# gtmpl
Template generate code

How to Use
===================

gtmpl maps a set of JSON data across a set of templates.

The data can be supplied in a configuration file with the `--data` flag.

The data can also be supplied on the command line with the `--cli` flag.

Example:

```

	$ ./gtmpl --cli '{"a":4}' --cli '{"b":8}' --data t1.json --cli '{"a":22}' \
		--debug echo_input --tmpl ./test1 --out ./out > out/test001.out

```

The data is merged in the order it is specified.  In the above example "a" will end being a numeric value of 22.

The templates are specified using the `--tmpl` flag. If you specify a file it will use a single template file. If you
specific a directory all the `*.tmpl` in the directory will be processed. It will **not** at at this time recursively
traverse decent the directory tree.

Each file will be placed in the `--out` directory with the .tmpl stripped off.   This means that `abc.go.tmpl` will
end up being `abc.go`.

An XML example
---------------

```

	$ ./gtmpl --cli '{"eventUUID":"99a2199d-473c-41b4-6e3a-794102710c90"}' \
		--data sampleAnimal.json --debug echo_input --tmpl ./test_xml \
		--out ./out > out/test003.out

```


A Sort of Realistic Example
----------------------------

The task is to substitute `string` in for a data type and generate the data type.
The template is GO code with an unknown data type inside the struct.

The template has the struct in it.

```
type DLL{{.type}} struct {
	next, prev *DLL{{.type}}
	data {{.type}}
}
```

	

