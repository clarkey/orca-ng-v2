package main

import (
	"fmt"
	"os"
	"strconv"
	"time"
	
	"github.com/orca-ng/orca/internal/config"
	"github.com/spf13/cobra"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "orca-keygen",
		Short: "ORCA key generation utility",
		Long:  `Generate secure keys for ORCA production deployment`,
	}
	
	var allCmd = &cobra.Command{
		Use:   "all",
		Short: "Generate all required keys",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("# ORCA Security Keys")
			fmt.Println("# Generated on:", time.Now().Format(time.RFC3339))
			fmt.Println("# IMPORTANT: Save these keys securely and never commit them to version control")
			fmt.Println()
			
			fmt.Printf("ENCRYPTION_KEY=%s\n", config.GenerateSecureKey())
			fmt.Printf("SESSION_SECRET=%s\n", config.GenerateSecureKey())
			fmt.Printf("INITIAL_ADMIN_PASSWORD=%s\n", config.GenerateSecurePassword(16))
			
			fmt.Println()
			fmt.Println("# Docker Compose Usage:")
			fmt.Println("# Save the above to a .env file and reference in docker-compose.yml")
			fmt.Println()
			fmt.Println("# Kubernetes Usage:")
			fmt.Println("# kubectl create secret generic orca-secrets \\")
			fmt.Println("#   --from-literal=encryption-key=$ENCRYPTION_KEY \\")
			fmt.Println("#   --from-literal=session-secret=$SESSION_SECRET \\")
			fmt.Println("#   --from-literal=initial-admin-password=$INITIAL_ADMIN_PASSWORD")
		},
	}
	
	var encryptionKeyCmd = &cobra.Command{
		Use:   "encryption-key",
		Short: "Generate encryption key",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(config.GenerateSecureKey())
		},
	}
	
	var sessionSecretCmd = &cobra.Command{
		Use:   "session-secret",
		Short: "Generate session secret",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(config.GenerateSecureKey())
		},
	}
	
	var passwordCmd = &cobra.Command{
		Use:   "password",
		Short: "Generate secure password",
		Run: func(cmd *cobra.Command, args []string) {
			length := 16
			if len(args) > 0 {
				if l, err := strconv.Atoi(args[0]); err == nil && l > 8 {
					length = l
				}
			}
			fmt.Println(config.GenerateSecurePassword(length))
		},
	}
	
	rootCmd.AddCommand(allCmd, encryptionKeyCmd, sessionSecretCmd, passwordCmd)
	
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}