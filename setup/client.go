package setup

import (
	"go-twilio-taskrouter/config"

	"github.com/twilio/twilio-go"
)

func InitTwilioClient() *twilio.RestClient {
	accountSid := config.GetEnv("TWILIO_ACCOUNT_SID")
	authToken := config.GetEnv("TWILIO_AUTH_TOKEN")
	return twilio.NewRestClientWithParams(twilio.ClientParams{
		Username: accountSid,
		Password: authToken,
	})
}