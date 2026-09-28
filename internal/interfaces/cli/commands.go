package cli

import (
	"context"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"os"
	"strings"

	"github.com/forgedash/forgedash/internal/application/ports"
	"github.com/forgedash/forgedash/internal/application/usecase"
	"github.com/forgedash/forgedash/internal/domain/account"
	"github.com/forgedash/forgedash/internal/infrastructure/auth"
	"github.com/forgedash/forgedash/internal/infrastructure/config"
	"github.com/forgedash/forgedash/internal/infrastructure/git"
	"github.com/forgedash/forgedash/internal/infrastructure/provider/forgejo"
	"github.com/forgedash/forgedash/internal/infrastructure/provider/github"
	"github.com/forgedash/forgedash/internal/infrastructure/provider/gitlab"
	"github.com/forgedash/forgedash/internal/infrastructure/provider/mock"
	"github.com/forgedash/forgedash/internal/interfaces/tui/screens"
)

var (
	cfgFile       string
	demoMode      bool
	scopedAccount string
	initialView   string
	forceUnified  bool

	Version   = "v0.1.0"
	CommitSHA = "dev"
	BuildDate = "unknown"
)

var RootCmd = &cobra.Command{
	Use:   "forgedash",
	Short: "Standalone multi-forge terminal dashboard for GitHub, GitLab, and Codeberg",
	Long: `ForgeDash brings the beloved gh-dash experience to every Git forge.
Aggregate and manage Pull Requests, Merge Requests, and Issues across
GitHub (github.com & GHE), GitLab (gitlab.com & self-hosted), and
Codeberg / Forgejo in one fast, keyboard-driven terminal dashboard.`,
	RunE: runDashboard,
}

func init() {
	RootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file path (default is $XDG_CONFIG_HOME/forgedash/config.yaml)")
	RootCmd.Flags().BoolVar(&demoMode, "demo", false, "run in offline demo mode with multi-forge mock data")
	RootCmd.Flags().StringVar(&scopedAccount, "account", "", "start scoped to a specific account ID")
	RootCmd.Flags().StringVar(&initialView, "view", "", "initial view mode: prs, issues, notifications")
	RootCmd.Flags().BoolVar(&forceUnified, "unified", false, "force unified view across all accounts")

	RootCmd.AddCommand(newAuthCmd())
	RootCmd.AddCommand(newAccountCmd())
	RootCmd.AddCommand(newConfigCmd())
	RootCmd.AddCommand(newDoctorCmd())
	RootCmd.AddCommand(newVersionCmd())
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func runDashboard(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load(cfgFile)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	if initialView != "" {
		cfg.DefaultView = initialView
	}
	if forceUnified {
		cfg.DefaultMode = "unified"
	}
	if scopedAccount != "" {
		cfg.DefaultMode = "scoped"
		cfg.ActiveAccount = scopedAccount
	}

	var clients []ports.ForgeClient

	if demoMode {
		clients = mock.BuildDemoClients()
	} else {
		credStore, err := auth.NewStore("")
		if err != nil {
			return fmt.Errorf("failed to initialize credential store: %w", err)
		}

		for _, acc := range cfg.Accounts {
			token, _ := credStore.Get(acc.AuthRef)
			if token == "" {
				token, _ = credStore.Get(string(acc.Provider))
			}

			switch acc.Provider {
			case account.ForgeGitHub:
				clients = append(clients, github.NewClient(acc, token))
			case account.ForgeGitLab:
				clients = append(clients, gitlab.NewClient(acc, token))
			case account.ForgeForgejo:
				clients = append(clients, forgejo.NewClient(acc, token))
			}
		}

		if len(clients) == 0 {
			fmt.Println("No accounts configured yet!")
			fmt.Println("• Run 'forgedash --demo' to test the full dashboard with mock data.")
			fmt.Println("• Run 'forgedash auth login' to connect an account.")
			fmt.Println("• Run 'forgedash config init' to create a configuration file.")
			return nil
		}
	}

	gitSvc := git.NewService("")
	dashboardUC := usecase.NewDashboardUseCase(clients)
	changeUC := usecase.NewChangeActionUseCase(dashboardUC, gitSvc)
	issueUC := usecase.NewIssueActionUseCase(dashboardUC)

	model := screens.NewDashboardModel(cfg, dashboardUC, changeUC, issueUC)
	p := tea.NewProgram(model, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		return fmt.Errorf("error running dashboard: %w", err)
	}
	return nil
}

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print ForgeDash version and build information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("ForgeDash %s (%s) built %s\n", Version, CommitSHA, BuildDate)
		},
	}
}

func newDoctorCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Check system prerequisites, network connectivity, and credentials",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(cfgFile)
			if err != nil {
				return err
			}

			credStore, _ := auth.NewStore("")
			var clients []ports.ForgeClient
			for _, acc := range cfg.Accounts {
				token, _ := credStore.Get(acc.AuthRef)
				switch acc.Provider {
				case account.ForgeGitHub:
					clients = append(clients, github.NewClient(acc, token))
				case account.ForgeGitLab:
					clients = append(clients, gitlab.NewClient(acc, token))
				case account.ForgeForgejo:
					clients = append(clients, forgejo.NewClient(acc, token))
				}
			}

			gitSvc := git.NewService("")
			docUC := usecase.NewDoctorUseCase(clients, gitSvc)
			report := docUC.RunDiagnostics(context.Background())

			fmt.Println("ForgeDash Health & Environment Diagnostics:")
			fmt.Println("-------------------------------------------")
			for _, item := range report.Items {
				status := "[✓ PASS]"
				if !item.Passed {
					status = "[✗ FAIL]"
				}
				fmt.Printf("%-8s %-25s: %s\n", status, item.Name, item.Message)
			}
			fmt.Println("-------------------------------------------")
			if report.AllValid {
				fmt.Println("All diagnostic checks passed. System ready!")
			} else {
				fmt.Println("Some checks failed or have warnings. Review suggestions above.")
			}
			return nil
		},
	}
}

func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage ForgeDash configuration file",
	}

	var customPath string
	initCmd := &cobra.Command{
		Use:   "init",
		Short: "Generate default commented ~/.config/forgedash/config.yaml",
		RunE: func(cmd *cobra.Command, args []string) error {
			path := customPath
			if path == "" {
				path = config.DefaultConfigPath()
			}
			if err := config.Init(path); err != nil {
				return err
			}
			fmt.Printf("Created default configuration at %s\n", path)
			return nil
		},
	}
	initCmd.Flags().StringVar(&customPath, "path", "", "custom destination path for config")
	cmd.AddCommand(initCmd)

	return cmd
}

func newAccountCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "account",
		Short: "Manage configured forge accounts",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List all configured accounts",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(cfgFile)
			if err != nil {
				return err
			}
			if len(cfg.Accounts) == 0 {
				fmt.Println("No accounts configured yet. Use 'forgedash auth login'.")
				return nil
			}
			fmt.Printf("%-15s %-10s %-20s %s\n", "ACCOUNT ID", "PROVIDER", "NAME", "HOST")
			for _, a := range cfg.Accounts {
				fmt.Printf("%-15s %-10s %-20s %s\n", a.ID, a.Provider, a.Name, a.Host)
			}
			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "switch <account-id>",
		Short: "Switch the active account in scoped mode",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(cfgFile)
			if err != nil {
				return err
			}
			target := args[0]
			found := false
			for _, a := range cfg.Accounts {
				if string(a.ID) == target {
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("account %s not found in configuration", target)
			}
			cfg.ActiveAccount = target
			cfg.DefaultMode = "scoped"
			if err := config.Save(cfgFile, cfg); err != nil {
				return err
			}
			fmt.Printf("Active account switched to %s (scoped mode)\n", target)
			return nil
		},
	})

	return cmd
}

func newAuthCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Manage authentication tokens",
	}

	var (
		provider string
		host     string
		token    string
		accName  string
	)

	loginCmd := &cobra.Command{
		Use:   "login",
		Short: "Authenticate with a Git forge account",
		RunE: func(cmd *cobra.Command, args []string) error {
			if provider == "" {
				provider = "github"
			}
			if host == "" {
				host = account.DefaultHost(account.ForgeType(provider))
			}
			if accName == "" {
				accName = fmt.Sprintf("%s (%s)", provider, host)
			}
			if token == "" {
				return fmt.Errorf("--token flag is required for non-interactive login")
			}

			store, err := auth.NewStore("")
			if err != nil {
				return err
			}

			accID := fmt.Sprintf("%s-%s", provider, strings.ReplaceAll(host, ".", "-"))
			if err := store.Set(accID, token); err != nil {
				return fmt.Errorf("failed to save credential: %w", err)
			}

			cfg, err := config.Load(cfgFile)
			if err != nil {
				return err
			}

			newAcc := account.Account{
				ID:       account.AccountID(accID),
				Name:     accName,
				Provider: account.ForgeType(provider),
				Host:     host,
				AuthRef:  accID,
			}

			updated := false
			for i, a := range cfg.Accounts {
				if a.ID == newAcc.ID {
					cfg.Accounts[i] = newAcc
					updated = true
					break
				}
			}
			if !updated {
				cfg.Accounts = append(cfg.Accounts, newAcc)
			}

			if err := config.Save(cfgFile, cfg); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}

			fmt.Printf("Successfully logged into %s as %s!\n", host, accName)
			return nil
		},
	}

	loginCmd.Flags().StringVarP(&provider, "provider", "p", "github", "forge provider (github, gitlab, forgejo)")
	loginCmd.Flags().StringVar(&host, "host", "", "forge hostname (default: provider default host)")
	loginCmd.Flags().StringVarP(&token, "token", "t", "", "authentication personal access token")
	loginCmd.Flags().StringVarP(&accName, "name", "n", "", "descriptive account name")

	cmd.AddCommand(loginCmd)

	cmd.AddCommand(&cobra.Command{
		Use:   "logout <account-id>",
		Short: "Remove authentication credentials for an account",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			accID := args[0]
			store, err := auth.NewStore("")
			if err != nil {
				return err
			}
			_ = store.Delete(accID)

			cfg, err := config.Load(cfgFile)
			if err != nil {
				return err
			}
			var remaining []account.Account
			for _, a := range cfg.Accounts {
				if string(a.ID) != accID {
					remaining = append(remaining, a)
				}
			}
			cfg.Accounts = remaining
			_ = config.Save(cfgFile, cfg)

			fmt.Printf("Logged out account %s.\n", accID)
			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "status",
		Short: "Display authentication status across accounts",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(cfgFile)
			if err != nil {
				return err
			}
			credStore, _ := auth.NewStore("")
			for _, acc := range cfg.Accounts {
				tok, _ := credStore.Get(acc.AuthRef)
				hasTok := "Token configured"
				if tok == "" {
					hasTok = "NO TOKEN"
				}
				fmt.Printf("• %s (%s, %s): %s\n", acc.Name, acc.Provider, acc.Host, hasTok)
			}
			return nil
		},
	})

	return cmd
}
