package auth

import (
	"net"
	"strconv"
	"strings"

	"github.com/gliderlabs/ssh"
	gossh "golang.org/x/crypto/ssh"

	"github.com/jumpserver/koko/pkg/common"
	"github.com/jumpserver/koko/pkg/jms-sdk-go/service"
	"github.com/jumpserver/koko/pkg/logger"
	"github.com/jumpserver/koko/pkg/sshd"
)

type SSHAuthFunc func(ctx ssh.Context, password, publicKey string) (res sshd.AuthStatus)

func SSHPasswordAndPublicKeyAuth(jmsService *service.JMService) SSHAuthFunc {
	return func(ctx ssh.Context, password, publicKey string) (res sshd.AuthStatus) {
		username := GetUsernameFromSSHCtx(ctx)
		authMethod := "publickey"
		action := actionAccepted
		res = sshd.AuthFailed
		if password != "" {
			authMethod = "password"
		}
		remoteAddr, _, _ := net.SplitHostPort(ctx.RemoteAddr().String())
		userAuthClient, ok := ctx.Value(ContextKeyClient).(*UserAuthClient)
		if !ok {
			newClient := jmsService.CloneClient()

			userClient := service.NewUserClient(
				service.UserClientUsername(username),
				service.UserClientRemoteAddr(remoteAddr),
				service.UserClientLoginType("T"),
				service.UserClientHttpClient(&newClient),
			)
			userAuthClient = &UserAuthClient{
				UserClient:  userClient,
				authOptions: make(map[string]authOptions),
			}
			ctx.SetValue(ContextKeyClient, userAuthClient)
		}
		userAuthClient.SetOption(service.UserClientPassword(password),
			service.UserClientPublicKey(publicKey))
		logger.Infof("SSH conn[%s] authenticating user %s %s", ctx.SessionID(), username, authMethod)
		user, authStatus := userAuthClient.Authenticate(ctx)
		switch authStatus {
		case authMFARequired:
			action = actionPartialAccepted
			res = sshd.AuthPartiallySuccessful
		case authSuccess:
			res = sshd.AuthSuccessful
			ctx.SetValue(ContextKeyUser, &user)
		case authConfirmRequired:
			required := true
			ctx.SetValue(ContextKeyConfirmRequired, &required)
			action = actionPartialAccepted
			res = sshd.AuthPartiallySuccessful
		default:
			action = actionFailed
		}
		logger.Infof("SSH conn[%s] %s %s for %s from %s", ctx.SessionID(),
			action, authMethod, username, remoteAddr)
		return

	}
}

func SSHKeyboardInteractiveAuth(ctx ssh.Context, challenger gossh.KeyboardInteractiveChallenge) (res sshd.AuthStatus) {
	if value, ok := ctx.Value(ContextKeyConfirmFailed).(*bool); ok && *value {
		return sshd.AuthFailed
	}
	username := GetUsernameFromSSHCtx(ctx)
	remoteAddr, _, _ := net.SplitHostPort(ctx.RemoteAddr().String())
	res = sshd.AuthFailed

	client, ok := ctx.Value(ContextKeyClient).(*UserAuthClient)
	if !ok {
		logger.Errorf("SSH conn[%s] user %s Mfa Auth failed: not found session client.",
			ctx.SessionID(), username)
		return
	}
	opts := client.GetMFAOptions()
	if len(opts) == 1 {
		if ok2 := client.SetAuthMFAType(opts[0]); !ok2 {
			logger.Errorf("SSH conn[%s] user %s select mfa choice failed",
				ctx.SessionID(), username)
			return
		}
		client.SetNextStage(StageMFACode)
	}
	instruction, question := CreateChallengerInstruction(opts)
	currentStage := client.CurrentStage()
	var authFunc func(string) sshd.AuthStatus
	var answers []string
	var err error
	switch currentStage {
	case StageConfirm:
		answers, err = challenger(username, confirmInstruction, []string{confirmQuestion}, []bool{true})
		if err != nil {
			client.CancelConfirm()
			logger.Errorf("SSH conn[%s] user %s happened err: %s", ctx.SessionID(), username, err)
			return
		}
		authFunc = func(answer string) (res sshd.AuthStatus) {
			res = sshd.AuthFailed
			switch strings.TrimSpace(strings.ToLower(answer)) {
			case "yes", "y", "":
				logger.Infof("SSH conn[%s] checking user %s login confirm", ctx.SessionID(), username)
				user, authStatus := client.CheckConfirm(ctx)
				switch authStatus {
				case authSuccess:
					res = sshd.AuthSuccessful
					ctx.SetValue(ContextKeyUser, &user)
					logger.Infof("SSH conn[%s] checking user %s login confirm success", ctx.SessionID(), username)
					return
				}
			case "no", "n":
				logger.Infof("SSH conn[%s] user %s cancel login", ctx.SessionID(), username)
				client.CancelConfirm()
			default:
				return
			}
			failed := true
			ctx.SetValue(ContextKeyConfirmFailed, &failed)
			logger.Infof("SSH conn[%s] checking user %s login confirm failed", ctx.SessionID(), username)
			return
		}
	case StageMFASelect:
		answers, err = challenger(username, instruction, []string{question}, []bool{true})
		if err != nil {
			logger.Errorf("SSH conn[%s] user %s happened err: %s", ctx.SessionID(), username, err)
			return
		}
		authFunc = func(answer string) (res sshd.AuthStatus) {
			res = sshd.AuthFailed
			index, err2 := strconv.Atoi(answer)
			if err2 != nil {
				logger.Errorf("SSH conn[%s] user %s input wrong answer: %s", ctx.SessionID(), username, err2)
				return
			}
			optIndex := index - 1
			if optIndex < 0 || optIndex >= len(opts) {
				logger.Errorf("SSH conn[%s] user %s input wrong index: %d", ctx.SessionID(), username, index)
				return
			}
			optType := opts[optIndex]
			if ok = client.SetAuthMFAType(optType); !ok {
				logger.Errorf("SSH conn[%s] select MFA choice %s failed", ctx.SessionID(), optType)
				return
			}
			res = sshd.AuthPartiallySuccessful
			client.SetNextStage(StageMFACode)
			return
		}
	case StageMFACode:
		answers, err = challenger(username, instruction, []string{question}, []bool{true})
		if err != nil {
			logger.Errorf("SSH conn[%s] user %s happened err: %s", ctx.SessionID(), username, err)
			return
		}
		authFunc = func(answer string) (res sshd.AuthStatus) {
			res = sshd.AuthFailed
			user, authStatus := client.CheckUserOTP(ctx, answer)
			switch authStatus {
			case authSuccess:
				res = sshd.AuthSuccessful
				ctx.SetValue(ContextKeyUser, &user)
				logger.Infof("SSH conn[%s] %s MFA for %s from %s", ctx.SessionID(),
					actionAccepted, username, remoteAddr)
			case authConfirmRequired:
				res = sshd.AuthPartiallySuccessful
				required := true
				ctx.SetValue(ContextKeyConfirmRequired, &required)
				logger.Infof("SSH conn[%s] %s MFA for %s from %s", ctx.SessionID(),
					actionPartialAccepted, username, remoteAddr)
			default:
				logger.Errorf("SSH conn[%s] %s MFA for %s from %s", ctx.SessionID(),
					actionFailed, username, remoteAddr)
			}
			return
		}
		logger.Infof("SSH conn[%s] checking user %s mfa code", ctx.SessionID(), username)
	default:
		return
	}
	if len(answers) != 1 {
		return
	}
	return authFunc(answers[0])
}

const (
	ContextKeyUser            = "CONTEXT_USER"
	ContextKeyClient          = "CONTEXT_CLIENT"
	ContextKeyConfirmRequired = "CONTEXT_CONFIRM_REQUIRED"
	ContextKeyConfirmFailed   = "CONTEXT_CONFIRM_FAILED"

	ContextKeyDirectLoginFormat = "CONTEXT_DIRECT_LOGIN_FORMAT"
)

type DirectLoginAssetReq struct {
	Username    string
	SysUserInfo string
	AssetInfo   string
}

func (d *DirectLoginAssetReq) IsUUIDString() bool {
	for _, item := range []string{d.SysUserInfo, d.AssetInfo} {
		if !common.ValidUUIDString(item) {
			return false
		}
	}
	return true
}

const (
	SeparatorATSign   = "@"
	SeparatorHashMark = "#"
)

func parseUserFormatBySeparator(s, Separator string) (DirectLoginAssetReq, bool) {
	authInfos := strings.Split(s, Separator)
	if len(authInfos) != 3 {
		return DirectLoginAssetReq{}, false
	}
	req := DirectLoginAssetReq{
		Username:    authInfos[0],
		SysUserInfo: authInfos[1],
		AssetInfo:   authInfos[2],
	}
	return req, true
}

func ParseDirectUserFormat(s string) (DirectLoginAssetReq, bool) {
	for _, separator := range []string{SeparatorATSign, SeparatorHashMark} {
		if req, ok := parseUserFormatBySeparator(s, separator); ok {
			return req, true
		}
	}
	return DirectLoginAssetReq{}, false
}

func GetUsernameFromSSHCtx(ctx ssh.Context) string {
	if directReq, ok := ctx.Value(ContextKeyDirectLoginFormat).(*DirectLoginAssetReq); ok {
		return directReq.Username
	}
	username := ctx.User()
	if req, ok := ParseDirectUserFormat(username); ok {
		username = req.Username
		ctx.SetValue(ContextKeyDirectLoginFormat, &req)
	}
	return username
}
