package main

import (
	"context"
	"errors"
	"log"
	"os"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

var (
	ErrNotFound      = errors.New("not found in db")
	ErrReadFailed    = errors.New("read from db failed")
	ErrWriteFailed   = errors.New("write to db failed")
	ErrNoRowsDeleted = errors.New("no rows deleted from db")
)

type repository struct {
	pool *pgxpool.Pool
}

func (repo *repository) close() {
	if repo.pool == nil {
		return
	}

	repo.pool.Close()
}

func (repo *repository) createRedirect(link, url string) (Redirect, error) {
	var rd Redirect

	query := `
	INSERT INTO redirects (link, url)
	VALUES ($1, $2)
	RETURNING id, link, url, access_code, created_at`

	row := repo.pool.QueryRow(context.Background(), query, link, url)

	err := row.Scan(&rd.id, &rd.Link, &rd.URL, &rd.AccessCode, &rd.CreatedAt)
	if err != nil {
		log.Println(err)

		return rd, ErrWriteFailed
	}

	return rd, nil
}

func (repo *repository) getRedirectByAccessCode(accessCode string) (Redirect, error) {
	var rd Redirect

	redirectQuery := `
	SELECT id, link, url, access_code, created_at
	FROM redirects
	WHERE access_code = $1
	AND deleted_at IS NULL`

	row := repo.pool.QueryRow(context.Background(), redirectQuery, accessCode)

	err := row.Scan(&rd.id, &rd.Link, &rd.URL, &rd.AccessCode, &rd.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return rd, ErrNotFound
		}

		log.Println(err)

		return rd, ErrReadFailed
	}

	visitsQuery := `
	SELECT ip_address, country, region_name, city, isp, created_at
	FROM visits
	WHERE redirect_id = $1
	`

	rows, err := repo.pool.Query(context.Background(), visitsQuery, rd.id)
	if err != nil {
		log.Println(err)

		return rd, ErrReadFailed
	}

	defer rows.Close()

	for rows.Next() {
		var v Visit

		err = rows.Scan(&v.IPAddress, &v.Country, &v.RegionName, &v.City, &v.ISP, &v.CreatedAt)
		if err != nil {
			log.Println(err)

			return rd, ErrReadFailed
		}

		rd.Visits = append(rd.Visits, v)
	}

	return rd, nil
}

func (repo *repository) getRedirectByLink(link string) (Redirect, error) {
	var rd Redirect

	query := `
	SELECT id, link, url, access_code, created_at
	FROM redirects
	WHERE link = $1
	AND deleted_at IS NULL`

	row := repo.pool.QueryRow(context.Background(), query, link)

	err := row.Scan(&rd.id, &rd.Link, &rd.URL, &rd.AccessCode, &rd.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return rd, ErrNotFound
		}

		log.Println(err)

		return rd, ErrReadFailed
	}

	return rd, nil
}

func (repo *repository) deleteRedirectByAccessCode(accessCode string) error {
	query := "UPDATE redirects SET deleted_at = NOW() WHERE access_code = $1 AND deleted_at IS NULL;"

	ct, err := repo.pool.Exec(context.Background(), query, accessCode)
	if err != nil {
		log.Println(err)

		return ErrWriteFailed
	}

	if ct.RowsAffected() == 0 {
		return ErrNoRowsDeleted
	}

	return nil
}

func (repo *repository) createVisit(rd Redirect, visit Visit) error {
	query := "INSERT INTO visits (redirect_id, ip_address, country, region_name, city, isp) VALUES ($1, $2, $3, $4, $5, $6 )"

	_, err := repo.pool.Exec(context.Background(), query, rd.id, visit.IPAddress, visit.Country, visit.RegionName, visit.City, visit.ISP)
	if err != nil {
		log.Println(err)

		return ErrWriteFailed
	}

	return nil
}

// Generates a new repository.
func newRepository() repository {
	pool := createDBPool()

	r := repository{pool: pool}

	return r
}

// Attempts to establish a DB connection.
// If it fails, it returns nil.
func createDBPool() *pgxpool.Pool {
	DBUrl := os.Getenv("db_url")

	conn, err := pgxpool.Connect(context.Background(), DBUrl)
	if err != nil {
		log.Println("Failed to establish DB connection")

		if DBUrl == "" {
			log.Println("db_url is empty")
		}

		return nil
	}

	return conn
}
