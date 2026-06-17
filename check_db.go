package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

func main() {
	connStr := "host=localhost user=postgres password=postgres123 dbname=genrent port=5432 sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	fmt.Println("--- USERS ---")
	rows, err := db.Query("SELECT id, email, role FROM users")
	if err != nil {
		log.Fatal(err)
	}
	for rows.Next() {
		var id int
		var email, role string
		rows.Scan(&id, &email, &role)
		fmt.Printf("ID: %d, Email: %s, Role: %s\n", id, email, role)
	}
	rows.Close()

	fmt.Println("\n--- VENDORS ---")
	rows, err = db.Query("SELECT id, company_name, verified FROM vendors")
	if err != nil {
		log.Fatal(err)
	}
	for rows.Next() {
		var id int
		var name string
		var verified bool
		rows.Scan(&id, &name, &verified)
		fmt.Printf("ID: %d, Name: %s, Verified: %v\n", id, name, verified)
	}
	rows.Close()

	fmt.Println("\n--- EQUIPMENT ---")
	rows, err = db.Query("SELECT id, name, available_quantity, availability_status FROM equipment")
	if err != nil {
		log.Fatal(err)
	}
	for rows.Next() {
		var id int
		var name, status string
		var qty int
		rows.Scan(&id, &name, &qty, &status)
		fmt.Printf("ID: %d, Name: %s, Qty: %d, Status: %s\n", id, name, qty, status)
	}
	rows.Close()

	fmt.Println("\n--- BOOKINGS ---")
	rows, err = db.Query("SELECT id, status, total_price, return_otp FROM bookings")
	if err != nil {
		log.Fatal(err)
	}
	for rows.Next() {
		var id int
		var status string
		var price float64
		var returnOtp sql.NullString
		rows.Scan(&id, &status, &price, &returnOtp)
		fmt.Printf("ID: %d, Status: %s, Price: %.2f, Return OTP: %s\n", id, status, price, returnOtp.String)
	}
	rows.Close()

	fmt.Println("\n--- DISPUTES ---")
	rows, err = db.Query("SELECT id, booking_id, claimed_amount, status FROM damage_disputes")
	if err != nil {
		log.Println("Disputes query err (maybe table not exist or empty):", err)
	} else {
		for rows.Next() {
			var id, bookingId int
			var amount float64
			var status string
			rows.Scan(&id, &bookingId, &amount, &status)
			fmt.Printf("ID: %d, BookingID: %d, Amount: %.2f, Status: %s\n", id, bookingId, amount, status)
		}
		rows.Close()
	}
}
