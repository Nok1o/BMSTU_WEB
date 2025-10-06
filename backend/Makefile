# Устанавливаем переменные для путей
GO_FILES=$(shell go list ./... | grep -v 'mocks' | grep -v 'proto')
EASYJSON_FILES=$(shell find gateway/internal/delivery/http/forms -name '*_easyjson.go')
MOCKGEN=go run github.com/golang/mock/mockgen
DELIVERY_PATH=internal/delivery
USECASE_PATH=internal/usecase
REPOSITORY_PATH=internal/repository
MODEL_PATH=gateway/internal/delivery/http/forms
EASYJSON=go run github.com/mailru/easyjson/easyjson

coverage.out.clean: coverage.out
	 cat coverage.out | grep -v '_easyjson.go:' > coverage.filtered.out


# Генерация отчета покрытия тестами
coverage.out:
	go test $(GO_FILES) -coverprofile=coverage.out
#-coverpkg=$(shell echo $(GO_FILES) | sed 's/\([^ ]*\)/\1,.../g') --exclude=$(EASYJSON_FILES)

# Выводить отчет по покрытию
.PHONY: summarize-coverage
summarize-coverage: coverage.out.clean
	go tool cover -func=coverage.filtered.out


# Генерация моков
.PHONY: mockgen
mockgen:
	sh ./generate_mocks.sh

# Очистка сгенерированных easyjson файлов
clean-easyjson:
	rm -f $(MODEL_PATH)/*_easyjson.go

# Генерация файлов easyjson
.PHONY: easyjson
easyjson: clean-easyjson
	$(EASYJSON) $(shell find $(MODEL_PATH) -name "*.go" ! -name "*_test.go")

# Генерация документации с помощью swag
gendock:
	swag init --parseDependency --parseInternal --parseDepth=2 --output ../docs/docs.go --dir .
