package services

import (
	"context"
	"fmt"
	"os"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"google.golang.org/api/option"
)

// FirebaseService handles interactions with Firebase Authentication
type FirebaseService interface {
	VerifyIDToken(ctx context.Context, idToken string) (*auth.Token, error)
}

type firebaseService struct {
	authClient *auth.Client
}

// NewFirebaseService initializes the Firebase app and returns the service
func NewFirebaseService() (FirebaseService, error) {
	ctx := context.Background()
	
	// Si GOOGLE_APPLICATION_CREDENTIALS no está seteado, Firebase intentará usar Application Default Credentials
	var app *firebase.App
	var err error
	
	credentialsEnv := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	credentialsJSON := os.Getenv("FIREBASE_CREDENTIALS_JSON")
	
	if credentialsJSON != "" {
		opt := option.WithCredentialsJSON([]byte(credentialsJSON))
		app, err = firebase.NewApp(ctx, nil, opt)
	} else if credentialsEnv != "" {
		// Check if it's a raw JSON string instead of a path
		if len(credentialsEnv) > 0 && credentialsEnv[0] == '{' {
			// Borramos la variable de entorno para que el SDK base no intente abrirla como archivo (file name too long)
			os.Unsetenv("GOOGLE_APPLICATION_CREDENTIALS")
			opt := option.WithCredentialsJSON([]byte(credentialsEnv))
			app, err = firebase.NewApp(ctx, nil, opt)
		} else {
			opt := option.WithCredentialsFile(credentialsEnv)
			app, err = firebase.NewApp(ctx, nil, opt)
		}
	} else {
		app, err = firebase.NewApp(ctx, nil)
	}

	if err != nil {
		return nil, fmt.Errorf("error initializing firebase app: %v", err)
	}

	authClient, err := app.Auth(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting firebase auth client: %v", err)
	}

	return &firebaseService{
		authClient: authClient,
	}, nil
}

// VerifyIDToken verifies a Firebase ID token and returns the token payload.
func (s *firebaseService) VerifyIDToken(ctx context.Context, idToken string) (*auth.Token, error) {
	token, err := s.authClient.VerifyIDToken(ctx, idToken)
	if err != nil {
		return nil, fmt.Errorf("invalid firebase token: %v", err)
	}
	return token, nil
}
