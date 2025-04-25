package handlers

import (
	"fmt"
	"time"
)

// generateEmailTemplate generates the HTML email template with the provided link.
func generateEmailTemplate(link string) string {
	year := time.Now().Year()
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Your Magic Login Link</title>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { text-align: center; margin-bottom: 30px; }
        .logo { max-width: 150px; margin-bottom: 20px; }
        .button { 
            display: inline-block; 
            padding: 12px 24px; 
            background-color: #4F46E5; 
            color: white; 
            text-decoration: none; 
            border-radius: 4px; 
            font-weight: bold; 
            margin: 20px 0;
            transition: background-color 0.3s ease;
        }
        .button:hover {
            background-color: #3730a3;
            color: white;
            text-decoration: none;
        }
        .footer { 
            margin-top: 30px; 
            font-size: 12px; 
            color: #666; 
            text-align: center;
        }
    </style>
</head>
<body>
    <div class="header">
        <h1>Welcome to Planning Poker!</h1>
    </div>
    
    <p>Hello,</p>
    
    <p>We received a request to sign in to Planning Poker using this email address. Click the button below to sign in.</p>
    
    <div style="text-align: center;">
        <a href="%s" class="button">Sign In Now</a>
    </div>
    
    <p>If you didn't request this link, you can safely ignore this email.</p>
    
    <div class="footer">
        <p>This link will expire in 24 hours and can only be used once.</p>
        <p>© %d Planning Poker. All rights reserved.</p>
    </div>
</body>
</html>
`, link, year)
}

