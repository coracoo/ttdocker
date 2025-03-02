package database

import (
    "database/sql"
    _ "github.com/mattn/go-sqlite3"
    "log"
    "os"
    "path/filepath"
)

var db *sql.DB

// InitDB 初始化数据库连接
func InitDB(dbPath string) error {
    // 确保数据目录存在
    dir := filepath.Dir(dbPath)
    if err := os.MkdirAll(dir, 0755); err != nil {
        return err
    }

    log.Printf("正在打开数据库: %s", dbPath)
    
    var err error
    db, err = sql.Open("sqlite3", dbPath)
    if err != nil {
        return err
    }

    // 测试数据库连接
    if err := db.Ping(); err != nil {
        return err
    }

    // 创建表
    return createTables()
}

// createTables 创建必要的数据库表
func createTables() error {
    // 创建注册表配置表
    _, err := db.Exec(`
    CREATE TABLE IF NOT EXISTS registries (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        name TEXT NOT NULL,
        url TEXT NOT NULL,
        username TEXT,
        password TEXT,
        is_default INTEGER DEFAULT 0,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
    )`)
    if err != nil {
        log.Printf("创建 registries 表失败: %v", err)
        return err
    }

    // 创建 Docker 代理配置表
    _, err = db.Exec(`
    CREATE TABLE IF NOT EXISTS docker_proxy (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        enabled INTEGER DEFAULT 0,
        http_proxy TEXT,
        https_proxy TEXT,
        no_proxy TEXT,
        registry_mirrors TEXT,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
    )`)
    if err != nil {
        log.Printf("创建 docker_proxy 表失败: %v", err)
        return err
    }

    // 创建应用商店表
    _, err = db.Exec(`
    CREATE TABLE IF NOT EXISTS applications (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        name TEXT NOT NULL,
        description TEXT,
        icon_url TEXT,
        category TEXT,
        version TEXT,
        image_name TEXT NOT NULL,
        port_mappings TEXT,
        environment_vars TEXT,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
    )`)
    if err != nil {
        return err
    }

    // 创建部署记录表
    _, err = db.Exec(`
    CREATE TABLE IF NOT EXISTS deployments (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        app_id INTEGER,
        container_id TEXT,
        status TEXT,
        port_mappings TEXT,
        environment_vars TEXT,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        FOREIGN KEY (app_id) REFERENCES applications(id)
    )`)

    return err
}

func GetDB() *sql.DB {
    return db
}

func Close() error {
    if db != nil {
        return db.Close()
    }
    return nil
}