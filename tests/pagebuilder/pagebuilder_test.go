package pagebuilderintegration

import (
	"context"
	"strings"
	"testing"

	"github.com/go-orz/orz"
	_ "github.com/go-orz/orz/drivers/sqlite"
	"gorm.io/gorm"
)

type pageBuilderUser struct {
	ID   uint `gorm:"primaryKey"`
	Name string
}

func (pageBuilderUser) TableName() string {
	return "page_builder_users"
}

type pageBuilderProfile struct {
	ID     uint `gorm:"primaryKey"`
	UserID uint `gorm:"uniqueIndex"`
	Bio    string
}

func (pageBuilderProfile) TableName() string {
	return "page_builder_profiles"
}

type pageBuilderUserView struct {
	ID   uint
	Name string
	Bio  string
}

type pageBuilderArticle struct {
	ID   uint `gorm:"primaryKey"`
	Name string
	Tags string `gorm:"type:json"`
}

func (pageBuilderArticle) TableName() string {
	return "page_builder_articles"
}

func TestExecuteAsTypedWithSelectAndJoinCountsSuccessfully(t *testing.T) {
	db := newTestSQLiteDB(t)
	if err := db.AutoMigrate(&pageBuilderUser{}, &pageBuilderProfile{}); err != nil {
		t.Fatalf("AutoMigrate returned error: %v", err)
	}

	if err := db.Create(&pageBuilderUser{ID: 1, Name: "alice"}).Error; err != nil {
		t.Fatalf("create user returned error: %v", err)
	}
	if err := db.Create(&pageBuilderProfile{ID: 1, UserID: 1, Bio: "engineer"}).Error; err != nil {
		t.Fatalf("create profile returned error: %v", err)
	}

	repo := orz.NewRepository[pageBuilderUser, uint](db)
	result, err := orz.ExecuteAsTyped[pageBuilderUserView](context.Background(), orz.Query(repo).
		Select("page_builder_users.id, page_builder_users.name, page_builder_profiles.bio").
		LeftJoin("page_builder_profiles", "page_builder_profiles.user_id = page_builder_users.id").
		SortByDesc("id", "id", "name"))
	if err != nil {
		t.Fatalf("ExecuteAsTyped returned error: %v", err)
	}

	if result.Total != 1 {
		t.Fatalf("expected total 1, got %d", result.Total)
	}
	if len(result.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(result.Items))
	}
	if result.Items[0].Bio != "engineer" {
		t.Fatalf("expected bio engineer, got %q", result.Items[0].Bio)
	}
}

func TestPageBuilderRejectsUnsafeSortFieldWithoutWhitelist(t *testing.T) {
	db := newTestSQLiteDB(t)
	repo := orz.NewRepository[pageBuilderUser, uint](db)

	_, err := orz.Query(repo).
		SortByDesc("id desc; drop table page_builder_users; --").
		Execute(context.Background())
	if err == nil {
		t.Fatal("expected sort validation error")
	}
	if !strings.Contains(err.Error(), "sort validation failed") {
		t.Fatalf("expected sort validation error, got %v", err)
	}
}

func TestPageBuilderRejectsUnsafeMatcherFieldName(t *testing.T) {
	db := newTestSQLiteDB(t)
	repo := orz.NewRepository[pageBuilderUser, uint](db)

	_, err := orz.Query(repo).
		Contains("name desc; drop table page_builder_users; --", "alice").
		Execute(context.Background())
	if err == nil {
		t.Fatal("expected matcher field validation error")
	}
	if !strings.Contains(err.Error(), "invalid query field") {
		t.Fatalf("expected query field validation error, got %v", err)
	}
}

func TestPageBuilderRejectsUnsafeKeywordFieldName(t *testing.T) {
	db := newTestSQLiteDB(t)
	repo := orz.NewRepository[pageBuilderUser, uint](db)

	_, err := orz.Query(repo).
		Keyword([]string{"name desc; drop table page_builder_users; --"}, "alice").
		Execute(context.Background())
	if err == nil {
		t.Fatal("expected keyword field validation error")
	}
	if !strings.Contains(err.Error(), "invalid keyword field") {
		t.Fatalf("expected keyword field validation error, got %v", err)
	}
}

func TestTagsMatcherWorksWithSQLiteJSONArrays(t *testing.T) {
	db := newTestSQLiteDB(t)
	if err := db.AutoMigrate(&pageBuilderArticle{}); err != nil {
		t.Fatalf("AutoMigrate returned error: %v", err)
	}

	if err := db.Create(&pageBuilderArticle{
		ID:   1,
		Name: "gorm guide",
		Tags: `["go","gorm"]`,
	}).Error; err != nil {
		t.Fatalf("create article returned error: %v", err)
	}
	if err := db.Create(&pageBuilderArticle{
		ID:   2,
		Name: "echo guide",
		Tags: `["go","echo"]`,
	}).Error; err != nil {
		t.Fatalf("create article returned error: %v", err)
	}

	repo := orz.NewRepository[pageBuilderArticle, uint](db)
	result, err := orz.Query(repo).
		Tags("tags", "gorm").
		Execute(context.Background())
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if result.Total != 1 {
		t.Fatalf("expected total 1, got %d", result.Total)
	}
	if len(result.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(result.Items))
	}
	if result.Items[0].Name != "gorm guide" {
		t.Fatalf("expected gorm guide, got %q", result.Items[0].Name)
	}
}

func newTestSQLiteDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := orz.ConnectDatabase(orz.DatabaseConfig{
		Type: orz.DatabaseSqlite,
		Sqlite: orz.SqliteConfig{
			Path: ":memory:",
		},
	})
	if err != nil {
		t.Fatalf("ConnectDatabase returned error: %v", err)
	}

	return db
}
