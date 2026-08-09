package main

import (
	"encoding/csv"
	"os"
)

func loadRecipients(filePath string, ch chan Recipient) error {
	defer close(ch) // Close the channel when done sending data
	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	r := csv.NewReader(f)
	records, err := r.ReadAll()
	if err != nil {
		return err
	}

	for _, record := range records[1:] { // Skip header row
		ch <- Recipient{
			Name:  record[0],
			Email: record[1],
		}
	}
	return nil
}
