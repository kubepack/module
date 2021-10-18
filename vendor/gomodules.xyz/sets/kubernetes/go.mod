module gomodules.xyz/sets/kubernetes

go 1.17

require (
	gomodules.xyz/sets v0.2.0
	k8s.io/apimachinery v0.21.1
)

require github.com/gogo/protobuf v1.3.2 // indirect

replace gomodules.xyz/sets => ./..
