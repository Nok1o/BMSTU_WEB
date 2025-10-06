#!/bin/bash

# Подсказка
if [[ "$1" == "-h" ]]; then
    echo "Использование: $0 [-p параллельность] [-r]"
    echo "  -p <число>   Количество параллельных тестов (по умолчанию 1)"
    echo "  -r           Случайный порядок тестов"
    exit 0
fi

# Параметры по умолчанию
PARALLEL=1
SHUFFLE=false

while getopts "p:r" opt; do
  case $opt in
    p) PARALLEL=$OPTARG ;;
    r) SHUFFLE=true ;;
  esac
done

# Allure
export ALLURE_OUTPUT_PATH="$(pwd)"
mkdir -p "$ALLURE_OUTPUT_PATH"

# Формируем команду
CMD="go test ./... --tags unit -v -p $PARALLEL"
$SHUFFLE && CMD="$CMD -shuffle=on"

echo "Запуск: $CMD"
$CMD | tee test.log

# Allure отчет
REPORT_DIR="$(pwd)/allure-report"
allure generate "$ALLURE_OUTPUT_PATH/allure-results" -o "$REPORT_DIR" --clean
allure open "$REPORT_DIR"