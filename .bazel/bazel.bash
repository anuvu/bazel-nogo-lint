#!/bin/bash

bzl() {
	TOPLEVEL=$(git rev-parse --show-toplevel);
	pushd .;
	cd ${TOPLEVEL};
	make -f Makefile.bazel build;
	popd
}
