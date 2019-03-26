package parallel

import (
	"fmt"
	"github.com/onsi/gomega"
	"github.com/prometheus/common/log"
	"github.com/theckman/go-flock"
	"io/ioutil"
	"strconv"
	"time"
)

var (
	fileLock     = flock.NewFlock("/tmp/cassandra-operator-test.lock")
	resourceFile = "/tmp/cassandra-operator-test-resource.txt"
)

type ResourceSemaphore struct {
	Limit int
}

func NewResourceSemaphore(limit int) *ResourceSemaphore {
	resourceSemaphore := ResourceSemaphore{Limit: limit}
	resourceSemaphore.setResourceAvailable(limit)
	return &resourceSemaphore
}

func NewUnInitialisedResourceSemaphore(limit int) *ResourceSemaphore {
	return &ResourceSemaphore{Limit: limit}
}

func (r *ResourceSemaphore) AcquireResource(size int) {
	log.Infof("Attempting to acquire %d from resource semaphore", size)
	err := r.withLock(func() (bool, error) {
		available, err := r.resourceAvailable()
		if err != nil {
			return false, err
		}

		if available >= size {
			r.setResourceAvailable(available - size)
			log.Infof("Acquired %d from resource semaphore. Total available: %d", size, available-size)
			return true, nil
		}
		log.Debugf("Not enough resource available. Needed %d, found %d", size, available)
		return false, nil
	})

	gomega.Expect(err).ToNot(gomega.HaveOccurred())
}

func (r *ResourceSemaphore) ReleaseResource(size int) {
	log.Infof("Attempting to release %d from resource semaphore", size)
	r.withLock(func() (bool, error) {
		available, err := r.resourceAvailable()
		if err != nil {
			return false, err
		}

		available += size
		if available > r.Limit {
			available = r.Limit
		}

		r.setResourceAvailable(available)
		log.Infof("Released %d from resource semaphore. Total available: %d", size, available)
		return true, nil
	})
}

func (r *ResourceSemaphore) withLock(action func() (bool, error)) error {
	maxRetryCount := 360 // 30 minutes
	retry := 0
	for retry < maxRetryCount {
		log.Debugf("Attempt %d", retry)
		locked, err := fileLock.TryLock()
		if err != nil {
			return fmt.Errorf("failed to acquire resource lock, %v", err)
		}

		log.Debugf("Acquired resource lock: %v", locked)
		if locked {
			actionSuccess, err := action()
			fileLock.Unlock()

			if err != nil {
				return err
			}

			if actionSuccess {
				break
			}
		}

		time.Sleep(5 * time.Second)
		retry++
	}

	if retry == maxRetryCount {
		return fmt.Errorf("timed out while trying to obtain lock")
	}

	return nil
}

func (r *ResourceSemaphore) setResourceAvailable(size int) {
	err := ioutil.WriteFile(resourceFile, []byte(strconv.Itoa(size)), 0644)
	if err != nil {
		log.Fatalf("Failed to write to resource file, %v", err)
	}
}

func (r *ResourceSemaphore) resourceAvailable() (int, error) {
	content, err := ioutil.ReadFile(resourceFile)
	if err != nil {
		return 0, fmt.Errorf("error while reading resource file content, %v", err)
	}
	resourceAvailable, err := strconv.Atoi(string(content))
	if err != nil {
		return 0, fmt.Errorf("unable to convert resource content to int. Content was: %s, %v", content, err)
	}
	return resourceAvailable, nil
}
