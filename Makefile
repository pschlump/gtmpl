
all:
	go build

build_linux:
		GOOS=linux go build -o gtmpl.linux . 

test: test001
	@echo PASS

test001:
	go build
	./gtmpl --cli '{"a":4}' --cli '{"b":8}' --data t1.json --cli '{"a":22}' --debug echo_input --tmpl ./test1 --out ./out > out/test001.out
	diff out/test001.out ref/test001.out

test002:
	go build
	./gtmpl --cli '{"a":4}' --cli '{"b":8}' --data t1.json --cli '{"a":22}' --debug echo_input --debug file_name --tmpl ./test1 --out ./out > out/test002.out
	diff out/test002.out ref/test002.out

#contract {{.Name}}Token is StandardToken {				
#	string public constant NAME = "{{.Name_UC}} Token";
test003_old:
	go build
	./gtmpl --cli '{"Name":"SampleCorp","Name_UC":"SAMPLECORP"}' --tmpl ./test2 --out ./out 

test003:
	go build
	./gtmpl --data test3.json --tmpl ./test2 --out ./out 

run_ex1:
	./gtmpl --cli '{"type":"string"}' --tmpl ./ex1 --out ./out 
	( cd out ; goimports -w *.go )
	( cd out ; go build )

run_ex1_good:
	go build
	./gtmpl --cli '{"type":"string"}' --data t1.json --debug proc_file --debug echo_input --debug file_name --tmpl ./ex1 --out ./out > out/ex1.001.out

blk00:
	go build
	./gtmpl --cli '{"a":4}' --cli '{"b":8}' --data t1.json --cli '{"a":22}' --debug echo_input --debug file_name --tmpl ./test1 --out ./out > out/test002.out
	diff out/test002.out ref/test002.out

blk01:
	go build
	./gtmpl --cli '{"a":4}' --cli '{"b":8}' --data t1.json --cli '{"a":22}' --debug echo_input --debug file_name --tmpl-list ./testdata/blk01/blk01.tmpl,./testdata/blk01/redef.tmpl --out ./out/blk01.out 
	diff out/blk01.out ref/blk01.out
