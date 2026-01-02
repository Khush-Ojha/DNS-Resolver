package main

import (
	"fmt"
	"log"
	"my-dns-resolver/dns"
	"net"
	"os"
	"sync"
	"time"
)

// --- CACHE IMPLEMENTATION ---
type CacheEntry struct {
	IPv4      string
	ExpiresAt time.Time
}

var (
	cache = make(map[string]CacheEntry)
	mu    sync.RWMutex
)

func getFromCache(domain string) (string, bool) {
	mu.RLock()
	defer mu.RUnlock()
	entry, found := cache[domain]
	if !found {
		return "", false
	}
	if time.Now().After(entry.ExpiresAt) {
		return "", false // Expired
	}
	return entry.IPv4, true
}

func saveToCache(domain string, ip string, ttl uint32) {
	mu.Lock()
	defer mu.Unlock()
	cache[domain] = CacheEntry{
		IPv4:      ip,
		ExpiresAt: time.Now().Add(time.Duration(ttl) * time.Second),
	}
}

// ----------------------------

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: go run main.go <domain>")
		return
	}
	targetDomain := os.Args[1]

	// We start timing the request to show off speed
	start := time.Now()

	ip, err := resolve(targetDomain)
	if err != nil {
		log.Fatalf("Failed: %v", err)
	}

	duration := time.Since(start)
	fmt.Printf("\n-----------------------------------------------\n")
	fmt.Printf("FINAL RESULT: %s -> %s\n", targetDomain, ip)
	fmt.Printf("Resolution took: %v\n", duration)
	fmt.Printf("-----------------------------------------------\n")
}

func resolve(domain string) (string, error) {
	// 1. Check Cache First
	if ip, found := getFromCache(domain); found {
		fmt.Printf("[CACHE HIT] %s is %s\n", domain, ip)
		return ip, nil
	}

	// 2. Start Querying from Root
	rootServer := "198.41.0.4"
	nameserver := rootServer

	// Loop until we find an answer
	for {
		fmt.Printf("Querying %s for %s...\n", nameserver, domain)
		response, err := sendQuery(domain, nameserver)
		if err != nil {
			return "", err
		}

		// 3. Check for Answers (Type A or CNAME)
		if len(response.Answers) > 0 {
			for _, answer := range response.Answers {
				// CASE A: We found the IP (Type 1)
				if answer.Type == 1 {
					ip := fmt.Sprintf("%d.%d.%d.%d", answer.RData[0], answer.RData[1], answer.RData[2], answer.RData[3])
					saveToCache(domain, ip, answer.TTL)
					return ip, nil
				}

				// CASE B: We found a CNAME (Type 5) - "Go look here instead"
				if answer.Type == 5 {
					cname := answer.ToName()
					fmt.Printf("   -> Found CNAME alias: %s. Restarting search...\n", cname)
					// Recursive call to resolve the new alias
					return resolve(cname)
				}
			}
		}

		// 4. Check for Referrals (Go ask someone else)
		// We look for the "NS" records in the Authority section, and their IP in Additional section
		if len(response.Authorities) > 0 {
			foundNewNS := false

			// Optimization: Look for the IP of the Name Server in the "Additional" section
			for _, additional := range response.Additionals {
				if additional.Type == 1 {
					nameserver = fmt.Sprintf("%d.%d.%d.%d", additional.RData[0], additional.RData[1], additional.RData[2], additional.RData[3])
					foundNewNS = true
					break
				}
			}

			// If we found a referral IP, loop and query that IP
			if foundNewNS {
				continue
			}
		}

		return "", fmt.Errorf("dead end: could not resolve %s", domain)
	}
}

func sendQuery(domain string, nameserver string) (*dns.Message, error) {
	query := dns.NewQuery(domain)
	queryBytes, err := query.ToBytes()
	if err != nil {
		return nil, err
	}
	// Using Port 53 (Standard DNS Port)
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
	conn.SetReadDeadline(time.Now().Add(2 * time.Second)) // Add timeout
	n, err := conn.Read(responseBuffer)
	if err != nil {
		return nil, err
	}

	return dns.FromBytes(responseBuffer[:n])
}
