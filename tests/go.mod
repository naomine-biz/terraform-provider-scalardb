module github.com/scalar-labs/terraform-provider-scalardb/tests

go 1.24.1

replace github.com/scalar-labs/terraform-provider-scalardb => ../

require (
	github.com/scalar-labs/terraform-provider-scalardb v0.0.0-00010101000000-000000000000
	google.golang.org/grpc v1.71.1
)

require (
	golang.org/x/net v0.34.0 // indirect
	golang.org/x/sys v0.30.0 // indirect
	golang.org/x/text v0.22.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250115164207-1a7da9e5054f // indirect
	google.golang.org/protobuf v1.36.6 // indirect
)
