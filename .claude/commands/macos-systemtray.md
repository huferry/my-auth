# Make the application available from macOS menu bar (system tray)

Add a macOS menu bar extra that shows accounts in a dropdown. Clicking an account copies its TOTP code to clipboard. If the code is valid for ≤4 seconds, wait for the next period before copying.

## Architecture

- Use **build tags** to separate CLI and tray entry points (not a flag), so the systray/CGo dependency doesn't affect CLI or cross-platform builds.
- Library: `fyne.io/systray` (maintained fork of getlantern/systray).
- Notification: `osascript display notification` (no extra dependency).
- Icon: 22x22 black-and-alpha PNG embedded via `//go:embed` — use `SetTemplateIcon` so macOS adapts for light/dark mode.

## Steps

1. **Add build constraint to `main.go`**: Insert `//go:build !tray` as the first line so it's excluded when building with `-tags tray`.

2. **Create `icon.png`**: Generate a 22x22 monochrome lock icon (black pixels + alpha transparency) with Python. This is a macOS template image.

3. **Create `icon.go`** (build tag `tray`): Uses `//go:embed icon.png` to embed the icon as `var appIcon []byte`.

4. **Create `tray.go`** (build tag `tray`): Contains:
   - `main()` → calls `systray.Run(onReady, onExit)`
   - `onReady()` → sets template icon, tooltip, calls `buildMenu()`
   - `buildMenu()` → calls `loadAccounts()`, adds a `MenuItem` per account with a goroutine on `ClickedCh` that calls `copyCode()`. Adds separator, "Refresh" (calls `systray.ResetMenu()` + `buildMenu()`), and "Quit".
   - `copyCode(secretB64)` → decodes secret, checks remaining seconds, if ≤4 sleeps until next period, generates TOTP, copies to clipboard, shows notification.
   - `showNotification(message)` → runs `osascript -e 'display notification "..." with title "my-auth"'`.

5. **Update `Makefile`**: Add `tray` target (`go build -tags tray -o my-auth-tray .`). Add macOS tray builds (arm64 + amd64) to `all` target. Add `$(APP)-tray` to `clean`.

6. **Add dependency**: `go get fyne.io/systray`

7. **Verify both builds**: `go build -o my-auth .` (CLI) and `go build -tags tray -o my-auth-tray .` (tray) must both succeed. Test CLI still works with `./my-auth -l`.
