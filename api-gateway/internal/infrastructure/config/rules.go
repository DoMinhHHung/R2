package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type RuleConfig struct {
	Routes    []RouteRule    `yaml:"routes"`
	RBACRules []RBACRuleItem `yaml:"rbac_rules"`
}

type RouteRule struct {
	Prefix      string   `yaml:"prefix"`
	Service     string   `yaml:"service"`
	Methods     []string `yaml:"methods"`
	RequireAuth bool     `yaml:"require_auth"`
	Version     string   `yaml:"version"`
}

type RBACRuleItem struct {
	Service string   `yaml:"service"`
	Method  string   `yaml:"method"`
	Path    string   `yaml:"path"`
	Roles   []string `yaml:"roles"`
}

func LoadRules() (*RuleConfig, error) {
	path := os.Getenv("GATEWAY_RULES_PATH")
	if path == "" {
		path = "configs/gateway-rules.yaml"
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read rules file %q: %w", path, err)
	}

	var cfg RuleConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse rules file: %w", err)
	}

	if len(cfg.Routes) == 0 {
		return nil, fmt.Errorf("rules file has no routes defined")
	}

	return &cfg, nil
}
