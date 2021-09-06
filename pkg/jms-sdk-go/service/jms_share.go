package service

import "github.com/jumpserver/koko/pkg/jms-sdk-go/model"

func (s *JMService) CreateShare(sessionId string, expired int) (res model.SharingSession, err error) {
	var postData struct {
		Session     string `json:"session"`
		ExpiredTime int    `json:"expired_time"`
	}
	postData.Session = sessionId
	postData.ExpiredTime = expired
	_, err = s.authClient.Post(ShareCreateURL, postData, &res)
	return
}
