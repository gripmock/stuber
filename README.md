# stuber

`stuber` is a Go package that uses the `github.com/bavix/gripmock` package to find stubs for gRPC requests. 

The `stuber` package provides a `Budgerigar` struct that contains a searcher and toggles. The searcher is a map that stores and retrieves used stubs by their UUID. The toggles are flags that control the behavior of the `Budgerigar`.

The `Budgerigar` struct has methods to search for and return stubs based on a given query. It also has a method to add stubs to the searcher.

The `stuber` package is designed to be used in conjunction with the `github.com/bavix/gripmock` package to create a mock gRPC server.

The `stuber` package is designed to be used as a dependency in other Go projects. It is released under the MIT license.
