package locks

import (
	"errors"
	"fmt"
	"pos-proxy/db"

	lock "github.com/bsm/redis-lock"
)

func LockFDM(productionNumber string) (*lock.Lock, error) {
	l, err := lock.ObtainLock(db.Redis, fmt.Sprintf("fdm_%s", productionNumber), nil)
	if err != nil {
		return &lock.Lock{}, err
	} else if l == nil {
		return &lock.Lock{}, errors.New("couldn't obtain fdm lock")
	}

	ok, err := l.Lock()
	if err != nil {
		return &lock.Lock{}, err
	} else if !ok {
		return &lock.Lock{}, errors.New("failed to acquire fdm lock")
	}

	return l, nil
}
