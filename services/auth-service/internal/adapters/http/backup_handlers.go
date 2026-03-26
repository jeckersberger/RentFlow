package http

import (
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/pkg/common/middleware"
)

// BackupHandlers holds backup-related HTTP handlers
type BackupHandlers struct {
	logger    logger.Logger
	backupDir string
	mu        sync.Mutex
}

// BackupInfo represents metadata about a backup file
type BackupInfo struct {
	Filename  string `json:"filename"`
	Size      int64  `json:"size"`
	SizeHuman string `json:"size_human"`
	CreatedAt string `json:"created_at"`
	Type      string `json:"type"`
	Status    string `json:"status"`
}

// NewBackupHandlers creates a new backup handlers instance
func NewBackupHandlers(log logger.Logger) *BackupHandlers {
	backupDir := os.Getenv("BACKUP_DIR")
	if backupDir == "" {
		backupDir = "/backups"
	}
	// Ensure backup directory exists
	os.MkdirAll(backupDir, 0755)

	return &BackupHandlers{
		logger:    log,
		backupDir: backupDir,
	}
}

// allDatabases returns the list of all service databases to back up
func allDatabases() []string {
	return []string{
		"auth_service",
		"inventory_service",
		"project_service",
		"scanner_service",
		"warehouse_service",
		"invoice_service",
		"document_service",
		"crew_service",
		"federation_service",
		"maintenance_service",
		"transport_service",
		"insurance_service",
		"workflow_service",
		"ai_service",
		"notification_service",
		"reporting_service",
		"audit_service",
		"expense_service",
	}
}

// CreateBackup handles POST /api/v1/system/backup
func (h *BackupHandlers) CreateBackup(w http.ResponseWriter, r *http.Request) {
	if !middleware.HasRole(r.Context(), "admin") {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "Admin role required")
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	timestamp := time.Now().UTC().Format("2006-01-02_15-04-05")
	backupName := fmt.Sprintf("rentflow_backup_%s", timestamp)
	backupPath := filepath.Join(h.backupDir, backupName)

	// Create temp directory for individual database dumps
	if err := os.MkdirAll(backupPath, 0755); err != nil {
		h.logger.Error("failed to create backup directory", err)
		writeError(w, http.StatusInternalServerError, "BACKUP_ERROR", "Failed to create backup directory")
		return
	}

	dbHost := envOrDefault("DB_HOST", "postgres")
	dbPort := envOrDefault("DB_PORT", "5432")
	dbUser := envOrDefault("DB_USER", "rentflow")
	dbPass := envOrDefault("DB_PASSWORD", "rentflow_dev")

	databases := allDatabases()
	var failedDBs []string

	for _, dbName := range databases {
		dumpFile := filepath.Join(backupPath, dbName+".sql.gz")

		cmd := exec.Command("pg_dump",
			"-h", dbHost,
			"-p", dbPort,
			"-U", dbUser,
			"-d", dbName,
			"--no-owner",
			"--no-acl",
			"--format=plain",
		)
		cmd.Env = append(os.Environ(), "PGPASSWORD="+dbPass)

		stdout, err := cmd.StdoutPipe()
		if err != nil {
			h.logger.Error("failed to create stdout pipe for "+dbName, err)
			failedDBs = append(failedDBs, dbName)
			continue
		}

		outFile, err := os.Create(dumpFile)
		if err != nil {
			h.logger.Error("failed to create dump file for "+dbName, err)
			failedDBs = append(failedDBs, dbName)
			continue
		}

		gzWriter := gzip.NewWriter(outFile)

		if err := cmd.Start(); err != nil {
			h.logger.Error("failed to start pg_dump for "+dbName, err)
			gzWriter.Close()
			outFile.Close()
			failedDBs = append(failedDBs, dbName)
			continue
		}

		if _, err := io.Copy(gzWriter, stdout); err != nil {
			h.logger.Error("failed to write dump for "+dbName, err)
			failedDBs = append(failedDBs, dbName)
		}

		gzWriter.Close()
		outFile.Close()

		if err := cmd.Wait(); err != nil {
			h.logger.Error("pg_dump failed for "+dbName, err)
			failedDBs = append(failedDBs, dbName)
			os.Remove(dumpFile)
		}
	}

	// Create tar.gz archive of all dumps
	archiveName := backupName + ".tar.gz"
	archivePath := filepath.Join(h.backupDir, archiveName)

	tarCmd := exec.Command("tar", "-czf", archivePath, "-C", h.backupDir, backupName)
	if err := tarCmd.Run(); err != nil {
		h.logger.Error("failed to create backup archive", err)
		writeError(w, http.StatusInternalServerError, "BACKUP_ERROR", "Failed to create backup archive")
		// Clean up temp directory
		os.RemoveAll(backupPath)
		return
	}

	// Clean up temp directory, keep only the archive
	os.RemoveAll(backupPath)

	// Get archive info
	stat, err := os.Stat(archivePath)
	if err != nil {
		h.logger.Error("failed to stat backup archive", err)
		writeError(w, http.StatusInternalServerError, "BACKUP_ERROR", "Backup created but failed to read info")
		return
	}

	status := "Erfolgreich"
	if len(failedDBs) > 0 {
		status = fmt.Sprintf("Teilweise erfolgreich (%d/%d)", len(databases)-len(failedDBs), len(databases))
		h.logger.Warn("some databases failed to backup", "failed", strings.Join(failedDBs, ", "))
	}

	h.logger.Info("backup created", "file", archiveName, "size", stat.Size(), "status", status)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"data": BackupInfo{
			Filename:  archiveName,
			Size:      stat.Size(),
			SizeHuman: humanSize(stat.Size()),
			CreatedAt: stat.ModTime().Format(time.RFC3339),
			Type:      "Manuell",
			Status:    status,
		},
		"message": "Backup erfolgreich erstellt",
	})
}

// RunBackup creates a backup programmatically (called before updates)
func (h *BackupHandlers) RunBackup() error {
	h.mu.Lock()
	defer h.mu.Unlock()

	timestamp := time.Now().UTC().Format("2006-01-02_15-04-05")
	backupName := fmt.Sprintf("rentflow_pre_update_%s", timestamp)
	backupPath := filepath.Join(h.backupDir, backupName)

	if err := os.MkdirAll(backupPath, 0755); err != nil {
		return fmt.Errorf("failed to create backup directory: %w", err)
	}

	dbHost := envOrDefault("DB_HOST", "postgres")
	dbPort := envOrDefault("DB_PORT", "5432")
	dbUser := envOrDefault("DB_USER", "rentflow")
	dbPass := envOrDefault("DB_PASSWORD", "rentflow_dev")

	for _, dbName := range allDatabases() {
		dumpFile := filepath.Join(backupPath, dbName+".sql.gz")
		cmd := exec.Command("pg_dump", "-h", dbHost, "-p", dbPort, "-U", dbUser, "-d", dbName, "--no-owner", "--no-acl", "--format=plain")
		cmd.Env = append(os.Environ(), "PGPASSWORD="+dbPass)
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			continue
		}
		outFile, err := os.Create(dumpFile)
		if err != nil {
			continue
		}
		gzWriter := gzip.NewWriter(outFile)
		if err := cmd.Start(); err != nil {
			gzWriter.Close()
			outFile.Close()
			continue
		}
		io.Copy(gzWriter, stdout)
		gzWriter.Close()
		outFile.Close()
		if err := cmd.Wait(); err != nil {
			os.Remove(dumpFile)
		}
	}

	archivePath := filepath.Join(h.backupDir, backupName+".tar.gz")
	tarCmd := exec.Command("tar", "-czf", archivePath, "-C", h.backupDir, backupName)
	if err := tarCmd.Run(); err != nil {
		os.RemoveAll(backupPath)
		return fmt.Errorf("failed to create archive: %w", err)
	}
	os.RemoveAll(backupPath)

	h.logger.Info("pre-update backup created", "file", backupName+".tar.gz")
	return nil
}

// ListBackups handles GET /api/v1/system/backups
func (h *BackupHandlers) ListBackups(w http.ResponseWriter, r *http.Request) {
	if !middleware.HasRole(r.Context(), "admin") {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "Admin role required")
		return
	}

	entries, err := os.ReadDir(h.backupDir)
	if err != nil {
		h.logger.Error("failed to read backup directory", err)
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"data":    []BackupInfo{},
			"message": "Backups retrieved successfully",
		})
		return
	}

	var backups []BackupInfo
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".tar.gz") {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}

		backupType := "Manuell"
		if strings.Contains(name, "_auto_") {
			backupType = "Automatisch"
		}

		backups = append(backups, BackupInfo{
			Filename:  name,
			Size:      info.Size(),
			SizeHuman: humanSize(info.Size()),
			CreatedAt: info.ModTime().Format(time.RFC3339),
			Type:      backupType,
			Status:    "Erfolgreich",
		})
	}

	// Sort by creation time, newest first
	sort.Slice(backups, func(i, j int) bool {
		return backups[i].CreatedAt > backups[j].CreatedAt
	})

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"data":    backups,
		"message": "Backups retrieved successfully",
	})
}

// DownloadBackup handles GET /api/v1/system/backups/{filename}
func (h *BackupHandlers) DownloadBackup(w http.ResponseWriter, r *http.Request) {
	if !middleware.HasRole(r.Context(), "admin") {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "Admin role required")
		return
	}

	filename := r.PathValue("filename")
	if filename == "" {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Filename is required")
		return
	}

	// Sanitize filename to prevent path traversal
	filename = filepath.Base(filename)
	if !strings.HasSuffix(filename, ".tar.gz") {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid backup filename")
		return
	}

	filePath := filepath.Join(h.backupDir, filename)
	stat, err := os.Stat(filePath)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Backup file not found")
		return
	}

	w.Header().Set("Content-Type", "application/gzip")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
	w.Header().Set("Content-Length", fmt.Sprintf("%d", stat.Size()))

	file, err := os.Open(filePath)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to open backup file")
		return
	}
	defer file.Close()

	io.Copy(w, file)
}

// RestoreBackup handles POST /api/v1/system/backups/{filename}/restore
func (h *BackupHandlers) RestoreBackup(w http.ResponseWriter, r *http.Request) {
	if !middleware.HasRole(r.Context(), "admin") {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "Admin role required")
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	filename := r.PathValue("filename")
	if filename == "" {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Filename is required")
		return
	}

	// Sanitize filename
	filename = filepath.Base(filename)
	if !strings.HasSuffix(filename, ".tar.gz") {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid backup filename")
		return
	}

	archivePath := filepath.Join(h.backupDir, filename)
	if _, err := os.Stat(archivePath); err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Backup file not found")
		return
	}

	// Extract archive to temp directory
	restoreDir := filepath.Join(h.backupDir, "restore_tmp")
	os.RemoveAll(restoreDir)
	if err := os.MkdirAll(restoreDir, 0755); err != nil {
		writeError(w, http.StatusInternalServerError, "RESTORE_ERROR", "Failed to create restore directory")
		return
	}
	defer os.RemoveAll(restoreDir)

	tarCmd := exec.Command("tar", "-xzf", archivePath, "-C", restoreDir)
	if err := tarCmd.Run(); err != nil {
		h.logger.Error("failed to extract backup archive", err)
		writeError(w, http.StatusInternalServerError, "RESTORE_ERROR", "Failed to extract backup archive")
		return
	}

	// Find the extracted directory (should be the backup name)
	extractedEntries, err := os.ReadDir(restoreDir)
	if err != nil || len(extractedEntries) == 0 {
		writeError(w, http.StatusInternalServerError, "RESTORE_ERROR", "Invalid backup archive structure")
		return
	}

	extractedDir := filepath.Join(restoreDir, extractedEntries[0].Name())

	dbHost := envOrDefault("DB_HOST", "postgres")
	dbPort := envOrDefault("DB_PORT", "5432")
	dbUser := envOrDefault("DB_USER", "rentflow")
	dbPass := envOrDefault("DB_PASSWORD", "rentflow_dev")

	databases := allDatabases()
	var failedDBs []string
	var restoredDBs []string

	for _, dbName := range databases {
		dumpFile := filepath.Join(extractedDir, dbName+".sql.gz")
		if _, err := os.Stat(dumpFile); err != nil {
			// Skip databases that don't have a dump in this backup
			continue
		}

		// Open and decompress the dump file
		f, err := os.Open(dumpFile)
		if err != nil {
			h.logger.Error("failed to open dump file for "+dbName, err)
			failedDBs = append(failedDBs, dbName)
			continue
		}

		gzReader, err := gzip.NewReader(f)
		if err != nil {
			f.Close()
			h.logger.Error("failed to decompress dump for "+dbName, err)
			failedDBs = append(failedDBs, dbName)
			continue
		}

		// Drop and recreate the database, then restore
		// First, terminate existing connections
		terminateCmd := exec.Command("psql",
			"-h", dbHost,
			"-p", dbPort,
			"-U", dbUser,
			"-d", "postgres",
			"-c", fmt.Sprintf("SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname='%s' AND pid <> pg_backend_pid();", dbName),
		)
		terminateCmd.Env = append(os.Environ(), "PGPASSWORD="+dbPass)
		terminateCmd.Run() // Ignore errors

		// Drop database
		dropCmd := exec.Command("psql",
			"-h", dbHost,
			"-p", dbPort,
			"-U", dbUser,
			"-d", "postgres",
			"-c", fmt.Sprintf("DROP DATABASE IF EXISTS %s;", dbName),
		)
		dropCmd.Env = append(os.Environ(), "PGPASSWORD="+dbPass)
		if err := dropCmd.Run(); err != nil {
			gzReader.Close()
			f.Close()
			h.logger.Error("failed to drop database "+dbName, err)
			failedDBs = append(failedDBs, dbName)
			continue
		}

		// Create database
		createCmd := exec.Command("psql",
			"-h", dbHost,
			"-p", dbPort,
			"-U", dbUser,
			"-d", "postgres",
			"-c", fmt.Sprintf("CREATE DATABASE %s;", dbName),
		)
		createCmd.Env = append(os.Environ(), "PGPASSWORD="+dbPass)
		if err := createCmd.Run(); err != nil {
			gzReader.Close()
			f.Close()
			h.logger.Error("failed to create database "+dbName, err)
			failedDBs = append(failedDBs, dbName)
			continue
		}

		// Restore from dump
		restoreCmd := exec.Command("psql",
			"-h", dbHost,
			"-p", dbPort,
			"-U", dbUser,
			"-d", dbName,
		)
		restoreCmd.Env = append(os.Environ(), "PGPASSWORD="+dbPass)
		restoreCmd.Stdin = gzReader

		if err := restoreCmd.Run(); err != nil {
			gzReader.Close()
			f.Close()
			h.logger.Error("failed to restore database "+dbName, err)
			failedDBs = append(failedDBs, dbName)
			continue
		}

		gzReader.Close()
		f.Close()
		restoredDBs = append(restoredDBs, dbName)
	}

	if len(failedDBs) > 0 {
		h.logger.Warn("some databases failed to restore",
			"failed", strings.Join(failedDBs, ", "),
			"restored", strings.Join(restoredDBs, ", "),
		)
	}

	h.logger.Info("backup restored",
		"file", filename,
		"restored", len(restoredDBs),
		"failed", len(failedDBs),
	)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"data": map[string]interface{}{
			"restored_databases": restoredDBs,
			"failed_databases":   failedDBs,
			"success":            len(failedDBs) == 0,
		},
		"message": "Backup wiederhergestellt",
	})
}

// UploadBackup handles POST /api/v1/system/backups/upload
func (h *BackupHandlers) UploadBackup(w http.ResponseWriter, r *http.Request) {
	if !middleware.HasRole(r.Context(), "admin") {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "Admin role required")
		return
	}

	// Limit upload to 2GB
	r.Body = http.MaxBytesReader(w, r.Body, 2<<30)

	if err := r.ParseMultipartForm(256 << 20); err != nil {
		writeError(w, http.StatusBadRequest, "UPLOAD_ERROR", "Failed to parse upload (max 2GB)")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "UPLOAD_ERROR", "No file provided")
		return
	}
	defer file.Close()

	// Validate file extension
	if !strings.HasSuffix(header.Filename, ".tar.gz") && !strings.HasSuffix(header.Filename, ".tgz") {
		writeError(w, http.StatusBadRequest, "UPLOAD_ERROR", "Only .tar.gz files are accepted")
		return
	}

	// Save to backup directory
	destName := filepath.Base(header.Filename)
	destPath := filepath.Join(h.backupDir, destName)

	// If file already exists, add timestamp
	if _, err := os.Stat(destPath); err == nil {
		ext := ".tar.gz"
		base := strings.TrimSuffix(destName, ext)
		if strings.HasSuffix(destName, ".tgz") {
			ext = ".tgz"
			base = strings.TrimSuffix(destName, ext)
		}
		destName = fmt.Sprintf("%s_%s%s", base, time.Now().UTC().Format("150405"), ext)
		destPath = filepath.Join(h.backupDir, destName)
	}

	dest, err := os.Create(destPath)
	if err != nil {
		h.logger.Error("failed to create upload file", err)
		writeError(w, http.StatusInternalServerError, "UPLOAD_ERROR", "Failed to save upload")
		return
	}
	defer dest.Close()

	written, err := io.Copy(dest, file)
	if err != nil {
		h.logger.Error("failed to write upload file", err)
		os.Remove(destPath)
		writeError(w, http.StatusInternalServerError, "UPLOAD_ERROR", "Failed to save upload")
		return
	}

	h.logger.Info("backup uploaded", "file", destName, "size", written)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"data": BackupInfo{
			Filename:  destName,
			Size:      written,
			SizeHuman: humanSize(written),
			CreatedAt: time.Now().UTC().Format(time.RFC3339),
			Type:      "Importiert",
			Status:    "Erfolgreich",
		},
		"message": "Backup erfolgreich importiert",
	})
}

// DeleteBackup handles DELETE /api/v1/system/backups/{filename}
func (h *BackupHandlers) DeleteBackup(w http.ResponseWriter, r *http.Request) {
	if !middleware.HasRole(r.Context(), "admin") {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "Admin role required")
		return
	}

	filename := r.PathValue("filename")
	if filename == "" {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Filename is required")
		return
	}

	filename = filepath.Base(filename)
	if !strings.HasSuffix(filename, ".tar.gz") {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid backup filename")
		return
	}

	filePath := filepath.Join(h.backupDir, filename)
	if _, err := os.Stat(filePath); err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Backup file not found")
		return
	}

	if err := os.Remove(filePath); err != nil {
		h.logger.Error("failed to delete backup", err)
		writeError(w, http.StatusInternalServerError, "DELETE_ERROR", "Failed to delete backup")
		return
	}

	h.logger.Info("backup deleted", "file", filename)
	w.WriteHeader(http.StatusNoContent)
}

// envOrDefault returns the environment variable value or a default
func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// humanSize formats bytes into human-readable size
func humanSize(bytes int64) string {
	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
	)
	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.1f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.1f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.1f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}
