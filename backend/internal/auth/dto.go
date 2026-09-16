package auth

type RegisterRequest struct {
	BusinessName string `json:"business_name"`
	OwnerName    string `json:"owner_name"`
	Phone        string `json:"phone"`
	Email        string `json:"email"`
	Password     string `json:"password"`
}

type LoginRequest struct {
	Phone    string `json:"phone"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Token    string         `json:"token"`
	User     PublicUser     `json:"user"`
	Business PublicBusiness `json:"business"`
}

type PublicUser struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Role       string `json:"role"`
	BusinessID string `json:"business_id"`
}

type PublicBusiness struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// MeResponse surfaces the current user plus the owning business.
type MeResponse struct {
	User     PublicUser     `json:"user"`
	Business PublicBusiness `json:"business"`
}