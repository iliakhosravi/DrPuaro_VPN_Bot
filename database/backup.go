package database

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"

	"techybat.org/go-vpn/vars"
)

// DumpPostgres generates a SQL dump of the database by shelling out to
// `pg_dump` (the standard, correct way to dump a Postgres database — it
// handles schema/type/sequence details a hand-rolled dumper would get
// wrong). The `postgresql-client` package must be installed in the runtime
// image for this to work (see Dockerfile).
func DumpPostgres(outputFile string) error {
	if _, err := exec.LookPath("pg_dump"); err != nil {
		return fmt.Errorf("pg_dump not found in PATH: %v", err)
	}

	host := vars.Get("POSTGRES_HOST")
	port := vars.Get("POSTGRES_PORT")
	user := vars.Get("POSTGRES_USER")
	pass := vars.Get("POSTGRES_PASS")
	dbname := vars.Get("POSTGRES_DB")
	sslmode := vars.Get("POSTGRES_SSLMODE")
	if sslmode == "" {
		sslmode = "disable"
	}
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, pass, dbname, sslmode)

	outfile, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("could not create dump file: %v", err)
	}
	defer outfile.Close()

	var stderr bytes.Buffer
	cmd := exec.Command("pg_dump", "--no-owner", "--no-privileges", "--clean", "--if-exists", dsn)
	cmd.Stdout = outfile
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("pg_dump failed: %v: %s", err, stderr.String())
	}

	return nil
}
