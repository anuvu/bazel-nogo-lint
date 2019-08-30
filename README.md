bazel-nogo-lint
===============

[bazel](https://bazel.build/) is a build system which supports incremental repeatable builds with
support for various languages and environments.

This project provides linter support [specifically for golang](https://github.com/bazelbuild/rules_go). 
bazel's way is to parse the source code and share that compiler state in the
[nogo](https://github.com/bazelbuild/rules_go/blob/master/go/nogo.rst) hook.

[golangci-lint](https://github.com/golangci/golangci-lint) has similar goals and
optimizations but is not compatible with _nogo_ directly.

This project attempts that port.
