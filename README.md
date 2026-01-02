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

### 2. Caching & Performance

_Notice the "Resolution took" time on the second run drops to near zero due to the in-memory cache._

![After Cache & CName](images/Screenshot%202026-01-02%20082502.png)

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
