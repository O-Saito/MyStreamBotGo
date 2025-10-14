compile: go build -ldflags="-X 'main.Version=1.0.0' -X 'main.BuildDate=$(date)' -X 'main.CommitHash=$(git rev-parse HEAD)'" -o build/mystreambot.exe .
build: go run .