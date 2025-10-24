package main

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/tidwall/buntdb"
)

type Cache struct {
	db      *buntdb.DB
	onerror func(string, string, error)
}

func NewCache(datafile string, onerror func(string, string, error)) (*Cache, error) {
	db, err := buntdb.Open(datafile)
	if err != nil {
		return nil, err
	}
	return &Cache{
		db:      db,
		onerror: onerror,
	}, nil
}

func (c *Cache) Close() {
	c.db.Close()
}

func (c *Cache) GetString(ctx context.Context, key string) (string, bool) {
	var out string
	if err := c.db.View(func(tx *buntdb.Tx) error {
		var err error
		out, err = tx.Get(key)
		return err
	}); err != nil {
		if !errors.Is(err, buntdb.ErrNotFound) {
			c.onerror("Get", key, err)
		}
		return "", false
	}
	return out, true
}

func (c *Cache) Get(ctx context.Context, key string, obj any) bool {
	v, ok := c.GetString(ctx, key)
	if !ok {
		return false
	}
	if err := json.Unmarshal([]byte(v), obj); err != nil {
		c.onerror("Unmarshal", key, err)
		return false
	}
	return true
}

func (c *Cache) SetString(ctx context.Context, key string, value string) {
	if err := c.db.Update(func(tx *buntdb.Tx) error {
		_, _, err := tx.Set(key, value, nil)
		return err
	}); err != nil {
		c.onerror("Set", key, err)
	}
}

func (c *Cache) Set(ctx context.Context, key string, obj any) {
	v, err := json.Marshal(obj)
	if err != nil {
		c.onerror("Marshal", key, err)
		return
	}
	c.SetString(ctx, key, string(v))
}

func (c *Cache) Delete(ctx context.Context, key string) {
	if err := c.db.Update(func(tx *buntdb.Tx) error {
		_, err := tx.Delete(key)
		return err
	}); err != nil {
		c.onerror("Delete", key, err)
	}
}
