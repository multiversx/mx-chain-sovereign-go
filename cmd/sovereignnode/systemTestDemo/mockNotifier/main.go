package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/multiversx/mx-chain-core-go/data/sovereign"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/multiversx/mx-chain-sovereign-bridge-go/cert"
	"github.com/urfave/cli"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// Before merging anything into feat/chain-go-sdk, please try a "stress" system test with a local testnet and this notifier.
// Steps:
// 1. Replace github.com/multiversx/mx-chain-communication-go from cmd/sovereignnode/systemTestDemo/go.mod with the one
// from this branch: sovereign-stress-test-branch.
// 2. Keep the config in variables.sh with at least 3 validators.
//
// If you are running with a local testnet and need the necessary certificate files to mock bridge operations, you
// can find them(certificate.crt + private_key.pem) within testnet environment setup at ~MultiversX/testnet/node/config

func main() {
	app := cli.NewApp()
	app.Name = "MultiversX sovereign chain mock notifier"
	app.Usage = "This tool serves as an observer notifier for a sovereign shard. It initiates the transmission of blocks " +
		"starting from an arbitrary nonce, with incoming events occurring every 3 blocks. Each incoming event comprises " +
		"an NFT and an ESDT transfer. The periodic transmission includes 2 NFTs (ASH-a642d1-01 & ASH-a642d1-02) and one " +
		"ESDT (WEGLD-bd4d79). To verify these entities, one can utilize the sovereign proxy at, for example, " +
		fmt.Sprintf("http://127.0.0.1:7950/address/%s/esdt", subscribedAddress) +
		"The blocks are sent with an arbitrary period between them."
	app.Flags = []cli.Flag{
		logLevel,
		grpcEnabled,
		sovereignBridgeCertificateFile,
		sovereignBridgeCertificatePkFile,
		notifiers,
	}
	app.Authors = []cli.Author{
		{
			Name:  "The MultiversX Team",
			Email: "contact@multiversx.com",
		},
	}

	app.Action = startAllNotifiers
	err := app.Run(os.Args)
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
}

func startAllNotifiers(ctx *cli.Context) error {
	err := initializeLogger(ctx)
	if err != nil {
		return err
	}

	notifiersMap := make(map[string]struct{})
	for _, notifier := range ctx.StringSlice(notifiers.Name) {
		notifiersMap[notifier] = struct{}{}
	}

	mockedGRPCServer, grpcServerConn, err := createAndStartGRPCServer(ctx)
	if err != nil {
		log.Error("cannot create grpc server", "error", err)
		return err
	}

	defer func() {
		grpcServerConn.Stop()
	}()

	if _, enabled := notifiersMap[chainMVX]; enabled {
		go func() {
			err = startMVXMockNotifier(mockedGRPCServer)
			if err != nil {
				log.Error(err.Error())
			}
		}()

	}

	if _, enabled := notifiersMap[chainETH]; enabled {
		go func() {
			err = startETHMockNotifier()
			if err != nil {
				log.Error(err.Error())
			}
		}()
	}

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM)

	<-interrupt
	log.Info("closing app at user's signal")

	return nil
}

func initializeLogger(ctx *cli.Context) error {
	logLevelFlagValue := ctx.GlobalString(logLevel.Name)
	return logger.SetLogLevel(logLevelFlagValue)
}

func createAndStartGRPCServer(ctx *cli.Context) (GRPCServer, GRPCServerConnection, error) {
	if !ctx.Bool(grpcEnabled.Name) {
		return NewDisabledMockServer(), NewDisabledGRPCServer(), nil
	}

	listener, err := net.Listen("tcp", grpcAddress)
	if err != nil {
		return nil, nil, err
	}

	tlsConfig, err := cert.LoadTLSServerConfig(cert.FileCfg{
		CertFile: getAbsolutePath(ctx.GlobalString(sovereignBridgeCertificateFile.Name)),
		PkFile:   getAbsolutePath(ctx.GlobalString(sovereignBridgeCertificatePkFile.Name)),
	})
	if err != nil {
		return nil, nil, err
	}
	tlsCredentials := credentials.NewTLS(tlsConfig)
	grpcServer := grpc.NewServer(
		grpc.Creds(tlsCredentials),
	)
	mockedServer := NewMockServer()
	sovereign.RegisterBridgeTxSenderServer(grpcServer, mockedServer)

	log.Info("starting grpc server...")

	go func() {
		for {
			if err = grpcServer.Serve(listener); err != nil {
				log.LogIfError(err)
				time.Sleep(time.Second)
			}
		}
	}()

	return mockedServer, grpcServer, nil
}

func getAbsolutePath(path string) string {
	if !strings.HasPrefix(path, "~") {
		return path
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Error("Error getting home directory: " + err.Error())
		return ""
	}
	return strings.Replace(path, "~", homeDir, 1)
}
