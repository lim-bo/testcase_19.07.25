package subs

import (
	"context"
	"errors"
	"log"
	"testcase/internal/errvalues"
	"testcase/models"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PgConnection interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	Ping(ctx context.Context) error
}

type Client struct {
	conn PgConnection
}

type DBConfig struct {
	Address  string
	User     string
	Password string
	DBName   string
	Options  map[string]string
}

func New(cfg *DBConfig) *Client {
	optsStr := ""
	if len(cfg.Options) != 0 {
		optsStr = "?"
		for k, v := range cfg.Options {
			optsStr += k + "=" + v
		}
	}
	p, err := pgxpool.New(context.Background(), "postgresql://"+cfg.User+":"+cfg.Password+"@"+cfg.Address+"/"+cfg.DBName+optsStr)
	if err != nil {
		log.Fatal(err)
	}
	err = p.Ping(context.Background())
	if err != nil {
		log.Fatal("ping error: " + err.Error())
	}
	return &Client{
		conn: p,
	}
}

func NewWithConn(conn PgConnection) *Client {
	err := conn.Ping(context.Background())
	if err != nil {
		log.Fatal("ping error: " + err.Error())
	}
	return &Client{
		conn: conn,
	}
}

// Creates a new subscription row in db
func (cli *Client) AddSub(s *models.Subscription) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	_, err := cli.conn.Exec(ctx, `INSERT INTO subscriptions (uid, name, cost, created_at, expires) VALUES
($1, $2, $3, $4, $5);`, s.UID, s.Name, s.Price, s.Start, s.Expires)
	if err != nil {
		return errors.New("error inserting sub: " + err.Error())
	}
	return nil
}

// Returns subscription by provided id, if there is no any
// returns ErrNoSuchRow
func (cli *Client) GetSub(id int) (*models.Subscription, error) {
	result := models.Subscription{
		ID: id,
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	row := cli.conn.QueryRow(ctx, `SELECT uid, name, cost, created_at, expires FROM subscriptions WHERE id = $1;`, id)
	if err := row.Scan(&result.UID, &result.Name, &result.Price, &result.Start, &result.Expires); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errvalues.ErrNoSuchRow
		}
		return nil, errors.New("error getting subscription: " + err.Error())
	}
	return &result, nil
}

// Takes new subscription info and updates row with provided id,
// if there is no any returns ErrNoSuchRow
func (cli *Client) UpdateSub(id int, s *models.Subscription) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	tag, err := cli.conn.Exec(ctx, `UPDATE subscriptions SET uid = $1, name = $2, cost = $3, created_at = $4, expires = $5 WHERE id = $6;`,
		s.UID, s.Name, s.Price, s.Start, s.Expires, id)
	if err != nil {
		return errors.New("error updating subscription: " + err.Error())
	} else if tag.RowsAffected() == 0 {
		return errvalues.ErrNoSuchRow
	}
	return nil
}

// Deletes row with provided id, if there is no any returns ErrNoSuchRow
func (cli *Client) DeleteSub(id int) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	tag, err := cli.conn.Exec(ctx, `DELETE FROM subscriptions WHERE id = $1;`, id)
	if err != nil {
		return errors.New("deleting sub error: " + err.Error())
	} else if tag.RowsAffected() == 0 {
		return errvalues.ErrNoSuchRow
	}
	return nil
}

// Takes opts for filtering, limit, order and offset settings and returns
// list of subscriptions. opts.Filter and opts.Order can be nil for unfiltered
// and unordered result
func (cli *Client) ListSubs(opts *models.ListOpts) ([]*models.Subscription, error) {
	query := squirrel.Select("id, name, uid, cost, created_at, expires").
		From("subscriptions").
		Offset(uint64(opts.Offset))
	if opts.Limit != 0 {
		query = query.Limit(uint64(opts.Limit))
	}
	if opts.Order != "" {
		query = query.OrderBy(opts.Order)
	}
	if opts.Filter != nil {
		query = query.Where(squirrel.Eq(opts.Filter))
	}
	sql, args, err := query.PlaceholderFormat(squirrel.Dollar).ToSql()
	if err != nil {
		return nil, errors.New("building query error: " + err.Error())
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()
	rows, err := cli.conn.Query(ctx, sql, args...)
	if err != nil {
		return nil, errors.New("getting subs list error: " + err.Error())
	}
	result := make([]*models.Subscription, 0, len(rows.RawValues()))
	for rows.Next() {
		s := models.Subscription{}
		err = rows.Scan(&s.ID, &s.Name, &s.UID, &s.Price, &s.Start, &s.Expires)
		if err != nil {
			return nil, errors.New("error converting rows error: " + err.Error())
		}
		result = append(result, &s)
	}
	return result, nil
}

// Returns sum of found rows with provided filter.
// If filter is nil, returns sum of all rows
func (cli *Client) PriceSum(filter map[string]interface{}) (int, error) {
	query := squirrel.Select("SUM(cost)").
		From("subscriptions").
		Where(squirrel.Eq(filter))
	if filter != nil {
		query = query.Where(squirrel.Eq(filter))
	}
	sql, args, err := query.PlaceholderFormat(squirrel.Dollar).ToSql()
	if err != nil {
		return 0, errors.New("building query error: " + err.Error())
	}
	var result int
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()
	if err = cli.conn.QueryRow(ctx, sql, args...).Scan(&result); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, errvalues.ErrNoSuchRow
		}
		return 0, errors.New("getting subs sum error: " + err.Error())
	}
	return result, nil
}
