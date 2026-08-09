# Concurrent Email Dispatcher in Go

A high-performance, concurrent bulk email dispatcher built in Go using the **Producer-Consumer pattern**, Go channels, `sync.WaitGroup` worker pools, and HTML/text templating. Designed to process recipient lists from CSV files and send emails concurrently via an SMTP server (e.g., [Mailpit](https://github.com/axllent/mailpit) for local testing).

---

## 📐 System Architecture

```
                       +------------------+
                       |    emails.csv    |
                       +--------+---------+
                                |
                                v
                      +-------------------+
                      |   loadRecipients  | (Producer Goroutine)
                      |   (producer.go)   |
                      +---------+---------+
                                |
                                |  Recipient Channel (chan Recipient)
                                v
               +---------------------------------+
               |   Worker Pool (workerCount = 5) |
               +--------+-------+-------+--------+
                        |       |       |
            +-----------+       |       +-----------+
            |                   v                   |
   +-----------------+ +-----------------+ +-----------------+
   | Worker 1        | | Worker 2..4     | | Worker 5        |
   | (consumer.go)   | | (consumer.go)   | | (consumer.go)   |
   +--------+--------+ +--------+--------+ +--------+--------+
            |                   |                   |
            +-----------+       |       +-----------+
                        |       v       |
                        v               v
                     +---------------------+
                     |  executeTemplate()  | (email.tmpl)
                     +----------+----------+
                                |
                                v
                     +---------------------+
                     |   net/smtp Mail     |
                     |  (localhost:1025)   |
                     +---------------------+
```

---

## 📁 Repository Structure

| File | Description |
| :--- | :--- |
| [`main.go`](file:///Users/yash/workspace/personal/email-dispatcher/main.go) | Entry point. Defines the `Recipient` struct, initializes channels, launches the producer goroutine, spawns 5 concurrent workers using `sync.WaitGroup`, and contains `executeTemplate`. |
| [`producer.go`](file:///Users/yash/workspace/personal/email-dispatcher/producer.go) | Implements `loadRecipients()`, which streams entries from a CSV file into the recipient channel and closes the channel when finished. |
| [`consumer.go`](file:///Users/yash/workspace/personal/email-dispatcher/consumer.go) | Implements `emailWorker()`, which consumes recipients from the channel, executes email templating, and dispatches messages via SMTP. |
| [`email.tmpl`](file:///Users/yash/workspace/personal/email-dispatcher/email.tmpl) | Go template containing email headers (`To`, `Subject`) and body text. |
| [`emails.csv`](file:///Users/yash/workspace/personal/email-dispatcher/emails.csv) | CSV dataset containing `Name` and `Email` columns (20 mock records included). |
| [`info.md`](file:///Users/yash/workspace/personal/email-dispatcher/info.md) | Quick reference command for starting Mailpit SMTP test server via Docker. |
| [`go.mod`](file:///Users/yash/workspace/personal/email-dispatcher/go.mod) | Go module definition (`github.com/yashtiwari22/email-dispatcher`, Go 1.26). |

---

## ⚙️ How It Works

### 1. Data Ingestion (Producer)
- The [`loadRecipients`](file:///Users/yash/workspace/personal/email-dispatcher/producer.go#L8) function opens `emails.csv`, parses the records using Go's `encoding/csv` package, skips the header row, and pushes `Recipient` structs into an unbuffered Go channel (`chan Recipient`).
- `defer close(ch)` guarantees that the channel is closed once all CSV rows are emitted, signaling workers to shut down gracefully.

### 2. Concurrent Processing (Worker Pool / Consumer)
- Main spawns 5 concurrent [`emailWorker`](file:///Users/yash/workspace/personal/email-dispatcher/consumer.go#L11) goroutines.
- Each worker continuously receives `Recipient` objects from the shared channel using a `for recipient := range ch` loop.
- Synchronization is tracked using `sync.WaitGroup`.

### 3. Dynamic Templating & Dispatch
- Each email body is rendered via [`executeTemplate`](file:///Users/yash/workspace/personal/email-dispatcher/main.go#L35) using [`email.tmpl`](file:///Users/yash/workspace/personal/email-dispatcher/email.tmpl).
- The formatted message is sent via standard `net/smtp.SendMail` to the target SMTP host (`localhost:1025`).
- A small throttle duration (`50ms`) is introduced per job to simulate realistic network delay and avoid hammering the SMTP provider.

---

## 🚀 Getting Started

### Prerequisites
- **Go**: Version 1.22+ installed
- **Docker**: For running Mailpit (SMTP Mock Server)

### Step 1: Start Mailpit (Local SMTP Mock Server)
Mailpit captures sent emails in a local web interface without sending actual emails to real recipients.

```bash
docker run -d \
  --restart unless-stopped \
  --name=mailpit \
  -p 8025:8025 \
  -p 1025:1025 \
  axllent/mailpit
```
- **SMTP Server**: `localhost:1025`
- **Mailpit Web UI**: `http://localhost:8025`

### Step 2: Run the Dispatcher

```bash
go run .
```

---

## 🛠 Tech Stack

- **Language**: Go (`go 1.26.5`)
- **Concurrency**: Goroutines, Unbuffered Channels (`chan Recipient`), `sync.WaitGroup`
- **Templating**: `html/template`
- **Networking**: `net/smtp`
- **Testing SMTP Server**: [Mailpit](https://github.com/axllent/mailpit)

---

## 💡 Codebase Observations & Optimization Opportunities

1. **Template Parsing Optimization**:
   - *Current*: [`executeTemplate`](file:///Users/yash/workspace/personal/email-dispatcher/main.go#L35) parses `email.tmpl` on every single email processed (`template.ParseFiles("email.tmpl")`).
   - *Improvement*: Parse the template once during application startup (e.g., `template.Must(template.ParseFiles("email.tmpl"))`) and reuse the parsed `*template.Template` across goroutines to eliminate unnecessary disk I/O and CPU overhead.

2. **SMTP Error Handling**:
   - *Current*: `emailWorker` invokes [`log.Fatal(err)`](file:///Users/yash/workspace/personal/email-dispatcher/consumer.go#L32) on SMTP sending failure, which terminates the entire process immediately.
   - *Improvement*: Replace `log.Fatal` with retry logic or log the error and send the failed job to a Dead Letter Queue (DLQ).

3. **Dead Letter Queue (DLQ)**:
   - As noted by the `// todo: add to dlq` comment in [`consumer.go`](file:///Users/yash/workspace/personal/email-dispatcher/consumer.go#L24), failed template parsing or SMTP delivery should push failed items into a dedicated retry/DLQ channel for further inspection or retry.

4. **SMTP Connection Reuse**:
   - *Current*: Opens a new SMTP connection per email with `smtp.SendMail`.
   - *Improvement*: Maintain persistent SMTP client connections or connection pooling per worker using `smtp.Client` (`Dial`, `Mail`, `Rcpt`, `Data`) to reduce TCP connection overhead.

5. **Configurability**:
   - Hardcoded parameters (`workerCount`, `smtpHost`, `smptPort`, sender email, file paths) can be extracted to environment variables or CLI flags (`flag` package or `envconfig`).
