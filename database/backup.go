package database

import (
	"database/sql"
	"fmt"
	"os"
	"strings"
)

// DumpMySQL generates a SQL dump of the database by querying MySQL directly
// through the existing connection. This works regardless of whether the binary
// is running on the host or inside a container.
func DumpMySQL(outputFile string) error {
	sqlDB, err := GetDB().DB()
	if err != nil {
		return fmt.Errorf("could not get sql.DB: %v", err)
	}

	outfile, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("could not create dump file: %v", err)
	}
	defer outfile.Close()

	fmt.Fprintln(outfile, "-- govpn MySQL dump")
	fmt.Fprintln(outfile, "SET FOREIGN_KEY_CHECKS=0;")
	fmt.Fprintln(outfile, "SET SQL_MODE='NO_AUTO_VALUE_ON_ZERO';")
	fmt.Fprintln(outfile)

	rows, err := sqlDB.Query("SHOW TABLES")
	if err != nil {
		return fmt.Errorf("could not list tables: %v", err)
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var table string
		if err := rows.Scan(&table); err != nil {
			return fmt.Errorf("could not scan table name: %v", err)
		}
		tables = append(tables, table)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("error iterating tables: %v", err)
	}

	for _, table := range tables {
		if err := dumpTable(sqlDB, outfile, table); err != nil {
			return fmt.Errorf("could not dump table %s: %v", table, err)
		}
	}

	fmt.Fprintln(outfile, "SET FOREIGN_KEY_CHECKS=1;")
	return nil
}

func dumpTable(db *sql.DB, out *os.File, table string) error {
	var tableName, createStmt string
	row := db.QueryRow(fmt.Sprintf("SHOW CREATE TABLE `%s`", table))
	if err := row.Scan(&tableName, &createStmt); err != nil {
		return fmt.Errorf("SHOW CREATE TABLE failed: %v", err)
	}

	fmt.Fprintf(out, "\n-- Table: %s\n", table)
	fmt.Fprintf(out, "DROP TABLE IF EXISTS `%s`;\n", table)
	fmt.Fprintf(out, "%s;\n\n", createStmt)

	rows, err := db.Query(fmt.Sprintf("SELECT * FROM `%s`", table))
	if err != nil {
		return fmt.Errorf("SELECT failed: %v", err)
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return fmt.Errorf("could not get columns: %v", err)
	}

	for rows.Next() {
		vals := make([]interface{}, len(cols))
		ptrs := make([]interface{}, len(cols))
		for i := range vals {
			ptrs[i] = &vals[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return fmt.Errorf("scan failed: %v", err)
		}

		parts := make([]string, len(cols))
		for i, v := range vals {
			parts[i] = sqlLiteral(v)
		}
		fmt.Fprintf(out, "INSERT INTO `%s` VALUES (%s);\n", table, strings.Join(parts, ", "))
	}

	return rows.Err()
}

func sqlLiteral(v interface{}) string {
	if v == nil {
		return "NULL"
	}
	switch val := v.(type) {
	case []byte:
		return "'" + escapeSQLString(string(val)) + "'"
	case string:
		return "'" + escapeSQLString(val) + "'"
	case int64:
		return fmt.Sprintf("%d", val)
	case float64:
		return fmt.Sprintf("%g", val)
	case bool:
		if val {
			return "1"
		}
		return "0"
	default:
		return fmt.Sprintf("'%v'", val)
	}
}

func escapeSQLString(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `'`, `\'`)
	s = strings.ReplaceAll(s, "\n", `\n`)
	s = strings.ReplaceAll(s, "\r", `\r`)
	s = strings.ReplaceAll(s, "\x00", `\0`)
	return s
}
