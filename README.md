# Go Recursive DNS Resolver

A high-performance, custom-built Recursive DNS Resolver written in Go.

Unlike standard DNS clients that rely on 3rd-party resolvers (like Google's `8.8.8.8` or Cloudflare's `1.1.1.1`), this tool performs the resolution process **from scratch**. It traverses the entire DNS hierarchy—starting from the Root Servers, moving to TLDs, and finally querying Authoritative Nameservers—to discover IP addresses manually.

It demonstrates deep systems programming concepts including raw UDP socket communication, binary packet parsing, and in-memory caching.

## 🚀 Features

- **Recursive Resolution Engine:** Implements the full "iterative" query logic (Root -> TLD -> Auth).
- **Raw Packet Manipulation:** Manually constructs and parses DNS headers and question/answer sections (handling bytes, not strings).
- **CNAME Handling:** Automatically follows canonical name aliases (e.g., resolving `www.facebook.com` -> `star-mini...` -> IP).
- **In-Memory Caching:** Implements a thread-safe cache with TTL (Time-To-Live) support to provide instant responses for repeated queries.
- **Custom Protocol Implementation:** Zero reliance on external DNS libraries (`net/http` or `github.com/miekg/dns` are NOT used).

## 📸 Demo & Screenshots

### 1. The Resolution Journey (Trace)

_This screenshot shows the resolver traversing from the Root Server down to the final IP._

![Before Cache](images/Screenshot%202026-01-01%20231006.png)

### 2. Advanced CNAME Resolution

_This screenshot demonstrates the resolver handling a complex CNAME alias. It detects that `www.facebook.com` points to `star-mini.c10r.facebook.com`, automatically restarts the entire recursive process for the new name, and successfully resolves the final IP._

![Complex CNAME Resolution](images/Screenshot%202026-01-02%20092158.png)

## 🛠️ Installation & Usage

### Prerequisites

- Go 1.20 or higher

### Running the Resolver

Clone the repository and run the `main.go` file with a target domain:

```bash
# Basic Usage
go run main.go google.com

# Test CNAME handling
go run main.go [www.facebook.com](https://www.facebook.com)
```

## The Working

This project bypasses the OS's default DNS stub resolver. Here is the lifecycle of a request:

1. Packet Construction: The program builds a raw DNS query packet (Header + Question) encoded in Big Endian binary format.

2. Root Query: It sends this packet via UDP to a hardcoded Root Server (198.41.0.4).

3. Iterative Loop:

   - The server replies not with the answer, but with a referral (Authoritative Section).

   - The program parses this referral to find the next Nameserver IP.

   - It repeats the query to this new server.

4. Handling Aliases (CNAME): If a server responds with a CNAME record (Type 5), the resolver detects this, extracts the alias domain, and restarts the recursion process for that new name.

5. Caching: Once an A Record (IP) is found, it is stored in a sync.RWMutex protected map with an expiration timestamp based on the record's TTL.

## Project Structure

```bash
├── main.go # Entry point, recursive loop logic, and caching layer
├── go.mod # Go module definition
└── dns/
└── message.go # Low-level packet parsing (Header, Question, Records) and byte manipulation
```
