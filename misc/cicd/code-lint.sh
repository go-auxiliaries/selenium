#!/bin/sh

set -e

st=0
for pkg in $(go list ./...); do
  local_pkg=$( echo "$pkg" | grep -oP "^github.com/go-auxiliaries/selenium/\K.*" || true)
  [ -z "$local_pkg" ] && continue

  echo "goimports -d $local_pkg"
  out=$(goimports -d "$local_pkg" 2>&1)
  echo "$out"
  [ -n "$out" ] && st=1 && break

  echo "gofmt -l -s $local_pkg"
  gofmt -l -s "$local_pkg"
  [ $? -ne 0 ] && st=1 && break

  echo "go vet $pkg/..."
  go vet "$pkg"/...
  [ $? -ne 0 ] && st=1 && break

  echo "golangci-lint run $local_pkg/..."
  golangci-lint run "$local_pkg"/...
  [ $? -ne 0 ] && st=1 && break
done
[ $st -ne 0 ] && exit $st

echo "goimports -d ./*.go"
out=$(goimports -d ./*.go 2>&1)
echo "$out"
[ -n "$out" ] && st=1 && exit $st

echo "gofmt -l -s ./*.go"
gofmt -l -s ./*.go
[ $? -ne 0 ] && st=1 && exit $st

echo "go vet ./*.go"
go vet ./*.go
[ $? -ne 0 ] && st=1 && exit $st

echo "golangci-lint run ./*.go"
golangci-lint run ./*.go
[ $? -ne 0 ] && st=1 && exit $st

exit $st
