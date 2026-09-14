package policy

type PasswordPolicy struct {
	MinLength          int
	MaxLength          int
	RequireUppercase   bool
	RequireLowercase   bool
	RequireNumbers     bool
	RequireSpecialChar bool
}

func DefaultPasswordPolicy() PasswordPolicy {
	return PasswordPolicy{
		MinLength:          8,
		MaxLength:          128,
		RequireUppercase:   true,
		RequireLowercase:   true,
		RequireNumbers:     true,
		RequireSpecialChar: true,
	}
}
