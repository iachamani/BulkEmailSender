package main

import (
	"bufio"
	"bytes"
	"html/template"
	"log"
	"net/smtp"
	"os"

	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type Message interface {
	Get() interface{}
}

type StringMessage string

func (s StringMessage) Get() interface{} {
	return string(s)
}

type TemplateMessage struct {
	*template.Template
}

func (t TemplateMessage) Get() interface{} {
	var tpl bytes.Buffer
	if t.Template == nil {
		return ""
	}
	err := t.Template.Execute(&tpl, nil)
	if err != nil {
		log.Fatal(err)
	}
	return tpl.String()
}

// convert type Message to bytes type
type messageWrapper struct {
	Message
}

func (mw messageWrapper) Bytes() []byte {
	var b []byte
	switch m := mw.Message.(type) {
	case StringMessage:
		b = []byte(m)
	case TemplateMessage:
		var tpl bytes.Buffer
		err := m.Template.Execute(&tpl, nil)
		if err != nil {
			log.Fatal(err)
		}
		b = tpl.Bytes()
	}
	return b
}
func main() {
	var message Message
	myApp := app.New()
	myWindow := myApp.NewWindow("BulkEmailSender")

	selected := widget.NewRadioGroup([]string{"html", "plaintxt"}, func(value string) {
		log.Println("User choice set to", value)
	})
	// check if the user has selected html or plaintxt and then call the appropriate function
	if selected.Selected == "html" {
		tmpl, err := template.ParseFiles("html/email.html")
		if err != nil {
			panic(err)
		}

		message = TemplateMessage{tmpl}

	} else {
		//read plaintext message from a file
		plaintxt, err := os.ReadFile("message/email.txt")
		if err != nil {
			panic(err)
		}
		message = StringMessage(plaintxt)
	}
	username := widget.NewEntry()
	username.SetPlaceHolder("Enter username...")
	hostname := widget.NewEntry()
	hostname.SetPlaceHolder("Enter hostname...")
	password := widget.NewPasswordEntry()
	password.SetPlaceHolder("Enter password...")
	port := widget.NewEntry()
	port.SetPlaceHolder("Enter port...")
	auth := smtp.PlainAuth("", username.Text, password.Text, hostname.Text)
	// read all the recipients from a file
	recipients := []string{}
	// Read the file line by line and append the lines to the recipients slice
	lines, err := readLines("recipients.txt")
	if err != nil {
		log.Fatalf("readLines: %s", err)
	}
	for _, line := range lines {
		recipients = append(recipients, line)
	}
	wrapper := messageWrapper{Message: message}
	bytes_message := wrapper.Bytes()
	Send := widget.NewButton("Send", func() {
		err := smtp.SendMail(hostname.Text+":"+port.Text, auth, username.Text, recipients, bytes_message)
		if err != nil {
			log.Fatal(err)
		}
	})
	myWindow.SetContent(container.NewVBox(selected, username, hostname, password, port, Send))
	myWindow.ShowAndRun()
}

func readLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}
