package services

import (
	"fmt"
	"log"
	"math/rand"
	"myproject/pkg/config"
	"myproject/pkg/model"
	"net/smtp"

	"strconv"
	"time"

	"github.com/jung-kurt/gofpdf"
	"github.com/xuri/excelize/v2"
)

type Services interface {
	GenerateOtp(length int) int

	SendEmailWithOTP(email string) (string, error)
	SendOrderConfirmationEmail(orderUUID string, amount float64, recipientEmail string) error
	SendOrderReturnConfirmationEmailUser(name string, amt float64, unit int, mail string) error
	SendOrderReturnConfirmationEmailToUser(name string, amt float64, unit int, mail string)
	SendOrderReturnConfirmationEmailVendor(name string, amt float64, unit int, mail string) error
	GenerateDailySalesReportExcel(orders []model.ListOrdersVendor, facts model.Salesfact, types string, id string) (string, error)
	GenerateDailySalesReportPDF(orders []model.ListOrdersVendor, facts model.Salesfact, types string, id string) (string, error)
	GenerateDailySalesReportPDFAdmin(orders []model.ListOrdersAdmin, facts model.Salesfact, types string, id string, ranges string) (string, error)
	GenerateDailySalesReportExcelAdmin(orders []model.ListOrdersAdmin, facts model.Salesfact, types string, id string, ranges string) (string, error)

	GenerateDailySalesReportExcelAdminside(orders []model.ListOrdersVendor, facts model.Salesfact, types string, id string, ranges, name, email, gst string) (string, error)
	GenerateDailySalesReportPDFAdminside(orders []model.ListOrdersVendor, facts model.Salesfact, types string, id string, ranges, name, email, gst string) (string, error)
}
type MyService struct {
	Config config.Config
}

func (s MyService) GenerateDailySalesReportPDFAdminside(orders []model.ListOrdersVendor, facts model.Salesfact, types string, id string, ranges, name, email, gst string) (string, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// Set font size for the title
	pdf.SetFont("Arial", "B", 14)
	title := "Adiecom Sales Report on " + ranges
	pdf.Cell(0, 8, title)
	pdf.Ln(10)

	// Add Vendor Details section
	pdf.SetFont("Arial", "", 10)
	pdf.Cell(0, 8, "Vendor Name: "+name)
	pdf.Ln(6)

	pdf.Cell(0, 8, "Vendor Email: "+email)
	pdf.Ln(6)

	pdf.Cell(0, 8, "Vendor GST: "+gst)
	pdf.Ln(10)

	// Add Salesfact data
	pdf.Cell(0, 8, fmt.Sprintf("Revenue: %.2f", facts.Revenue))
	pdf.Ln(6)

	pdf.Cell(0, 8, fmt.Sprintf("Total Discount: %.2f", facts.TotalDiscount))
	pdf.Ln(6)

	pdf.Cell(0, 8, fmt.Sprintf("Total Sales: %.2f", facts.TotalSales))
	pdf.Ln(6)

	pdf.Cell(0, 8, fmt.Sprintf("Total Orders: %d", facts.TotalOrders))
	pdf.Ln(10)

	// Define headers and column widths for order data
	headers := []string{"Date", "Product Name", "Qty", "Amount", "Product ID", "Discount", "Coupon Amt", "Coupon Code", "Order ID"}
	colWidths := []float64{18, 30, 10, 18, 18, 22, 18, 18, 20}

	// Set header font
	pdf.SetFont("Arial", "B", 9)
	pdf.SetFillColor(220, 220, 220) // Light grey background for headers
	for i, header := range headers {
		pdf.CellFormat(colWidths[i], 7, header, "1", 0, "C", true, 0, "")
	}
	pdf.Ln(7)

	// Set font for data and reduce cell size
	pdf.SetFont("Arial", "", 9)

	// Add rows of data
	for _, order := range orders {
		timestamp := order.Date
		t, err := time.Parse(time.RFC3339, timestamp)
		if err != nil {
			fmt.Println("Error parsing time:", err)
		}
		date := t.Format("2006-01-02")

		row := []string{
			date,
			order.Name,
			fmt.Sprintf("%d", order.Unit),
			fmt.Sprintf("%.2f", order.Amount),
			order.Pid,
			fmt.Sprintf("%.2f", order.Discount),
			fmt.Sprintf("%.2f", order.CouponAmt),
			order.CouponCode,
			fmt.Sprintf("%.8s", order.Oid), // Shortened Order ID
		}

		for i, value := range row {
			pdf.CellFormat(colWidths[i], 6, value, "1", 0, "", false, 0, "")
		}
		pdf.Ln(6)
	}

	// Generate file name and save the PDF
	rand.Seed(time.Now().UnixNano())
	randomNumber := rand.Intn(101)
	fileName := fmt.Sprintf("%s_Sales_Report_%d_%s.pdf", types, randomNumber, id)
	if err := pdf.OutputFileAndClose(fileName); err != nil {
		return "", fmt.Errorf("failed to save PDF file: %w", err)
	}

	return fileName, nil
}

func (s MyService) GenerateDailySalesReportExcelAdminside(orders []model.ListOrdersVendor, facts model.Salesfact, types string, id string, ranges, name, email, gst string) (string, error) {
	file := excelize.NewFile()
	sheet := "Sales Report"
	file.SetSheetName(file.GetSheetName(0), sheet)

	// Add company name and date range
	companyName := "Adiecom Sales Report on " + ranges
	file.SetCellValue(sheet, "A1", companyName)
	file.MergeCell(sheet, "A1", "F1") // Merge cells for title

	// Add Vendor details section
	file.SetCellValue(sheet, "A2", "Vendor Name")
	file.SetCellValue(sheet, "B2", name)

	file.SetCellValue(sheet, "A3", "Vendor Email")
	file.SetCellValue(sheet, "B3", email)

	file.SetCellValue(sheet, "A4", "Vendor GST")
	file.SetCellValue(sheet, "B4", gst)

	// Add Salesfact data starting from row 6
	file.SetCellValue(sheet, "A6", "Revenue")
	file.SetCellValue(sheet, "B6", facts.Revenue)

	file.SetCellValue(sheet, "A7", "Total Discount")
	file.SetCellValue(sheet, "B7", facts.TotalDiscount)

	file.SetCellValue(sheet, "A8", "Total Sales")
	file.SetCellValue(sheet, "B8", facts.TotalSales)

	file.SetCellValue(sheet, "A9", "Total Orders")
	file.SetCellValue(sheet, "B9", facts.TotalOrders)

	// Set starting row for order data
	startingRowForOrders := 11 // Adjust starting row for orders after salesfact data

	// Set headers for order data
	headers := []string{"Date", "Product Name", "Qty", "Amount", "Product ID", "Discount", "Coupon Amt", "Coupon Code", "Order ID"}
	for i, header := range headers {
		cell := fmt.Sprintf("%s%d", string(rune('A'+i)), startingRowForOrders)
		file.SetCellValue(sheet, cell, header)
	}

	// Fill data for each order
	for i, order := range orders {
		row := startingRowForOrders + i + 1 // Start from the row after headers
		timestamp := order.Date

		// Parse the timestamp
		t, err := time.Parse(time.RFC3339, timestamp)
		if err != nil {
			fmt.Println("Error parsing time:", err)
		}

		date := t.Format("2006-01-02")

		file.SetCellValue(sheet, fmt.Sprintf("A%d", row), date)
		file.SetCellValue(sheet, fmt.Sprintf("B%d", row), order.Name)
		file.SetCellValue(sheet, fmt.Sprintf("C%d", row), order.Unit)
		file.SetCellValue(sheet, fmt.Sprintf("D%d", row), order.Amount)
		file.SetCellValue(sheet, fmt.Sprintf("E%d", row), order.Pid)

		file.SetCellValue(sheet, fmt.Sprintf("F%d", row), order.Discount)
		file.SetCellValue(sheet, fmt.Sprintf("G%d", row), order.CouponAmt)
		file.SetCellValue(sheet, fmt.Sprintf("H%d", row), order.CouponCode)
		file.SetCellValue(sheet, fmt.Sprintf("I%d", row), fmt.Sprintf("%.8s", order.Oid)) // Shortened Order ID
	}

	// Set column widths for compact view
	columnWidths := map[string]float64{
		"A": 15,
		"B": 25,
		"C": 10,
		"D": 15,
		"E": 15,
		"F": 20,
		"G": 10,
		"H": 15,
		"I": 15,
		"J": 15,
	}
	for col, width := range columnWidths {
		file.SetColWidth(sheet, col, col, width)
	}

	// Generate a random number for the file name
	rand.Seed(time.Now().UnixNano())
	randomNumber := rand.Intn(101)

	// Save the file
	fileName := fmt.Sprintf("%s_Sales_Report_%d_%s.xlsx", types, randomNumber, id)
	err := file.SaveAs(fileName)
	if err != nil {
		return "", fmt.Errorf("failed to save Excel file: %w", err)
	}

	return fileName, nil
}

// ////////////
func (s MyService) GenerateDailySalesReportPDFAdmin(orders []model.ListOrdersAdmin, facts model.Salesfact, types string, id string, ranges string) (string, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// Set smaller font size for the title and content
	pdf.SetFont("Arial", "B", 14)
	P := "Adiecom  Sales Report on " + ranges
	pdf.Cell(0, 8, P)
	pdf.Ln(10)

	pdf.SetFont("Arial", "", 10)
	pdf.Cell(0, 8, fmt.Sprintf("Revenue: %.2f", facts.Revenue))
	pdf.Ln(6)

	pdf.Cell(0, 8, fmt.Sprintf("Total Discount: %.2f", facts.TotalDiscount))
	pdf.Ln(6)

	pdf.Cell(0, 8, fmt.Sprintf("Total Sales: %.2f", facts.TotalSales))
	pdf.Ln(6)

	pdf.Cell(0, 8, fmt.Sprintf("Total Orders: %d", facts.TotalOrders))
	pdf.Ln(10)

	// Define headers and use smaller font size
	headers := []string{"Date", "Product Name", "Qty", "Amount", "Product ID", "Vendor", "Discount", "Coupon Amt", "Coupon Code", "Order ID"}

	// Smaller column widths for compact content
	colWidths := []float64{18, 30, 10, 18, 18, 22, 18, 18, 20, 20}

	// Set header font
	pdf.SetFont("Arial", "B", 9)
	pdf.SetFillColor(220, 220, 220) // Light grey background for headers
	for i, header := range headers {
		pdf.CellFormat(colWidths[i], 7, header, "1", 0, "C", true, 0, "")
	}
	pdf.Ln(7)

	// Set font for data and reduce cell size
	pdf.SetFont("Arial", "", 9)

	// Add rows of data with compact spacing
	for _, order := range orders {
		timestamp := order.Date
		t, err := time.Parse(time.RFC3339, timestamp)
		if err != nil {
			fmt.Println("Error parsing time:", err)
		}
		date := t.Format("2006-01-02")

		row := []string{
			date,
			order.Name,
			fmt.Sprintf("%d", order.Unit),
			fmt.Sprintf("%.2f", order.Amount),
			order.Pid,
			order.VName,
			fmt.Sprintf("%.2f", order.Discount),
			fmt.Sprintf("%.2f", order.CouponAmt),
			order.CouponCode,
			fmt.Sprintf("%.8s", order.Oid), // Shortened Order ID
		}

		for i, value := range row {
			pdf.CellFormat(colWidths[i], 6, value, "1", 0, "", false, 0, "")
		}
		pdf.Ln(6)
	}

	// Generate file name and save the PDF
	rand.Seed(time.Now().UnixNano())
	randomNumber := rand.Intn(101)
	fileName := fmt.Sprintf("%s_Sales_Report_%d_%s.pdf", types, randomNumber, id)
	if err := pdf.OutputFileAndClose(fileName); err != nil {
		return "", fmt.Errorf("failed to save PDF file: %w", err)
	}

	return fileName, nil
}

func (s MyService) GenerateDailySalesReportPDF(orders []model.ListOrdersVendor, facts model.Salesfact, types string, id string) (string, error) {

	pdf := gofpdf.New("P", "mm", "A4", "")

	pdf.AddPage()

	pdf.SetFont("Arial", "", 12)

	pdf.Cell(0, 10, fmt.Sprintf("Revenue: %.2f", facts.Revenue))
	pdf.Ln(10)

	pdf.Cell(0, 10, fmt.Sprintf("Total Discount: %.2f", facts.TotalDiscount))
	pdf.Ln(10)

	pdf.Cell(0, 10, fmt.Sprintf("Total Sales: %.2f", facts.TotalSales))
	pdf.Ln(10)

	pdf.Cell(0, 10, fmt.Sprintf("Total Orders: %d", facts.TotalOrders))
	pdf.Ln(20)

	headers := []string{"Check Date", "Product Name", "Quantity", "Amount", "Product ID", "discount", "coupon amt", "coupon code", "wallet amt", "Order id"}
	for _, header := range headers {
		pdf.Cell(30, 10, header)
	}
	pdf.Ln(10)

	for _, order := range orders {
		timestamp := order.Date

		// Parse the timestamp
		t, err := time.Parse(time.RFC3339, timestamp)
		if err != nil {
			fmt.Println("Error parsing time:", err)

		}

		date := t.Format("2006-01-02")
		fmt.Println("checking the date is changed !!!", date)
		pdf.Cell(25, 10, date)

		pdf.Cell(20, 10, order.Name)
		pdf.Cell(10, 10, fmt.Sprintf("%d", order.Unit))
		pdf.Cell(10, 10, fmt.Sprintf("%.2f", order.Amount))

		pdf.Cell(10, 10, order.Pid)

		pdf.Cell(10, 10, fmt.Sprintf("%.2f", order.Discount))
		pdf.Cell(10, 10, fmt.Sprintf("%.2f", order.CouponAmt))
		pdf.Cell(10, 10, order.CouponCode)
		pdf.Cell(10, 10, fmt.Sprintf("%.2f", order.WalletAmt))
		pdf.Cell(20, 10, order.Oid)
		pdf.Ln(10)
	}
	if id == "" {
		id = "Admin_Report"
	}
	rand.Seed(time.Now().UnixNano())

	randomNumber := rand.Intn(101)

	fileName := fmt.Sprintf("%s_Sales_Report_%d_%s.pdf", types, randomNumber, id)
	if err := pdf.OutputFileAndClose(fileName); err != nil {
		return "", fmt.Errorf("failed to save PDF file: %w", err)
	}

	return fileName, nil
}
func (s MyService) GenerateDailySalesReportExcelAdmin(orders []model.ListOrdersAdmin, facts model.Salesfact, types string, id string, ranges string) (string, error) {
	file := excelize.NewFile()
	sheet := "Sales Report"
	file.SetSheetName(file.GetSheetName(0), sheet)

	// Add company name and date range
	companyName := "Adiecom Sales Report on " + ranges
	file.SetCellValue(sheet, "A1", companyName)

	// Add Salesfact data at the top
	file.SetCellValue(sheet, "A3", "Revenue")
	file.SetCellValue(sheet, "B3", facts.Revenue)

	file.SetCellValue(sheet, "A4", "Total Discount")
	file.SetCellValue(sheet, "B4", facts.TotalDiscount)

	file.SetCellValue(sheet, "A5", "Total Sales")
	file.SetCellValue(sheet, "B5", facts.TotalSales)

	file.SetCellValue(sheet, "A6", "Total Orders")
	file.SetCellValue(sheet, "B6", facts.TotalOrders)

	// Add a row of space
	startingRowForOrders := 8 // Adjust starting row for orders

	// Set headers
	headers := []string{"Date", "Product Name", "Qty", "Amount", "Product ID", "Vendor", "Discount", "Coupon Amt", "Coupon Code", "Order ID"}
	for i, header := range headers {
		cell := fmt.Sprintf("%s%d", string(rune('A'+i)), startingRowForOrders)
		file.SetCellValue(sheet, cell, header)
	}

	// Fill data
	for i, order := range orders {
		row := startingRowForOrders + i + 1 // Start from the row after headers
		timestamp := order.Date

		// Parse the timestamp
		t, err := time.Parse(time.RFC3339, timestamp)
		if err != nil {
			fmt.Println("Error parsing time:", err)
		}

		date := t.Format("2006-01-02")

		file.SetCellValue(sheet, fmt.Sprintf("A%d", row), date)
		file.SetCellValue(sheet, fmt.Sprintf("B%d", row), order.Name)
		file.SetCellValue(sheet, fmt.Sprintf("C%d", row), order.Unit)
		file.SetCellValue(sheet, fmt.Sprintf("D%d", row), order.Amount)
		file.SetCellValue(sheet, fmt.Sprintf("E%d", row), order.Pid)
		file.SetCellValue(sheet, fmt.Sprintf("F%d", row), order.VName)
		file.SetCellValue(sheet, fmt.Sprintf("G%d", row), order.Discount)
		file.SetCellValue(sheet, fmt.Sprintf("H%d", row), order.CouponAmt)
		file.SetCellValue(sheet, fmt.Sprintf("I%d", row), order.CouponCode)
		file.SetCellValue(sheet, fmt.Sprintf("J%d", row), fmt.Sprintf("%.8s", order.Oid)) // Shortened Order ID
	}

	// Set column widths for compact view
	columnWidths := map[string]float64{
		"A": 15,
		"B": 25,
		"C": 10,
		"D": 15,
		"E": 15,
		"F": 20,
		"G": 10,
		"H": 15,
		"I": 15,
		"J": 15,
	}
	for col, width := range columnWidths {
		file.SetColWidth(sheet, col, col, width)
	}

	// Generate a random number between 0 and 100
	rand.Seed(time.Now().UnixNano())
	randomNumber := rand.Intn(101)

	// Save the file
	fileName := fmt.Sprintf("%s_Sales_Report_%d_%s.xlsx", types, randomNumber, id)
	err := file.SaveAs(fileName)
	if err != nil {
		return "", fmt.Errorf("failed to save Excel file: %w", err)
	}

	return fileName, nil
}

func (s MyService) GenerateDailySalesReportExcel(orders []model.ListOrdersVendor, facts model.Salesfact, types string, id string) (string, error) {

	file := excelize.NewFile()
	sheet := "Sales Report"
	file.SetSheetName(file.GetSheetName(0), sheet)

	// Add Salesfact data at the top
	file.SetCellValue(sheet, "A1", "Revenue")
	file.SetCellValue(sheet, "B1", facts.Revenue)

	file.SetCellValue(sheet, "A2", "Total Discount")
	file.SetCellValue(sheet, "B2", facts.TotalDiscount)

	file.SetCellValue(sheet, "A3", "Total Sales")
	file.SetCellValue(sheet, "B3", facts.TotalSales)

	file.SetCellValue(sheet, "A4", "Total Orders")
	file.SetCellValue(sheet, "B4", facts.TotalOrders)

	// Add a row of space
	startingRowForOrders := 6 // Two rows after the last fact row

	// Set headers
	headers := []string{"Check Date", "Product Name", "Quantity", "Amount", "Product ID", "discount", "coupon amt", "coupon code", "wallet amt", "Order id"}
	for i, header := range headers {
		cell := fmt.Sprintf("%s%d", string(rune('A'+i)), startingRowForOrders)
		file.SetCellValue(sheet, cell, header)
	}

	// Fill data
	for i, order := range orders {
		row := startingRowForOrders + i + 1 // Start from the row after headers
		timestamp := order.Date

		// Parse the timestamp
		t, err := time.Parse(time.RFC3339, timestamp)
		if err != nil {
			fmt.Println("Error parsing time:", err)

		}

		date := t.Format("2006-01-02")
		fmt.Println("checking the date is changed !!!", date)

		file.SetCellValue(sheet, fmt.Sprintf("A%d", row), date)
		file.SetCellValue(sheet, fmt.Sprintf("B%d", row), order.Name)
		file.SetCellValue(sheet, fmt.Sprintf("C%d", row), order.Unit)
		file.SetCellValue(sheet, fmt.Sprintf("D%d", row), order.Amount)
		file.SetCellValue(sheet, fmt.Sprintf("E%d", row), order.Pid)

		file.SetCellValue(sheet, fmt.Sprintf("F%d", row), order.Discount)
		file.SetCellValue(sheet, fmt.Sprintf("G%d", row), order.CouponAmt)
		file.SetCellValue(sheet, fmt.Sprintf("H%d", row), order.CouponCode)
		file.SetCellValue(sheet, fmt.Sprintf("I%d", row), order.WalletAmt)
		file.SetCellValue(sheet, fmt.Sprintf("J%d", row), order.Oid)
	}
	rand.Seed(time.Now().UnixNano())

	// Generate a random number between 0 and 100
	randomNumber := rand.Intn(101)
	// Save the file
	fileName := fmt.Sprintf("%s_Sales_Report_%d_%s.xlsx", types, randomNumber, id)
	err := file.SaveAs(fileName)
	if err != nil {
		return "", fmt.Errorf("failed to save Excel file: %w", err)
	}

	return fileName, nil
}

func (s MyService) SendOrderConfirmationEmail(orderUUID string, amount float64, recipientEmail string) error {
	fmt.Println("this is in the SendOrderConfirmationEmail !!!--", orderUUID, amount, recipientEmail)
	// Message.
	subject := "Order Confirmation"
	body := fmt.Sprintf("Your order has been placed successfully!\nOrder UUID: %s\nAmount: RS%.2f", orderUUID, amount)
	message := fmt.Sprintf("Subject: %s\r\n\r\n%s", subject, body)

	// Authentication.
	SMTPemail := s.Config.SMTPemail
	SMTPpass := s.Config.Password
	auth := smtp.PlainAuth("", SMTPemail, SMTPpass, "smtp.gmail.com")
	fmt.Println("this is my mail !_+_++_+!", SMTPemail, "!+!+!+!+", SMTPpass)
	// Sending email.
	err := smtp.SendMail("smtp.gmail.com:587", auth, SMTPemail, []string{recipientEmail}, []byte(message))
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}
func (s MyService) SendOrderReturnConfirmationEmailToUser(name string, amt float64, unit int, recipientEmail string) {
	fmt.Println("this is in the SendOrderReturnConfirmationEmail !!!--", name, amt, recipientEmail)
	// Message.
	subject := "Vendor has Cancelled your order"
	body := fmt.Sprintf("Your order %s has been placed for returning!\nunits: %d\nAmount: RS%.2f", name, unit, amt)
	message := fmt.Sprintf("Subject: %s\r\n\r\n%s", subject, body)

	// Authentication.
	SMTPemail := s.Config.SMTPemail
	SMTPpass := s.Config.Password
	auth := smtp.PlainAuth("", SMTPemail, SMTPpass, "smtp.gmail.com")
	fmt.Println("this is my mail !_+_++_+!", SMTPemail, "!+!+!+!+", SMTPpass)
	// Sending email.
	err := smtp.SendMail("smtp.gmail.com:587", auth, SMTPemail, []string{recipientEmail}, []byte(message))
	if err != nil {
		fmt.Errorf("failed to send email: %w", err)
	}

}
func (s MyService) SendOrderReturnConfirmationEmailUser(name string, amt float64, unit int, recipientEmail string) error {
	fmt.Println("this is in the SendOrderReturnConfirmationEmail !!!--", name, amt, recipientEmail)
	// Message.
	subject := "Order item returned"
	body := fmt.Sprintf("Your order %s has been placed for returning!\nunits: %d\nAmount: RS%.2f", name, unit, amt)
	message := fmt.Sprintf("Subject: %s\r\n\r\n%s", subject, body)

	// Authentication.
	SMTPemail := s.Config.SMTPemail
	SMTPpass := s.Config.Password
	auth := smtp.PlainAuth("", SMTPemail, SMTPpass, "smtp.gmail.com")
	fmt.Println("this is my mail !_+_++_+!", SMTPemail, "!+!+!+!+", SMTPpass)
	// Sending email.
	err := smtp.SendMail("smtp.gmail.com:587", auth, SMTPemail, []string{recipientEmail}, []byte(message))
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}
func (s MyService) SendOrderReturnConfirmationEmailVendor(name string, amt float64, unit int, recipientEmail string) error {
	fmt.Println("this is in the SendOrderReturnConfirmationEmail !!!--", name, amt, recipientEmail)
	// Message.
	subject := "Customer placed for return"
	body := fmt.Sprintf("A  order %s has been placed for returning!\nunits: %d\nAmount: RS%.2f", name, unit, amt)
	message := fmt.Sprintf("Subject: %s\r\n\r\n%s", subject, body)

	// Authentication.
	SMTPemail := s.Config.SMTPemail
	SMTPpass := s.Config.Password
	auth := smtp.PlainAuth("", SMTPemail, SMTPpass, "smtp.gmail.com")
	fmt.Println("this is my mail !_+_++_+!", SMTPemail, "!+!+!+!+", SMTPpass)
	// Sending email.
	err := smtp.SendMail("smtp.gmail.com:587", auth, SMTPemail, []string{recipientEmail}, []byte(message))
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

func (s MyService) GenerateOtp(length int) int {
	rand.Seed(time.Now().UnixNano())

	// Generate a random number between 10000 and 99999
	randomNum := rand.Intn(90000) + 10000

	fmt.Println("Random 5-digit number:", randomNum)
	return randomNum
}

func (s MyService) SendEmailWithOTP(email string) (string, error) {
	// Generate OTP
	otp := strconv.Itoa(s.GenerateOtp(6))

	// Construct email message
	message := fmt.Sprintf("Subject: OTP for Verification\n\nYour OTP is: %s", otp)
	fmt.Println("this is my email  !!!!!", s.Config.SMTPemail, "this is my email  !!!!!", s.Config.Password)

	SMTPemail := s.Config.SMTPemail
	SMTPpass := s.Config.Password
	auth := smtp.PlainAuth("", "adithyanunni258@gmail.com", SMTPpass, "smtp.gmail.com")

	// Send email using SMTP server
	err := smtp.SendMail("smtp.gmail.com:587", auth, SMTPemail, []string{email}, []byte(message))
	if err != nil {
		log.Println("Error sending email:", err)
		return "", err
	}

	return otp, nil
}
