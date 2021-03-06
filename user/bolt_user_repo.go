package user

import (
	"encoding/json"
	"errors"

	"github.com/boltdb/bolt"
)

func NewBoltRepo(b *bolt.DB) *BoltUserRepo {
	return &BoltUserRepo{b}
}

type BoltUserRepo struct {
	*bolt.DB
}

var bucket = []byte("users")

func (r *BoltUserRepo) UserForID(id string) (*User, error) {
	var u *User
	err := r.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(bucket)
		if bucket == nil {
			return nil
		}
		su := bucket.Get([]byte(id))
		if len(su) == 0 {
			return nil
		}
		var err error
		u, err = r.deserializeUser(su)
		return err

	})
	return u, err
}

func (r *BoltUserRepo) ListUsers() ([]User, error) {
	users := make([]User, 0)
	err := r.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(bucket)
		if bucket == nil {
			return nil
		}
		return bucket.ForEach(func(k []byte, v []byte) error {
			u, err := r.deserializeUser(v)
			if err != nil {
				return err
			}
			users = append(users, *u)
			return nil
		})
	})
	return users, err
}

func (r *BoltUserRepo) SaveUser(u User) error {
	if len(u.ID) == 0 {
		return errors.New("user id must be set to save user")
	}
	su, err := r.serializeUser(u)
	if err != nil {
		return err
	}
	return r.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists(bucket)
		if err != nil {
			return err
		}
		return bucket.Put([]byte(u.ID), su)
	})
}

func (r *BoltUserRepo) serializeUser(u User) ([]byte, error) {
	return json.Marshal(u)
}

func (r *BoltUserRepo) deserializeUser(b []byte) (*User, error) {
	var u User
	if err := json.Unmarshal(b, &u); err != nil {
		return nil, err
	}
	return &u, nil
}
