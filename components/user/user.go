package user

import (
	"crypto/tls"
	"encoding/base64"
	"errors"
	"fmt"
	"main/components/platform/encryption"
	"main/components/platform/postgresmanager"
	"time"

	"github.com/lib/pq"
	"github.com/matcornic/hermes/v2"
	uuid "github.com/satori/go.uuid"
	gomail "gopkg.in/mail.v2"
)

var SendingEmail = ""
var SendingPassword = ""
var SMTPServer = ""
var SMTPPort = 0

type User struct {
	ID               string         `json:"id"`
	Name             string         `json:"name"`
	Email            string         `json:"email"`
	ReceivingEmail   string         `json:"receivingEmail"`
	Password         string         `json:"password"`
	Domain           string         `json:"domain"`
	Port             int            `json:"port"`
	Folder           string         `json:"folder"`
	Tags             pq.StringArray `gorm:"type:text[]" json:"tags"`
	BlackListSenders pq.StringArray `gorm:"type:text[]" json:"blackListSenders"`
	StartTime        time.Time      `json:"startTime"`
	SummaryCount     int            `json:"summaryCount"`
	LastUID          string         `json:"lastUID"`
	UpdateInterval   string         `json:"updateInterval"`
}

var h = hermes.Hermes{
	Product: hermes.Product{
		Name:      "MailDruid",
		Copyright: "Made by Akhil Datla and Alexander Ott",
	},
}

func CreateUser(name, email, receivingEmail, password, domain string, port int) error {
	id := uuid.NewV4()
	encryptedPass := base64.RawStdEncoding.EncodeToString(encryption.Encrypt(password))
	err := postgresmanager.Save(&User{ID: id.String(), Name: name, Email: email, ReceivingEmail: receivingEmail, Password: encryptedPass, Domain: domain, Port: port})
	return err
}

func AuthUser(email, password string) (string, error) {
	var user User
	err := postgresmanager.Query(&User{Email: email}, &user)
	if err != nil {
		return "", err
	}
	if err != nil {
		return "", err
	}
	rawPass, err := base64.RawStdEncoding.DecodeString(user.Password)
	if err != nil {
		return "", err
	}
	if string(encryption.Decrypt(rawPass)) == password {
		return user.ID, nil
	} else {
		return "", errors.New("invalid email or password")
	}
}

func ReadUser(id string) (*User, error) {
	var user *User
	err := postgresmanager.Query(User{ID: id}, &user)
	return user, err
}

func UpdateUser(id, name, email, receivingEmail, oldPassword, newPassword, domain, folder string, port int) []error {
	errorList := make([]error, 0)
	user, err := ReadUser(id)
	if err != nil {
		errorList = append(errorList, err)
	}

	if name != "" {
		err := postgresmanager.Update(&user, User{Name: name})
		if err != nil {
			errorList = append(errorList, err)
		}
	}

	if email != "" {
		err := postgresmanager.Update(&user, User{Email: email})
		if err != nil {
			errorList = append(errorList, err)
		}
	}

	if receivingEmail != "" {
		err := postgresmanager.Update(&user, User{ReceivingEmail: receivingEmail})
		if err != nil {
			errorList = append(errorList, err)
		}
	}

	if oldPassword != "" && newPassword != "" {
		rawPass, err := base64.RawStdEncoding.DecodeString(user.Password)
		if err != nil {
			errorList = append(errorList, err)
		}
		checkPass := string(encryption.Decrypt(rawPass)) == oldPassword
		if checkPass {
			encryptedPass := base64.RawStdEncoding.EncodeToString(encryption.Encrypt(newPassword))
			err := postgresmanager.Update(&user, User{Password: encryptedPass})
			if err != nil {
				errorList = append(errorList, err)
			}
		} else {
			errorList = append(errorList, errors.New("old password is invalid"))
		}

	}

	if name != "" {
		err := postgresmanager.Update(&user, User{Name: name})
		if err != nil {
			errorList = append(errorList, err)
		}
	}

	if domain != "" {
		err := postgresmanager.Update(&user, User{Domain: domain})
		if err != nil {
			errorList = append(errorList, err)
		}
	}

	if folder != "" {
		err := postgresmanager.Update(&user, User{Folder: folder})
		if err != nil {
			errorList = append(errorList, err)
		}
	}

	if port != 0 {
		err := postgresmanager.Update(&user, User{Port: port})
		if err != nil {
			errorList = append(errorList, err)
		}
	}

	return errorList
}

func DeleteUser(id string) error {
	user, err := ReadUser(id)
	if err != nil {
		return err
	} else {
		err := postgresmanager.Delete(user)
		return err
	}
}

func UpdateFolder(id string, folder string) error {
	user, err := ReadUser(id)
	if err != nil {
		return err
	} else {
		err := postgresmanager.Update(&user, User{Folder: folder})
		return err
	}
}

func UpdateTags(id string, tags []string) error {
	user, err := ReadUser(id)
	if err != nil {
		return err
	}
	err = postgresmanager.Update(&user, User{Tags: tags})
	if err != nil {
		return err
	}
	return nil
}

func UpdateBlackListSenders(id string, blackListSenders []string) error {
	user, err := ReadUser(id)
	if err != nil {
		return err
	}
	err = postgresmanager.Update(&user, User{BlackListSenders: blackListSenders})
	if err != nil {
		return err
	}
	return nil
}

func UpdateStartTime(id string, startTime string) error {
	user, err := ReadUser(id)
	if err != nil {
		return err
	}

	var startTimeParsed time.Time
	if startTime != "" {
		startTimeParsed, err = time.Parse(time.RFC3339, startTime)
		if err != nil {
			return err
		}
	} else {
		startTimeParsed = time.Time{}
	}

	err = postgresmanager.Update(&user, User{StartTime: startTimeParsed})
	if err != nil {
		return err
	}
	return nil
}

func UpdateSummaryCount(id string, count int) error {
	user, err := ReadUser(id)
	if err != nil {
		return err
	}
	err = postgresmanager.Update(&user, User{SummaryCount: count})
	if err != nil {
		return err
	}
	return nil
}

func (u *User) SendEmail(summary, fileName, errorMessage string) error {
	m := gomail.NewMessage()

	// Set E-Mail sender
	m.SetHeader("From", SendingEmail)

	// Set E-Mail receivers
	m.SetHeader("To", u.ReceivingEmail)

	// Set E-Mail subject
	m.SetHeader("Subject", "MailDruid Summary")

	email := hermes.Email{
		Body: hermes.Body{
			Name:      u.Name,
			Signature: "Regards",
			Intros: []string{
				fmt.Sprintf("Below is the summary and word cloud for: %s\n", u.Tags),
				summary,
				errorMessage,
			},
			Outros: []string{
				"If you have any questions, please contact me.",
			},
		},
	}

	// Generate an HTML email with the provided contents (for modern clients)
	emailBody, err := h.GenerateHTML(email)
	if err != nil {
		return err
	}

	if err != nil {
		return err
	}

	// Set E-Mail body. You can set plain text or html with text/html
	m.SetBody("text/html", emailBody)
	if fileName != "" {
		m.Embed(fileName)
	}

	// Settings for SMTP server
	d := gomail.NewDialer(SMTPServer, SMTPPort, SendingEmail, SendingPassword)

	// This is only needed when SSL/TLS certificate is not valid on server.
	// In production this should be set to false.
	d.TLSConfig = &tls.Config{ServerName: SMTPServer}

	// Now send E-Mail
	err = d.DialAndSend(m)
	return err
}
