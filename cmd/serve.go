package cmd

import (
	"os"
	"os/signal"
	"syscall"
	// "time"

	"github.com/spf13/cobra"

	"github.com/bartmika/mothership-server/internal/controllers"
	// "github.com/bartmika/mothership-server/utils"
)

func init() {
	// The following are optional and will have defaults placed when missing.
	serveCmd.Flags().StringVarP(&ipAddress, "ip", "i", "localhost", "The ip address to bind this server to")
	serveCmd.Flags().IntVarP(&port, "port", "p", 50051, "The port to run this server on")
	serveCmd.Flags().StringVarP(&databaseUrl, "database_url", "d", os.Getenv("MOTHERSHIP_SERVER_DATABASE_URL"), "The database URL to run this server on")
	serveCmd.Flags().StringVarP(&hmacSecret, "hmac_secret", "s", os.Getenv("MOTHERSHIP_SERVER_HMAC_SECRET"), "The secret key to use in this server")

	// Make this sub-command part of our application.
	rootCmd.AddCommand(serveCmd)
}

func doServe() {
	// Convert the user inputted integer value to be a `time.Duration` type.

	// Setup our server.
	server := controllers.New(ipAddress, port, databaseUrl, hmacSecret)

	// DEVELOPERS CODE:
	// The following code will create an anonymous goroutine which will have a
	// blocking chan `sigs`. This blocking chan will only unblock when the
	// golang app receives a termination command; therfore the anyomous
	// goroutine will run and terminate our running application.
	//
	// Special Thanks:
	// (1) https://gobyexample.com/signals
	// (2) https://guzalexander.com/2017/05/31/gracefully-exit-server-in-go.html
	//
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs // Block execution until signal from terminal gets triggered here.
		server.StopMainRuntimeLoop()
	}()
	server.RunMainRuntimeLoop()
}

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Run the gRPC server",
	Long:  `Run the gRPC server to allow other services to access this application`,
	Run: func(cmd *cobra.Command, args []string) {
		// Defensive code. ...
		// Do nothing for now...

		// Execute our command with our validated inputs.
		doServe()
	},
}
