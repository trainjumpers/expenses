server: 
  cd server && air

install-server: 
  cd server && go install github.com/air-verse/air@latest
  cd server && go mod tidy
