package subs_test

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"testcase/internal/errvalues"
	"testcase/internal/subs"
	"testcase/models"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pashagolub/pgxmock/v2"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestAddSub(t *testing.T) {
	t.Parallel()
	pool, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		pool.Close()
	})
	cli := subs.NewWithConn(pool)
	start, _ := time.Parse("01-2006", "07-2025")
	exp, _ := time.Parse("01-2006", "08-2025")
	sub := &models.Subscription{
		Name:    "yandex",
		Price:   400,
		UID:     uuid.New(),
		Start:   start,
		Expires: &exp,
	}
	query := regexp.QuoteMeta(`INSERT INTO subscriptions (uid, name, cost, created_at, expires) VALUES ($1, $2, $3, $4, $5);`)
	t.Run("successful", func(t *testing.T) {
		pool.ExpectExec(query).
			WillReturnResult(pgxmock.NewResult("INSERT", 1)).
			WithArgs(sub.UID, sub.Name, sub.Price, sub.Start, sub.Expires)
		err = cli.AddSub(sub)
		assert.NoError(t, err)
	})
	t.Run("with error", func(t *testing.T) {
		pool.ExpectExec(query).
			WithArgs(sub.UID, sub.Name, sub.Price, sub.Start, sub.Expires).
			WillReturnError(errors.New("db error"))
		err = cli.AddSub(sub)
		assert.Error(t, err)
	})
}

func TestGetSub(t *testing.T) {
	t.Parallel()
	pool, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		pool.Close()
	})
	cli := subs.NewWithConn(pool)
	start, _ := time.Parse("01-2006", "07-2025")
	exp, _ := time.Parse("01-2006", "08-2025")
	sub := &models.Subscription{
		ID:      1,
		Name:    "yandex",
		Price:   400,
		UID:     uuid.New(),
		Start:   start,
		Expires: &exp,
	}
	query := regexp.QuoteMeta(`SELECT uid, name, cost, created_at, expires FROM subscriptions WHERE id = $1;`)
	t.Run("successful", func(t *testing.T) {
		pool.ExpectQuery(query).
			WithArgs(1).
			WillReturnRows(pgxmock.NewRows([]string{"uid", "name", "cost", "created_at", "expires"}).
				AddRow(sub.UID, sub.Name, sub.Price, sub.Start, sub.Expires))
		result, err := cli.GetSub(1)
		assert.NoError(t, err)
		assert.Equal(t, sub, result)
	})
	t.Run("db error", func(t *testing.T) {
		pool.ExpectQuery(query).
			WithArgs(1).
			WillReturnError(errors.New("db error"))
		_, err := cli.GetSub(1)
		assert.Error(t, err)
	})
	t.Run("No row with such id", func(t *testing.T) {
		pool.ExpectQuery(query).
			WithArgs(1).
			WillReturnError(pgx.ErrNoRows)
		_, err := cli.GetSub(1)
		assert.ErrorIs(t, err, errvalues.ErrNoSuchRow)
	})
}

func TestUpdateSub(t *testing.T) {
	t.Parallel()
	pool, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		pool.Close()
	})
	cli := subs.NewWithConn(pool)
	start, _ := time.Parse("01-2006", "07-2025")
	exp, _ := time.Parse("01-2006", "08-2025")
	sub := &models.Subscription{
		Name:    "yandex",
		Price:   400,
		UID:     uuid.New(),
		Start:   start,
		Expires: &exp,
	}
	id := 1
	query := regexp.QuoteMeta(`UPDATE subscriptions SET uid = $1, name = $2, cost = $3, created_at = $4, expires = $5 WHERE id = $6`)
	t.Run("successful", func(t *testing.T) {
		pool.ExpectExec(query).
			WithArgs(sub.UID, sub.Name, sub.Price, sub.Start, sub.Expires, id).
			WillReturnResult(pgxmock.NewResult("UPDATE", 1))
		err := cli.UpdateSub(id, sub)
		assert.NoError(t, err)
	})
	t.Run("No row with such id", func(t *testing.T) {
		pool.ExpectExec(query).
			WithArgs(sub.UID, sub.Name, sub.Price, sub.Start, sub.Expires, id).
			WillReturnResult(pgxmock.NewResult("UPDATE", 0))
		err := cli.UpdateSub(id, sub)
		assert.ErrorIs(t, err, errvalues.ErrNoSuchRow)
	})
	t.Run("db error", func(t *testing.T) {
		pool.ExpectExec(query).
			WithArgs(sub.UID, sub.Name, sub.Price, sub.Start, sub.Expires, id).
			WillReturnError(errors.New("db error"))
		err := cli.UpdateSub(id, sub)
		assert.Error(t, err)
	})
}

func TestDelete(t *testing.T) {
	t.Parallel()
	pool, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		pool.Close()
	})
	cli := subs.NewWithConn(pool)
	id := 1
	query := regexp.QuoteMeta(`DELETE FROM subscriptions WHERE id = $1;`)
	t.Run("successful", func(t *testing.T) {
		pool.ExpectExec(query).
			WithArgs(id).
			WillReturnResult(pgxmock.NewResult("DELETE", 1))
		err := cli.DeleteSub(id)
		assert.NoError(t, err)
	})
	t.Run("No row with such id", func(t *testing.T) {
		pool.ExpectExec(query).
			WithArgs(id).
			WillReturnResult(pgxmock.NewResult("DELETE", 0))
		err := cli.DeleteSub(id)
		assert.ErrorIs(t, err, errvalues.ErrNoSuchRow)
	})
	t.Run("db error", func(t *testing.T) {
		pool.ExpectExec(query).
			WithArgs(id).
			WillReturnError(errors.New("db error"))
		err := cli.DeleteSub(id)
		assert.Error(t, err)
	})
}

// Setting up testcontainer for integrational test
func setupTestDB(t *testing.T) subs.DBConfig {
	container, err := postgres.Run(context.Background(), "postgres:17",
		postgres.WithUsername("test_user"),
		postgres.WithDatabase("url_shortener"),
		postgres.WithPassword("test_password"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	if err != nil {
		t.Fatal("error running test container: " + err.Error())
	}
	connStr, err := container.ConnectionString(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	pgxpoolCfg, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		t.Fatal("error parsing config: " + err.Error())
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), pgxpoolCfg)
	if err != nil {
		t.Fatal("error connecting to container: " + err.Error())
	}
	_, err = pool.Exec(context.Background(), `CREATE TABLE IF NOT EXISTS subscriptions (
    id SERIAL PRIMARY KEY,
    uid UUID NOT NULL, 
    name text NOT NULL,
    cost INTEGER NOT NULL,
    created_at DATE CHECK (EXTRACT(DAY FROM created_at) = 1) NOT NULL,
    expires TIMESTAMPTZ
);`)
	if err != nil {
		t.Fatal("error setting migrations: " + err.Error())
	}
	pool.Close()
	t.Cleanup(func() {
		container.Terminate(context.Background())
	})
	return subs.DBConfig{
		Address:  pgxpoolCfg.ConnConfig.Host + ":" + strconv.FormatUint(uint64(pgxpoolCfg.ConnConfig.Port), 10),
		Password: "test_password",
		User:     "test_user",
		DBName:   "url_shortener",
	}
}

func TestIntegrational(t *testing.T) {
	t.Parallel()
	cfg := setupTestDB(t)
	cli := subs.New(&cfg)
	start, _ := time.Parse("01-2006", "07-2025")
	exp, _ := time.Parse("01-2006", "08-2025")
	uid := uuid.New()
	for i := 0; i < 10; i++ {
		sub := &models.Subscription{
			Name:    fmt.Sprintf("name #%d", i),
			Price:   i * 100,
			UID:     uid,
			Start:   start,
			Expires: &exp,
		}
		err := cli.AddSub(sub)
		if err != nil {
			t.Fatal(err)
		}
	}

	t.Run("successfully listed", func(t *testing.T) {
		t.Parallel()
		result, err := cli.ListSubs(&models.ListOpts{
			Limit:  10,
			Offset: 0,
			Filter: nil,
			Order:  "id",
		})
		assert.NoError(t, err)
		for _, s := range result {
			t.Log(*s)
		}
	})
	t.Run("listed with filters", func(t *testing.T) {
		t.Parallel()
		filter := make(map[string]interface{})
		filter["id"] = 2

		result, err := cli.ListSubs(&models.ListOpts{
			Limit:  10,
			Offset: 0,
			Filter: filter,
			Order:  "id",
		})
		assert.NoError(t, err)
		if len(result) != 1 || result[0].ID != 2 {
			t.Error("unmatched expectations")
		}
	})
	t.Run("listed with limit and offset", func(t *testing.T) {
		t.Parallel()
		result, err := cli.ListSubs(&models.ListOpts{
			Limit:  5,
			Offset: 3,
			Filter: nil,
			Order:  "id",
		})
		assert.NoError(t, err)
		if len(result) != 5 {
			t.Error("unmatched expectations")
		}
		for _, s := range result {
			t.Log(*s)
		}
	})
	t.Run("got price sum", func(t *testing.T) {
		t.Parallel()
		expected := 4500
		sum, err := cli.PriceSum(nil)
		assert.NoError(t, err)
		assert.Equal(t, sum, expected)
	})
}
