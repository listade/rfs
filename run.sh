#!/bin/bash

killall -9 labfs

./labfs -s node.1 -p 5001 &
./labfs -s node.2 -p 5002 &
./labfs -s node.3 -p 5003 &