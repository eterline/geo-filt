#!/bin/bash

name="EterLine (Andrew)"
year="2025"
project="geo-filt"

LICENSE_TEXT="// Copyright (c) $year $name
// This file is part of $project.
// Licensed under the GNU AFFERO GENERAL PUBLIC LICENSE. See the LICENSE file for details.

"

find . -type f -name "*.go" | while read -r file; do
    if ! grep -q "Copyright (c) $year $name" "$file"; then
        tmpfile=$(mktemp)
        echo "$LICENSE_TEXT" > "$tmpfile"
        cat "$file" >> "$tmpfile"
        mv "$tmpfile" "$file"
        go fmt $file
        echo "LICENSE added to: $file"
    fi
done