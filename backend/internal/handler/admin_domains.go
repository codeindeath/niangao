package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/niangao/backend/internal/model"
)

// domainIcons maps each top-level domain to an icon identifier.
var domainIcons = map[model.Domain]string{
	model.DomainCareer:       "briefcase",
	model.DomainRelationship: "heart",
	model.DomainCognition:    "brain",
	model.DomainLife:         "home",
	model.DomainEmotion:      "smile",
}

// domainHierarchyItem represents a top-level domain with its sub-domains
// in the hierarchy endpoint response.
type domainHierarchyItem struct {
	Name        string              `json:"name"`
	DisplayName string              `json:"display_name"`
	Parent      *string             `json:"parent"`
	Icon        string              `json:"icon"`
	SubDomains  []domainSubItem     `json:"sub_domains"`
}

type domainSubItem struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
}

// RegisterAdminDomainRoutes registers admin domain routes on the given
// admin group (which should already have RequireAdmin middleware applied).
func RegisterAdminDomainRoutes(admin *gin.RouterGroup, db *pgxpool.Pool) {
	admin.GET("/domains", func(c *gin.Context) {
		getDomainHierarchy(c)
	})
	admin.GET("/domains/stats", func(c *gin.Context) {
		getDomainStats(c, db)
	})
	admin.POST("/domains", func(c *gin.Context) {
		createDomain(c, db)
	})
	admin.PUT("/domains/:name", func(c *gin.Context) {
		updateDomain(c, db)
	})
	admin.POST("/domains/:name/sub", func(c *gin.Context) {
		addSubDomain(c, db)
	})
	admin.PUT("/domains/reorder", func(c *gin.Context) {
		reorderDomains(c, db)
	})
	admin.POST("/domains/:name/disable", func(c *gin.Context) {
		toggleDomain(c, db, false)
	})
	admin.POST("/domains/:name/enable", func(c *gin.Context) {
		toggleDomain(c, db, true)
	})
}

// ============================================================
// GET /admin/domains — Return full domain hierarchy
// ============================================================

func getDomainHierarchy(c *gin.Context) {
	var domains []domainHierarchyItem

	for _, d := range model.DomainSortOrder {
		displayName := model.ValidDomains[d]
		if displayName == "" {
			displayName = string(d)
		}
		icon := domainIcons[d]
		if icon == "" {
			icon = "default"
		}

		subs := model.SubDomainsByParent[d]
		subItems := make([]domainSubItem, 0, len(subs))
		for _, s := range subs {
			subDisplay := model.ValidSubDomains[s]
			if subDisplay == "" {
				subDisplay = string(s)
			}
			subItems = append(subItems, domainSubItem{
				Name:        string(s),
				DisplayName: subDisplay,
			})
		}

		domains = append(domains, domainHierarchyItem{
			Name:        string(d),
			DisplayName: displayName,
			Parent:      nil,
			Icon:        icon,
			SubDomains:  subItems,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"domains": domains,
	})
}

// ============================================================
// GET /admin/domains/stats — Domain experience counts
// ============================================================

type domainStat struct {
	Domain string `json:"domain"`
	Count  int64  `json:"count"`
}

func getDomainStats(c *gin.Context, db *pgxpool.Pool) {
	ctx := c.Request.Context()

	rows, err := db.Query(ctx, `SELECT domain, COUNT(*) FROM experiences GROUP BY domain`)
	if err != nil {
		log.Printf("getDomainStats error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取领域统计失败"})
		return
	}
	defer rows.Close()

	var stats []domainStat
	for rows.Next() {
		var s domainStat
		if err := rows.Scan(&s.Domain, &s.Count); err != nil {
			continue
		}
		stats = append(stats, s)
	}

	if stats == nil {
		stats = []domainStat{}
	}

	c.JSON(http.StatusOK, gin.H{
		"stats": stats,
	})
}

// ============================================================
// PUT /admin/domains/:name — Update domain display name / icon
// ============================================================

type updateDomainRequest struct {
	DisplayName string `json:"display_name"`
	Icon        string `json:"icon"`
}

func updateDomain(c *gin.Context, db *pgxpool.Pool) {
	name := c.Param("name")
	var req updateDomainRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求格式错误"})
		return
	}
	// Domains are currently hardcoded in model.ValidDomains
	// For now, this endpoint updates metadata stored in system_config
	key := "domain_display_" + name
	valueJSON, _ := json.Marshal(map[string]string{
		"display_name": req.DisplayName,
		"icon":         req.Icon,
	})
	_, err := db.Exec(c.Request.Context(),
		`INSERT INTO system_config (key, value) VALUES ($1, $2)
		 ON CONFLICT (key) DO UPDATE SET value=$2, updated_at=NOW()`,
		key, valueJSON,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("update domain: %v", err)})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// ============================================================
// POST /admin/domains/:name/disable|enable — Toggle domain
// ============================================================

func toggleDomain(c *gin.Context, db *pgxpool.Pool, active bool) {
	name := c.Param("name")
	status := "disabled"
	if active {
		status = "enabled"
	}
	key := "domain_active_" + name
	valueJSON, _ := json.Marshal(active)
	_, err := db.Exec(c.Request.Context(),
		`INSERT INTO system_config (key, value) VALUES ($1, $2)
		 ON CONFLICT (key) DO UPDATE SET value=$2, updated_at=NOW()`,
		key, valueJSON,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("%s domain: %v", status, err)})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok", "active": active})
}

// ============================================================
// POST /admin/domains — Create new top-level domain
// ============================================================

type createDomainRequest struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Icon        string `json:"icon"`
}

func createDomain(c *gin.Context, db *pgxpool.Pool) {
	var req createDomainRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请提供领域名称"})
		return
	}
	key := "domain_display_" + req.Name
	valueJSON, _ := json.Marshal(map[string]string{
		"display_name": req.DisplayName,
		"icon":         req.Icon,
	})
	_, err := db.Exec(c.Request.Context(),
		`INSERT INTO system_config (key, value) VALUES ($1, $2)
		 ON CONFLICT (key) DO UPDATE SET value=$2, updated_at=NOW()`,
		key, valueJSON,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("create domain: %v", err)})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"status": "created", "name": req.Name})
}

// ============================================================
// POST /admin/domains/:name/sub — Add sub-domain to parent
// ============================================================

type addSubDomainRequest struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
}

func addSubDomain(c *gin.Context, db *pgxpool.Pool) {
	parentName := c.Param("name")
	var req addSubDomainRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请提供子领域名称"})
		return
	}
	key := "sub_domains_" + parentName

	// Load existing sub-domains
	var existing []string
	var rawVal []byte
	err := db.QueryRow(c.Request.Context(),
		`SELECT value FROM system_config WHERE key=$1`, key,
	).Scan(&rawVal)
	if err == nil {
		json.Unmarshal(rawVal, &existing)
	}
	existing = append(existing, req.Name)
	valueJSON, _ := json.Marshal(existing)
	_, err = db.Exec(c.Request.Context(),
		`INSERT INTO system_config (key, value) VALUES ($1, $2)
		 ON CONFLICT (key) DO UPDATE SET value=$2, updated_at=NOW()`,
		key, valueJSON,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("add sub-domain: %v", err)})
		return
	}

	// Also store display name
	displayKey := "sub_domain_display_" + parentName + "_" + req.Name
	displayJSON, _ := json.Marshal(map[string]string{"display_name": req.DisplayName})
	db.Exec(c.Request.Context(),
		`INSERT INTO system_config (key, value) VALUES ($1, $2)
		 ON CONFLICT (key) DO UPDATE SET value=$2, updated_at=NOW()`,
		displayKey, displayJSON,
	)

	c.JSON(http.StatusCreated, gin.H{"status": "created", "parent": parentName, "sub": req.Name})
}

// ============================================================
// PUT /admin/domains/reorder — Reorder domains
// ============================================================

type reorderDomainsRequest struct {
	ParentName string   `json:"parent_name"`
	Names      []string `json:"names"`
}

func reorderDomains(c *gin.Context, db *pgxpool.Pool) {
	var req reorderDomainsRequest
	if err := c.ShouldBindJSON(&req); err != nil || len(req.Names) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请提供 names"})
		return
	}
	parentKey := req.ParentName
	if parentKey == "" {
		parentKey = "root"
	}
	key := "domain_order_" + parentKey
	valueJSON, _ := json.Marshal(req.Names)
	_, err := db.Exec(c.Request.Context(),
		`INSERT INTO system_config (key, value) VALUES ($1, $2)
		 ON CONFLICT (key) DO UPDATE SET value=$2, updated_at=NOW()`,
		key, valueJSON,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("reorder domains: %v", err)})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
