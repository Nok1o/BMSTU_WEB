#!/bin/bash

DOC_ROOT="/usr/local/var/www/documentation"
CSS_FILE="/usr/local/var/www/static/markdown.css"

# Функция для преобразования MD в HTML
convert_md() {
    local md_file="$1"
    local html_file="${md_file}.html"

    pandoc -s "$md_file" -c "$CSS_FILE" -o "$html_file"
    echo "Converted: $md_file -> $html_file"
}

# Первоначальное преобразование всех файлов
find "$DOC_ROOT" -name '*.md' | while read -r file; do
    convert_md "$file"
done

