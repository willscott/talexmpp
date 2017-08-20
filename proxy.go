package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"net"
	"os"
	"sync"

	"./xmpp"
	talekcommon "github.com/privacylab/talek/common"
	"github.com/privacylab/talek/libtalek"
)

func main() {
	serverPort := flag.Int("port", 5222, "port number to listen on")
	talekConfig := flag.String("config", "talek.conf", "Talek client configuration")

	flag.Parse()

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", *serverPort))
	if err != nil {
		panic(err)
	}
	defer listener.Close()

	var contacts = make(map[string]chan<- []byte)
	var messagebus = make(chan xmpp.Message)
	var connectbus = make(chan xmpp.Connect)
	var disconnectbus = make(chan xmpp.Disconnect)

	talekconf := libtalek.ClientConfigFromFile(*talekConfig)
	if talekconf == nil {
		panic("Could not load talek configuration file")
	}
	leader := talekcommon.NewFrontendRPC("rpc", talekconf.FrontendAddr)
	backend := libtalek.NewClient("talexmpp", *talekconf, leader)

	am := AccountManager{Online: contacts, Backend: backend, lock: &sync.Mutex{}}

	if _, err := os.Stat("./cert.pem"); err != nil && os.IsNotExist(err) {
		MakeSelfSignedCert()
	}
	var cert, _ = tls.LoadX509KeyPair("./cert.pem", "./key.pem")
	var tlsConfig = tls.Config{
		MinVersion:   tls.VersionTLS10,
		Certificates: []tls.Certificate{cert},
		CipherSuites: []uint16{tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA,
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA,
			tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,
			tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA},
	}

	xmppServer := &xmpp.Server{
		Accounts:   am,
		ConnectBus: connectbus,
		Extensions: []xmpp.Extension{
			&xmpp.NormalMessageExtension{MessageBus: messagebus},
			&xmpp.RosterExtension{Accounts: am},
			&GlueExtension{},
			&RosterManagementExtension{Accounts: am, Client: backend},
		},
		DisconnectBus: disconnectbus,
		Domain:        "localhost",
		TLSConfig:     &tlsConfig,
	}

	go am.routeRoutine(messagebus)
	go am.connectRoutine(connectbus)
	go am.disconnectRoutine(disconnectbus)

	for {
		conn, err := listener.Accept()
		if err != nil {
			panic(err)
		}

		go xmppServer.TCPAnswer(conn)
	}
}
