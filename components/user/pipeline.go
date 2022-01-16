package user

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"main/components/platform/email"
	"main/components/platform/encryption"
	"main/components/platform/postgresmanager"
	"main/components/platform/wordcloud"
	"strings"
)

func (u *User) GenerateSummaryandWordCloud() (string, string, error) {

	if len(u.Tags) == 0 {
		return "", "", errors.New("no tags")
	}

	rawPass, err := base64.RawStdEncoding.DecodeString(u.Password)
	if err != nil {
		return  "", "", err
	}
	im, err := email.Init(u.Email, string(encryption.Decrypt(rawPass)), u.Domain, u.Port)
	if err != nil {
		return  "", "", err
	}

	email.SelectFolder(im, u.Folder)

	uidMap := make(map[string]int)
	err = json.Unmarshal([]byte(u.LastUID), &uidMap)
	if err != nil && !strings.Contains(err.Error(), "unexpected end of JSON input") {
		return  "", "", err
	}

	for _, tag := range u.Tags {
		if _, ok := uidMap[tag]; !ok || uidMap[tag] == 0 {
			uidMap[tag] = 1
		}
	}

	lowestUID := uidMap[u.Tags[0]]
	
	if len(u.Tags) > 1 {
		for _, tag := range u.Tags {
			if uidMap[tag] < lowestUID {
				lowestUID = uidMap[tag]
			}
		}
	}

	uids, err := im.GetUIDs("999999999:*")

	if err != nil {
		return  "", "", err
	}
	if len(uids) == 0 {
		return  "", "", errors.New("no emails found")
	}

	if uids[len(uids)-1] < lowestUID {
		lowestUID = uids[len(uids)-1]
	}

	emails, uidList, err := email.GetEmails(im, u.Folder, lowestUID)
	if err != nil {
		return  "", "", err
	}

	m := make(map[string]int)

	for _, tag := range u.Tags {
		m[tag] = uidList[len(uidList)-1]
	}

	bytes, err := json.Marshal(m)
	if err != nil {
		return  "", "", err
	}

	postgresmanager.Update(u, &User{LastUID: string(bytes)})

	filteredEmails, err := email.FilterEmailsByTag(im, emails, u.Tags, u.BlackListSenders, u.StartTime)
	if err != nil {
		return  "", "", err
	}

	body := email.AggregateEmailBody(filteredEmails)

	if len(body) == 0 {
		return  "", "", errors.New("no summary")

	}

	summarizedText := wordcloud.Summarize(body, u.SummaryCount)

	keywords := wordcloud.ExtractKeyWords(summarizedText)

	fileName, err := wordcloud.GenerateWordCloud(keywords)

	if err != nil {
		return  "", "", err
	}

	return summarizedText, fileName, nil
}
