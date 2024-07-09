package main

import (
	"context"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/mikioh/ipaddr"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	"hornbill/pkg/allocator"
	"hornbill/pkg/daemon"
	"hornbill/pkg/pb"
	"hornbill/pkg/rpcconn"
	"log"
	"net"
	"os"
	"os/exec"
	"strings"
	"time"
)

func main() {
	err := godotenv.Load()
	if err == nil {
		log.Println("Loaded environment variables from .env file")
	}

	err = godotenv.Load("config.ini")
	if err == nil {
		log.Println("Loaded environment variables from config.ini")
	}

	listenAddr := os.Getenv("LISTEN_ADDR")
	if listenAddr == "" {
		listenAddr = ":3000"
	}

	wgInitScript := os.Getenv("WG_INIT_EXEC")
	if wgInitScript != "" {
		wgInitScriptSplit := strings.Split(wgInitScript, " ")
		command := exec.Command(wgInitScriptSplit[0], wgInitScriptSplit[1:]...)
		workingDir, err := os.Getwd()
		if err != nil {
			panic(fmt.Errorf("failed getting working directory while executing initialize script: %w", err))
		}
		command.Dir = workingDir
		command.Stdout = os.Stdout
		command.Stderr = os.Stderr
		err = command.Run()
		if err != nil {
			panic(fmt.Errorf("failed executing initialize script: %w", err))
		}
	}

	wgInterfaceName := os.Getenv("WG_INTERFACE_NAME")
	if wgInterfaceName == "" {
		log.Fatalln("interface name not provided")
	}

	wgPublicKeyFile := os.Getenv("WG_PUBLIC_KEY_FILE")
	if wgPublicKeyFile == "" {
		log.Fatalln("WireGuard public key file not provided")
	}
	wgPublicKeyData, err := os.ReadFile(wgPublicKeyFile)
	if err != nil {
		log.Fatalln("Error reading WireGuard public key file:", err)
	}

	wgPublicKey, err := wgtypes.ParseKey(string(wgPublicKeyData))
	if err != nil {
		log.Fatalln("Error parsing WireGuard public key file:", err)
	}

	wgPublicAddress := os.Getenv("WG_PUBLIC_ADDRESS")
	wgAllowedAddress := strings.Split(os.Getenv("WG_ALLOWED_CIDR"), ",")
	wgConfig := daemon.WireGuardConfig{
		InterfaceName:  wgInterfaceName,
		PublicKey:      wgPublicKey,
		PublicAddress:  wgPublicAddress,
		AllowedAddress: wgAllowedAddress,
	}

	tcpServer, err := net.Listen("tcp", listenAddr)
	if err != nil {
		panic(err)
	}

	wireGuard, err := daemon.NewWireGuard(wgConfig)
	if err != nil {
		panic(err)
	}

	server, err := rpcconn.NewServer()
	if err != nil {
		panic(err)
	}
	parse, err := ipaddr.Parse(os.Getenv("WG_NETWORK_CIDR"))
	if err != nil {
		panic(err)
	}
	pos := parse.Pos()
	if pos == nil {
		log.Fatalf("IP parse failed")
	}

	alloc := allocator.NewAllocator(pos.Prefix.IPNet)
	daemonServer := daemon.Server{
		Allocator: alloc,
		WireGuard: wireGuard,
	}
	go func() {
		time.Sleep(30 * time.Second)
		for {
			_, tickError := daemonServer.Tick(context.TODO(), &pb.Empty{})
			if tickError != nil {
				log.Println("daemon server tick error:", tickError.Error())
			}
			time.Sleep(15 * time.Minute)
		}
	}()

	pb.RegisterDaemonServer(server, &daemonServer)

	log.Println("Listening on " + listenAddr)
	err = server.Serve(tcpServer)
	if err != nil {
		panic(err)
	}

}
