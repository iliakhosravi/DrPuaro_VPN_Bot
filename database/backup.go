package database

import (
	"fmt"
	"os"
	"os/exec"

	"techybat.org/go-vpn/vars"
)

func DumpMySQL(outputFile string) error {
	user := vars.Get("MYSQL_USER")
	password := vars.Get("MYSQL_PASS")
	// host := vars.Get("MYSQL_HOST")
	containerName := vars.Get("MYSQL_CONTAINER")
	dbname := vars.Get("MYSQL_DB")

	// mysqldump command with options
	cmd := exec.Command("sudo", "-S", "docker", "exec", containerName, "mysqldump", "-u"+user, "-p"+password, dbname)

	// Create output file
	outfile, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("could not create dump file: %v", err)
	}
	defer outfile.Close()

	// Set the output of the command to the file
	cmd.Stdout = outfile

	// Run the mysqldump command
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("mysqldump command failed: %v", err)
	}

	return nil
}
