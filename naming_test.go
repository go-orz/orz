package orz

import (
	"testing"

	"gorm.io/gorm"
)

func TestCamelToSnakeHandlesCommonInitialisms(t *testing.T) {
	cases := map[string]string{
		"UserID":           "user_id",
		"HTTPServerConfig": "http_server_config",
		"products.UserID":  "products.user_id",
	}

	for input, want := range cases {
		if got := CamelToSnake(input); got != want {
			t.Fatalf("CamelToSnake(%q) = %q, want %q", input, got, want)
		}
	}
}

type namingCategory struct {
	ID uint `gorm:"primaryKey"`
}

func TestRepositoryUsesGormTableNameStrategy(t *testing.T) {
	repo := NewRepository[namingCategory, uint](&gorm.DB{})
	if got := repo.GetTableName(); got != "naming_categories" {
		t.Fatalf("expected table name naming_categories, got %q", got)
	}
}

func TestNewRepositoryFromAppSupportsAppGetDatabase(t *testing.T) {
	app := NewApp()
	app.SetDatabase(&gorm.DB{})

	repo, err := NewRepositoryFromApp[namingCategory, uint](app)
	if err != nil {
		t.Fatalf("NewRepositoryFromApp returned error: %v", err)
	}
	if repo == nil {
		t.Fatal("expected repository")
	}
}
