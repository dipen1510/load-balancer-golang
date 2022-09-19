package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
)

type LoadBalancer struct {
	port            string
	roundRobinCount int
	servers         []Server
}

type Server interface {
	Address() string
	IsAlive() bool
	Serve(re http.ResponseWriter, r *http.Request)
}
type simpleServer struct {
	adr   string
	proxy *httputil.ReverseProxy
}

func newServer(add string) *simpleServer {
	serverUrl, err := url.Parse(add)
	handleError(err)
	return &simpleServer{
		adr:   add,
		proxy: httputil.NewSingleHostReverseProxy(serverUrl),
	}
}

func newLoadBalancer(port string, servers []Server) *LoadBalancer {
	return &LoadBalancer{
		port:            port,
		roundRobinCount: 0,
		servers:         servers,
	}
}

func (s *simpleServer) Address() string {
	return s.adr
}

func (s *simpleServer) IsAlive() bool {
	return true
}

func (s *simpleServer) Serve(rw http.ResponseWriter, r *http.Request) {
	s.proxy.ServeHTTP(rw, r)
}

func main() {
	servers := []Server{
		newServer("https://www.facebook.com"),
		newServer("https://www.google.com"),
		newServer("https://www.bing.com"),
	}
	lb := newLoadBalancer("8000", servers)
	handleRedirect := func(rw http.ResponseWriter, req *http.Request) {
		lb.serverProxy(rw, req)
	}
	http.HandleFunc("/", handleRedirect)
	fmt.Printf("serving requests at 'localhost:%s'\n", lb.port)
	http.ListenAndServe(":"+lb.port, nil)
}

func (l *LoadBalancer) getAddress() Server {
	server := l.servers[l.roundRobinCount%len(l.servers)]
	for !server.IsAlive() {
		l.roundRobinCount++
		server = l.servers[l.roundRobinCount%len(l.servers)]
	}
	l.roundRobinCount++
	return server
}

func (l *LoadBalancer) serverProxy(rw http.ResponseWriter, r *http.Request) {
	targetServer := l.getAddress()
	fmt.Println("forwarding rewuest to add %q", targetServer.Address())
	targetServer.Serve(rw, r)
}

func handleError(err error) {
	if err != nil {
		fmt.Errorf("error: %e\n", err)
		os.Exit(1)
	}
}
