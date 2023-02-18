#!/bin/bash

killall -9 labfs
rm node.*

source proto.sh
go build .

mb=$((1024*1024))

dd if=/dev/zero of=node.1 bs=$mb count=150
dd if=/dev/zero of=node.2 bs=$mb count=150
dd if=/dev/zero of=node.3 bs=$mb count=150
dd if=/dev/urandom of=test.bin bs=$mb count=300

source run.sh

cat test.bin | ./labfs -w test.bin
./labfs -r test.bin > test.bin.1

diff test.bin test.bin.1