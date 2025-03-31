package emailer

import (
    "context"
    "errors"
    "fmt"
    "io/ioutil"
    "path/filepath"
    "runtime"
    "text/template"
    "time"

    // AWS Packages

    "github.com/aws/aws-sdk-go-v2/service/ses"
    "github.com/aws/aws-sdk-go-v2/service/ses/types"
)

type EmailSender interface {
    SendEmail(ctx context.Context, params *ses.SendEmailInput) (*ses.SendEmailOutput, error)
}

type SESClient struct {
    Client EmailSender
}

type EmailData struct {
    Name        string
    Email       string
    Subject     string
    Message     string
    Timestamp   string
}

func NewSESClient(client EmailSender) *SESClient {
    return &SESClient {
        Client: client,
    }
}

// Sends a contact from email
func (s *SESClient) SendContactEmail(ctx context.Context, toEmail, fromEmail string, data EmailData) error {
    // Set timestamp if it hasnt be intializaed
    if data.Timestamp == "" {
        data.Timestamp = time.Now().Format(time.RFC3339) // Y/M/D H/M/S
    }

    textContent := s.formatTextEmail(data)

    htmlContent, err := s.formatHTMLEmail(data)
    if err != nil {
        return fmt.Errorf("Error formatting HTML email: %w", err)
    }

    // SES Input

    input := &ses.SendEmailInput {
        Destination: &types.Destination {
            ToAddresses: []string{toEmail},
        },
        Message: &types.Messages {
            Body: &types.Body {
                Text: &types.Content {
                    Data: &textContent,
                },
                HTML: &types.Content {
                    Data: &htmlContent,
                },
            },
            Subject: &types.Content {
                Data: &data.Subject,
            },
        },
        Source: &fromEmail,
    }
    
    _, err = s.Client.SendEmail(ctx, input)
    
    if err != nil {
        return fmt.Errorf("Error sending email: %w", err)
    }

    return nil
}

    // Send Alert Email 

    func (s *SESClient) SendAlertEmail(ctx context.Context, fromEmail string, toEmails []string, err error, originalData interface{}) error {
	if err == nil {
		return errors.New("no error provided for alert email")
	}

	// Create alert email subject
	subject := fmt.Sprintf("⚠️ Contact Form Submission Failure - %s", time.Now().Format(time.RFC3339))

	// Create alert email body
	body := fmt.Sprintf(`
            ⚠️ FORM SUBMISSION FAILURE ALERT ⚠️

            Timestamp: %s
            Error: %s

            Original Form Data:
            %+v
        `, time.Now().Format(time.RFC3339), err.Error(), originalData)

	// Create SES input
	input := &ses.SendEmailInput{
		Destination: &types.Destination{
			ToAddresses: toEmails,
		},
		Message: &types.Message{
			Body: &types.Body{
				Text: &types.Content{
					Data: &body,
				},
			},
			Subject: &types.Content{
				Data: &subject,
			},
		},
		Source: &fromEmail,
	}

	// Send email
	_, err = s.Client.SendEmail(ctx, input)
	if err != nil {
		return fmt.Errorf("error sending alert email: %w", err)
	}

	return nil
    }

// Format text version of the email
    func (s *SESClient) formatTextEmail(data EmailData) string {
	return fmt.Sprintf(`
            NEW MESSAGE

            Name: %s
            Email: %s

            Subject: %s

            Message:
            %s

            Timestamp: %s
        `, data.Name, data.Email, data.Subject, data.Message, data.Timestamp)
    }

// Format HTML
    func (s *SESClient) formatHTMLEmail(data EmailData) (string, error) {
	// Get the location of the email template
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return "", errors.New("unable to determine current file path")
	}

	// Navigate to the templates directory
	templateDir := filepath.Join(filepath.Dir(filename), "../templates")
	templatePath := filepath.Join(templateDir, "email.html")

	// Read the template file
	templateContent, err := ioutil.ReadFile(templatePath)
	if err != nil {
		// Fallback to a simple inline template if file can't be read
		return s.formatFallbackHTMLEmail(data), nil
	}

	// Parse the template
	tmpl, err := template.New("email").Parse(string(templateContent))
	if err != nil {
		return "", fmt.Errorf("error parsing email template: %w", err)
	}

	// Execute the template with the data
	var htmlBuffer strings.Builder
	err = tmpl.Execute(&htmlBuffer, data)
	if err != nil {
		return "", fmt.Errorf("error executing email template: %w", err)
	}

	return htmlBuffer.String(), nil
    }

    // Fallback HTML for when HTML template cannot be loaded
    func (s *SESClient) formatFallbackHTMLEmail(data EmailData) string {
	return fmt.Sprintf(`
            <!DOCTYPE html>
            <html>
            <head>
                <meta charset="UTF-8">
                <style>
                    body { font-family: Arial, sans-serif; }
                    .container { max-width: 600px; margin: 0 auto; padding: 20px; }
                    .header { background-color: #f5f5f5; padding: 10px; }
                    .content { padding: 20px 0; }
                    .footer { font-size: 12px; color: #777; }
                </style>
            </head>
            <body>
                <div class="container">
                    <div class="header">
                        <h2>New Contact Form Message</h2>
                    </div>
                    <div class="content">
                        <p><strong>Name:</strong> %s</p>
                        <p><strong>Email:</strong> %s</p>
                        <p><strong>Subject:</strong> %s</p>
                        <h3>Message:</h3>
                        <p>%s</p>
                    </div>
                    <div class="footer">
                        <p>Timestamp: %s</p>
                    </div>
                </div>
            </body>
            </html>
        `, data.Name, data.Email, data.Subject, data.Message, data.Timestamp)
    }   
