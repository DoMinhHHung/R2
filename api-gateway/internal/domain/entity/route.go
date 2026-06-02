package entity

type Route struct {
	Prefix      string
	ServiceURL  string
	Methods     []string
	RequireAuth bool
	Roles       []string
	Version     string
}

type Claims struct {
	UserID string
	Roles  []string
	APIKey string
}
