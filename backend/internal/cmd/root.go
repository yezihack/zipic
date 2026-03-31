package cmd

import (
	"fmt"
	"io/fs"
	"net/http"
	"os"

	"zipic/internal/config"
	"zipic/internal/handler"
	"zipic/internal/version"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile    string
	port       int
	address    string
	mode       string
	staticFS   fs.FS // Embedded frontend filesystem
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "zipic",
	Short: "Zipic - Image compression service",
	Long:  `Zipic is a high-performance image compression service supporting JPG, PNG, and WebP formats.`,
	Run:   runServer,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// The staticFS parameter is the embedded frontend filesystem (optional).
func Execute(embeddedFS fs.FS) {
	staticFS = embeddedFS
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Persistent flags (available to all subcommands)
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default is ./config.yaml)")

	// Local flags (only for this command)
	rootCmd.Flags().IntVarP(&port, "port", "p", 8040, "server port")
	rootCmd.Flags().StringVarP(&address, "address", "a", "0.0.0.0", "server bind address")
	rootCmd.Flags().StringVarP(&mode, "mode", "m", "release", "gin mode (debug/release/test)")

	// Bind flags to viper
	viper.BindPFlag("server.port", rootCmd.Flags().Lookup("port"))
	viper.BindPFlag("server.address", rootCmd.Flags().Lookup("address"))
	viper.BindPFlag("server.mode", rootCmd.Flags().Lookup("mode"))
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Search for config in current directory
		viper.AddConfigPath(".")
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
	}

	// Read in environment variables that match
	viper.SetEnvPrefix("ZIPIC")
	viper.AutomaticEnv()

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

func runServer(cmd *cobra.Command, args []string) {
	// Load config
	cfg := config.Load()

	// Get port (priority: flag > env > config > default)
	serverPort := viper.GetInt("server.port")
	if serverPort == 0 {
		serverPort = cfg.Server.Port
	}

	// Get address
	serverAddress := viper.GetString("server.address")
	if serverAddress == "" {
		serverAddress = cfg.Server.Address
	}

	// Get mode
	serverMode := viper.GetString("server.mode")
	if serverMode == "" {
		serverMode = cfg.Server.Mode
	}

	// Set gin mode
	gin.SetMode(serverMode)

	// Create router
	r := gin.Default()

	// Configure CORS
	r.Use(corsMiddleware())

	// Create handlers
	imageHandler := handler.NewImageHandler()

	// Start cleanup scheduler
	go handler.StartCleanupScheduler(imageHandler.UploadDir())

	// Register routes
	handler.RegisterRoutes(r, imageHandler)

	// Serve static files (SPA) if embedded FS is available
	if staticFS != nil {
		setupSPA(r, staticFS)
	}

	// Start server
	addr := fmt.Sprintf("%s:%d", serverAddress, serverPort)
	fmt.Printf("Zipic %s starting on http://%s\n", version.Version, addr)
	if err := r.Run(addr); err != nil {
		fmt.Fprintln(os.Stderr, "Failed to start server:", err)
		os.Exit(1)
	}
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func setupSPA(r *gin.Engine, staticFS fs.FS) {
	// Serve assets
	r.GET("/assets/*filepath", func(c *gin.Context) {
		filepath := c.Param("filepath")
		c.FileFromFS("assets"+filepath, http.FS(staticFS))
	})

	// Serve favicon
	r.GET("/favicon.svg", func(c *gin.Context) {
		data, err := fs.ReadFile(staticFS, "favicon.svg")
		if err != nil {
			c.Status(404)
			return
		}
		c.Data(200, "image/svg+xml", data)
	})

	// SPA fallback - serve index.html for all unmatched routes (except /api, /health, /version)
	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path
		// Skip API routes and special endpoints
		if len(path) >= 4 && path[:4] == "/api" ||
			path == "/health" || path == "/version" {
			c.JSON(404, gin.H{"code": 404, "msg": "Not found"})
			return
		}
		// Serve index.html for SPA routing
		data, err := fs.ReadFile(staticFS, "index.html")
		if err != nil {
			c.String(500, "Failed to read index.html")
			return
		}
		c.Data(200, "text/html; charset=utf-8", data)
	})
}