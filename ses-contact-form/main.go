package main

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "strings"
    "time"
    
    // Local Packages
    
    "ses-contact-form/config"
    "ses-contact-form/emailer"
    "ses-contact-form/validator"

    // AWS Packages
    "github.com/aws/aws-lambda-go/events"
    "github.com/aws/aws-lambda-go/lambda"
    // "github.com/aws/aws-sdk-go-v2/aws"
    awsConfig "github.com/aws/aws-sdk-go-v2/config"
    "github.com/aws/aws-sdk-go-v2/service/ses"
)

type FormData struct {
    Name    string `json:"name"`
    Email   string `json:"email"`
    Subject string `json:"subject"`
    Message string `json:"message"`

}

type ResponseBody struct {
    Message string `json:"message,omitempty"`
    Error   string `json:"error,omitempty"`
    
}

// Handles Lambda Function invocation
func HandleRequest(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
    // Load local config
    cfg := config.LoadConfig()

    // Init res headers
    headers := map[string]string{
        "Content-Type": "application/json",
    }

    // Parse req body
    var data FormData
    
    if err := json.Unmarshal([]byte(request.Body), &data); err != nil {
        log.Printf("Error parsing request body: %v", err)
        return createErrorResponse(400, "Invalid request format", err, headers), nil
    }

    // Log incoming email from form data
    log.Printf("Received contact form submission from: %s", data.Email)
    
    originalData := data

    // Validate input
    
    if err := validateFormData(data); err != nil {
        log.Printf("Validation error: %v", err)

        // Send alert email for validation errors
        go sendAlertEmail(ctx, cfg, err, originalData)

        return createErrorResponse(400, err.Error(), nil, headers), nil
    }

    // Sanitize input
        
    data = sanitizeFormData(data)

    // Init SES Client
    sesClient, err := initSESClient(ctx, cfg.Region)
    if err != nil {
        log.Printf("Error initializing SES client: %v", err)
        return createErrorResponse(500, "Internal Server Error", err, headers), nil
    }

    // Create emailer client
    emailClient := emailer.NewSESClient(sesClient)
    
    // Prep email data
    emailData := emailer.EmailData {
        Name:       data.Name,
        Email:      data.Email,
        Subject:    data.Subject,
        Message:    data.Message,
        Timestamp:  time.Now().Format(time.RFC3339), // Year - Month - Day - Hour - Min - Sec
    }

    // Send mail
    if err := emailClient.SendContactEmail(ctx, cfg.ToEmail, cfg.FromEmail, emailData); err != nil {
        log.Printf("Error sending email: %v", err)

        // Send alert email for send err
        go sendAlertEmail(ctx, cfg, err, originalData)

        return createErrorResponse(500, "Failed to send email", err, headers), nil
    }

    // Return success
    responseBody, _ := json.Marshal(ResponseBody{
        Message: "Email sent successfully!",
    })

    return events.APIGatewayProxyResponse {
        StatusCode: 200,
        Headers:    headers,
        Body:       string(responseBody),
    }, nil

}

func validateFormData(data FormData) error {
    // Basic - Check for required fields
    if err := validator.ValidateRequired("name", data.Name); err != nil {
        return err
    }

    if err := validator.ValidateRequired("email", data.Name); err != nil {
        return err
    }


    if err := validator.ValidateEmail(data.Email); err != nil {
        return err
    }

    return nil
}

func sanitizeFormData(data FormData) FormData {
    return FormData {
        Name:       validator.SanitizeInput(data.Name),
        Email:      validator.SanitizeInput(data.Email),
        Subject:    validator.SanitizeInput(data.Subject),
        Message:    validator.SanitizeInput(data.Message),
    }
}

func initSESClient(ctx context.Context, region string) (*ses.Client, error) {
    awsCfg, err := awsConfig.LoadDefaultConfig(ctx, awsConfig.WithRegion(region))
    if err != nil {
        return nil, fmt.Errorf("Error loading AWS config: %w", err)
    }

    return ses.NewFromConfig(awsCfg), nil
}

func sendAlertEmail(ctx context.Context, cfg config.Config, err error, data interface{}) {
    sesClient, initErr := initSESClient(ctx, cfg.Region)
    if initErr != nil {
        log.Printf("Error initializing SES client for alert: %v", initErr)
        return
    }

    emailClient := emailer.NewSESClient(sesClient)
    if alertErr := emailClient.SendAlertEmail(ctx, cfg.FromEmail, cfg.AlertEmails, err, data); alertErr != nil {
        log.Printf("Error sending alert email: %v", alertErr)
    }
}

func createErrorResponse(statusCode int, message string, err error, headers map[string]string) events.APIGatewayProxyResponse {
    errMsg := message
    if err != nil && strings.Contains(message, "%") {
        errMsg = fmt.Sprintf(message,err.Error())
    }

    responseBody, _ := json.Marshal(ResponseBody {
        Error: errMsg,
    })

    return events.APIGatewayProxyResponse {
        StatusCode: statusCode,
        Headers:    headers,
        Body:       string(responseBody),
    }
}

func defaultString(val, defaultVal string) string {
    if val == "" {
        return defaultVal
    }
    return val
}

func main() {
    lambda.Start(HandleRequest)   
}
