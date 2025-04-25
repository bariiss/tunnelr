package cmd

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/bariiss/tunnelr/internal/client"
	"github.com/coder/websocket"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile  string
	domain   string
	port     int
	targetIP string
)

// rootCmd is executed with `tunnelr` (no sub-command needed)
var rootCmd = &cobra.Command{
	Use:   "tunnelr",
	Short: "Expose a local port through a WebSocket tunnel",
	RunE:  run,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func init() {
	home, _ := os.UserHomeDir()
	defCfg := filepath.Join(home, ".config", "tunnelr", "config.yaml")

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", defCfg, "config file")
	rootCmd.Flags().StringVarP(&domain, "domain", "d", "", "tunnel server domain (saved)")
	rootCmd.Flags().IntVarP(&port, "port", "p", 8080, "local port to expose")
	rootCmd.Flags().StringVarP(&targetIP, "target", "t", "127.0.0.1", "local host to forward (if empty defaults to 127.0.0.1)")

	cobra.OnInitialize(initConfig)
}

// initConfig loads domain from flag, env or config file
func initConfig() {
	viper.SetConfigFile(cfgFile)

	// ENV değişkenlerini içe aktar – örn. TUNNELR_DOMAIN
	viper.SetEnvPrefix("tunnelr")
	viper.AutomaticEnv()
	_ = viper.BindEnv("domain")

	// Varsa config.yaml oku (okuyamasa da sorun değil)
	_ = viper.ReadInConfig()

	// CLI flag en yüksek önceliğe sahip
	if domain != "" {
		viper.Set("domain", domain)
		_ = os.MkdirAll(filepath.Dir(cfgFile), 0o755)
		_ = viper.WriteConfigAs(cfgFile) // kaydet / güncelle
	}
}

// run is the main function that establishes a WebSocket connection to the tunnel server
func run(cmd *cobra.Command, args []string) error {
	domain = viper.GetString("domain")
	serverURL := fmt.Sprintf("wss://%s/register", domain)

	// build final URL
	u, _ := url.Parse(serverURL)
	if len(args) > 0 { // optional first arg == desired sub-domain
		q := u.Query()
		q.Set("sub", args[0])
		u.RawQuery = q.Encode()
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	conn, _, err := websocket.Dial(ctx, u.String(), &websocket.DialOptions{
		CompressionMode: websocket.CompressionDisabled,
	})
	if err != nil {
		return fmt.Errorf("dial: %w", err)
	}

	// first message = allocated host
	_, hostBytes, err := conn.Read(ctx)
	if err != nil {
		return fmt.Errorf("handshake read: %w", err)
	}
	publicURL := "https://" + string(hostBytes)
	log.Printf("🆕 public URL → %s", publicURL)

	target := fmt.Sprintf("%s:%d", targetIP, port)
	fwd := client.NewForwarder(conn, target)
	log.Printf("✅ connected — forwarding http://%s", target)

	return fwd.Serve(ctx)
}

// Execute is called from main.go
func Execute() { _ = rootCmd.Execute() }
