package backend

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.hasen.dev/vbeam"
	"go.hasen.dev/vbolt"
	"golang.org/x/crypto/bcrypt"
)

var jwtKey []byte
var ErrLoginFailure = errors.New("LoginFailure")
var ErrAuthFailure = errors.New("AuthFailure")

type Claims struct {
	UserId       int    `json:"userId"`                 // User ID (-1 for anonymous code sessions, >0 for logged-in users)
	SessionToken string `json:"sessionToken,omitempty"` // UUID of CodeSession (only for userId=-1)
	jwt.RegisteredClaims
}

var appDb *vbolt.DB

func SetupAuth(app *vbeam.Application) {
	// Get JWT secret from environment, generate one if not set
	jwtSecret := os.Getenv("JWT_SECRET_KEY")
	if jwtSecret == "" {
		// Check if this is production (could be detected by other env vars)
		if os.Getenv("ENVIRONMENT") == "production" || os.Getenv("PROD") == "true" {
			log.Fatal("JWT_SECRET_KEY must be set in production environment")
		}

		token, err := generateToken(32)
		if err != nil {
			log.Fatal("error generating JWT secret")
		}
		jwtSecret = token
		log.Println("Generated JWT secret. Set JWT_SECRET_KEY environment variable for production.")
	}

	// Validate JWT secret strength
	if len(jwtSecret) < 16 {
		log.Fatal("JWT secret must be at least 16 characters long")
	}

	jwtKey = []byte(jwtSecret)

	// Register essential auth API endpoints
	app.HandleFunc("/api/login", loginHandler)
	app.HandleFunc("/api/logout", logoutHandler)
	app.HandleFunc("/api/refresh", refreshTokenHandler)

	// Register Google OAuth endpoints
	app.HandleFunc("/api/login/google", googleLoginHandler)
	app.HandleFunc("/api/google/callback", googleCallbackHandler)

	// Setup Google OAuth configuration
	err := SetupGoogleOAuth()
	if err != nil {
		log.Printf("Google OAuth setup failed: %v", err)
		log.Println("Google login will not be available. Set GOOGLE_CLIENT_ID and GOOGLE_CLIENT_SECRET to enable.")
	}

	appDb = app.DB
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		vbeam.RespondError(w, errors.New("login call must be POST"))
		return
	}

	var credentials LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
		vbeam.RespondError(w, ErrLoginFailure)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	var user User
	var passHash []byte

	vbolt.WithReadTx(appDb, func(tx *vbolt.Tx) {
		userId := GetUserId(tx, credentials.Email)
		if userId == 0 {
			return
		}
		user = GetUser(tx, userId)
		passHash = GetPassHash(tx, userId)
	})

	if user.Id == 0 {
		LogWarnWithRequest(r, LogCategoryAuth, "Login attempt with unknown email", map[string]interface{}{
			"email": credentials.Email,
		})
		json.NewEncoder(w).Encode(LoginResponse{Success: false, Error: "Invalid credentials"})
		return
	}

	err := bcrypt.CompareHashAndPassword(passHash, []byte(credentials.Password))
	if err != nil {
		LogWarnWithRequest(r, LogCategoryAuth, "Login attempt with invalid password", map[string]interface{}{
			"userId": user.Id,
			"email":  user.Email,
		})
		json.NewEncoder(w).Encode(LoginResponse{Success: false, Error: "Invalid credentials"})
		return
	}

	// Migrate anonymous code session if user had one before logging in
	migratedSession, _ := migrateAnonymousCodeSession(r, user.Id)
	if migratedSession != "" {
		LogInfoWithRequest(r, LogCategoryAuth, "Migrated code session during login", map[string]interface{}{
			"userId":       user.Id,
			"sessionToken": migratedSession,
		})
	}

	token, err := generateAuthJwt(user, w)
	if err != nil {
		LogErrorWithRequest(r, LogCategoryAuth, "Failed to generate JWT token", map[string]interface{}{
			"userId": user.Id,
			"error":  err.Error(),
		})
		json.NewEncoder(w).Encode(LoginResponse{Success: false, Error: "Failed to generate token"})
		return
	}

	// Log successful login
	LogInfoWithRequest(r, LogCategoryAuth, "User login successful", map[string]interface{}{
		"userId": user.Id,
		"email":  user.Email,
	})

	var resp AuthResponse
	vbolt.WithReadTx(appDb, func(tx *vbolt.Tx) {
		resp = GetAuthResponseFromUser(tx, user)
	})
	json.NewEncoder(w).Encode(LoginResponse{Success: true, Token: token, Auth: resp})
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	// Try to get user info before clearing the cookie
	user, _ := AuthenticateRequest(r)

	// Delete refresh token from database if present
	if cookie, err := r.Cookie("refreshToken"); err == nil && cookie.Value != "" {
		vbolt.WithWriteTx(appDb, func(tx *vbolt.Tx) {
			DeleteRefreshToken(tx, cookie.Value)
			vbolt.TxCommit(tx)
		})
	}

	// Clear UserCodeSessionsBkt entry if user had code session access
	if user.Id > 0 {
		vbolt.WithWriteTx(appDb, func(tx *vbolt.Tx) {
			// Delete user's code session mapping
			vbolt.Delete(tx, UserCodeSessionsBkt, user.Id)
			vbolt.TxCommit(tx)
		})
	}

	// Clear auth token cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "authToken",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Expires:  time.Unix(0, 0),
	})

	// Clear refresh token cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "refreshToken",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Expires:  time.Unix(0, 0),
	})

	// Log logout event
	if user.Id != 0 {
		LogInfoWithRequest(r, LogCategoryAuth, "User logout", map[string]interface{}{
			"userId": user.Id,
			"email":  user.Email,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

func generateToken(n int) (string, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func generateAuthJwt(user User, w http.ResponseWriter) (tokenString string, err error) {
	expirationTime := time.Now().Add(24 * time.Hour) // 24 hour expiry
	claims := &Claims{
		UserId: user.Id,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err = token.SignedString(jwtKey)
	if err != nil {
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "authToken",
		Value:    tokenString,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   60 * 60 * 24, // 24 hours
	})

	// Create and set refresh token (30 days)
	var refreshToken RefreshToken
	vbolt.WithWriteTx(appDb, func(tx *vbolt.Tx) {
		refreshToken, err = CreateRefreshToken(tx, user.Id, 30*24*time.Hour)
		if err != nil {
			return
		}

		// Update last login
		user.LastLogin = time.Now()
		vbolt.Write(tx, UsersBkt, user.Id, &user)
		vbolt.TxCommit(tx)
	})

	if err != nil {
		return
	}

	// Set refresh token cookie (30 days)
	http.SetCookie(w, &http.Cookie{
		Name:     "refreshToken",
		Value:    refreshToken.Token,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   60 * 60 * 24 * 30, // 30 days
	})

	return
}

// migrateAnonymousCodeSession checks if the request has an anonymous code session (userId=-1)
// and migrates it to the newly authenticated user by storing the session in UserCodeSessionsBkt.
// Returns the sessionToken if migration occurred, or empty string if no migration needed.
func migrateAnonymousCodeSession(r *http.Request, newUserId int) (sessionToken string, err error) {
	// Try to get existing authToken cookie
	authCookie, cookieErr := r.Cookie("authToken")
	if cookieErr != nil || authCookie.Value == "" {
		// No existing JWT, nothing to migrate
		return "", nil
	}

	// Parse the existing JWT
	token, parseErr := jwt.ParseWithClaims(authCookie.Value, &Claims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return jwtKey, nil
	})

	// If parsing failed or token is invalid, nothing to migrate
	if parseErr != nil || !token.Valid {
		return "", nil
	}

	// Check if this is an anonymous code session (userId=-1)
	claims, ok := token.Claims.(*Claims)
	if !ok || claims.UserId != -1 || claims.SessionToken == "" {
		// Not an anonymous code session, nothing to migrate
		return "", nil
	}

	// Migrate: store session token in UserCodeSessionsBkt
	vbolt.WithWriteTx(appDb, func(tx *vbolt.Tx) {
		vbolt.Write(tx, UserCodeSessionsBkt, newUserId, &claims.SessionToken)
		vbolt.TxCommit(tx)
	})

	LogInfo(LogCategoryAuth, "Migrated anonymous code session to user account", map[string]interface{}{
		"newUserId":    newUserId,
		"sessionToken": claims.SessionToken,
	})

	return claims.SessionToken, nil
}

func GetAuthUser(ctx *vbeam.Context) (user User, err error) {
	if len(ctx.Token) == 0 {
		return user, ErrAuthFailure
	}
	token, err := jwt.ParseWithClaims(ctx.Token, &Claims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return jwtKey, nil
	})
	if err != nil || !token.Valid {
		LogWarn(LogCategoryAuth, "JWT parsing failed in GetAuthUser", map[string]interface{}{
			"error": err,
			"valid": token != nil && token.Valid,
		})
		return
	}

	if claims, ok := token.Claims.(*Claims); ok {
		// Check if this is anonymous code session (userId=-1)
		if claims.UserId == -1 {
			// Load code session from database
			var session CodeSession
			vbolt.Read(ctx.Tx, CodeSessionsBkt, claims.SessionToken, &session)

			if session.Token == "" {
				LogWarn(LogCategoryAuth, "Code session not found in database", map[string]interface{}{
					"sessionToken": claims.SessionToken,
				})
				return user, errors.New("code session not found")
			}

			// Validate the session (check code not revoked/expired/grace period)
			var accessCode AccessCode
			vbolt.Read(ctx.Tx, AccessCodesBkt, session.Code, &accessCode)

			if accessCode.Code == "" {
				LogWarn(LogCategoryAuth, "Access code not found in database", map[string]interface{}{
					"code": session.Code,
				})
				return user, errors.New("access code not found")
			}

			if accessCode.IsRevoked {
				LogWarn(LogCategoryAuth, "Access code is revoked", map[string]interface{}{
					"code": accessCode.Code,
				})
				return user, errors.New("access code revoked")
			}

			// Check expiration with grace period
			now := time.Now()
			if !session.GracePeriodUntil.IsZero() {
				// In grace period - check if grace has expired
				if now.After(session.GracePeriodUntil) {
					return user, errors.New("access code grace period expired")
				}
			} else if now.After(accessCode.ExpiresAt) {
				// Code expired - would need grace period but can't write in read-only tx
				// The session cleanup job will handle setting grace period on write operations
				return user, errors.New("access code expired")
			}

			// Return pseudo-user for code sessions
			user = User{
				Id:    -1, // Special ID for anonymous code sessions
				Email: "anonymous@code-session",
				Name:  "Access Code User",
			}
			return user, nil
		}

		// Regular user JWT (userId > 0)
		user = GetUser(ctx.Tx, claims.UserId)
	}
	return
}

// AuthContext represents the authentication context for a request
// It can be either a regular user (via JWT) or a code session (via access code)
type AuthContext struct {
	User            User         // Actual user (if JWT auth), or pseudo-user with Id=-1 for code sessions
	IsCodeAuth      bool         // True if authenticated via access code
	CodeSession     *CodeSession // Session info if anonymous code auth (userId=-1)
	UserCodeSession *CodeSession // Session info if logged-in user with code access (userId>0)
	AccessCode      *AccessCode  // Access code info if code auth
}

// GetAuthFromRequest checks JWT authentication (supports both user and code sessions)
// Returns AuthContext with user or code session information
// This is used in HTTP handlers that have access to the request object
func GetAuthFromRequest(r *http.Request, db *vbolt.DB) (authCtx AuthContext, err error) {
	// Read authToken cookie (works for both regular users and code sessions)
	authCookie, authErr := r.Cookie("authToken")
	if authErr != nil || authCookie.Value == "" {
		return authCtx, ErrAuthFailure
	}

	// Parse JWT token
	token, tokenErr := jwt.ParseWithClaims(authCookie.Value, &Claims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return jwtKey, nil
	})

	if tokenErr != nil || !token.Valid {
		return authCtx, ErrAuthFailure
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return authCtx, ErrAuthFailure
	}

	// Check if this is anonymous code session (userId=-1)
	if claims.UserId == -1 {
		// Load code session from database using sessionToken from JWT claims
		var session CodeSession
		var accessCode AccessCode

		vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
			vbolt.Read(tx, CodeSessionsBkt, claims.SessionToken, &session)

			if session.Token == "" {
				return
			}

			vbolt.Read(tx, AccessCodesBkt, session.Code, &accessCode)

			if accessCode.Code == "" || accessCode.IsRevoked {
				return
			}

			// Update LastSeen timestamp
			session.LastSeen = time.Now()
			vbolt.Write(tx, CodeSessionsBkt, session.Token, &session)

			vbolt.TxCommit(tx)
		})

		if session.Token == "" || accessCode.Code == "" {
			return authCtx, ErrAuthFailure
		}

		// Create pseudo-user for anonymous code sessions
		authCtx.User = User{
			Id:    -1, // Special ID for anonymous code sessions
			Email: "anonymous@code-session",
			Name:  "Access Code User",
		}
		authCtx.IsCodeAuth = true
		authCtx.CodeSession = &session
		authCtx.AccessCode = &accessCode
		return authCtx, nil
	}

	// Logged-in user JWT (userId > 0)
	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		// Load user
		authCtx.User = GetUser(tx, claims.UserId)

		if authCtx.User.Id == 0 {
			return
		}

		// Check if user has additional code session access
		var sessionToken string
		vbolt.Read(tx, UserCodeSessionsBkt, authCtx.User.Id, &sessionToken)

		if sessionToken != "" {
			// User has code-based access in addition to normal auth
			var session CodeSession
			vbolt.Read(tx, CodeSessionsBkt, sessionToken, &session)

			if session.Token != "" {
				// Load access code for the session
				var accessCode AccessCode
				vbolt.Read(tx, AccessCodesBkt, session.Code, &accessCode)

				if accessCode.Code != "" && !accessCode.IsRevoked {
					// Update LastSeen timestamp
					session.LastSeen = time.Now()
					vbolt.Write(tx, CodeSessionsBkt, session.Token, &session)

					authCtx.UserCodeSession = &session
					authCtx.AccessCode = &accessCode
				}
			}
		}

		vbolt.TxCommit(tx)
	})

	if authCtx.User.Id == 0 {
		return authCtx, ErrAuthFailure
	}

	// Set IsCodeAuth based on whether user has code session access
	authCtx.IsCodeAuth = (authCtx.UserCodeSession != nil)
	return authCtx, nil
}

func refreshTokenHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		vbeam.RespondError(w, errors.New("refresh call must be POST"))
		return
	}

	// Get refresh token from cookie
	cookie, err := r.Cookie("refreshToken")
	if err != nil || cookie.Value == "" {
		LogWarnWithRequest(r, LogCategoryAuth, "Refresh attempt without token", nil)
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "No refresh token provided",
		})
		return
	}

	var user User
	var validToken RefreshToken

	// Validate refresh token and get user
	vbolt.WithReadTx(appDb, func(tx *vbolt.Tx) {
		var valid bool
		validToken, valid = ValidateRefreshToken(tx, cookie.Value)
		if !valid {
			return
		}

		user = GetUser(tx, validToken.UserId)
	})

	if user.Id == 0 {
		LogWarnWithRequest(r, LogCategoryAuth, "Refresh attempt with invalid token", nil)
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Invalid or expired refresh token",
		})
		return
	}

	// Update last used timestamp
	vbolt.WithWriteTx(appDb, func(tx *vbolt.Tx) {
		UpdateRefreshTokenLastUsed(tx, validToken.Id)
		vbolt.TxCommit(tx)
	})

	// Generate new JWT
	token, err := generateAuthJwt(user, w)
	if err != nil {
		LogErrorWithRequest(r, LogCategoryAuth, "Failed to generate JWT during refresh", map[string]interface{}{
			"userId": user.Id,
			"error":  err.Error(),
		})
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Failed to generate token",
		})
		return
	}

	// Log successful refresh
	LogInfoWithRequest(r, LogCategoryAuth, "Token refresh successful", map[string]interface{}{
		"userId": user.Id,
		"email":  user.Email,
	})

	w.Header().Set("Content-Type", "application/json")
	var resp AuthResponse
	vbolt.WithReadTx(appDb, func(tx *vbolt.Tx) {
		resp = GetAuthResponseFromUser(tx, user)
	})
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"token":   token,
		"auth":    resp,
	})
}
