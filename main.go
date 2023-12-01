package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"

	"golang.org/x/crypto/ssh"
)

func main() {
	// export VAULT_SERVER=
	vs := os.Getenv("VAULT_SERVER")
	// export KEY_PATH=
	kp := os.Getenv("KEY_PATH")


	sshClient, err := connectSSH(vs, kp)
	if err != nil {
		panic(err)
	}

	defer sshClient.Close()

	fmt.Println("ssh client connected")

	http.HandleFunc("/", handleRequest(sshClient))
	log.Fatal(http.ListenAndServe(":8200", nil)) // Listen on port 8200
}

func handleRequest(client *ssh.Client) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		fmt.Printf("request: %+v\n", req)
		// Dial the SSH connection once and keep it open
		remoteConn, err := client.Dial("tcp", "0.0.0.0:8200")
		if err != nil {
			http.Error(w, "Failed to connect to remote server: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer remoteConn.Close()

		// Forward the HTTP request through the SSH tunnel
		err = forwardRequest(remoteConn, req)
		if err != nil {
			http.Error(w, "Failed to forward request: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Copy the response back to the client
		io.Copy(w, remoteConn)
	}
}

func forwardRequest(remoteConn net.Conn, req *http.Request) error {
	// Forward the request method, URL, and headers
	outReq, err := http.NewRequest(req.Method, "http://127.0.0.1:8200"+req.URL.String(), req.Body)
	if err != nil {
		return err
	}
	outReq.Header = req.Header

	fmt.Printf("outReq: %+v\n", outReq)

	// Send the request through the SSH tunnel
	resp, err := http.DefaultClient.Do(outReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Copy the response body back through the SSH tunnel
	_, err = io.Copy(remoteConn, resp.Body)
	return err
}

func connectSSH(server string, keyPath string) (*ssh.Client, error) {
	key, err := ioutil.ReadFile(keyPath)
	if err != nil {
		return nil, err
	}

	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil, err
	}

	config := &ssh.ClientConfig{
		User: "jeffreynaef",
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // Note: Insecure; for production, use a proper host key
	}

	client, err := ssh.Dial("tcp", server, config)
	if err != nil {
		return nil, err
	}

	return client, nil
}
