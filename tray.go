//go:build tray

package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"time"

	"github.com/atotto/clipboard"
	"fyne.io/systray"
)

func main() {
	systray.Run(onReady, onExit)
}

func onReady() {
	systray.SetTemplateIcon(appIcon, appIcon)
	systray.SetTooltip("my-auth")
	buildMenu()
}

func onExit() {}

func buildMenu() {
	accounts, err := loadAccounts()
	if err != nil {
		item := systray.AddMenuItem("No accounts found", "")
		item.Disable()
	} else {
		usage := loadUsage()
		sort.SliceStable(accounts, func(i, j int) bool {
			return usage[accounts[i].Name] > usage[accounts[j].Name]
		})
		for _, acct := range accounts {
			item := systray.AddMenuItem(acct.Name, "Click to copy TOTP code")
			secret := acct.Secret
			name := acct.Name
			go func() {
				for range item.ClickedCh {
					copyCode(secret)
					incrementUsage(name)
				}
			}()
		}
	}

	systray.AddSeparator()

	mRefresh := systray.AddMenuItem("Refresh", "Reload accounts")
	go func() {
		for range mRefresh.ClickedCh {
			systray.ResetMenu()
			buildMenu()
		}
	}()

	mQuit := systray.AddMenuItem("Quit", "Quit my-auth")
	go func() {
		<-mQuit.ClickedCh
		systray.Quit()
	}()
}

func copyCode(secretB64 string) {
	secret, err := base64.StdEncoding.DecodeString(secretB64)
	if err != nil {
		showNotification("Error decoding secret")
		return
	}

	now := time.Now()
	remaining := 30 - (now.Unix() % 30)

	if remaining <= 4 {
		time.Sleep(time.Duration(remaining+1) * time.Second)
		now = time.Now()
	}

	code := generateTOTP(secret, now)

	if err := clipboard.WriteAll(code); err != nil {
		showNotification("Clipboard error: " + err.Error())
		return
	}

	remaining = 30 - (time.Now().Unix() % 30)
	showNotification(fmt.Sprintf("Copied: %s (%ds remaining)", code, remaining))
}

func usagePath() string {
	d, err := configDir()
	if err != nil {
		return ""
	}
	return filepath.Join(d, "usage.json")
}

func loadUsage() map[string]int {
	p := usagePath()
	if p == "" {
		return nil
	}
	data, err := os.ReadFile(p)
	if err != nil {
		return make(map[string]int)
	}
	var usage map[string]int
	if err := json.Unmarshal(data, &usage); err != nil {
		return make(map[string]int)
	}
	return usage
}

func incrementUsage(name string) {
	p := usagePath()
	if p == "" {
		return
	}
	usage := loadUsage()
	usage[name]++
	data, err := json.Marshal(usage)
	if err != nil {
		return
	}
	os.WriteFile(p, data, 0600)
}

func showNotification(message string) {
	exec.Command("osascript", "-e",
		fmt.Sprintf(`display notification "%s" with title "my-auth"`, message),
	).Run()
}
