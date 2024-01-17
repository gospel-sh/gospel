package orm

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/fs"
	"io/ioutil"
	"log"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"text/template"
)

type ConfigError struct {
	msg string
}

type Migration struct {
	Version     int    `json:"version"`
	Description string `json:"description"`
	Content     string `json:"content"`
	FileName    string `json:"filename"`
}

func (self *ConfigError) Error() string {
	return self.msg
}

type MigrationConfig struct {
	VersionTable *VersionTableConfig `json:"versionTable"`
}

type VersionTableConfig struct {
	Name          string `json:"name"`
	VersionColumn string `json:"versionColumn"`
}

type MigrationContext struct {
	Migration Migration
	Config    *MigrationConfig
	DBType    string
}

type MigrationManager struct {
	Config         *MigrationConfig
	Path           string
	FS             fs.FS
	UpMigrations   map[int]Migration
	DownMigrations map[int]Migration
	latestVersion  int
	DB             *sql.DB
	DBType         string
}

func (self *MigrationManager) LatestVersion() int {
	return self.latestVersion
}

func (self *MigrationManager) CurrentVersion() (int, error) {

	//we test the SQL connection
	_, err := self.DB.Exec(`
  	      SELECT 'Hello, World';
	    `)

	if err != nil {
		return 0, err
	}

	rows, err := self.DB.Query(fmt.Sprintf(`
        SELECT
            %s
        FROM
            %s
        LIMIT 1;
    `, self.Config.VersionTable.VersionColumn, self.Config.VersionTable.Name))

	if err != nil {
		return 0, nil
	}

	defer rows.Close()

	if !rows.Next() {
		//nothing stored in the version table so far
		return 0, nil
	}

	var version int
	rows.Scan(&version)

	return version, nil
}

// Migrates the database to the current head version
func (self *MigrationManager) Migrate(version int) error {
	currentVersion, _ := self.CurrentVersion()
	relevantMigrations := make([]Migration, 0, 10)
	up := true
	if version != -1 && version < currentVersion {
		up = false
	}
	if up {
		keys := make([]int, 0)
		for key := range self.UpMigrations {
			if key > currentVersion {
				// if an explicit version number is given,
				// we include only revisions up to this version (inclusive)
				if version == -1 || key <= version {
					keys = append(keys, key)
				}
			}
		}
		sort.Ints(keys)
		for _, k := range keys {
			relevantMigrations = append(relevantMigrations, self.UpMigrations[k])
		}
	} else {
		keys := make([]int, 0)
		for key := range self.DownMigrations {
			if key <= currentVersion {
				// if an explicit version number is given,
				// we include only revisions up to this version (non-inclusive)
				if version == -1 || key > version {
					keys = append(keys, key)
				}
			}
		}
		sort.Sort(sort.Reverse(sort.IntSlice(keys)))
		for _, k := range keys {
			relevantMigrations = append(relevantMigrations, self.DownMigrations[k])
		}
	}
	tx, err := self.DB.Begin()
	if err != nil {
		return err
	}
	err = self.ExecuteMigrations(tx, relevantMigrations)
	if err != nil {
		tx.Rollback()
	} else {
		tx.Commit()
	}

	return err
}

// Executes a list of migrations
func (self *MigrationManager) ExecuteMigrations(tx *sql.Tx, migrations []Migration) error {
	for _, migration := range migrations {

		log.Printf("Executing migration %v\n", migration.FileName)

		templ, err := template.New(migration.FileName).Parse(migration.Content)

		if err != nil {
			return fmt.Errorf("cannot load migration: %v", err)
		}

		context := &MigrationContext{
			Config:    self.Config,
			Migration: migration,
			DBType:    self.DBType,
		}

		output := bytes.NewBuffer(nil)
		if err := templ.Execute(output, context); err != nil {
			return err
		}

		if _, err := tx.Exec(output.String()); err != nil {
			return err
		}

	}
	return nil
}

// Load the migrations from the "migrations" subfolder
func (self *MigrationManager) LoadMigrations() error {
	self.latestVersion = 0
	fileInfos, err := fs.ReadDir(self.FS, self.Path)
	if err != nil {
		return err
	}
	re := regexp.MustCompile("(?i)^(\\d+)_(up|down)_(.*)\\.sql$")
	for _, fileInfo := range fileInfos {
		subMatches := re.FindStringSubmatch(fileInfo.Name())
		if subMatches == nil {
			continue
		}
		version, _ := strconv.Atoi(subMatches[1])
		if version > self.latestVersion {
			self.latestVersion = version
		}
		direction := subMatches[2]
		description := subMatches[3]
		migrationFileName := filepath.Join(self.Path, fileInfo.Name())
		content, err := fs.ReadFile(self.FS, migrationFileName)
		if err != nil {
			continue
		}
		migration := Migration{
			Description: description,
			Content:     string(content),
			Version:     version,
			FileName:    migrationFileName,
		}
		if direction == "up" {
			self.UpMigrations[version] = migration
		} else {
			self.DownMigrations[version] = migration
		}
	}
	return nil
}

func (self *MigrationManager) LoadConfig(ConfigPath string) error {
	var config *MigrationConfig
	filePath := path.Join(ConfigPath, "config.json")
	file, err := self.FS.Open(filePath)

	if err != nil {
		return err
	}

	fileContent, err := ioutil.ReadAll(file)

	if err != nil {
		return err
	}

	if err := json.Unmarshal(fileContent, &config); err != nil {
		return fmt.Errorf("unmarshal %v: %v", filePath, err)
	}

	self.Config = config

	return nil
}

func MakeMigrationManager(configPath string, fS fs.FS, db *sql.DB, dbType string) (*MigrationManager, error) {

	migrationManager := &MigrationManager{
		Path:   configPath,
		DB:     db,
		DBType: dbType,
		FS:     fS,
	}

	err := migrationManager.LoadConfig(configPath)

	if err != nil {
		return nil, fmt.Errorf("migrationManager.LoadConfig: %v", err)
	}

	migrationManager.UpMigrations = make(map[int]Migration)
	migrationManager.DownMigrations = make(map[int]Migration)

	err = migrationManager.LoadMigrations()

	if err != nil {
		return nil, err
	}

	return migrationManager, nil
}
