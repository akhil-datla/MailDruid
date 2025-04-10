package main

import (
	"flag"
	"main/components/platform/encryption"
	"main/components/platform/postgresmanager"
	"main/components/scheduler"
	"main/components/user"
	"main/confidential"
	"main/server"
	"os"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/pterm/pterm"
)

type jwtCustomClaims struct {
	Value string
	jwt.StandardClaims
}

func main() {

	banner()

	dbuserPtr := flag.String("dbuser", "postgres", "PostgreSQL Database user")
	dbpassPtr := flag.String("dbpass", "password", "PostgreSQL Database password")
	dbnamePtr := flag.String("dbname", "postgres", "PostgreSQL Database name")
	portPtr := flag.String("httpport", "8080", "Port to run the server on")
	sendingEmailPtr := flag.String("email", "", "Email to send messages from")
	sendingEmailPasswordPtr := flag.String("pass", "", "Password for the sending email")
	smtpServerPtr := flag.String("domain", "", "SMTP server to send messages from")
	smtpPortPtr := flag.Int("smtpport", 0, "SMTP port to send messages from")
	loggerPtr := flag.Bool("log", false, "Enable HTTP request logging")
	signingKeyPtr := flag.String("signingkey", "", "Signing key for JWT")
	encryptionKeyPtr := flag.String("encryptionkey", "", "Encryption key for password encryption")

	flag.Parse()

	if *sendingEmailPtr == "" || *sendingEmailPasswordPtr == "" || *smtpServerPtr == "" || *smtpPortPtr == 0 || *signingKeyPtr == "" || *encryptionKeyPtr == "" {
		pterm.Error.Println("Please provide all the required parameters: email, password, smtp domain, smtp port, signing key, encryption key")
		os.Exit(1)
	}

	confidential.SigningKey = []byte(*signingKeyPtr)
	confidential.EncryptionKey = []byte(*encryptionKeyPtr)

	encryption.Encrypt("test")
	sclaims := jwt.StandardClaims{
		ExpiresAt: time.Now().Add(time.Second * 30).Unix(),
	}

	claims := &jwtCustomClaims{
		Value:          "test",
		StandardClaims: sclaims,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	_, err := token.SignedString(confidential.SigningKey)
	if err != nil {
		pterm.Error.Println("Error signing token: ", err)
		os.Exit(1)
	}

	err = postgresmanager.Open(*dbnamePtr, *dbuserPtr, *dbpassPtr)
	if err != nil {
		pterm.Error.Println("Error opening database: ", err)
		os.Exit(1)
	}

	err = postgresmanager.AutoCreateStruct(&user.User{})

	if err != nil {
		pterm.Error.Println("Error creating user table: ", err)
		os.Exit(1)
	}

	user.SendingEmail = *sendingEmailPtr
	user.SendingPassword = *sendingEmailPasswordPtr
	user.SMTPServer = *smtpServerPtr
	user.SMTPPort = *smtpPortPtr

	scheduler.Cleanup()

	scheduler.ScheduleTasks()

	server.Start(*portPtr, *loggerPtr)

}

func banner() {
	pterm.DefaultCenter.Print(pterm.DefaultHeader.WithFullWidth().WithBackgroundStyle(pterm.NewStyle(pterm.BgLightBlue)).WithMargin(10).Sprint("MailDruid"))
	pterm.Info.Println("Made by Akhil Datla")
}
