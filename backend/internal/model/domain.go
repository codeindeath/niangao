package model

import (
	"context"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ============================================================
// 动态领域体系 — DB 驱动, 10×10 上限
// ============================================================

type Domain string
type SubDomain string

// DomainEntry is a row from the domains table.
type DomainEntry struct {
	Name        string  `json:"name"`
	DisplayName string  `json:"display_name"`
	ParentName  *string `json:"parent_name"`
	SortOrder   int     `json:"sort_order"`
	Active      bool    `json:"active"`
}

// DomainCatalog holds the loaded domain tree and provides lookup methods.
type DomainCatalog struct {
	mu            sync.RWMutex
	domains       map[string]DomainEntry   // name → entry
	subByParent   map[string][]DomainEntry // parent_name → children
	parents       []DomainEntry            // top-level, sorted
	displayByName map[string]string        // name → display_name
}

var globalCatalog *DomainCatalog

// InitDomainCatalog loads domains from DB and sets the global catalog.
func InitDomainCatalog(ctx context.Context, db *pgxpool.Pool) error {
	rows, err := db.Query(ctx,
		`SELECT name, display_name, parent_name, sort_order, active
		 FROM domains WHERE active=true ORDER BY sort_order`)
	if err != nil {
		return err
	}
	defer rows.Close()

	c := &DomainCatalog{
		domains:       make(map[string]DomainEntry),
		subByParent:   make(map[string][]DomainEntry),
		displayByName: make(map[string]string),
	}

	for rows.Next() {
		var e DomainEntry
		if err := rows.Scan(&e.Name, &e.DisplayName, &e.ParentName, &e.SortOrder, &e.Active); err != nil {
			return err
		}
		c.domains[e.Name] = e
		c.displayByName[e.Name] = e.DisplayName

		if e.ParentName == nil {
			c.parents = append(c.parents, e)
		} else {
			c.subByParent[*e.ParentName] = append(c.subByParent[*e.ParentName], e)
		}
	}

	globalCatalog = c
	return nil
}

// GetCatalog returns the global domain catalog (nil if not initialized).
func GetCatalog() *DomainCatalog { return globalCatalog }

// --- Read methods ---

func (c *DomainCatalog) IsValidDomain(name string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	e, ok := c.domains[name]
	return ok && e.ParentName == nil
}

func (c *DomainCatalog) IsValidSubDomain(name string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	e, ok := c.domains[name]
	return ok && e.ParentName != nil
}

func (c *DomainCatalog) SubDomainBelongsToParent(parent, child string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	e, ok := c.domains[child]
	if !ok || e.ParentName == nil {
		return false
	}
	return *e.ParentName == parent
}

func (c *DomainCatalog) DisplayName(name string) string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if d, ok := c.displayByName[name]; ok {
		return d
	}
	return name
}

func (c *DomainCatalog) Parents() []DomainEntry {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.parents
}

func (c *DomainCatalog) SubDomains(parentName string) []DomainEntry {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.subByParent[parentName]
}

func (c *DomainCatalog) AllNames() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	names := make([]string, 0, len(c.domains))
	for n := range c.domains {
		names = append(names, n)
	}
	return names
}

// --- Write methods for auto-creation ---

// CanCreateSubDomain checks if parent has room for a new subdomain (max 10).
func (c *DomainCatalog) CanCreateSubDomain(parentName string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.subByParent[parentName]) < 10
}

// CanCreateParent checks if there's room for a new top-level domain (max 10).
func (c *DomainCatalog) CanCreateParent() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.parents) < 10
}

// CreateSubDomain inserts a new subdomain into DB and updates catalog.
func (c *DomainCatalog) CreateSubDomain(ctx context.Context, db *pgxpool.Pool, name, displayName, parentName string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	nextSort := len(c.subByParent[parentName]) + 1
	_, err := db.Exec(ctx,
		`INSERT INTO domains (name, display_name, parent_name, sort_order)
		 VALUES ($1,$2,$3,$4)
		 ON CONFLICT (name) DO UPDATE SET active=true, display_name=$2`,
		name, displayName, parentName, nextSort)
	if err != nil {
		return err
	}

	e := DomainEntry{Name: name, DisplayName: displayName, ParentName: &parentName, SortOrder: nextSort, Active: true}
	c.domains[name] = e
	c.displayByName[name] = displayName
	c.subByParent[parentName] = append(c.subByParent[parentName], e)
	return nil
}

// Reload refreshes the catalog from DB.
func (c *DomainCatalog) Reload(ctx context.Context, db *pgxpool.Pool) error {
	return InitDomainCatalog(ctx, db)
}

// ============================================================
// Compatibility layer — delegates to global catalog
// ============================================================

func IsValidDomain(d Domain) bool {
	cat := GetCatalog()
	if cat == nil {
		return false
	}
	return cat.IsValidDomain(string(d))
}

func IsValidSubDomain(d SubDomain) bool {
	cat := GetCatalog()
	if cat == nil {
		return false
	}
	return cat.IsValidSubDomain(string(d))
}

func SubDomainBelongsToParent(parent Domain, child SubDomain) bool {
	cat := GetCatalog()
	if cat == nil {
		return false
	}
	return cat.SubDomainBelongsToParent(string(parent), string(child))
}

// DomainDisplay returns Chinese display name for a domain/subdomain name.
func DomainDisplay(name string) string {
	cat := GetCatalog()
	if cat == nil {
		return name
	}
	return cat.DisplayName(name)
}
