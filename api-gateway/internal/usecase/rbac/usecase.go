package rbac

import (
	"context"
	"strings"

	"github.com/DoMinhHHung/Rental/internal/domain/entity"
	"github.com/DoMinhHHung/Rental/internal/infrastructure/config"
)

type UseCase struct {
	rules []entity.Permission
}

func New(ruleItems []config.RBACRuleItem) *UseCase {
	rules := make([]entity.Permission, 0, len(ruleItems))
	for _, item := range ruleItems {
		rules = append(rules, entity.Permission{
			Service: item.Service,
			Method:  item.Method,
			Path:    item.Path,
			Roles:   item.Roles,
		})
	}
	return &UseCase{rules: rules}
}

func (u *UseCase) Check(_ context.Context, perm entity.Permission, roles []string) bool {
	for _, rule := range u.rules {
		if rule.Service != perm.Service {
			continue
		}
		if rule.Method != "*" && rule.Method != perm.Method {
			continue
		}
		if !pathMatch(rule.Path, perm.Path) {
			continue
		}
		for _, allowed := range rule.Roles {
			if allowed == "*" {
				return true
			}
			for _, r := range roles {
				if r == allowed {
					return true
				}
			}
		}
	}
	return false
}

func pathMatch(pattern, path string) bool {
	if pattern == "*" {
		return true
	}
	if strings.HasSuffix(pattern, "/*") {
		prefix := strings.TrimSuffix(pattern, "/*")
		return strings.HasPrefix(path, prefix)
	}
	return pattern == path
}

// func defaultRules() []entity.Permission {
// 	return []entity.Permission{
// 		{Service: "user", Method: "POST", Path: "/api/v1/auth/*", Roles: []string{"*"}},
// 		{Service: "user", Method: "*", Path: "/api/v1/users/*", Roles: []string{"admin", "user"}},
// 		{Service: "property", Method: "GET", Path: "/api/v1/properties/*", Roles: []string{"*"}},
// 		{Service: "property", Method: "POST", Path: "/api/v1/properties/*", Roles: []string{"landlord", "admin"}},
// 		{Service: "property", Method: "PUT", Path: "/api/v1/properties/*", Roles: []string{"landlord", "admin"}},
// 		{Service: "property", Method: "DELETE", Path: "/api/v1/properties/*", Roles: []string{"admin"}},
// 		{Service: "booking", Method: "*", Path: "/api/v1/bookings/*", Roles: []string{"user", "landlord", "admin"}},
// 		{Service: "payment", Method: "*", Path: "/api/v1/payments/*", Roles: []string{"user", "landlord", "admin"}},
// 		{Service: "notification", Method: "*", Path: "/api/v1/notifications/*", Roles: []string{"user", "landlord", "admin"}},
// 	}
// }
