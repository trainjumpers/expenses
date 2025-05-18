[working-directory: 'server']
@install-server: 
  go install github.com/air-verse/air@latest
  go mod tidy

[working-directory: 'server']
@server:
  air
