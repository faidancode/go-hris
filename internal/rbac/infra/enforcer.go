package infra

import "github.com/casbin/casbin/v2"

func NewEnforcer(modelPath string) (*casbin.Enforcer, error) {
	return casbin.NewEnforcer(modelPath)
}
