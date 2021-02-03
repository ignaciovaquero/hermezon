package boltdb

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/igvaquero18/hermezon/utils"
	"github.com/stretchr/testify/assert"
	bolt "go.etcd.io/bbolt"
)

const (
	dbPath string = "./hermezon_test.db"
)

func TestSave(t *testing.T) {
	db, _ := bolt.Open(dbPath, 0600, &bolt.Options{Timeout: 1 * time.Second})
	client := &Client{
		DB:     db,
		Logger: &utils.DefaultLogger{},
	}

	testCases := []struct {
		name, key, value, bucket string
		err                      error
	}{
		{
			name:   "Save a key and a value in a non-existing bucket",
			key:    "some_key",
			value:  "some_value",
			bucket: "some_bucket",
			err:    nil,
		},
		{
			name:   "Save a key and a value in an existing bucket",
			key:    "some_key",
			value:  "some_value",
			bucket: "some_bucket",
			err:    nil,
		},
		{
			name:   "Save a key and a value in an empty bucket",
			key:    "some_key",
			value:  "some_value",
			bucket: "",
			err:    fmt.Errorf("empty bucket"),
		},
		{
			name:   "Save a value with an empty key in a new bucket",
			key:    "",
			value:  "some_value",
			bucket: "some_other_bucket",
			err:    fmt.Errorf("empty key"),
		},
		{
			name:   "Save an empty value in a key in an existing bucket",
			key:    "some_key",
			value:  "",
			bucket: "some_bucket",
			err:    nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(tt *testing.T) {
			if err := client.Save(tc.key, tc.value, tc.bucket); tc.err != nil {
				assert.Error(tt, err)
			} else {
				assert.NoError(tt, err)
				var value string
				err = client.View(func(tx *bolt.Tx) error {
					b := tx.Bucket([]byte(tc.bucket))
					assert.NotNil(tt, b)
					value = string(b.Get([]byte(tc.key)))
					return nil
				})
				assert.NoError(tt, err)
				assert.Equal(tt, tc.value, value)
			}
		})
	}
	os.Remove(dbPath)
}

func TestDelete(t *testing.T) {
	db, _ := bolt.Open(dbPath, 0600, &bolt.Options{Timeout: 1 * time.Second})
	client := &Client{
		DB:     db,
		Logger: &utils.DefaultLogger{},
	}
	const bucket = "some_bucket"
	client.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucket([]byte(bucket))
		if err != nil {
			os.Remove(dbPath)
			t.Fatalf("error creating bucket")
		}
		return nil
	})

	testCases := []struct {
		name, key, bucket string
	}{
		{
			name:   "Delete a key in an existing bucket",
			key:    "some_key",
			bucket: bucket,
		},
		{
			name:   "Delete a key in a non-existing bucket",
			key:    "some_key",
			bucket: "some_other_bucket",
		},
		{
			name:   "Delete a key in an empty bucket",
			key:    "some_key",
			bucket: "",
		},
		{
			name:   "Delete an empty key in an existing bucket",
			key:    "",
			bucket: bucket,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(tt *testing.T) {
			client.Update(func(tx *bolt.Tx) error {
				b := tx.Bucket([]byte(bucket))
				assert.NotNil(tt, b)
				if tc.key != "" {
					err := b.Put([]byte(tc.key), []byte("value"))
					assert.NoError(tt, err)
				}
				return nil
			})
			assert.NoError(tt, client.Delete(tc.key, tc.bucket))
		})
	}
	os.Remove(dbPath)
}

func TestGet(t *testing.T) {
	db, _ := bolt.Open(dbPath, 0600, &bolt.Options{Timeout: 1 * time.Second})
	client := &Client{
		DB:     db,
		Logger: &utils.DefaultLogger{},
	}
	const bucket = "some_bucket"

	testCases := []struct {
		name, key, value, bucket string
	}{
		{
			name:   "Get value for a key in an existing bucket",
			key:    "some_key",
			value:  "some_value",
			bucket: bucket,
		},
		{
			name:   "Get value for a key in an empty bucket",
			key:    "some_key",
			value:  "",
			bucket: "",
		},
		{
			name:   "Get value for an empty key in an existing bucket",
			key:    "",
			value:  "",
			bucket: bucket,
		},
		{
			name:   "Get value for a key in a non-existing bucket",
			key:    "some_key",
			value:  "",
			bucket: "some_other_bucket",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(tt *testing.T) {
			if tc.bucket == bucket {
				client.Update(func(tx *bolt.Tx) error {
					b, err := tx.CreateBucketIfNotExists([]byte(tc.bucket))
					if err != nil {
						return nil
					}
					if b != nil {
						b.Put([]byte(tc.key), []byte(tc.value))
					}
					return nil
				})
			}
			value, err := client.Get(tc.key, tc.bucket)
			assert.NoError(tt, err)
			assert.Equal(tt, tc.value, value)
		})
	}
	os.Remove(dbPath)
}

func TestGetAll(t *testing.T) {
	db, _ := bolt.Open(dbPath, 0600, &bolt.Options{Timeout: 1 * time.Second})
	client := &Client{
		DB:     db,
		Logger: &utils.DefaultLogger{},
	}
	const bucket = "some_bucket"
	const otherBucket = "some_other_bucket"

	testCases := []struct {
		name, bucket string
		expected     map[string]string
	}{
		{
			name:   "Get all values in an existing bucket",
			bucket: bucket,
			expected: map[string]string{
				"some_key":       "some_value",
				"some_other_key": "some_other_value",
			},
		},
		{
			name:     "Get empty values in an existing bucket",
			bucket:   otherBucket,
			expected: map[string]string{},
		},
		{
			name:     "Get all values in an empty bucket",
			bucket:   "",
			expected: map[string]string{},
		},
		{
			name:     "Get value for a key in a non-existing bucket",
			bucket:   "non_existing_bucket",
			expected: map[string]string{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(tt *testing.T) {
			if tc.bucket == bucket || tc.bucket == otherBucket {
				client.Update(func(tx *bolt.Tx) error {
					b, err := tx.CreateBucketIfNotExists([]byte(tc.bucket))
					if err != nil {
						return nil
					}
					if b != nil {
						for k, v := range tc.expected {
							b.Put([]byte(k), []byte(v))
						}
					}
					return nil
				})
			}
			actual, err := client.GetAll(tc.bucket)
			assert.NoError(tt, err)
			assert.Equal(tt, actual, tc.expected)
		})
	}
	os.Remove(dbPath)
}
