package migrate

import (
	"database/sql"
	"embed"
	"fmt"
	"log"
	"sort"
	"strings"

	"gorm.io/gorm"
)

//go:embed *.sql
var sqlMigrations embed.FS

// AutoMigrate runs SQL migrations from embedded files
func AutoMigrate(db *gorm.DB) error {
	log.Println("Running SQL migrations...")

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get database connection: %w", err)
	}

	db.Exec(`
			CREATE TABLE IF NOT EXISTS schema_migrations (
				version VARCHAR(255) PRIMARY KEY,
				applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			)
		`)

	var applied []string
	db.Raw("SELECT version FROM schema_migrations ORDER BY version").Scan(&applied)
	appliedSet := make(map[string]bool)
	for _, v := range applied {
		appliedSet[v] = true
	}

	files, err := sqlMigrations.ReadDir(".")
	if err != nil {
		return fmt.Errorf("failed to read migrations: %w", err)
	}

	var fileNames []string
	for _, f := range files {
		if !f.IsDir() && len(f.Name()) > 4 && f.Name()[len(f.Name())-4:] == ".sql" {
			fileNames = append(fileNames, f.Name())
		}
	}
	sort.Strings(fileNames)

	for _, fileName := range fileNames {
		version := fileName[:len(fileName)-4]
		if appliedSet[version] {
			log.Printf("  Skipping %s (already applied)", version)
			continue
		}

		log.Printf("  Applying %s...", version)

		content, err := sqlMigrations.ReadFile(fileName)
		if err != nil {
			return fmt.Errorf("failed to read migration %s: %w", fileName, err)
		}

		if err := executeMigration(sqlDB, string(content)); err != nil {
			return fmt.Errorf("migration %s failed: %w", fileName, err)
		}

		db.Exec("INSERT INTO schema_migrations (version) VALUES (?)", version)
		log.Printf("  ✓ Applied %s", version)
	}

	log.Println("✓ SQL migrations completed successfully")
	return nil
}

// executeMigration runs SQL migration in a transaction
func executeMigration(db *sql.DB, sqlContent string) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Split SQL into statements, preserving dollar-quoted strings
	statements := splitSQL(sqlContent)

	log.Printf("  Executing %d statements", len(statements))
	for i, stmt := range statements {
		// Debug: log first few characters of each statement
		if len(stmt) > 50 {
			log.Printf("  Statement %d: %s...", i, stmt[:50])
		} else {
			log.Printf("  Statement %d: %s", i, stmt)
		}
		if _, err := tx.Exec(stmt); err != nil {
			return fmt.Errorf("statement failed: %w\nStatement: %s", err, stmt)
		}
	}

	return tx.Commit()
}

// splitSQL splits SQL into statements, preserving dollar-quoted strings
func splitSQL(sql string) []string {
	var statements []string
	var current strings.Builder
	var inDollarQuote bool
	var dollarTag strings.Builder

	// Process character by character
	for i := 0; i < len(sql); i++ {
		char := sql[i]

		// Check for single-line comments
		if char == '-' && i+1 < len(sql) && sql[i+1] == '-' && !inDollarQuote {
			// Skip the rest of the line
			for i < len(sql) && sql[i] != '\n' {
				i++
			}
			continue
		}

		// Check for dollar quote start/end
		if char == '$' {
			// Look ahead to see if this is a dollar quote
			if i+1 < len(sql) && sql[i+1] == '$' {
				// Found $$ delimiter
				if inDollarQuote && dollarTag.String() == "$$" {
					// Closing the dollar quote
					inDollarQuote = false
					dollarTag.Reset()
					current.WriteByte(char)
					i++ // Skip next $
					current.WriteByte(sql[i])
					continue
				} else if !inDollarQuote {
					// Opening a new dollar quote
					inDollarQuote = true
					dollarTag.WriteString("$$")
					current.WriteByte(char)
					i++ // Skip next $
					current.WriteByte(sql[i])
					continue
				}
			}
		}

		current.WriteByte(char)

		// Only split on semicolons when not in dollar quote
		if char == ';' && !inDollarQuote {
			s := strings.TrimSpace(current.String())
			if s != "" {
				statements = append(statements, s)
			}
			current.Reset()
		}
	}

	// Handle remaining content
	s := strings.TrimSpace(current.String())
	if s != "" {
		statements = append(statements, s)
	}

	return statements
}
