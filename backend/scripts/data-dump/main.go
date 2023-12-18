package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	ctx := context.Background()
	var dbURL string = fmt.Sprintf("postgresql://%s:%s@%s:%s/%s", "postgres", "eicdev", "localhost", "15432", "electra")
	dbpool, err := NewPG(ctx, dbURL)
	if err != nil {
		fmt.Printf("Failed init DB")
		return
	}
	defer dbpool.Close()

    lines, err := ReadCsv("data.csv")
    if err != nil {
        panic(err)
    }

	idx := 1

    // Loop through lines & turn into object
    for _, line := range lines {
        data := Material{
            No: idx,
            Plant: line[1],
			Area: line[2],
			Name: line[3],
			Specification: Specification{
				Capacity: CleanData(line[4]),
				Voltage: CleanData(line[5]),
				Current: CleanData(line[6]),
				RPM: CleanData(line[7]),
			},
			Maker: line[8],
			SerialNumber: line[9],
			StartingCurrent: StartingCurrent{
				When: line[10],
				Check: line[11],
			},
			RotorBar: RotorBar{
				CheckDate: line[12],
				CheckStatus: line[13],
				Reason: line[14],
				Remark: line[15],
			},
			Frame: line[16],
			Type: line[17],
			Size: Size{
				ShaftDiameter: CleanData(line[21]),
				BaseWidth: CleanData(line[22]),
				BaseLength: CleanData(line[23]),
				C: CleanData(line[24]),
				E: CleanData(line[25]),
				H: CleanData(line[26]),
			},

        }
		if err := dbpool.InsertUser(ctx, data); err != nil{
			fmt.Printf("Error Insert Material: %+v\n", err)
			return
		}
        fmt.Printf("Material: %+v\n", data)
		idx++
    }
}

//unused function, for future usage
func CleanData(data string) (defaultVal float32) { //convert - to 0 value
	defaultVal = 0
	if strings.Contains(data, "-") {
		return defaultVal
	}

	value, err := strconv.ParseFloat(data, 32)
	if err != nil {
		return defaultVal
	}

	return float32(value)
}

// ReadCsv accepts a file and returns its content as a multi-dimentional type
// with lines and each column. Only parses to string type.
func ReadCsv(filename string) ([][]string, error) {

    // Open CSV file
    f, err := os.Open(filename)
    if err != nil {
        return [][]string{}, err
    }
    defer f.Close()

    // Read File into a Variable
    lines, err := csv.NewReader(f).ReadAll()
    if err != nil {
        return [][]string{}, err
    }

    return lines, nil
}

type postgres struct {
	db *pgxpool.Pool
}

var (
	pgInstance *postgres
	pgOnce     sync.Once
)

func NewPG(ctx context.Context, connString string) (*postgres, error) {
	pgOnce.Do(func() {
		db, err := pgxpool.New(ctx, connString)
		if err != nil {
			fmt.Println("unable to create connection pool: ", err)
			return  
		}

		pgInstance = &postgres{db}
	})

	return pgInstance, nil
}

func (pg *postgres) Ping(ctx context.Context) error {
	return pg.db.Ping(ctx)
}

func (pg *postgres) Close() {
	pg.db.Close()
}

func (pg *postgres) InsertUser(ctx context.Context, material Material) error {
	query := `INSERT INTO list_materials
	(id, qcode, plant, area, category, name, capacity, voltage, current, rpm, shaft_diameter, base_width, base_length, c, e, h, maker)
	VALUES(@id, @qcode, @plant, @area, @category, @name, @capacity, @voltage, @current, @rpm, @shaft_diameter, @base_width, @base_length, @c, @e, @h, @maker)`
	// query := `INSERT INTO users (name, email) VALUES (@userName, @userEmail)`
	args := pgx.NamedArgs{
	  "qcode": material.Qcode,
	  "plant": material.Plant,
	  "area": material.Area,
	  "category": "HV Motor",
	  "name": material.Name,
	  "capacity": material.Specification.Capacity,
	  "voltage": material.Specification.Voltage,
	  "current": material.Specification.Current,
	  "rpm": material.Specification.RPM,
	  "shaft_diameter": material.Size.ShaftDiameter,
	  "base_width": material.Size.BaseWidth,
	  "base_length": material.Size.BaseLength,
	  "c": material.Size.C,
	  "e": material.Size.E,
	  "h": material.Size.H,
	  "maker": material.Maker,
	  "id": material.No,
	}
	_, err := pg.db.Exec(ctx, query, args)
	if err != nil {
	  return fmt.Errorf("unable to insert row: %w", err)
	}
  
	return nil
  }
  