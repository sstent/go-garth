#!/bin/bash

set -e

echo "--- Running End-to-End CLI Tests ---"

echo "Testing garth --help"
go run go-garth/cmd/garth --help

echo "Testing garth auth status"
go run go-garth/cmd/garth auth status

echo "Testing garth activities list"
go run go-garth/cmd/garth activities list --limit 5

echo "Testing garth health sleep"
go run go-garth/cmd/garth health sleep --from 2024-01-01 --to 2024-01-02

echo "Testing garth stats distance"
go run go-garth/cmd/garth stats distance --year

echo "Testing garth health vo2max"
go run go-garth/cmd/garth health vo2max --from 2024-01-01 --to 2024-01-02

echo "Testing garth health hr-zones"
go run go-garth/cmd/garth health hr-zones

echo "--- End-to-End CLI Tests Passed ---"