run: build
	@./bin/app

build:
	@go build -o bin/app cmd/app/main.go

css:
	@./tailwindcss -i views/css/app.css -o public/styles.css --watch

templ:
	@templ generate --watch --proxy=http://localhost:8818

ngrok:
	@ngrok http http://localhost:8818

browse:
	@boltbrowser data/leeg.db

test:
	@go test ./model