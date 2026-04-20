Read CLAUDE.md first. Then:

1. Initialize a Go module named `github.com/yourorg/f5xc-sec-events`
2. Create the full directory tree from CLAUDE.md's Directory Guide
3. Create a .env.example with: F5XC_API_KEY=, F5XC_TENANT=f5-sa, F5XC_NAMESPACE=s-iannetta
4. Create .gitignore that ignores: .env, *.env, bin/, vendor/
5. Create a minimal go.mod
6. In cmd/f5xc-sec/main.go create a skeleton main() that:
   - Parses flags: --window (default "1h"), --namespace, --lb, --serve, --port (default 8080)
   - Loads config from internal/config
   - Prints "F5 XC Security Events Tool" and exits cleanly
7. Make sure `go build ./...` succeeds with no errors
