package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var (
	systemdUser    string
	systemdGroup   string
	systemdWorkDir string
	systemdPort    int
)

// systemdCmd represents the systemd command
var systemdCmd = &cobra.Command{
	Use:   "systemd",
	Short: "Generate systemd service configuration",
	Long:  `Generate a systemd service unit file for running Zipic as a Linux service.`,
	Run:   generateSystemd,
}

func init() {
	rootCmd.AddCommand(systemdCmd)

	systemdCmd.Flags().StringVarP(&systemdUser, "user", "u", "www", "user to run the service")
	systemdCmd.Flags().StringVarP(&systemdGroup, "group", "g", "www", "group to run the service")
	systemdCmd.Flags().StringVarP(&systemdWorkDir, "workdir", "w", "/opt/zipic", "working directory")
	systemdCmd.Flags().IntVarP(&systemdPort, "port", "p", 8040, "service port")
}

func generateSystemd(cmd *cobra.Command, args []string) {
	// Get executable name
	execName := "zipic"
	if len(args) > 0 {
		execName = args[0]
	}

	// Generate service file
	serviceContent := fmt.Sprintf(`[Unit]
Description=Zipic Image Compression Service
Documentation=https://github.com/yezihack/zipic
After=network.target

[Service]
Type=simple
User=%s
Group=%s
WorkingDirectory=%s
ExecStart=%s/%s --port %d
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal
LimitNOFILE=65536

# Environment variables
Environment=ZIPIC_SERVER_ADDRESS=127.0.0.1
Environment=ZIPIC_SERVER_PORT=%d
Environment=ZIPIC_SERVER_MODE=release

[Install]
WantedBy=multi-user.target
`, systemdUser, systemdGroup, systemdWorkDir, systemdWorkDir, execName, systemdPort, systemdPort)

	// Print the service file
	fmt.Println("Generated systemd service file:")
	fmt.Println(strings.Repeat("-", 50))
	fmt.Println(serviceContent)
	fmt.Println(strings.Repeat("-", 50))

	// Print installation instructions
	fmt.Println("\nInstallation instructions:")
	fmt.Println("1. Save the above content to: /etc/systemd/system/zipic.service")
	fmt.Println("2. Run: sudo systemctl daemon-reload")
	fmt.Println("3. Run: sudo systemctl enable zipic")
	fmt.Println("4. Run: sudo systemctl start zipic")
	fmt.Println("5. Check status: sudo systemctl status zipic")
	fmt.Println("6. View logs: sudo journalctl -u zipic -f")

	// Optionally save to file
	saveFlag, _ := cmd.Flags().GetBool("save")
	if saveFlag {
		savePath := filepath.Join(systemdWorkDir, "zipic.service")
		if err := os.WriteFile(savePath, []byte(serviceContent), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to save service file: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("\nService file saved to: %s\n", savePath)
	}
}