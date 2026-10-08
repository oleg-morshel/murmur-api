package auth_postgres

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/oleg-morshel/murmur-api/internal/core/domain"
	core_errors "github.com/oleg-morshel/murmur-api/internal/core/errors"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

type testPool struct {
	*pgxpool.Pool
}

func (testPool) OpTimeout() time.Duration { return 5 * time.Second }

var testRepo *AuthRepository

func TestMain(m *testing.M) {
	os.Exit(run(m))
}

func run(m *testing.M) int {
	ctx := context.Background()

	scripts, err := filepath.Glob("../../../../../migrations/*.up.sql")
	if err != nil || len(scripts) == 0 {
		log.Printf("migrations not found (err=%v, files=%d)", err, len(scripts))
		return 1
	}

	container, err := postgres.Run(ctx, "postgres:16-alpine",
		postgres.WithDatabase("murmur_test"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		postgres.WithInitScripts(scripts...),
		postgres.BasicWaitStrategies(),
	)
	if err != nil {
		log.Printf("start postgres container: %v", err)
		return 1
	}
	defer func() {
		if err := testcontainers.TerminateContainer(container); err != nil {
			log.Printf("terminate container: %v", err)
		}
	}()

	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		log.Printf("connection string: %v", err)
		return 1
	}

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Printf("create pool: %v", err)
		return 1
	}
	defer pool.Close()

	testRepo = NewAuthRepository(testPool{pool})

	return m.Run()
}

func newUser() *domain.User {
	n := time.Now().UnixNano()
	return &domain.User{
		Username:     fmt.Sprintf("user%d", n),
		Email:        fmt.Sprintf("user%d@example.com", n),
		PasswordHash: "$2a$10$hashhashhashhashhashhashhashhashhashhashhashhashhash",
	}
}

func TestAuthRepository_CreateAndGetByEmail(t *testing.T) {
	ctx := context.Background()
	user := newUser()

	id, err := testRepo.Create(ctx, user)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if id <= 0 {
		t.Fatalf("Create returned id = %d, want > 0", id)
	}

	got, err := testRepo.GetByEmail(ctx, user.Email)
	if err != nil {
		t.Fatalf("GetByEmail: %v", err)
	}

	if got.ID != id {
		t.Errorf("ID = %d, want %d", got.ID, id)
	}
	if got.Username != user.Username {
		t.Errorf("Username = %q, want %q", got.Username, user.Username)
	}
	if got.Email != user.Email {
		t.Errorf("Email = %q, want %q", got.Email, user.Email)
	}
	if got.PasswordHash != user.PasswordHash {
		t.Errorf("PasswordHash = %q, want %q", got.PasswordHash, user.PasswordHash)
	}
	if got.CreatedAt.IsZero() || got.UpdatedAt.IsZero() {
		t.Errorf("timestamps not set: created=%v updated=%v", got.CreatedAt, got.UpdatedAt)
	}
}

func TestAuthRepository_Create_DuplicateEmail(t *testing.T) {
	ctx := context.Background()
	first := newUser()

	if _, err := testRepo.Create(ctx, first); err != nil {
		t.Fatalf("first Create: %v", err)
	}

	second := newUser()
	second.Email = first.Email

	_, err := testRepo.Create(ctx, second)
	if !errors.Is(err, core_errors.ErrConflict) {
		t.Fatalf("second Create error = %v, want ErrConflict", err)
	}
}

func TestAuthRepository_GetByEmail_NotFound(t *testing.T) {
	_, err := testRepo.GetByEmail(context.Background(), "nobody-here@example.com")
	if !errors.Is(err, core_errors.ErrNotFound) {
		t.Fatalf("GetByEmail error = %v, want ErrNotFound", err)
	}
}

func TestAuthRepository_GetById(t *testing.T) {
	ctx := context.Background()
	user := newUser()

	id, err := testRepo.Create(ctx, user)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	t.Run("found", func(t *testing.T) {
		got, err := testRepo.GetById(ctx, id)
		if err != nil {
			t.Fatalf("GetById: %v", err)
		}
		if got.ID != id || got.Email != user.Email || got.Username != user.Username {
			t.Errorf("user = %+v, want id %d email %q username %q", got, id, user.Email, user.Username)
		}
	})

	t.Run("not found", func(t *testing.T) {
		_, err := testRepo.GetById(ctx, 999999999)
		if !errors.Is(err, core_errors.ErrNotFound) {
			t.Fatalf("GetById error = %v, want ErrNotFound", err)
		}
	})
}
