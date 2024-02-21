package main

import (
	"io"
	"log"
	"os"

	"github.com/emersion/go-imap/v2"
	"github.com/emersion/go-imap/v2/imapclient"
	_ "github.com/emersion/go-message/charset"
	"github.com/emersion/go-message/mail"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()

	imapServer := os.Getenv("IMAP_SERVER")
	imapUsername := os.Getenv("IMAP_USERNAME")
	imapPassword := os.Getenv("IMAP_PASSWORD")

	c, err := imapclient.DialTLS(imapServer, nil)
	if err != nil {
		log.Fatal("Client connection failed: " + err.Error())
	}

	defer c.Close()

	err = c.Login(imapUsername, imapPassword).Wait()
	if err != nil {
		log.Fatal("Client login failed: " + err.Error())
	}

	log.Println("Mailboxes: ")
	mailboxes, err := c.List("", "*", nil).Collect()
	if err != nil {
		log.Fatal("Client list failed: " + err.Error())
	}

	for _, m := range mailboxes {
		log.Println("* " + m.Mailbox)
	}

	mbox, err := c.Select("INBOX", nil).Wait()
	if err != nil {
		log.Fatal("Client select failed: " + err.Error())
	}

	seqSet := imap.SeqSetNum(mbox.NumMessages - 2)
	fetchOptions := &imap.FetchOptions{
		BodySection: []*imap.FetchItemBodySection{{}},
	}
	fetchCmd := c.Fetch(seqSet, fetchOptions)
	defer fetchCmd.Close()

	msg := fetchCmd.Next()
	if msg == nil {
		log.Fatalf("FETCH command did not return any message")
	}

	// Find the body section in the response
	var bodySection imapclient.FetchItemDataBodySection
	ok := false
	for {
		item := msg.Next()
		if item == nil {
			break
		}
		bodySection, ok = item.(imapclient.FetchItemDataBodySection)
		if ok {
			break
		}
	}
	if !ok {
		log.Fatalf("FETCH command did not return body section")
	}

	// Read the message via the go-message library
	mr, err := mail.CreateReader(bodySection.Literal)
	if err != nil {
		log.Fatalf("failed to create mail reader: %v", err)
	}

	// Print a few header fields
	h := mr.Header
	if date, err := h.Date(); err != nil {
		log.Printf("failed to parse Date header field: %v", err)
	} else {
		log.Printf("Date: %v", date)
	}
    if from, err := h.AddressList("From"); err != nil {
        log.Printf("failed to parse From header field: %v", err)
    } else {
        log.Printf("From: %v", from)
    }
	if to, err := h.AddressList("To"); err != nil {
		log.Printf("failed to parse To header field: %v", err)
	} else {
		log.Printf("To: %v", to)
	}
	if subject, err := h.Text("Subject"); err != nil {
		log.Printf("failed to parse Subject header field: %v", err)
	} else {
		log.Printf("Subject: %v", subject)
	}

	// Process the message's parts
	for {
		p, err := mr.NextPart()
		if err == io.EOF {
			break
		} else if err != nil {
			log.Fatalf("failed to read message part: %v", err)
		}

		switch h := p.Header.(type) {
		case *mail.InlineHeader:
			// This is the message's text (can be plain-text or HTML)
			b, _ := io.ReadAll(p.Body)
            ct, _, _ := h.ContentType()
            if ct == "text/plain" {
                log.Printf("%v", string(b))
            }

            /*
            log.Printf("Content-Type: %v", ct)
			log.Printf("Inline text: %v", string(b))
            */
		case *mail.AttachmentHeader:
			// This is an attachment
			filename, _ := h.Filename()
			log.Printf("Attachment: %v", filename)

			// Save the attachment to a file
            /*
			f, err := os.Create(filename)
			if err != nil {
				log.Fatalf("failed to create attachment file: %v", err)
			}

			if _, err := io.Copy(f, p.Body); err != nil {
				log.Fatalf("failed to write attachment file: %v", err)
			}

			if err := f.Close(); err != nil {
				log.Fatalf("failed to close attachment file: %v", err)
			}
            */
		}
	}

	if err := fetchCmd.Close(); err != nil {
		log.Fatalf("FETCH command failed: %v", err)
	}
}
