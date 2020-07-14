#!/bin/bash

PWD=$(pwd)
export CANIS_ROOT=$PWD
export PATH=$PWD/bin:$PATH

export CGO_CFLAGS=-I"${CANIS_ROOT}"/include
