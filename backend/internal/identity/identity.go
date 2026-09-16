package identity

import "context"

type contextKey string

const ctxUserKey contextKey = "identity.user"

// Principal carries the authenticated request identity.
type Principal struct {
	UserID     string
	BusinessID string
	Role       string
}

func (p *Principal) IsOwner() bool      { return p.Role == "OWNER" }
func (p *Principal) IsAdmin() bool      { return p.Role == "ADMIN" }
func (p *Principal) IsTechnician() bool { return p.Role == "TECHNICIAN" }
func (p *Principal) CanManageBusiness() bool {
	return p.IsOwner()
}
func (p *Principal) CanManageTechnicians() bool { return p.IsOwner() || p.IsAdmin() }
func (p *Principal) CanManageCustomers() bool   { return p.IsOwner() || p.IsAdmin() }
func (p *Principal) CanCreateJobs() bool        { return p.IsOwner() || p.IsAdmin() }

func WithPrincipal(ctx context.Context, p *Principal) context.Context {
	return context.WithValue(ctx, ctxUserKey, p)
}

func PrincipalFrom(ctx context.Context) (*Principal, bool) {
	p, ok := ctx.Value(ctxUserKey).(*Principal)
	return p, ok
}

// MustPrincipal returns the principal; used only behind RequireAuth.
func MustPrincipal(ctx context.Context) *Principal {
	p, ok := PrincipalFrom(ctx)
	if !ok {
		panic("identity: principal missing from context")
	}
	return p
}