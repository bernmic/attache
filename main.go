package main

import (
	"flag"
	"fmt"
	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/emersion/go-message"
	_ "github.com/emersion/go-message/charset"
	"github.com/emersion/go-message/mail"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var (
	context Context = Context{}
)

type Context struct {
	username string
	password string
	server   string
	port     int
	tls      bool
	exclude  []string
	path     string
	client   *client.Client
}

func main() {
	parseArguments()
	if context.username == "" || context.password == "" || context.server == "" {
		fmt.Println("Missing arguments")
		flag.Usage()
		os.Exit(-1)
	}

	s := fmt.Sprintf("%s:%d", context.server, context.port)
	c, err := client.DialTLS(s, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer c.Logout()

	context.client = c

	if err := c.Login(context.username, context.password); err != nil {
		log.Fatal(err)
	}

	mailboxes := make(chan *imap.MailboxInfo, 100)
	done := make(chan error, 1)
	go func() {
		done <- c.List("", "*", mailboxes)
	}()

	log.Println("Mailboxes:")
	for m := range mailboxes {
		context.readAttachments(m)
	}

	if err := <-done; err != nil {
		log.Fatal(err)
	}
}

func (c *Context) readAttachments(m *imap.MailboxInfo) {
	// if in exclude list, return
	for _, f := range c.exclude {
		if m.Name == f {
			return
		}
	}
	log.Println("* " + m.Name)
	mbox, err := c.client.Select(m.Name, false)
	if err != nil {
		log.Fatal(err)
	}
	if mbox.Messages == 0 {
		return
	}

	log.Printf("Flags for %s: %s\n", m.Name, mbox.Flags)
	seqset := new(imap.SeqSet)
	seqset.AddRange(1, mbox.Messages)

	// Get the whole message body
	var section imap.BodySectionName
	items := []imap.FetchItem{section.FetchItem()}

	messages := make(chan *imap.Message, 10)
	done := make(chan error, 1)
	go func() {
		done <- c.client.Fetch(seqset, items, messages)
	}()

	for msg := range messages {
		saveAttachments(msg)
	}

	if err := <-done; err != nil {
		log.Fatal(err)
	}
}

func saveAttachments(m *imap.Message) {
	var section imap.BodySectionName
	r := m.GetBody(&section)
	if r == nil {
		log.Println("NO BODY")
		return
	}
	// Create a new mail reader
	mr, err := mail.CreateReader(r)
	if err != nil {
		log.Println(err)
		return
	}

	// Print some info about the message
	/*
	header := mr.Header
	if date, err := header.Date(); err == nil {
		log.Println("Date:", date)
	}
	if from, err := header.AddressList("From"); err == nil {
		log.Println("From:", from)
	}
	if to, err := header.AddressList("To"); err == nil {
		log.Println("To:", to)
	}
	if subject, err := header.Subject(); err == nil {
		log.Println("Subject:", subject)
	}
*/
	from, err := mr.Header.AddressList("From")
	if err != nil {
		log.Println(err)
		return
	}
	if len(from) < 1 {
		log.Printf("No sender in Mail")
		return
	}
	date, err := mr.Header.Date()
	if err != nil {
		log.Println(err)
		return
	}
	dir := fmt.Sprintf("%s/%s/%s", context.path, from[0].Address, date.Format("2006-01-02T03-04-05"))
	dir, err = filepath.Abs(dir)
	_, err = os.Stat(dir)
	if err == nil || os.IsExist(err) {
		log.Printf("%s exists. Skip it.\n", dir)
		return
	}

	// Process each message's part
	for {
		p, err := mr.NextPart()
		if err == io.EOF {
			log.Printf("Error is %v", err.(type))
			break
		} else if err != nil {
			log.Printf("Error is %w", err)
			_, et := err.(message.UnknownCharsetError)
			if !et {
				log.Println(err)
				// continue
			}
			log.Printf("NextPart-Error: %v\n", err)
			return
			//continue
		}

		switch h := p.Header.(type) {
		case *mail.InlineHeader:
			// This is the message's text (can be plain-text or HTML)
			//b, _ := ioutil.ReadAll(p.Body)
			//log.Println("Got text: %v", string(b))
		case *mail.AttachmentHeader:
			// This is an attachment
			filename, _ := h.Filename()
			log.Printf("Got attachment: %v\n", filename)
			if strings.Trim(filename, " ") == "" {
				filename = "unknown"
			}
			// Create dir if needed
			if os.MkdirAll(dir, os.ModePerm) != nil {
				log.Printf("Error creating dir %s: %v\n", dir, err)
				continue
			}
			// Create file with attachment name
			file, err := os.Create(dir + "/" + filename)
			if err != nil {
				log.Printf("Error creating file %s/%s: %v\n", dir, filename, err)
				continue
			}
			// using io.Copy instead of io.ReadAll to avoid insufficient memory issues
			size, err := io.Copy(file, p.Body)
			if err != nil {
				log.Printf("Error reading attachment %s: %v\n", filename, err)
				continue
			}
			log.Printf("Saved %v bytes into %v\n", size, dir + "/" + filename)
		}
	}
}

func parseArguments() {
	flag.StringVar(&context.username, "username", "", "username for the imap server")
	flag.StringVar(&context.password, "password", "", "password for the imap server")
	flag.StringVar(&context.server, "server", "", "imap server name")
	flag.IntVar(&context.port, "port", 993, "Port of the imap server")
	flag.BoolVar(&context.tls, "tls", true, "TLS encrypted session")
	flag.StringVar(&context.path, "path", "", "Path where the attachments should be saved")
	var exclude string
	flag.StringVar(&exclude, "exclude", "Spam,Trash,Deleted Messages", "Comma separated list of IMAP folders to be excluded")

	flag.Parse()

	val, ok := os.LookupEnv("ATTACHE_USERNAME")
	if context.username == "" && ok {
		context.username = val
	}
	val, ok = os.LookupEnv("ATTACHE_PASSWORD")
	if context.password == "" && ok {
		context.password = val
	}
	val, ok = os.LookupEnv("ATTACHE_SERVER")
	if context.server == "" && ok {
		context.server = val
	}
	val, ok = os.LookupEnv("ATTACHE_PATH")
	if context.path == "" && ok {
		context.path = val
	}

	val, ok = os.LookupEnv("ATTACHE_EXCLUDE")
	if exclude == "" && ok {
		exclude = val
	}
	if exclude != "" {
		context.exclude = strings.Split(exclude, ",")
	}
}
