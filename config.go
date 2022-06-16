package qsgpm

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/quicksight/types"
	gv "github.com/hashicorp/go-version"
	gc "github.com/kayac/go-config"
)

type Config struct {
	RequiredVersion string `yaml:"required_version"`

	CreateOnly       bool          `yaml:"create_only"`
	User             *UserConfig   `yaml:"user"`
	Groups           []string      `yaml:"groups"`
	CustomPermission string        `yaml:"custom_permission"`
	Rules            []*RuleConfig `yaml:"rules"`

	versionConstraints gv.Constraints
}

func (cfg *Config) Load(path string) error {
	if path == "" {
		return errors.New("no config")
	}
	if err := gc.LoadWithEnv(cfg, path); err != nil {
		return err
	}
	return cfg.Restrict()
}

func (cfg *Config) Restrict() error {
	if cfg.RequiredVersion != "" {
		constraints, err := gv.NewConstraint(cfg.RequiredVersion)
		if err != nil {
			return fmt.Errorf("required_version has invalid format: %w", err)
		}
		cfg.versionConstraints = constraints
	}
	for i, rule := range cfg.Rules {
		rule.User = cfg.User.Merge(rule.User)
		rule.Groups = append(rule.Groups, cfg.Groups...)
		rule.CustomPermission = coalesceString(rule.CustomPermission, cfg.CustomPermission)
		if err := rule.Restrict(); err != nil {
			return fmt.Errorf("rules[%d]: %w", i, err)
		}
	}
	return nil
}

func (cfg *Config) GetCustomPermissionName(user *User) *string {
	for _, rule := range cfg.Rules {
		if name, ok := rule.GetCustomPermissionName(user); ok {
			return &name
		}
	}
	return nil
}

func (cfg *Config) GetGroupNames(user *User) ([]string, bool) {
	for _, rule := range cfg.Rules {
		if groups, ok := rule.GetGroupNames(user); ok {
			return groups, true
		}
	}
	return nil, false
}

func (cfg *Config) GetNamespaces() []string {
	m := make(map[string]struct{}, 1+len(cfg.Rules))
	m[strings.TrimSpace(cfg.User.Namespace)] = struct{}{}
	for _, rule := range cfg.Rules {
		m[strings.TrimSpace(rule.User.Namespace)] = struct{}{}
	}
	namespaces := make([]string, 0, len(m))
	for namespace := range m {
		if len(namespace) != 0 {
			namespaces = append(namespaces, namespace)
		}
	}
	return namespaces
}

// ValidateVersion validates a version satisfies required_version.
func (c *Config) ValidateVersion(version string) error {
	if c.versionConstraints == nil {
		log.Println("[warn] required_version is empty. Skip checking required_version.")
		return nil
	}
	versionParts := strings.SplitN(version, "-", 2)
	v, err := gv.NewVersion(versionParts[0])
	if err != nil {
		log.Printf("[warn] Invalid version format \"%s\". Skip checking required_version.", version)
		// invalid version string (e.g. "current") always allowed
		return nil
	}
	if !c.versionConstraints.Check(v) {
		return fmt.Errorf("version %s does not satisfy constraints required_version: %s", version, c.versionConstraints)
	}
	return nil
}

type RuleConfig struct {
	User             *UserConfig `yaml:"user"`
	Groups           []string    `yaml:"groups"`
	CustomPermission string      `yaml:"custom_permission"`
}

func (cfg *RuleConfig) Restrict() error {
	if err := cfg.User.Restrict(); err != nil {
		return fmt.Errorf("user: %w", err)
	}
	groups := make(map[string]struct{}, len(cfg.Groups))
	for _, group := range cfg.Groups {
		groups[group] = struct{}{}
	}
	cfg.Groups = make([]string, 0, len(groups))
	for group := range groups {
		cfg.Groups = append(cfg.Groups, group)
	}
	return nil
}

func (cfg *RuleConfig) GetCustomPermissionName(user *User) (string, bool) {
	if cfg.CustomPermission == "" {
		return "", false
	}
	if !cfg.User.Match(user) {
		return "", false
	}
	return cfg.CustomPermission, true
}

func (cfg *RuleConfig) GetGroupNames(user *User) ([]string, bool) {
	if len(cfg.Groups) == 0 {
		return nil, false
	}
	if !cfg.User.Match(user) {
		return nil, false
	}
	return cfg.Groups, true
}

type UserConfig struct {
	IdentityType      string `yaml:"identity_type"`
	SessionNameSuffix string `yaml:"session_name_suffix"`
	EmailSuffix       string `yaml:"email_suffix"`
	Namespace         string `yaml:"namespace"`
	IAMRoleName       string `yaml:"iam_role_name"`
	Role              string `yaml:"role"`

	identityType types.IdentityType
	role         types.UserRole
}

func coalesceString(strs ...string) string {
	for _, str := range strs {
		if str != "" {
			return str
		}
	}
	return ""
}

func (cfg *UserConfig) Clone() *UserConfig {
	cloned := *cfg
	return &cloned
}

func (cfg *UserConfig) Merge(other *UserConfig) *UserConfig {
	cloned := cfg.Clone()
	cloned.IdentityType = coalesceString(cfg.IdentityType, other.IdentityType)
	cloned.SessionNameSuffix = coalesceString(cfg.SessionNameSuffix, other.SessionNameSuffix)
	cloned.EmailSuffix = coalesceString(cfg.EmailSuffix, other.EmailSuffix)
	cloned.Namespace = coalesceString(cfg.Namespace, other.Namespace)
	cloned.IAMRoleName = coalesceString(cfg.IAMRoleName, other.IAMRoleName)
	cloned.Role = coalesceString(cfg.Role, other.Role)
	return cloned
}

func (cfg *UserConfig) Restrict() error {
	if cfg.IdentityType != "" {
		if err := cfg.validateIdentityType(); err != nil {
			return err
		}
	}
	if cfg.Role != "" {
		if err := cfg.validateRole(); err != nil {
			return err
		}
	}
	return nil
}

func (cfg *UserConfig) validateIdentityType() error {
	var t types.IdentityType
	identityTypes := t.Values()
	values := make([]string, 0, len(identityTypes))
	for _, identityType := range identityTypes {
		value := string(identityType)
		if strings.EqualFold(value, cfg.IdentityType) {
			cfg.identityType = identityType
			return nil
		}
		values = append(values, value)
	}
	return fmt.Errorf("given IdentityType: %s is not one of %s or %s", cfg.IdentityType, strings.Join(values[:len(values)-1], ", "), values[len(values)-1])
}

func (cfg *UserConfig) validateRole() error {
	var r types.UserRole
	roles := r.Values()
	values := make([]string, 0, len(roles))
	for _, role := range roles {
		value := string(role)
		if strings.EqualFold(value, cfg.Role) {
			cfg.role = role
			return nil
		}
		values = append(values, value)
	}
	return fmt.Errorf("given Role: %s is not one of %s or %s", cfg.Role, strings.Join(values[:len(values)-1], ", "), values[len(values)-1])
}

func (cfg *UserConfig) Match(user *User) bool {
	if cfg.identityType != "" {
		if user.IdentityType != cfg.identityType {
			return false
		}
	}
	sessionName := user.SessionName()
	if cfg.SessionNameSuffix != "" {
		if !strings.HasSuffix(sessionName, cfg.SessionNameSuffix) {
			return false
		}
	}
	var email string
	if user.Email != nil {
		email = *user.Email
	}
	if cfg.EmailSuffix != "" {
		if !strings.HasSuffix(email, cfg.EmailSuffix) {
			return false
		}
	}
	if cfg.Namespace != "" {
		if user.Namespace != cfg.Namespace {
			return false
		}
	}
	iamRoleName := user.IAMRoleName()
	if cfg.IAMRoleName != "" {
		if iamRoleName != cfg.IAMRoleName {
			return false
		}
	}
	if cfg.role != "" {
		if user.Role != cfg.role {
			return false
		}
	}
	return true
}

func NewDefaultConfig() *Config {
	return &Config{}
}
