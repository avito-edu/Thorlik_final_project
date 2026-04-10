package repository

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Thorlik/marketplace/internal/model/entity"
	"github.com/Thorlik/marketplace/internal/pkg/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func setupUserRepositoryTest(t *testing.T) (*userRepository, sqlmock.Sqlmock, func()) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)

	logger := zap.NewNop()
	txManager := db.NewTransactionManager(mockDB, logger)
	qManager := db.NewQueryManager(mockDB, logger)

	repo := &userRepository{
		txManager: txManager,
		qManager:  qManager,
		logger:    logger,
	}

	cleanup := func() {
		mockDB.Close()
	}

	return repo, mock, cleanup
}

func TestUserRepository_Create_Success(t *testing.T) {
	repo, mock, cleanup := setupUserRepositoryTest(t)
	defer cleanup()

	user := &entity.User{
		Email:        "test@example.com",
		PasswordHash: "hashedpassword",
		FirstName:    "Ivan",
		LastName:     "Khorolsky",
		Role:         "user",
	}

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "email", "password_hash", "first_name", "last_name", "role", "created_at", "updated_at"}).
		AddRow(1, user.Email, user.PasswordHash, user.FirstName, user.LastName, user.Role, now, now)

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO users`).
		WithArgs(user.Email, user.PasswordHash, user.FirstName, user.LastName, user.Role).
		WillReturnRows(rows)
	mock.ExpectCommit()

	result, err := repo.Create(context.Background(), user)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, int64(1), result.ID)
	assert.Equal(t, user.Email, result.Email)
	assert.Equal(t, user.FirstName, result.FirstName)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_Create_DuplicateEmail(t *testing.T) {
	repo, mock, cleanup := setupUserRepositoryTest(t)
	defer cleanup()

	user := &entity.User{
		Email:        "existing@example.com",
		PasswordHash: "hashedpassword",
		FirstName:    "Ivan",
		LastName:     "Khorolsky",
		Role:         "user",
	}

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO users`).
		WithArgs(user.Email, user.PasswordHash, user.FirstName, user.LastName, user.Role).
		WillReturnError(errors.New("duplicate key value violates unique constraint"))
	mock.ExpectRollback()

	result, err := repo.Create(context.Background(), user)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, ErrUserAlreadyExists, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_Create_DBError(t *testing.T) {
	repo, mock, cleanup := setupUserRepositoryTest(t)
	defer cleanup()

	user := &entity.User{
		Email:        "test@example.com",
		PasswordHash: "hashedpassword",
		FirstName:    "Ivan",
		LastName:     "Khorolsky",
		Role:         "user",
	}

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO users`).
		WithArgs(user.Email, user.PasswordHash, user.FirstName, user.LastName, user.Role).
		WillReturnError(errors.New("connection refused"))
	mock.ExpectRollback()

	result, err := repo.Create(context.Background(), user)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_Create_BeginTransactionError(t *testing.T) {
	repo, mock, cleanup := setupUserRepositoryTest(t)
	defer cleanup()

	user := &entity.User{
		Email:        "test@example.com",
		PasswordHash: "hashedpassword",
		FirstName:    "Ivan",
		LastName:     "Khorolsky",
		Role:         "user",
	}

	mock.ExpectBegin().WillReturnError(errors.New("connection error"))

	result, err := repo.Create(context.Background(), user)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_GetByID_Success(t *testing.T) {
	repo, mock, cleanup := setupUserRepositoryTest(t)
	defer cleanup()

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "email", "password_hash", "first_name", "last_name", "role", "created_at", "updated_at"}).
		AddRow(1, "test@example.com", "hash", "Ivan", "Khorolsky", "user", now, now)

	mock.ExpectQuery(`SELECT id, email, password_hash, first_name, last_name, role, created_at, updated_at FROM users WHERE id = \$1`).
		WithArgs(int64(1)).
		WillReturnRows(rows)

	result, err := repo.GetByID(context.Background(), 1)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, int64(1), result.ID)
	assert.Equal(t, "test@example.com", result.Email)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_GetByID_NotFound(t *testing.T) {
	repo, mock, cleanup := setupUserRepositoryTest(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT id, email, password_hash, first_name, last_name, role, created_at, updated_at FROM users WHERE id = \$1`).
		WithArgs(int64(999)).
		WillReturnError(sql.ErrNoRows)

	result, err := repo.GetByID(context.Background(), 999)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, ErrUserNotFound, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_GetByID_DBError(t *testing.T) {
	repo, mock, cleanup := setupUserRepositoryTest(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT id, email, password_hash, first_name, last_name, role, created_at, updated_at FROM users WHERE id = \$1`).
		WithArgs(int64(1)).
		WillReturnError(errors.New("database error"))

	result, err := repo.GetByID(context.Background(), 1)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_GetByEmail_Success(t *testing.T) {
	repo, mock, cleanup := setupUserRepositoryTest(t)
	defer cleanup()

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "email", "password_hash", "first_name", "last_name", "role", "created_at", "updated_at"}).
		AddRow(1, "test@example.com", "hash", "Ivan", "Khorolsky", "user", now, now)

	mock.ExpectQuery(`SELECT id, email, password_hash, first_name, last_name, role, created_at, updated_at FROM users WHERE email = \$1`).
		WithArgs("test@example.com").
		WillReturnRows(rows)

	result, err := repo.GetByEmail(context.Background(), "test@example.com")

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "test@example.com", result.Email)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_GetByEmail_NotFound(t *testing.T) {
	repo, mock, cleanup := setupUserRepositoryTest(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT id, email, password_hash, first_name, last_name, role, created_at, updated_at FROM users WHERE email = \$1`).
		WithArgs("notfound@example.com").
		WillReturnError(sql.ErrNoRows)

	result, err := repo.GetByEmail(context.Background(), "notfound@example.com")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, ErrUserNotFound, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_GetAll_Success(t *testing.T) {
	repo, mock, cleanup := setupUserRepositoryTest(t)
	defer cleanup()

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "email", "password_hash", "first_name", "last_name", "role", "created_at", "updated_at"}).
		AddRow(1, "user1@example.com", "hash1", "Ivan", "Khorolsky", "user", now, now).
		AddRow(2, "user2@example.com", "hash2", "Ivan", "Ivanov", "user", now, now)

	mock.ExpectQuery(`SELECT id, email, password_hash, first_name, last_name, role, created_at, updated_at FROM users ORDER BY created_at DESC LIMIT \$1 OFFSET \$2`).
		WithArgs(10, 0).
		WillReturnRows(rows)

	result, err := repo.GetAll(context.Background(), 10, 0)

	require.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "user1@example.com", result[0].Email)
	assert.Equal(t, "user2@example.com", result[1].Email)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_GetAll_Empty(t *testing.T) {
	repo, mock, cleanup := setupUserRepositoryTest(t)
	defer cleanup()

	rows := sqlmock.NewRows([]string{"id", "email", "password_hash", "first_name", "last_name", "role", "created_at", "updated_at"})

	mock.ExpectQuery(`SELECT id, email, password_hash, first_name, last_name, role, created_at, updated_at FROM users ORDER BY created_at DESC LIMIT \$1 OFFSET \$2`).
		WithArgs(10, 0).
		WillReturnRows(rows)

	result, err := repo.GetAll(context.Background(), 10, 0)

	require.NoError(t, err)
	assert.Empty(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_GetAll_DBError(t *testing.T) {
	repo, mock, cleanup := setupUserRepositoryTest(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT id, email, password_hash, first_name, last_name, role, created_at, updated_at FROM users ORDER BY created_at DESC LIMIT \$1 OFFSET \$2`).
		WithArgs(10, 0).
		WillReturnError(errors.New("database error"))

	result, err := repo.GetAll(context.Background(), 10, 0)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_Update_Success(t *testing.T) {
	repo, mock, cleanup := setupUserRepositoryTest(t)
	defer cleanup()

	now := time.Now()
	user := &entity.User{
		ID:        1,
		FirstName: "Ivan",
		LastName:  "Updated",
	}

	rows := sqlmock.NewRows([]string{"id", "email", "password_hash", "first_name", "last_name", "role", "created_at", "updated_at"}).
		AddRow(user.ID, "test@example.com", "hash", user.FirstName, user.LastName, "user", now, now)

	mock.ExpectBegin()
	mock.ExpectQuery(`UPDATE users SET first_name = \$1, last_name = \$2 WHERE id = \$3`).
		WithArgs(user.FirstName, user.LastName, user.ID).
		WillReturnRows(rows)
	mock.ExpectCommit()

	result, err := repo.Update(context.Background(), user)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "test@example.com", result.Email)
	assert.Equal(t, user.LastName, result.LastName)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_Update_NotFound(t *testing.T) {
	repo, mock, cleanup := setupUserRepositoryTest(t)
	defer cleanup()

	user := &entity.User{
		ID:        999,
		FirstName: "Ivan",
		LastName:  "Khorolsky",
	}

	mock.ExpectBegin()
	mock.ExpectQuery(`UPDATE users SET first_name = \$1, last_name = \$2 WHERE id = \$3`).
		WithArgs(user.FirstName, user.LastName, user.ID).
		WillReturnError(sql.ErrNoRows)
	mock.ExpectRollback()

	result, err := repo.Update(context.Background(), user)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, ErrUserNotFound, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_Update_DBError(t *testing.T) {
	repo, mock, cleanup := setupUserRepositoryTest(t)
	defer cleanup()

	user := &entity.User{
		ID:        1,
		FirstName: "Ivan",
		LastName:  "Khorolsky",
	}

	mock.ExpectBegin()
	mock.ExpectQuery(`UPDATE users SET first_name = \$1, last_name = \$2 WHERE id = \$3`).
		WithArgs(user.FirstName, user.LastName, user.ID).
		WillReturnError(errors.New("database error"))
	mock.ExpectRollback()

	result, err := repo.Update(context.Background(), user)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_UpdateRole_Success(t *testing.T) {
	repo, mock, cleanup := setupUserRepositoryTest(t)
	defer cleanup()

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE users SET role = \$1 WHERE id = \$2`).
		WithArgs("user", int64(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := repo.UpdateRole(context.Background(), 1, "user")

	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_UpdateRole_NotFound(t *testing.T) {
	repo, mock, cleanup := setupUserRepositoryTest(t)
	defer cleanup()

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE users SET role = \$1 WHERE id = \$2`).
		WithArgs("user", int64(999)).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectRollback()

	err := repo.UpdateRole(context.Background(), 999, "user")

	assert.Error(t, err)
	assert.Equal(t, ErrUserNotFound, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_UpdateRole_DBError(t *testing.T) {
	repo, mock, cleanup := setupUserRepositoryTest(t)
	defer cleanup()

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE users SET role = \$1 WHERE id = \$2`).
		WithArgs("user", int64(1)).
		WillReturnError(errors.New("database error"))
	mock.ExpectRollback()

	err := repo.UpdateRole(context.Background(), 1, "user")

	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_UpdatePassword_Success(t *testing.T) {
	repo, mock, cleanup := setupUserRepositoryTest(t)
	defer cleanup()

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE users SET password_hash = \$1 WHERE id = \$2`).
		WithArgs("newhash", int64(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := repo.UpdatePassword(context.Background(), 1, "newhash")

	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_UpdatePassword_NotFound(t *testing.T) {
	repo, mock, cleanup := setupUserRepositoryTest(t)
	defer cleanup()

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE users SET password_hash = \$1 WHERE id = \$2`).
		WithArgs("newhash", int64(999)).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectRollback()

	err := repo.UpdatePassword(context.Background(), 999, "newhash")

	assert.Error(t, err)
	assert.Equal(t, ErrUserNotFound, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_Delete_Success(t *testing.T) {
	repo, mock, cleanup := setupUserRepositoryTest(t)
	defer cleanup()

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM users WHERE id = \$1`).
		WithArgs(int64(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := repo.Delete(context.Background(), 1)

	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_Delete_NotFound(t *testing.T) {
	repo, mock, cleanup := setupUserRepositoryTest(t)
	defer cleanup()

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM users WHERE id = \$1`).
		WithArgs(int64(999)).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectRollback()

	err := repo.Delete(context.Background(), 999)

	assert.Error(t, err)
	assert.Equal(t, ErrUserNotFound, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_Delete_DBError(t *testing.T) {
	repo, mock, cleanup := setupUserRepositoryTest(t)
	defer cleanup()

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM users WHERE id = \$1`).
		WithArgs(int64(1)).
		WillReturnError(errors.New("database error"))
	mock.ExpectRollback()

	err := repo.Delete(context.Background(), 1)

	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_Count_Success(t *testing.T) {
	repo, mock, cleanup := setupUserRepositoryTest(t)
	defer cleanup()

	rows := sqlmock.NewRows([]string{"count"}).AddRow(42)

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM users`).
		WillReturnRows(rows)

	count, err := repo.Count(context.Background())

	require.NoError(t, err)
	assert.Equal(t, int64(42), count)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_Count_DBError(t *testing.T) {
	repo, mock, cleanup := setupUserRepositoryTest(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM users`).
		WillReturnError(errors.New("database error"))

	count, err := repo.Count(context.Background())

	assert.Error(t, err)
	assert.Equal(t, int64(0), count)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_Update_CommitError(t *testing.T) {
	repo, mock, cleanup := setupUserRepositoryTest(t)
	defer cleanup()

	now := time.Now()
	user := &entity.User{
		ID:        1,
		FirstName: "Ivan",
		LastName:  "Khorolsky",
	}

	rows := sqlmock.NewRows([]string{"id", "email", "password_hash", "first_name", "last_name", "role", "created_at", "updated_at"}).
		AddRow(user.ID, "test@example.com", "hash", user.FirstName, user.LastName, "user", now, now)

	mock.ExpectBegin()
	mock.ExpectQuery(`UPDATE users SET first_name = \$1, last_name = \$2 WHERE id = \$3`).
		WithArgs(user.FirstName, user.LastName, user.ID).
		WillReturnRows(rows)
	mock.ExpectCommit().WillReturnError(errors.New("commit failed"))

	result, err := repo.Update(context.Background(), user)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}
