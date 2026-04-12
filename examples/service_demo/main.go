package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/go-orz/orz"
	_ "github.com/go-orz/orz/drivers/sqlite"
	"gorm.io/gorm"
)

// User 用户模型
type User struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	Username  string    `gorm:"uniqueIndex;size:50" json:"username"`
	Email     string    `gorm:"size:100" json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (User) TableName() string {
	return "users"
}

// UserProfile 用户档案模型
type UserProfile struct {
	ID     uint   `gorm:"primarykey" json:"id"`
	UserID uint   `gorm:"index" json:"user_id"`
	Bio    string `gorm:"size:500" json:"bio"`
	Avatar string `gorm:"size:200" json:"avatar"`
}

func (UserProfile) TableName() string {
	return "user_profiles"
}

// UserService 用户服务，继承 Service 基类
// 负责处理用户相关的复杂业务逻辑
type UserService struct {
	*orz.Service
	userRepo    orz.Repository[User, uint]
	profileRepo orz.Repository[UserProfile, uint]
}

// NewUserService 创建用户服务
func NewUserService(db *gorm.DB) *UserService {
	return &UserService{
		Service:     orz.NewService(db),
		userRepo:    orz.NewRepository[User, uint](db),
		profileRepo: orz.NewRepository[UserProfile, uint](db),
	}
}

// CreateUserWithProfile 创建用户和档案（业务逻辑：一个用户必须有档案）
func (s *UserService) CreateUserWithProfile(ctx context.Context, user *User, bio string) error {
	// 业务逻辑：检查用户名是否已存在（使用改进的查询）
	exists, err := s.userRepo.Exists(ctx, []orz.Matcher{
		orz.NewMatcher("username", user.Username, orz.MatcherEqual),
	})
	if err != nil {
		return fmt.Errorf("failed to check username: %w", err)
	}
	if exists {
		return fmt.Errorf("username %s already exists", user.Username)
	}

	// 在事务中创建用户和档案
	return s.Transaction(ctx, func(txCtx context.Context) error {
		// 创建用户
		if err := s.userRepo.Create(txCtx, user); err != nil {
			return fmt.Errorf("failed to create user: %w", err)
		}

		// 创建用户档案
		profile := &UserProfile{
			UserID: user.ID,
			Bio:    bio,
			Avatar: "default.png",
		}
		if err := s.profileRepo.Create(txCtx, profile); err != nil {
			return fmt.Errorf("failed to create user profile: %w", err)
		}

		return nil
	})
}

// GetUserWithProfile 获取用户及其档案信息（业务逻辑：组合数据）
func (s *UserService) GetUserWithProfile(ctx context.Context, userID uint) (map[string]interface{}, error) {
	user, err := s.userRepo.FindById(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// 查找用户档案
	profiles, err := s.profileRepo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get profiles: %w", err)
	}

	var userProfile *UserProfile
	for _, profile := range profiles {
		if profile.UserID == userID {
			userProfile = &profile
			break
		}
	}

	result := map[string]interface{}{
		"user": user,
	}
	if userProfile != nil {
		result["profile"] = userProfile
	}

	return result, nil
}

// GetUserStats 获取用户统计信息（业务逻辑：复杂统计）
func (s *UserService) GetUserStats(ctx context.Context) (map[string]interface{}, error) {
	userCount, err := s.userRepo.Count(ctx)
	if err != nil {
		return nil, err
	}

	profileCount, err := s.profileRepo.Count(ctx)
	if err != nil {
		return nil, err
	}

	users, err := s.userRepo.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"total_users":     userCount,
		"total_profiles":  profileCount,
		"completion_rate": float64(profileCount) / float64(userCount) * 100,
		"users":           users,
	}, nil
}

// BulkCreateUsers 批量创建用户（业务逻辑：批量处理）
func (s *UserService) BulkCreateUsers(ctx context.Context, users []User) error {
	// 业务逻辑：验证所有用户名都不重复
	existingUsers, err := s.userRepo.FindAll(ctx)
	if err != nil {
		return err
	}

	existingUsernames := make(map[string]bool)
	for _, u := range existingUsers {
		existingUsernames[u.Username] = true
	}

	for _, user := range users {
		if existingUsernames[user.Username] {
			return fmt.Errorf("username %s already exists", user.Username)
		}
	}

	// 在事务中批量创建
	return s.Transaction(ctx, func(txCtx context.Context) error {
		return s.userRepo.CreateInBatches(txCtx, users, 10)
	})
}

// DemoApp 演示应用
type DemoApp struct {
	userService *UserService
}

func (a *DemoApp) Configure(app *orz.App) error {
	// 获取数据库并自动迁移
	db := app.GetDatabase()
	if err := db.AutoMigrate(&User{}, &UserProfile{}); err != nil {
		return err
	}

	// 创建用户服务
	a.userService = NewUserService(db)

	fmt.Println("=== Service 基类使用演示 ===")
	fmt.Println()

	return a.demonstrateServiceUsage(db)
}

func (a *DemoApp) demonstrateServiceUsage(db *gorm.DB) error {
	// 演示1: 创建用户和档案（业务逻辑）
	fmt.Println("1. 创建用户和档案（业务逻辑）:")
	ctx := orz.WithDB(context.Background(), db)

	user1 := &User{
		Username: "alice",
		Email:    "alice@example.com",
	}

	if err := a.userService.CreateUserWithProfile(ctx, user1, "我是Alice，一个软件工程师"); err != nil {
		return fmt.Errorf("创建用户和档案失败: %w", err)
	}
	fmt.Printf("   创建用户和档案成功: %+v\n", user1)

	// 演示2: 批量创建用户（带业务验证）
	fmt.Println("\n2. 批量创建用户（带业务验证）:")
	users := []User{
		{Username: "bob", Email: "bob@example.com"},
		{Username: "charlie", Email: "charlie@example.com"},
	}

	if err := a.userService.BulkCreateUsers(ctx, users); err != nil {
		return fmt.Errorf("批量创建用户失败: %w", err)
	}
	fmt.Printf("   批量创建用户成功: %d 个用户\n", len(users))

	// 演示3: 获取用户和档案信息（组合数据）
	fmt.Println("\n3. 获取用户和档案信息（组合数据）:")
	userWithProfile, err := a.userService.GetUserWithProfile(ctx, user1.ID)
	if err != nil {
		return fmt.Errorf("获取用户档案失败: %w", err)
	}
	fmt.Printf("   用户档案: %+v\n", userWithProfile)

	// 演示4: 复杂统计（业务逻辑）
	fmt.Println("\n4. 复杂统计（业务逻辑）:")
	stats, err := a.userService.GetUserStats(ctx)
	if err != nil {
		return fmt.Errorf("获取用户统计失败: %w", err)
	}
	fmt.Printf("   用户统计: %+v\n", stats)

	// 演示5: 创建更多用户档案，测试事务
	fmt.Println("\n5. 创建更多用户档案（测试事务）:")
	user2 := &User{Username: "david", Email: "david@example.com"}
	if err := a.userService.CreateUserWithProfile(ctx, user2, "我是David，一个产品经理"); err != nil {
		return fmt.Errorf("创建用户档案失败: %w", err)
	}

	user3 := &User{Username: "eve", Email: "eve@example.com"}
	if err := a.userService.CreateUserWithProfile(ctx, user3, "我是Eve，一个设计师"); err != nil {
		return fmt.Errorf("创建用户档案失败: %w", err)
	}

	// 最终统计
	finalStats, err := a.userService.GetUserStats(ctx)
	if err != nil {
		return fmt.Errorf("获取最终统计失败: %w", err)
	}
	fmt.Printf("\n=== 最终统计 ===\n")
	fmt.Printf("   用户总数: %v\n", finalStats["total_users"])
	fmt.Printf("   档案总数: %v\n", finalStats["total_profiles"])
	fmt.Printf("   档案完成率: %.2f%%\n", finalStats["completion_rate"])

	return nil
}

func main() {
	configMap := map[string]interface{}{
		"log": map[string]interface{}{
			"level":   "info",
			"encode":  "console",
			"console": true,
		},
		"database": map[string]interface{}{
			"type": "sqlite",
			"sqlite": map[string]interface{}{
				"path": ":memory:",
			},
		},
	}

	app := &DemoApp{}

	framework, err := orz.NewFramework(
		orz.WithConfigMap(configMap),
		orz.WithLoggerFromConfig(),
		orz.WithDatabase(),
		orz.WithApplication(app),
	)
	if err != nil {
		log.Fatal("框架初始化失败:", err)
	}

	if err := framework.Run(); err != nil {
		log.Fatal("运行失败:", err)
	}
}
