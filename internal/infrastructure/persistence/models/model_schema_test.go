package models

import (
	"sync"
	"testing"

	"gorm.io/gorm/schema"
)

func TestGormModelsParse(t *testing.T) {
	for _, model := range All() {
		model := model
		t.Run(schemaName(t, model), func(t *testing.T) {
			if _, err := schema.Parse(model, &sync.Map{}, schema.NamingStrategy{}); err != nil {
				t.Fatalf("parse GORM model: %v", err)
			}
		})
	}
}

func schemaName(t *testing.T, model any) string {
	t.Helper()

	parsed, err := schema.Parse(model, &sync.Map{}, schema.NamingStrategy{})
	if err != nil {
		return "unparseable"
	}
	return parsed.Name
}
