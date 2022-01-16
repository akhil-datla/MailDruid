package email

import (
	"fmt"
	"log"
	"strings"
	"time"

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

func FilterEmailsByTag(im *imap.Dialer, emails map[int]*imap.Email, tags, blackListSenders []string, startTime time.Time) (map[int]*imap.Email, error) {
	var err error

	filteredUids, err := filterTags(emails, tags, blackListSenders, startTime)

	if err != nil {
		return nil, err
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

func getSender(emails imap.EmailAddresses) (string, error) {
	for address := range emails {
		return address, nil
	}
	return "", fmt.Errorf("no sender found")
}

func checkBlackListSenders(email *imap.Email, blackListSenders []string) bool {
	for _, sender := range blackListSenders {
		address, err := getSender(email.From)
		if err != nil {
			log.Printf("error getting sender: %v", err)
		}
		if address == sender {
			return true
		}
	}
	return false
}

func checkTime(email *imap.Email, startTime time.Time) bool {
	if startTime.IsZero() {
		return true
	}
	
	if email.Sent.After(startTime) {
		return true
	}

	return false
}

func filterTags(emails map[int]*imap.Email, tags, blackListSenders []string, startTime time.Time) ([]int, error) {
	if len(tags) == 0 {
		return nil, fmt.Errorf("no tags provided")
	}
	filteredUIDs := make([]int, 0)
	for _, email := range emails {
		for _, tag := range tags {
			lowerTag := strings.ToLower(tag)
			lowerSubject := strings.ToLower(email.Subject)
			if strings.Contains(lowerSubject, lowerTag) && !checkBlackListSenders(email, blackListSenders) && checkTime(email, startTime) {
				filteredUIDs = append(filteredUIDs, email.UID)
			}
		}
	}
	return filteredUIDs, nil
}
