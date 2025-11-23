package cache

import (
	"L0/internal/interfaces"
	"L0/internal/models"
	"sync"
	"time"
)

var _ interfaces.OrderCache = (*Cache)(nil)

type Cache struct {
	mu              sync.RWMutex
	orders          map[string]*cacheEntry
	maxSize         int
	defaultTTL      time.Duration
	cleanupInterval time.Duration
	stopCleanup     chan struct{}
}

type cacheEntry struct {
	order     *models.Order
	createdAt time.Time
	expiresAt time.Time
}

func New() *Cache {
	c := &Cache{
		orders:          make(map[string]*cacheEntry),
		maxSize:         1000,
		defaultTTL:      24 * time.Hour,
		cleanupInterval: 1 * time.Hour,
		stopCleanup:     make(chan struct{}),
	}

	go c.startCleanup()
	return c
}

func (c *Cache) Set(order *models.Order) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.orders) >= c.maxSize {
		c.evictOldest()
	}

	c.orders[order.OrderUID] = &cacheEntry{
		order:     order,
		createdAt: time.Now(),
		expiresAt: time.Now().Add(c.defaultTTL),
	}
}

func (c *Cache) Get(orderUID string) (*models.Order, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.orders[orderUID]
	if !exists {
		return nil, false
	}

	if time.Now().After(entry.expiresAt) {
		return nil, false
	}

	return entry.order, true
}

func (c *Cache) GetAll() map[string]*models.Order {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make(map[string]*models.Order)
	for key, entry := range c.orders {
		if time.Now().Before(entry.expiresAt) {
			result[key] = entry.order
		}
	}
	return result
}

func (c *Cache) LoadFromRepository(getAllOrders func() ([]*models.Order, error)) error {
	orders, err := getAllOrders()
	if err != nil {
		return err
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.orders = make(map[string]*cacheEntry)

	now := time.Now()
	for _, order := range orders {
		if len(c.orders) >= c.maxSize {
			break
		}

		c.orders[order.OrderUID] = &cacheEntry{
			order:     order,
			createdAt: now,
			expiresAt: now.Add(c.defaultTTL),
		}
	}

	return nil
}

func (c *Cache) Count() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	count := 0
	for _, entry := range c.orders {
		if time.Now().Before(entry.expiresAt) {
			count++
		}
	}
	return count
}

func (c *Cache) Stop() {
	close(c.stopCleanup)
}

func (c *Cache) evictOldest() {
	var oldestKey string
	var oldestTime time.Time

	for key, entry := range c.orders {
		if oldestKey == "" || entry.createdAt.Before(oldestTime) {
			oldestKey = key
			oldestTime = entry.createdAt
		}
	}

	if oldestKey != "" {
		delete(c.orders, oldestKey)
	}
}

func (c *Cache) startCleanup() {
	ticker := time.NewTicker(c.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.cleanupExpired()
		case <-c.stopCleanup:
			return
		}
	}
}

func (c *Cache) cleanupExpired() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for key, entry := range c.orders {
		if now.After(entry.expiresAt) {
			delete(c.orders, key)
		}
	}
}
