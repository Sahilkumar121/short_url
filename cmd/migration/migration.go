package main

import (
	"fmt"
	"log"
	"os"

	"github.com/Sahilkumar121/short_url/internal/config"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {

	cfg, err := config.MustLoad()
	if err != nil {
		log.Fatalf("error occur %v \n", err)
	}

	if len(os.Args) < 2 {
		log.Fatalln("argument required <up | down>")
	}

	m, err := migrate.New("file://migrations", cfg.DB)
	if err != nil {
		log.Fatalf("failed to migrate %v \n", err)
	}

	switch os.Args[1] {
	case "up":
		m.Up()
	case "down":
		m.Steps(-1)
	default:
		log.Fatalln("unknown argument")
	}

	fmt.Println("migrate successfully...")
}
