package main

import (
	"bytes"
	"fmt"
	"html/template"
	"sync"
)

type Recipient struct {
	Name  string
	Email string
}

func main() {

	fmt.Println("welcome to email dispatcher")

	recipientChannel := make(chan Recipient)

	go func() {
		loadRecipients("./emails.csv", recipientChannel)
	}()

	var wg sync.WaitGroup
	workerCount := 5
	for i := 1; i <= workerCount; i++ {
		wg.Add(1)
		go emailWorker(i, recipientChannel, &wg)
	}

	wg.Wait()
}

func executeTemplate(r Recipient) (string, error) {
	t, err := template.ParseFiles("email.tmpl")
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	err = t.Execute(&buf, r)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}
