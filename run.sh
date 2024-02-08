#!/bin/bash
set -a # export all variables from .env
source .env
set +a

go run .
