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
make build          # Build CLI for current platform
make tray           # Build menu bar app (macOS only)
make all            # Build all variants for all platforms
```

## Storage

Accounts are stored encrypted at:

| OS      | Path                                          |
|---------|-----------------------------------------------|
| macOS   | `~/Library/Application Support/my-auth/`      |
| Linux   | `~/.config/my-auth/`                          |
| Windows | `%AppData%\my-auth\`                          |

## Release Notes

### v1.1.1

- Tray menu now sorts accounts by usage frequency (most used on top)

### v1.1.0

- macOS menu bar app: click the lock icon to see accounts, click an account to copy its TOTP code to clipboard
- Automatically waits for a fresh code if the current one expires in ≤4 seconds
- Native macOS notifications on code copy
- Build with `make tray` (uses build tags to keep CLI builds unaffected)

### v1.0.0

Initial release.

- Import accounts from encrypted CSV files with password-based decryption
- List stored accounts
- Generate TOTP codes with countdown timer
- Substring matching for account lookup
- Copy codes to clipboard with `-c` flag
- Add new accounts manually with base32 secret input
- Encrypted local storage (AES-256-GCM with Argon2 key derivation)
- Cross-platform builds: macOS (arm64/amd64), Linux (amd64), Windows (amd64)
