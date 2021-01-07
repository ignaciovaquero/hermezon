package main

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
