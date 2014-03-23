package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
)

var (
	// Listening web server options
	listenAddress      = flag.String("listen", "", "Specify an address to accept HTTP requests, e.g. \":8000\"")
	tlsListenAddress   = flag.String("tls-listen", ":8443", "Specify an address to accept HTTPS requests, e.g. \":8443\"")
	tlsCertificateFile = flag.String("tls-cert", "proxy.crt", "Path to the TLS certificate chain to use")
	tlsPrivateKeyFile  = flag.String("tls-key", "proxy.key", "Path to the private key for the TLS certificate")

	// Remote web server options
	remoteUrl = flag.String("remote", "http://localhost:8080", "HTTP URL to forward incoming hooks to, upon successful mirroring")

	// Git options
	mirrorPath = flag.String("mirror-path", "/tmp/mirror", "Directory to which git repositories should be mirrored")
)

func usage() {
	fmt.Fprintln(os.Stderr, "Receives git webhooks, keeps a local mirror of the repo up-to-date, then forwards the webhook to another server.\n")
	fmt.Fprintln(os.Stderr, "Usage:", os.Args[0])
	flag.PrintDefaults()
	os.Exit(2)
}

func startListening(handler http.Handler, address, tlsAddress, tlsCertFile, tlsKeyFile string) {
	isRunning := false
	if *listenAddress != "" {
		go serveHttp(address, handler)
		isRunning = true
	}
	if *tlsListenAddress != "" {
		go serveTls(tlsAddress, tlsCertFile, tlsKeyFile, handler)
		isRunning = true
	}
	if !isRunning {
		log.Fatal("Quitting as both HTTP and TLS were disabled")
	}
}

func main() {
	// Get the command line options
	flag.Usage = usage
	flag.Parse()

	// Show some basic config info
	log.Println("Git repositories will be mirrored to: ", *mirrorPath)
	log.Println("Webhook requests will be forwarded to:", *remoteUrl)

	// Start the listening web server
	handler, err := NewHandler(*mirrorPath, *remoteUrl)
	if err != nil {
		log.Fatal("Invalid config:", err)
	}
	startListening(handler, *listenAddress, *tlsListenAddress, *tlsCertificateFile, *tlsPrivateKeyFile)

	// Wait for our eventual death
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)
	<-c
	log.Println("Shutting down...")
}
