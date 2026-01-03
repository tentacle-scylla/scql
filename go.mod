module github.com/tentacle-scylla/scql

go 1.21

require (
	github.com/antlr4-go/antlr/v4 v4.13.0
	github.com/tentacle-scylla/scql/gen/cqldata v0.0.0
	github.com/tentacle-scylla/scql/gen/parser v0.0.0
	github.com/urfave/cli/v2 v2.27.1
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/cpuguy83/go-md2man/v2 v2.0.2 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/xrash/smetrics v0.0.0-20201216005158-039620a65673 // indirect
	golang.org/x/exp v0.0.0-0-20230515195305-f3d0a9c9a5cc // indirect
)

replace github.com/tentacle-scylla/scql/gen/parser => ./gen/parser

replace github.com/tentacle-scylla/scql/gen/cqldata => ./gen/cqldata
