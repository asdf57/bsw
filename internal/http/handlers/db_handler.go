package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
)

// GetDbHealth godoc
// @Summary Health check
// @Tags health
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]string
// @Router /api/v1/health [get]
func (h *Handlers) GetDbHealth(c *gin.Context) {
	// Run simple query to verify connectivity
	var result int
	if err := h.Db.DB.Raw("SELECT 1").Scan(&result).Error; err != nil {
		log.Printf("health check failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok", "value": result})
}

// BackupDB godoc
// @Summary      Trigger a database backup
// @Description  Starts a Postgres backup job (pg_dump)
// @Tags         admin
// @Produce      json
// @Success      202  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /api/v1/admin/backup [post]
func (h *Handlers) BackupDB(c *gin.Context) {
	filename := fmt.Sprintf("backup_%s.dump", time.Now().UTC().Format("20060102T150405Z"))
	path := filepath.Join("/backups", filename)

	if err := os.MkdirAll("/backups", os.ModePerm); err != nil {
		log.Printf("failed to create backup directory: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create backup directory"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Minute)
	defer cancel()

	cmd := exec.CommandContext(ctx,
		"pg_dump",
		"-Fc",
		"-h", os.Getenv("PGHOST"),
		"-p", os.Getenv("PGPORT"),
		"-U", os.Getenv("PGUSER"),
		"-d", os.Getenv("PGDATABASE"),
		"-f", path,
	)
	cmd.Env = append(os.Environ(), "PGPASSWORD="+os.Getenv("PGPASSWORD"))

	out, err := cmd.CombinedOutput()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to run backup command", "output": string(out), "errorInfo": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"backup": filename})
}

// GetBackup godoc
// @Summary      Download a database backup
// @Description  Downloads a Postgres backup file
// @Tags         admin
// @Produce      application/octet-stream
// @Param        filename path string true "Backup filename"
// @Success      200  {file}  string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /api/v1/admin/backup/{filename} [get]
func (h *Handlers) GetBackup(c *gin.Context) {
	filename := c.Param("filename")

	fmt.Printf("Exporting backup for %s", filename)

	path := filepath.Join("/backups", filename)

	if _, err := os.Stat(path); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "backup not found"})
		return
	}

	c.FileAttachment(path, filename)
}
