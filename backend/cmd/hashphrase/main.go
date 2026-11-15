package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/gmatiasvc/air-quality-trujillo/backend/internal/auth"
)

func main() {
	saltFlag := flag.String("salt", "", "Optional hex salt (generates a random 16-byte salt if omitted)")
	phraseFlag := flag.String("phrase", "", "Passphrase to hash")
	flag.Parse()

	phrase := *phraseFlag
	if phrase == "" && flag.NArg() > 0 {
		phrase = flag.Arg(0)
	}
	if phrase == "" {
		fmt.Fprintln(os.Stderr, "Usage: go run ./cmd/hashphrase -phrase \"your secret passphrase\" [-salt hex]")
		os.Exit(1)
	}

	salt := *saltFlag
	if salt == "" {
		var err error
		salt, err = auth.GenerateRandomSaltHex()
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to generate salt: %v\n", err)
			os.Exit(1)
		}
	}

	hashHex, err := auth.DeriveHashHex(phrase, salt)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to derive Argon2id hash: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("ADMIN_PASSPHRASE_SALT=%s\n", salt)
	fmt.Printf("ADMIN_PASSPHRASE_HASH=%s\n", hashHex)
}
