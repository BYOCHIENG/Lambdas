package main

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ses"
)

// MockSESClient implements the EmailSender interface for testing
type MockSESClient struct{}

// SendEmail simulates sending an email without connecting to AWS
func (m *MockSESClient) SendEmail(ctx context.Context, params *ses.SendEmailInput, optFns ...func(*ses.Options)) (*ses.SendEmailOutput, error) {
	// Log what would have been sent
	log.Printf("MOCK: Would send email to: %v", params.Destination.ToAddresses)
	log.Printf("MOCK: From: %v", *params.Source)
	log.Printf("MOCK: Subject: %v", *params.Message.Subject.Data)
	
	// Return success
	return &ses.SendEmailOutput{
		MessageId: aws.String("mock-message-id-12345"),
	}, nil
}
