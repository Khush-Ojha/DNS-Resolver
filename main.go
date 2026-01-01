package main

import (
	"fmt"
	"log"
	"my-dns-resolver/dns"
	"net"
	"os"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: go run main.go <domain>")
		return
	}
	targetDomain := os.Args[1]
	rootServer := "198.41.0.4"
	fmt.Printf("Starting resolution for %s...\n", targetDomain)
	ip, err := resolve(targetDomain, rootServer)
	if err != nil {
		log.Fatalf("Failed: %v", err)
	}
	fmt.Printf("SUCCESS! IP for %s is %s\n", targetDomain, ip)
}

func resolve(domain string, nameserver string) (string, error) {
	for {
		fmt.Printf("Querying %s...\n", nameserver)
		response, err := sendQuery(domain, nameserver)
		if err != nil {
			return "", err
		}
		if len(response.Answers) > 0 {
			for _, answer := range response.Answers {
				if answer.Type == 1 {
					return fmt.Sprintf("%d.%d.%d.%d", answer.RData[0], answer.RData[1], answer.RData[2], answer.RData[3]), nil
				}
			}
		}
		if len(response.Additionals) > 0 {
			foundNewNS := false
			for _, additional := range response.Additionals {
				if additional.Type == 1 {
					nameserver = fmt.Sprintf("%d.%d.%d.%d", additional.RData[0], additional.RData[1], additional.RData[2], additional.RData[3])
					foundNewNS = true
					break
				}
			}
			if foundNewNS {
				continue
			}
		}
		return "", fmt.Errorf("could not resolve")
	}
}

func sendQuery(domain string, nameserver string) (*dns.Message, error) {
	query := dns.NewQuery(domain)
	queryBytes, err := query.ToBytes()
	if err != nil {
		return nil, err
	}
	conn, err := net.Dial("udp", nameserver+":53")
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	_, err = conn.Write(queryBytes)
	if err != nil {
		return nil, err
	}
	responseBuffer := make([]byte, 1024)
	n, err := conn.Read(responseBuffer)
	if err != nil {
		return nil, err
	}
	return dns.FromBytes(responseBuffer[:n])
}
