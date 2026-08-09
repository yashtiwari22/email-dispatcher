package main

import (
	"fmt"
	"log"
	"net/smtp"
	"sync"
	"time"
)

func emailWorker(id int, ch chan Recipient, wg *sync.WaitGroup) {

	defer wg.Done()
	for recipient := range ch {
		smtpHost := "localhost"
		smptPort := "1025"

		// formattedMessage := fmt.Sprintf("To: %s\r\nSubject: Test Email\r\n\r\n%s\r\n", recipient.Email, "Just Testing Our Email Dispatcher")
		// msg := []byte(formattedMessage)

		msg, err := executeTemplate(recipient)
		if err != nil {
			fmt.Printf("Worker %d: Error parsing template for %s", id, recipient.Email)
			// todo: add to dlq
			continue
		}

		fmt.Printf("Worker %d: Sending email to %s\n", id, recipient.Email)

		err = smtp.SendMail(smtpHost+":"+smptPort, nil, "yash.tiwari4238@gmail.com", []string{recipient.Email}, []byte(msg))
		if err != nil {
			log.Fatal(err)
		}

		time.Sleep(50 * time.Millisecond)
		fmt.Printf("Worker %d: Sent email to %s\n", id, recipient.Email)
	}
}
