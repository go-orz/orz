package pagebuilderintegration

import (
	"context"
	"testing"

	"github.com/go-orz/orz"
)

type repositoryUpdateUser struct {
	ID   uint `gorm:"primaryKey"`
	Name string
}

func (repositoryUpdateUser) TableName() string {
	return "repository_update_users"
}

type repositoryNoID struct {
	Name string
}

func TestUpdateByIdUsesExplicitIDCondition(t *testing.T) {
	db := newTestSQLiteDB(t)
	if err := db.AutoMigrate(&repositoryUpdateUser{}); err != nil {
		t.Fatalf("AutoMigrate returned error: %v", err)
	}

	repo := orz.NewRepository[repositoryUpdateUser, uint](db)
	ctx := context.Background()
	if err := repo.Create(ctx, &repositoryUpdateUser{ID: 1, Name: "alice"}); err != nil {
		t.Fatalf("create alice returned error: %v", err)
	}
	if err := repo.Create(ctx, &repositoryUpdateUser{ID: 2, Name: "bob"}); err != nil {
		t.Fatalf("create bob returned error: %v", err)
	}

	if err := repo.UpdateById(ctx, &repositoryUpdateUser{ID: 1, Name: "alice-updated"}); err != nil {
		t.Fatalf("UpdateById returned error: %v", err)
	}

	first, err := repo.FindById(ctx, 1)
	if err != nil {
		t.Fatalf("FindById(1) returned error: %v", err)
	}
	second, err := repo.FindById(ctx, 2)
	if err != nil {
		t.Fatalf("FindById(2) returned error: %v", err)
	}
	if first.Name != "alice-updated" {
		t.Fatalf("expected first user to be updated, got %+v", first)
	}
	if second.Name != "bob" {
		t.Fatalf("expected second user to remain unchanged, got %+v", second)
	}
}

func TestUpdateByIdRequiresEntityIDField(t *testing.T) {
	db := newTestSQLiteDB(t)
	repo := orz.NewRepository[repositoryNoID, uint](db)

	err := repo.UpdateById(context.Background(), &repositoryNoID{Name: "invalid"})
	if err == nil {
		t.Fatal("expected UpdateById to fail without ID field")
	}
}
