package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	plistLabel = "com.ykpythemind.totalmix-mac-volume-key"
	plistName  = plistLabel + ".plist"
)

func plistDir() string {
	return filepath.Join(os.Getenv("HOME"), "Library", "LaunchAgents")
}

func plistPath() string {
	return filepath.Join(plistDir(), plistName)
}

func selfPath() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.EvalSymlinks(exe)
}

func installDaemon() error {
	binPath, err := selfPath()
	if err != nil {
		return fmt.Errorf("could not determine binary path: %w", err)
	}

	if err := os.MkdirAll(plistDir(), 0755); err != nil {
		return fmt.Errorf("could not create LaunchAgents dir: %w", err)
	}

	plist := strings.Join([]string{
		`<?xml version="1.0" encoding="UTF-8"?>`,
		`<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">`,
		`<plist version="1.0">`,
		`<dict>`,
		`	<key>Label</key>`,
		`	<string>` + plistLabel + `</string>`,
		`	<key>ProgramArguments</key>`,
		`	<array>`,
		`		<string>` + binPath + `</string>`,
		`	</array>`,
		`	<key>RunAtLoad</key>`,
		`	<true/>`,
		`	<key>KeepAlive</key>`,
		`	<true/>`,
		`	<key>StandardOutPath</key>`,
		`	<string>/tmp/totalmix-mac-volume-key.log</string>`,
		`	<key>StandardErrorPath</key>`,
		`	<string>/tmp/totalmix-mac-volume-key.log</string>`,
		`</dict>`,
		`</plist>`,
		``,
	}, "\n")

	if err := os.WriteFile(plistPath(), []byte(plist), 0644); err != nil {
		return fmt.Errorf("could not write plist: %w", err)
	}

	if err := exec.Command("launchctl", "load", plistPath()).Run(); err != nil {
		return fmt.Errorf("launchctl load failed: %w", err)
	}

	fmt.Printf("Installed and started.\n")
	fmt.Printf("  Binary: %s\n", binPath)
	fmt.Printf("  Plist:  %s\n", plistPath())
	fmt.Printf("  Log:    /tmp/totalmix-mac-volume-key.log\n")
	return nil
}

func uninstallDaemon() error {
	// Stop first (ignore error if not loaded)
	exec.Command("launchctl", "unload", plistPath()).Run()

	if err := os.Remove(plistPath()); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("could not remove plist: %w", err)
	}

	fmt.Println("Uninstalled. Binary not removed — delete it manually if needed.")
	return nil
}
