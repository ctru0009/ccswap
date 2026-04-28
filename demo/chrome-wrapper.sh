#!/bin/bash
exec "$(dirname "$0")/chrome" --no-sandbox "$@"
