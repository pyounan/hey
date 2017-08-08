package locks

import (
	"errors"
	"fmt"
	lock "github.com/bsm/redis-lock"
	"github.com/go-redis/redis"
)

func LockTerminal(terminal string) error {
	client := redis.NewClient(&redis.Options{
		Network: "tcp",
		Addr:    "127.0.0.1:6379",
	})
	defer client.Close()
	/*lockOpts := &lock.LockOptions{
		WaitTimeout: 3 * time.Second,
	}*/
	lock, err := lock.ObtainLock(client, fmt.Sprintf("terminal_%s", terminal), nil)
	if err != nil {
		return err
	} else if lock == nil {
		return errors.New("Couldn't obtain terminal lock.")
	}

	ok, err := lock.Lock()
	if err != nil {
		return err
	} else if !ok {
		return errors.New("Failed to acquire lock")
	}

	return nil
}

func UnlockTerminal(terminal string) error {
	client := redis.NewClient(&redis.Options{
		Network: "tcp",
		Addr:    "127.0.0.1:6379",
	})
	defer client.Close()

	lock, err := lock.ObtainLock(client, fmt.Sprintf("terminal_%s", terminal), nil)
	if err != nil {
		return err
	} else if lock == nil {
		return errors.New("Couldn't obtain terminal lock.")
	}

	err = lock.Unlock()
	if err != nil {
		return err
	}

	return nil
}
