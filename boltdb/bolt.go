package boltdb

import (
	"time"

	"github.com/pkg/errors"
	bolt "go.etcd.io/bbolt"
)

// KeyValueStorage is an interface for abstracting away the key-value storage layer
// from the application itself.
type KeyValueStorage interface {
	// Save saves a key-value pair to the database, at a particular bucket
	Save(key, value, bucket string) error
	// Delete deletes an object by Key
	Delete(key, bucket string) error
	// Get gets a value from a key
	Get(key, bucket string) (string, error)
	// GetAll gets all values, returning them in an array of strings
	GetAll(bucket string) (map[string]string, error)
	// Close closes the database
	Close() error
}

// Client is a client that interacts with bolt. It implements the
// KeyValueStorage interface
type Client struct {
	*bolt.DB
	Logger
}

// NewClient creates a new bolt Client. A path to a database must be passed in.
func NewClient(databasePath string, log Logger) (*Client, error) {
	if log == nil {
		log = &defaultLogger{}
	}
	db, err := bolt.Open(databasePath, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return nil, err
	}
	return &Client{
		DB:     db,
		Logger: log,
	}, nil
}

// Save saves a key-value pair to the database, at a particular bucket
func (c *Client) Save(key, value, bucket string) error {
	return c.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(bucket))
		if err != nil {
			return errors.Wrap(err, "create bucket error")
		}
		return b.Put([]byte(key), []byte(value))
	})
}

// Delete deletes an object by Key
func (c *Client) Delete(key, bucket string) error {
	return c.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			c.Errorw("bucket doesn't exist", "bucket", bucket)
			return nil
		}
		return b.Delete([]byte(key))
	})
}

// Get gets a value from a key
func (c *Client) Get(key, bucket string) (string, error) {
	var value string
	err := c.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			c.Debugw("bucket not found", "bucket", bucket)
			return nil
		}
		value = string(b.Get([]byte(key)))
		return nil
	})
	return value, err
}

// GetAll gets all values, returning them in a map of keys and values of strings
func (c *Client) GetAll(bucket string) (map[string]string, error) {
	results := make(map[string]string)
	err := c.View(func(tx *bolt.Tx) error {
		c.Debugw("getting all elements in bucket", "bucket", bucket)
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			c.Debugw("bucket not found", "bucket", bucket)
			return nil
		}
		return b.ForEach(func(k, v []byte) error {
			results[string(k)] = string(v)
			return nil
		})
	})
	return results, err
}
