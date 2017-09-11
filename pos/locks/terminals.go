package locks

import (
	"errors"
	"fmt"
	"pos-proxy/db"
	"strconv"

	lock "github.com/bsm/redis-lock"
	"github.com/go-redis/redis"
)

// LockTerminal creates a key for this terminal in Redis,
// and lock this terminal for a certain cashier.
// If Terminal is already locked, returns the id of the cashier
// currently locking the terminal.
func LockTerminal(terminalID int, cashierID int) (int, error) {
	lockName := fmt.Sprintf("terminals_lock")
	/*lockOpts := &lock.LockOptions{
		WaitTimeout: 3 * time.Second,
	}*/
	lock, err := lock.ObtainLock(db.Redis, lockName, nil)
	if err != nil {
		return 0, err
	} else if lock == nil {
		return 0, errors.New("couldn't obtain terminal lock")
	}

	ok, err := lock.Lock()
	if err != nil {
		return 0, err
	} else if !ok {
		return 0, errors.New("Failed to acquire lock")
	}
	defer lock.Lock()

	key := fmt.Sprintf("terminal_%d", terminalID)
	otherCashier := 0
	err = db.Redis.Watch(func(tx *redis.Tx) error {
		_, err = tx.Pipelined(func(pipe redis.Pipeliner) error {
			val, err := tx.Get(key).Result()
			if err != nil && err != redis.Nil {
				return err
			}
			if val == "" {
				pipe.Set(key, cashierID, 0)
				return nil
			}
			n, err := strconv.Atoi(val)
			if err != nil {
				return err
			} else if n != cashierID {
				otherCashier = n
				return errors.New("invoice key already exists")
			}
			return nil
		})
		if err != nil && err != redis.Nil {
			return err
		}
		return nil
	}, key)

	if err != nil && err != redis.Nil {
		return otherCashier, err
	}
	return otherCashier, nil
}

// UnlockTerminal deletes terminal key from redis
// and makes this terminal available for other cashiers
func UnlockTerminal(terminalID int) {
	key := fmt.Sprintf("terminal_%d", terminalID)
	db.Redis.Del(key)
}
