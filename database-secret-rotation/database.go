package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"
)

type DatabaseParameters struct {
	hostname string
	port     string
	name     string
	timeout  time.Duration
}

type DatabaseCredentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Database struct {
	connection      *sql.DB
	connectionMutex sync.Mutex
	parameters      DatabaseParameters
}

type Product struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// NewDatabase establishes a database connection with the given Vault credentials
func NewDatabase(
	ctx context.Context,
	parameters DatabaseParameters,
	credentials DatabaseCredentials,
) (*Database, error) {
	database := &Database{
		connection:      nil,
		connectionMutex: sync.Mutex{},
		parameters:      parameters,
	}

	return database, nil
}

// Reconnect will be called periodically to refresh the database connection
// since the dynamic credentials expire after some time, it will:
//  1. construct a connection string using the given credentials
//  2. establish a database connection
//  3. close & replace the existing connection with the new one behind a mutex
func (db *Database) Reconnect(ctx context.Context, credentials DatabaseCredentials) error {
	ctx, cancelCtxFunc := context.WithTimeout(ctx, db.parameters.timeout)
	defer cancelCtxFunc()

	log.Printf(
		"connecting to %q database @ %s:%s with username %q",
		db.parameters.name,
		db.parameters.hostname,
		db.parameters.port,
		credentials.Username,
	)

	connectionString := fmt.Sprintf(
		"host=%s port=%s dbname=%s user=%s password=%s sslmode=false",
		db.parameters.hostname,
		db.parameters.port,
		db.parameters.name,
		credentials.Username,
		credentials.Password,
	)

	connection, err := sql.Open("postgres", connectionString)

	if err != nil {
		return fmt.Errorf("unable to open database connection: %w", err)
	}

	// wait until the database is ready or timeout expires
	for {
		err = connection.Ping()
		if err != nil {
			break
		}
		select {
		case <-time.After(500 * time.Millisecond):
			continue
		case <-ctx.Done():
			return fmt.Errorf("failed to successfully ping database before context timeout: %w", err)
		}
	}

	db.closeReplaceConnection(connection)

	log.Printf("connecting to %q database: success!", db.parameters.name)

	return nil

}

func (db *Database) closeReplaceConnection(newDB *sql.DB) {
	db.connectionMutex.Lock()
	defer db.connectionMutex.Unlock()

	if db.connection != nil {
		_ = db.connection.Close()
	}

	// replace with new
	db.connection = newDB
}

func (db *Database) Close() error {
	/* */ db.connectionMutex.Lock()
	defer db.connectionMutex.Unlock()

	if db.connection != nil {
		return db.connection.Close()
	}

	return nil
}

func (db *Database) GetProducts(ctx context.Context) ([]Product, error) {
	db.connectionMutex.Lock()
	defer db.connectionMutex.Unlock()

	const query = "SELECT id, name FROM products"

	rows, err := db.connection.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to execute %q query: %w", query, err)
	}

	defer func() {
		_ = rows.Close()
	}()

	var products []Product

	for rows.Next() {
		var product Product
		if err := rows.Scan(
			&product.ID,
			&product.Name,
		); err != nil {
			return nil, fmt.Errorf("failed to scan table row for %q query: %w", query, err)
		}
		products = append(products, product)

	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error after scanning %q query: %w", query, err)
	}

	return products, nil
}
