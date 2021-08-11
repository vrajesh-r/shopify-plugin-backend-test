package mailer

import (
	"fmt"

	"github.com/getbread/shopify_plugin_backend/service/spb_config"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"github.com/sirupsen/logrus"
)

var sendgridConfig spb_config.SendgridConfig

func InitMailer(config spb_config.SendgridConfig) {
	sendgridConfig = config
}
func emailSendRequest(m *mail.SGMailV3) error {
	apiKey := sendgridConfig.SendgridMiltonApiKey.Unmask()
	if apiKey == "" {
		logrus.Error("(EmailSendRequest) SendGrid API key not found, please set SENDGRID_MILTON_API_KEY in your environment")
		return fmt.Errorf("No SendGrid API key")
	}

	request := sendgrid.GetRequest(apiKey, "/v3/mail/send", "https://api.sendgrid.com")
	request.Method = "POST"
	request.Body = mail.GetRequestBody(m)
	response, err := sendgrid.API(request)
	if err != nil {
		return err
	} else if response.StatusCode >= 400 {
		return fmt.Errorf(response.Body)
	}
	return nil
}

func SendPasswordResetLink(email, resetLink string) error {
	template := sendgridConfig.SendgridMiltonPasswordResetTemplate
	if template == "" {
		logrus.Error("(SendPasswordResetLink) SendGrid template ID not found, please set SENDGRID_MILTON_PASSWORD_RESET_TEMPLATE in your environment")
		return fmt.Errorf("No SendGrid template ID")
	}

	from := mail.NewEmail("Bread", "no-reply@getbread.com")
	to := mail.NewEmail("Merchant", email)
	content := mail.NewContent("text/html", " ")
	m := mail.NewV3MailInit(from, "Password Reset", to, content)
	m.Personalizations[0].SetSubstitution("-resetlink-", resetLink)
	m.SetTemplateID(template)

	err := emailSendRequest(m)
	if err != nil {
		return err
	}
	return nil
}
