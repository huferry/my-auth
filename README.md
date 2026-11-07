# My Auth

A private second-factor authenticator CLI.

## Usage

```
my-auth -i <file>       Import tokens from an encrypted CSV file
my-auth -l              List accounts
my-auth -a <name>       Show TOTP code for an account (substring match)
my-auth -a <name> -c    Show TOTP code and copy to clipboard
my-auth -n              Add a new account manually
```

## Import

Import tokens from a CSV file with the format `name,encrypted_seed,salt,iv`:

```
my-auth -i tokens.csv
```

You will be prompted for the encryption password. After import, no password is needed for subsequent operations.

## Build

```
make build          # Build for current platform
make all            # Build for all platforms (macOS arm64/amd64, Linux amd64, Windows amd64)
```

## Storage

Accounts are stored encrypted at:

| OS      | Path                                          |
|---------|-----------------------------------------------|
| macOS   | `~/Library/Application Support/my-auth/`      |
| Linux   | `~/.config/my-auth/`                          |
| Windows | `%AppData%\my-auth\`                          |
