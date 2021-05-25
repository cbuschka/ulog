package journal

import (
	"encoding/binary"
	"encoding/json"
	bolt "go.etcd.io/bbolt"
)

const bucketName = "MessagesV1"

type Journal struct {
	db *bolt.DB
}

func NewJournal() (*Journal, error) {
	db, err := bolt.Open("./tmp/my.db", 0640, nil)
	if err != nil {
		return nil, err
	}

	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(bucketName))
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &Journal{db: db}, nil
}

func (journal *Journal) ListMessages(begin int, limit int, callback func(message *Message) (bool, error)) error {

	return journal.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bucketName))
		cursor := bucket.Cursor()
		count := 0
		var key, value []byte
		if begin <= 0 {
			key, value = cursor.First()
		} else {
			key, value = cursor.Seek(itob(uint64(begin)))
		}
		for ; key != nil; key, value = cursor.Next() {
			if limit != -1 && count >= limit {
				break
			}
			var message Message
			err := json.Unmarshal(value, &message)
			if err != nil {
				return err
			}
			goon, err := callback(&message)
			if err != nil {
				return err
			}
			if !goon {
				return nil
			}
			count = count + 1
		}
		return nil
	})
}

func (journal *Journal) WriteMessage(message *Message) error {

	return journal.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bucketName))

		id, _ := bucket.NextSequence()
		message.Id = id

		buf, err := json.Marshal(message)
		if err != nil {
			return err
		}

		err = bucket.Put(itob(id), buf)
		if err != nil {
			return err
		}

		return nil
	})
}

func (journal *Journal) Close() error {
	return journal.db.Close()
}

func itob(v uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, v)
	return b
}

func btoi(b []byte) uint64 {
	v := binary.BigEndian.Uint64(b)
	return v
}
