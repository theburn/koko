package auth

import (
	"context"
	"github.com/jumpserver/koko/pkg/jms-sdk-go/service"
	"time"

	"github.com/jumpserver/koko/pkg/jms-sdk-go/model"
	"github.com/jumpserver/koko/pkg/logger"
)

type authOptions struct {
	MFAType string
	Url     string
}

type UserAuthClient struct {
	*service.UserClient

	authOptions map[string]authOptions

	mfaTypes    []string
	stageStatus Stage

	selectedMFAType string
}

func (u *UserAuthClient) SetOption(setters ...service.UserClientOption) {
	u.UserClient.SetOption(setters...)
}

func (u *UserAuthClient) Authenticate(ctx context.Context) (user model.User, authStatus StatusAuth) {
	authStatus = authFailed
	resp, err := u.UserClient.GetAPIToken()
	if err != nil {
		logger.Errorf("User %s Authenticate err: %s", u.Opts.Username, err)
		return
	}
	if resp.Err != "" {
		switch resp.Err {
		case ErrLoginConfirmWait:
			logger.Infof("User %s login need confirmation", u.Opts.Username)
			authStatus = authConfirmRequired
			u.stageStatus = StageConfirm
		case ErrMFARequired:
			u.mfaTypes = nil
			for _, choiceType := range resp.Data.Choices {
				u.authOptions[choiceType] = authOptions{
					MFAType: choiceType,
					Url:     resp.Data.Url,
				}
				u.mfaTypes = append(u.mfaTypes, choiceType)
			}
			logger.Infof("User %s login need MFA", u.Opts.Username)
			authStatus = authMFARequired
			u.stageStatus = StageMFASelect
		default:
			logger.Errorf("User %s login err: %s", u.Opts.Username, resp.Err)
		}
		return
	}
	if resp.Token != "" {
		return resp.User, authSuccess
	}
	return
}

func (u *UserAuthClient) CheckUserOTP(ctx context.Context, code string) (user model.User, authStatus StatusAuth) {
	authStatus = authFailed
	authData, ok := u.authOptions[u.selectedMFAType]
	if !ok {
		logger.Errorf("User %s use %s check MFA not found", u.Opts.Username, u.selectedMFAType)
		return
	}
	data := map[string]interface{}{
		"code":        code,
		"remote_addr": u.Opts.RemoteAddr,
		"login_type":  u.Opts.LoginType,
		"type":        authData.MFAType,
	}

	resp, err := u.UserClient.SendOTPRequest(&service.OTPRequest{
		ReqURL:  authData.Url,
		ReqBody: data,
	})
	if err != nil {
		logger.Errorf("User %s use %s check MFA err: %s", u.Opts.Username, authData.MFAType, err)
		return
	}
	if resp.Err != "" {
		logger.Errorf("User %s use %s check MFA err: %s", u.Opts.Username, authData.MFAType, resp.Err)
		return
	}
	if resp.Msg == "ok" {
		logger.Infof("User %s check MFA success, check if need admin confirm", u.Opts.Username)
		return u.Authenticate(ctx)
	}
	logger.Errorf("User %s failed to use %s check MFA", u.Opts.Username, authData.MFAType)
	return
}

func (u *UserAuthClient) GetMFAOptions() []string {
	return u.mfaTypes
}

func (u *UserAuthClient) SetAuthMFAType(mfaType string) error {
	logger.Infof("User select mfa type %s", mfaType)
	u.selectedMFAType = mfaType
	return u.UserClient.SelectMFAChoice(mfaType)
}

func (u *UserAuthClient) SetNextStage(next Stage) {
	u.stageStatus = next
}

func (u *UserAuthClient) CurrentStage() Stage {
	return u.stageStatus
}

func (u *UserAuthClient) GetSelectedMFAType() string {
	return u.selectedMFAType
}

const (
	ErrLoginConfirmWait     = "login_confirm_wait"
	ErrLoginConfirmRejected = "login_confirm_rejected"
	ErrLoginConfirmRequired = "login_confirm_required"
	ErrMFARequired          = "mfa_required"
	ErrPasswordFailed       = "password_failed"
)

func (u *UserAuthClient) CheckConfirm(ctx context.Context) (user model.User, authStatus StatusAuth) {
	t := time.NewTicker(5 * time.Second)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			logger.Errorf("User %s exit and cancel confirmation", u.Opts.Username)
			u.CancelConfirm()
			return
		case <-t.C:
			resp, err := u.UserClient.CheckConfirmAuthStatus()
			if err != nil {
				logger.Errorf("User %s check confirm err: %s", u.Opts.Username, err)
				return
			}
			if resp.Err != "" {
				switch resp.Err {
				case ErrLoginConfirmWait:
					logger.Infof("User %s still wait confirm", u.Opts.Username)
					continue
				case ErrLoginConfirmRejected:
					logger.Infof("User %s confirmation was rejected by admin", u.Opts.Username)
				default:
					logger.Infof("User %s confirmation was rejected by err: %s", u.Opts.Username, resp.Err)
				}
				return
			}
			if resp.Msg == "ok" {
				logger.Infof("User %s confirmation was accepted", u.Opts.Username)
				return u.Authenticate(ctx)
			}
		}
	}
}

func (u *UserAuthClient) CancelConfirm() {
	err := u.UserClient.CancelConfirmAuth()
	if err != nil {
		logger.Errorf("Cancel User %s confirmation err: %s", u.Opts.Username, err)
		return
	}
	logger.Infof("Cancel User %s confirmation success", u.Opts.Username)
}

type StatusAuth int64

const (
	authSuccess StatusAuth = iota + 1
	authFailed
	authMFARequired
	authConfirmRequired
)

type Stage int

const (
	StageMFASelect Stage = iota + 1
	StageMFACode
	StageConfirm
)
