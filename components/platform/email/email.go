package email

import (
	"fmt"
	"strings"

	"github.com/BrianLeishman/go-imap"
)

func Init(email, password, domain string, port int) (*imap.Dialer, error) {
	im, err := imap.New(email, password, domain, port)
	if err != nil {
		return nil, err
	}
	return im, nil
}

func GetFolders(im *imap.Dialer) ([]string, error) {
	folders, err := im.GetFolders()
	if err != nil {
		return nil, err
	}
	return folders, nil
}

func SelectFolder(im *imap.Dialer, folder string) error {
	err := im.SelectFolder(folder)
	if err != nil {
		return err
	}
	return nil
}

func GetEmails(im *imap.Dialer, folder string, count int) (map[int]*imap.Email, []int, error) {

	uids, err := im.GetUIDs(fmt.Sprintf("%d:*", count))
	if err != nil {
		return nil, nil, err
	}

	emails, err := im.GetEmails(uids...)
	if err != nil {
		return nil, nil, err
	}
	return emails, uids, nil
}

func FilterEmailsByTag(im *imap.Dialer, emails map[int]*imap.Email, tags ...string) (map[int]*imap.Email, error) {
	filteredUids := make([]int, 0)
	for _, email := range emails {
		for _, t := range tags {
			if strings.Contains(email.Subject, t) {
				filteredUids = append(filteredUids, email.UID)
			}
		}
	}

	if len(filteredUids) == 0 {
		return nil, fmt.Errorf("no emails found with tags: %v", tags)
	}

	filteredEmails, err := im.GetEmails(filteredUids...)
	if err != nil {
		return nil, err
	}
	return filteredEmails, nil
}

func AggregateEmailBody(emails map[int]*imap.Email) string {
	body := ""
	for _, email := range emails {
		body += email.Text + ". "
	}
	return body
}
