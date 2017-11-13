package locks

import (
	"errors"
	"fmt"
	"pos-proxy/db"
	"time"

	lock "github.com/bsm/redis-lock"
)

// LockFDM creates a lock on the FDM connection to prevent race conditions
func LockFDM(productionNumber string) (*lock.Locker, error) {
	lockOptions := &lock.Options{
		WaitTimeout: 4 * time.Second,
	}

	l, err := lock.ObtainLock(db.Redis, fmt.Sprintf("fdm_%s", productionNumber), lockOptions)
	if err != nil {
		return &lock.Locker{}, err
	} else if l == nil {
		return &lock.Locker{}, errors.New("couldn't obtain fdm lock")
	}

	ok, err := l.Lock()
	if err != nil {
		return &lock.Locker{}, err
	} else if !ok {
		return &lock.Locker{}, errors.New("failed to acquire fdm lock")
	}

	return l, nil
}
