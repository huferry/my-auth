//go:build !tray

package main

import (
	"bufio"
	"encoding/base32"
	"encoding/base64"
	"flag"
	"fmt"
	"os"
	"strings"
	"syscall"
	"time"

	"github.com/atotto/clipboard"
	"golang.org/x/term"
)

func main() {
	listFlag := flag.Bool("l", false, "list accounts")
	accountFlag := flag.String("a", "", "show auth code for account")
	importFlag := flag.String("i", "", "import tokens from CSV file")
	copyFlag := flag.Bool("c", false, "copy code to clipboard")
	newFlag := flag.Bool("n", false, "add new account")
	flag.BoolVar(listFlag, "list", false, "list accounts")
	flag.StringVar(accountFlag, "account", "", "show auth code for account")
	flag.StringVar(importFlag, "import", "", "import tokens from CSV file")
	flag.BoolVar(copyFlag, "copy", false, "copy code to clipboard")
	flag.BoolVar(newFlag, "new", false, "add new account")

	flag.Parse()

	var err error
	switch {
	case *importFlag != "":
		err = doImport(*importFlag)
	case *listFlag:
		err = doList()
	case *newFlag:
		err = doNew()
	case *accountFlag != "":
		err = doShowCode(*accountFlag, *copyFlag)
	default:
		flag.Usage()
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func readPassword(prompt string) (string, error) {
	fmt.Print(prompt)
	pw, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Println()
	if err != nil {
		return "", err
	}
	return string(pw), nil
}

func readLine(prompt string) (string, error) {
	fmt.Print(prompt)
	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		return "", fmt.Errorf("no input")
	}
	return strings.TrimSpace(scanner.Text()), scanner.Err()
}

func doImport(csvFile string) error {
	records, err := parseTokensCSV(csvFile)
	if err != nil {
		return fmt.Errorf("parsing CSV: %w", err)
	}

	password, err := readPassword("Password: ")
	if err != nil {
		return fmt.Errorf("reading password: %w", err)
	}

	accounts, err := importAccounts(records, password)
	if err != nil {
		return err
	}

	if err := saveAccounts(accounts); err != nil {
		return fmt.Errorf("saving accounts: %w", err)
	}

	fmt.Printf("Imported %d accounts.\n", len(accounts))
	return nil
}

func doList() error {
	accounts, err := loadAccounts()
	if err != nil {
		return err
	}

	for _, a := range accounts {
		fmt.Println(a.Name)
	}
	return nil
}

func doNew() error {
	name, err := readLine("Account name: ")
	if err != nil {
		return err
	}
	if name == "" {
		return fmt.Errorf("name cannot be empty")
	}

	secretStr, err := readLine("Secret (base32): ")
	if err != nil {
		return err
	}

	// Validate base32
	secretStr = strings.ToUpper(strings.ReplaceAll(secretStr, " ", ""))
	secret, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(secretStr)
	if err != nil {
		return fmt.Errorf("invalid base32 secret: %w", err)
	}

	accounts, err := loadAccounts()
	if err != nil {
		// If no store yet, start fresh
		accounts = nil
	}

	accounts = append(accounts, Account{
		Name:   name,
		Secret: base64.StdEncoding.EncodeToString(secret),
	})

	if err := saveAccounts(accounts); err != nil {
		return err
	}

	fmt.Printf("Added %q.\n", name)
	return nil
}

func doShowCode(search string, copyToClipboard bool) error {
	accounts, err := loadAccounts()
	if err != nil {
		return err
	}

	search = strings.ToLower(search)
	var match *Account
	for i, a := range accounts {
		if strings.Contains(strings.ToLower(a.Name), search) {
			match = &accounts[i]
			break
		}
	}

	if match == nil {
		return fmt.Errorf("account %q not found", search)
	}

	secret, err := base64.StdEncoding.DecodeString(match.Secret)
	if err != nil {
		return fmt.Errorf("decoding secret: %w", err)
	}

	code := generateTOTP(secret, time.Now())
	remaining := 30 - (time.Now().Unix() % 30)

	fmt.Printf("%s: %s (%ds remaining)\n", match.Name, code, remaining)

	if copyToClipboard {
		if err := clipboard.WriteAll(code); err != nil {
			return fmt.Errorf("copying to clipboard: %w", err)
		}
		fmt.Println("Copied to clipboard.")
	}

	return nil
}
