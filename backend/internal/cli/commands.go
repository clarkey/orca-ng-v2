package cli

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/orca-ng/orca/internal/database"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/term"
)

func NewLoginCmd() *cobra.Command {
	var serverURL string

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Login to ORCA server",
		Long:  `Authenticate with an ORCA server and save session credentials.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Prompt for username
			fmt.Print("Username: ")
			var username string
			fmt.Scanln(&username)

			// Prompt for password (hidden)
			fmt.Print("Password: ")
			passwordBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
			if err != nil {
				return fmt.Errorf("failed to read password: %w", err)
			}
			fmt.Println() // New line after password input
			password := string(passwordBytes)

			// Create client
			client := NewClient(serverURL)

			// Attempt login
			fmt.Println("Logging in...")
			loginResp, err := client.Login(username, password)
			if err != nil {
				return fmt.Errorf("login failed: %w", err)
			}

			// Save session
			store, err := NewSessionStore()
			if err != nil {
				return fmt.Errorf("failed to create session store: %w", err)
			}

			session := &SessionInfo{
				Token:     loginResp.Token,
				ExpiresAt: loginResp.ExpiresAt,
				ServerURL: serverURL,
				Username:  loginResp.User.Username,
			}

			if err := store.Save(session); err != nil {
				return fmt.Errorf("failed to save session: %w", err)
			}

			fmt.Printf("Successfully logged in as %s\n", loginResp.User.Username)
			if loginResp.User.IsAdmin {
				fmt.Println("Admin privileges: Yes")
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&serverURL, "server", "s", "http://localhost:8080", "ORCA server URL")

	return cmd
}

func NewLogoutCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "Logout from ORCA server",
		Long:  `Clear stored session credentials.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := NewSessionStore()
			if err != nil {
				return fmt.Errorf("failed to create session store: %w", err)
			}

			// Load existing session
			session, err := store.Load()
			if err != nil {
				return fmt.Errorf("failed to load session: %w", err)
			}

			if session != nil {
				// Try to logout from server
				client := NewClient(session.ServerURL)
				client.SetToken(session.Token)
				
				if err := client.Logout(); err != nil {
					// Don't fail if server logout fails, just warn
					fmt.Printf("Warning: Failed to logout from server: %v\n", err)
				}
			}

			// Delete local session
			if err := store.Delete(); err != nil {
				return fmt.Errorf("failed to delete session: %w", err)
			}

			fmt.Println("Successfully logged out")
			return nil
		},
	}
}

func NewStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show current session status",
		Long:  `Display information about the current session.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := NewSessionStore()
			if err != nil {
				return fmt.Errorf("failed to create session store: %w", err)
			}

			session, err := store.Load()
			if err != nil {
				return fmt.Errorf("failed to load session: %w", err)
			}

			if session == nil {
				fmt.Println("Not logged in")
				return nil
			}

			fmt.Printf("Server: %s\n", session.ServerURL)
			fmt.Printf("Username: %s\n", session.Username)
			fmt.Printf("Session expires: %s\n", session.ExpiresAt.Local().Format(time.RFC3339))

			// Check if session is still valid
			client := NewClient(session.ServerURL)
			client.SetToken(session.Token)

			user, err := client.GetCurrentUser()
			if err != nil {
				fmt.Println("Session status: Invalid (failed to verify)")
				return nil
			}

			fmt.Println("Session status: Valid")
			if user.IsAdmin {
				fmt.Println("Admin privileges: Yes")
			} else {
				fmt.Println("Admin privileges: No")
			}

			return nil
		},
	}
}

func NewUserCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "user",
		Short: "Manage ORCA users",
		Long:  `Commands for managing ORCA users.`,
	}

	// Add subcommands
	cmd.AddCommand(newUserListCmd())
	cmd.AddCommand(newUserCreateCmd())
	cmd.AddCommand(newUserResetPasswordCmd())

	return cmd
}

func newUserListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all users",
		Long:  `Display a list of all ORCA users.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Load session
			session, err := loadSession()
			if err != nil {
				return err
			}

			// TODO: Implement user list API call
			fmt.Println("User list functionality not yet implemented")
			_ = session
			return nil
		},
	}
}

func newUserCreateCmd() *cobra.Command {
	var username string
	var isAdmin bool

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new user",
		Long:  `Create a new ORCA user.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Load session
			session, err := loadSession()
			if err != nil {
				return err
			}

			// TODO: Implement user create API call
			fmt.Println("User create functionality not yet implemented")
			_ = session
			_ = username
			_ = isAdmin
			return nil
		},
	}

	cmd.Flags().StringVarP(&username, "username", "u", "", "Username for the new user")
	cmd.Flags().BoolVarP(&isAdmin, "admin", "a", false, "Grant admin privileges")
	cmd.MarkFlagRequired("username")

	return cmd
}

func newUserResetPasswordCmd() *cobra.Command {
	var username string
	var password string

	cmd := &cobra.Command{
		Use:   "reset-password",
		Short: "Reset a user's password",
		Long:  `Reset the password for an ORCA user. This command requires admin privileges.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Prompt for password if not provided
			if password == "" {
				fmt.Print("New password: ")
				passwordBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
				if err != nil {
					return fmt.Errorf("failed to read password: %w", err)
				}
				fmt.Println()
				password = string(passwordBytes)

				// Confirm password
				fmt.Print("Confirm password: ")
				confirmBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
				if err != nil {
					return fmt.Errorf("failed to read password confirmation: %w", err)
				}
				fmt.Println()

				if string(confirmBytes) != password {
					return fmt.Errorf("passwords do not match")
				}
			}

			// For local admin password reset (when not logged in)
			if username == "admin" {
				// Check if we're running with the --local flag
				localReset, _ := cmd.Flags().GetBool("local")
				if localReset {
					// Direct database connection for local admin reset
					dbURL := viper.GetString("DATABASE_URL")
					if dbURL == "" {
						dbURL = os.Getenv("DATABASE_URL")
						if dbURL == "" {
							return fmt.Errorf("DATABASE_URL environment variable not set")
						}
					}

					config := database.DatabaseConfig{
						Driver: "postgres",
						DSN:    dbURL,
					}
					
					db, err := database.NewGormConnection(config)
					if err != nil {
						return fmt.Errorf("failed to connect to database: %w", err)
					}
					defer db.Close()

					ctx := context.Background()
					if err := db.UpdateUserPassword(ctx, username, password); err != nil {
						return fmt.Errorf("failed to update password: %w", err)
					}

					fmt.Printf("Successfully reset password for user: %s\n", username)
					return nil
				}
			}

			// Load session for remote password reset
			session, err := loadSession()
			if err != nil {
				return fmt.Errorf("not logged in - use 'orca-cli login' first or use --local flag for admin password reset")
			}

			// TODO: Implement remote password reset API call
			fmt.Println("Remote password reset functionality not yet implemented")
			_ = session
			return nil
		},
	}

	cmd.Flags().StringVarP(&username, "username", "u", "", "Username for password reset")
	cmd.Flags().StringVarP(&password, "password", "p", "", "New password (will prompt if not provided)")
	cmd.Flags().Bool("local", false, "Perform local database update (requires DATABASE_URL)")
	cmd.MarkFlagRequired("username")

	return cmd
}

func NewConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage ORCA CLI configuration",
		Long:  `Commands for managing ORCA CLI configuration.`,
	}

	cmd.AddCommand(newConfigShowCmd())

	return cmd
}

func newConfigShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "Show configuration",
		Long:  `Display current CLI configuration and session information.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			configDir, err := getConfigDir()
			if err != nil {
				return fmt.Errorf("failed to get config directory: %w", err)
			}

			fmt.Printf("Configuration directory: %s\n", configDir)

			// Check for session
			store, err := NewSessionStore()
			if err != nil {
				return fmt.Errorf("failed to create session store: %w", err)
			}

			session, err := store.Load()
			if err != nil {
				return fmt.Errorf("failed to load session: %w", err)
			}

			if session != nil {
				fmt.Println("\nActive session:")
				fmt.Printf("  Server: %s\n", session.ServerURL)
				fmt.Printf("  Username: %s\n", session.Username)
				fmt.Printf("  Expires: %s\n", session.ExpiresAt.Local().Format(time.RFC3339))
			} else {
				fmt.Println("\nNo active session")
			}

			return nil
		},
	}
}

func loadSession() (*SessionInfo, error) {
	store, err := NewSessionStore()
	if err != nil {
		return nil, fmt.Errorf("failed to create session store: %w", err)
	}

	session, err := store.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load session: %w", err)
	}

	if session == nil {
		return nil, fmt.Errorf("not logged in - use 'orca-cli login' first")
	}

	return session, nil
}

func init() {
	viper.SetEnvPrefix("ORCA")
	viper.AutomaticEnv()
}