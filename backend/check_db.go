package main

import (
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type User struct {
	ID    uint
	Email string
	Role  string
}

type Vendor struct {
	ID          uint
	CompanyName string
	Verified    bool
}

type Equipment struct {
	ID                 uint
	VendorID           uint
	CategoryID         uint
	Name               string
	TotalQuantity      int
	AvailableQuantity  int
	AvailabilityStatus string
}

type Booking struct {
	ID          uint
	CustomerID  uint
	EquipmentID *uint
	GeneratorID *uint
	Status      string
	TotalPrice  float64
	ReturnOTP   string
}

type DamageDispute struct {
	ID            uint
	BookingID     uint
	ClaimedAmount float64
	Status        string
}

type Category struct {
	ID   uint
	Name string
}

func main() {
	dsn := "host=localhost user=postgres password=postgres123 dbname=genrent port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("--- USERS ---")
	var users []User
	db.Raw("SELECT id, email, role FROM users ORDER BY id").Scan(&users)
	for _, u := range users {
		fmt.Printf("ID: %d, Email: %s, Role: %s\n", u.ID, u.Email, u.Role)
	}

	fmt.Println("\n--- VENDORS ---")
	var vendors []Vendor
	db.Raw("SELECT id, company_name, verified FROM vendors ORDER BY id").Scan(&vendors)
	for _, v := range vendors {
		fmt.Printf("ID: %d, Name: %s, Verified: %v\n", v.ID, v.CompanyName, v.Verified)
	}

	fmt.Println("\n--- CATEGORIES ---")
	var cats []Category
	db.Raw("SELECT id, name FROM equipment_categories ORDER BY id").Scan(&cats)
	for _, c := range cats {
		fmt.Printf("ID: %d, Name: %s\n", c.ID, c.Name)
	}

	fmt.Println("\n--- EQUIPMENT ---")
	var equipment []Equipment
	db.Raw("SELECT id, vendor_id, category_id, name, total_quantity, available_quantity, availability_status FROM equipment ORDER BY id").Scan(&equipment)
	for _, e := range equipment {
		fmt.Printf("ID: %d, VendorID: %d, CatID: %d, Name: %s, TotalQty: %d, AvailQty: %d, Status: %s\n",
			e.ID, e.VendorID, e.CategoryID, e.Name, e.TotalQuantity, e.AvailableQuantity, e.AvailabilityStatus)
	}

	fmt.Println("\n--- BOOKINGS ---")
	var bookings []Booking
	db.Raw("SELECT id, customer_id, equipment_id, generator_id, status, total_price, return_otp FROM bookings ORDER BY id").Scan(&bookings)
	for _, b := range bookings {
		equipID := "nil"
		genID := "nil"
		if b.EquipmentID != nil {
			equipID = fmt.Sprintf("%d", *b.EquipmentID)
		}
		if b.GeneratorID != nil {
			genID = fmt.Sprintf("%d", *b.GeneratorID)
		}
		fmt.Printf("ID: %d, CustID: %d, EquipID: %s, GenID: %s, Status: %s, Price: %.2f, ReturnOTP: %s\n",
			b.ID, b.CustomerID, equipID, genID, b.Status, b.TotalPrice, b.ReturnOTP)
	}

	fmt.Println("\n--- DISPUTES ---")
	var disputes []DamageDispute
	db.Raw("SELECT id, booking_id, claimed_amount, status FROM damage_disputes ORDER BY id").Scan(&disputes)
	for _, d := range disputes {
		fmt.Printf("ID: %d, BookingID: %d, Amount: %.2f, Status: %s\n", d.ID, d.BookingID, d.ClaimedAmount, d.Status)
	}

	// Check equipment schema
	fmt.Println("\n--- EQUIPMENT TABLE COLUMNS ---")
	var cols []struct {
		ColName  string
		DataType string
	}
	db.Raw("SELECT column_name as col_name, data_type FROM information_schema.columns WHERE table_name = 'equipment' ORDER BY ordinal_position").Scan(&cols)
	for _, c := range cols {
		fmt.Printf("  %s (%s)\n", c.ColName, c.DataType)
	}

	// Check bookings schema
	fmt.Println("\n--- BOOKINGS TABLE COLUMNS ---")
	var bCols []struct {
		ColName  string
		DataType string
	}
	db.Raw("SELECT column_name as col_name, data_type FROM information_schema.columns WHERE table_name = 'bookings' ORDER BY ordinal_position").Scan(&bCols)
	for _, c := range bCols {
		fmt.Printf("  %s (%s)\n", c.ColName, c.DataType)
	}

	// Try directly inserting a booking to see the exact error
	fmt.Println("\n--- TRYING DIRECT BOOKING INSERT ---")
	type EqID = uint
	eqID := EqID(7)
	catID := uint(2)
	result := db.Raw(`
		INSERT INTO bookings 
		(customer_id, equipment_id, category_id, start_date, end_date, total_price, rental_price, mobilization_fee, demobilization_fee, advance_amount, advance_paid, status, address, notes, created_at, updated_at)
		VALUES (24, ?, ?, '2026-07-01', '2026-07-06', 12500, 12500, 0, 0, 3750, false, 'requested', 'Test Address Mumbai', '', NOW(), NOW())
		RETURNING id
	`, eqID, catID)
	if result.Error != nil {
		fmt.Println("INSERT ERROR:", result.Error)
	} else {
		fmt.Println("Direct insert OK!")
	}
}
