package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/emersion/go-imap/v2"
	"github.com/emersion/go-imap/v2/imapclient"
	_ "github.com/emersion/go-message/charset"
	"github.com/emersion/go-message/mail"
	"github.com/joho/godotenv"
	"github.com/rivo/tview"
)

type Email struct {
	From        []*mail.Address
	To          []*mail.Address
	Date        time.Time
	Subject     string
	Body        map[string]string
	Attachments map[string][]byte
}

var app *tview.Application
var emails []Email = make([]Email, 0)

func main() {
	godotenv.Load()
	app = tview.NewApplication()

	previewView := tview.NewTextView().SetWordWrap(true).SetChangedFunc(func() {
		app.Draw()
	})
	previewView.SetBorder(true).SetTitle("Preview")

	list := tview.NewList().ShowSecondaryText(false)
	list.SetBorder(true).SetTitle("Messages")
	list.AddItem("Quit", "Press to exit", 'q', func() {
		app.Stop()
	})

	list.SetChangedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
		if index >= len(emails) {
			return
		}

		previewView.SetText(emails[index].Body["text/plain"])
	})

	go fetchMails(list)
	flex := tview.NewFlex().
		AddItem(list, 0, 1, true).
		AddItem(previewView, 0, 1, false)

	if err := app.SetRoot(flex, true).Run(); err != nil {
		panic(err)
	}

}

func fetchMails(list *tview.List) {
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

	_, err = c.List("", "*", nil).Collect()
	if err != nil {
		log.Fatal("Client list failed: " + err.Error())
	}

	mbox, err := c.Select("INBOX", nil).Wait()
	if err != nil {
		log.Fatal("Client select failed: " + err.Error())
	}

	seqSet := new(imap.SeqSet)
	seqSet.AddRange(mbox.NumMessages-100, mbox.NumMessages)
	fetchOptions := &imap.FetchOptions{
		BodySection: []*imap.FetchItemBodySection{{}},
	}
	fetchCmd := c.Fetch(*seqSet, fetchOptions)
	defer fetchCmd.Close()

	i := 1
	for {
		msg := fetchCmd.Next()
		if msg == nil {
			break
		}

		emailData, err := getEmailData(msg)
		if err != nil {
			log.Fatalf("Get email data failed!")
		}

		emails = append(emails, *emailData)

		//log.Printf("EmailData: %v", emailData.Subject)
		listItem := fmt.Sprintf("%d. [%s] %s (%s)", i, emailData.From[0].Address, emailData.Subject, emailData.Date)
		list.InsertItem(-2, listItem, "", 0, nil)

		if i == 1 {
			list.SetCurrentItem(0)
		}

		app.Draw()

		i += 1
	}

	if err := fetchCmd.Close(); err != nil {
		log.Fatalf("FETCH command failed: %v", err)
	}
}

func getEmailData(msg *imapclient.FetchMessageData) (*Email, error) {
	if msg == nil {
		return nil, errors.New("Email is null")
	}

	var bodySection *imapclient.FetchItemDataBodySection = nil
	for {
		item := msg.Next()
		if item == nil {
			break
		}

		section, ok := item.(imapclient.FetchItemDataBodySection)
		if ok {
			bodySection = &section
			break
		}
	}
	if bodySection == nil {
		return nil, errors.New("Message without body section")
	}

	mr, err := mail.CreateReader(bodySection.Literal)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Failed to create mail reader: %v", err))
	}

	h := mr.Header
	date, err := h.Date()
	if err != nil {
		return nil, errors.New(fmt.Sprintf("failed to parse Date header field: %v", err))
	}

	from, err := h.AddressList("From")
	if err != nil {
		return nil, errors.New(fmt.Sprintf("failed to parse From header field: %v", err))
	}

	to, err := h.AddressList("To")
	if err != nil {
		return nil, errors.New(fmt.Sprintf("failed to parse To header field: %v", err))
	}

	subject, err := h.Text("Subject")
	if err != nil {
		return nil, errors.New(fmt.Sprintf("failed to parse Subject header field: %v", err))
	}

	email := new(Email)
	email.Date = date
	email.From = from
	email.To = to
	email.Subject = subject
	email.Body = make(map[string]string)
	email.Attachments = make(map[string][]byte)

	for {
		p, err := mr.NextPart()
		if err == io.EOF {
			break
		} else if err != nil {
			log.Fatalf("failed to read message part: %v", err)
		}

		switch h := p.Header.(type) {
		case *mail.InlineHeader:
			b, _ := io.ReadAll(p.Body)
			ct, _, _ := h.ContentType()
			email.Body[ct] = string(b)
		case *mail.AttachmentHeader:
			filename, _ := h.Filename()
			buf := new(bytes.Buffer)
			buf.ReadFrom(p.Body)
			email.Attachments[filename] = buf.Bytes()
		}
	}

	return email, nil
}
