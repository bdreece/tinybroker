#!/bin/bash

protoc -I=. --go_out=../../.. packet.proto
