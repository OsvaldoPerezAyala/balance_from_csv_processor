package requesthandler

import (
	"balance_from_csv_processor/models"
	"balance_from_csv_processor/repository"
	"bytes"
	"encoding/base64"
	"encoding/csv"
	"errors"
	"fmt"
	"github.com/go-mail/mail"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
}
func ProcessCSV(c echo.Context) error {

	var body = new(models.Request)

	if err := c.Bind(&body); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "error_parsing: "+err.Error())
	}

	if err := validate.Struct(body); err != nil {
		var validationErrors validator.ValidationErrors
		errors.As(err, &validationErrors)
		var errMsg string
		for _, fieldErr := range validationErrors {
			errMsg += fmt.Sprintf("Campo '%s' %s; ", fieldErr.Field(), fieldErr.Tag())
		}
		return echo.NewHTTPError(http.StatusBadRequest, "error_validation: "+errMsg)
	}

	fmt.Println("Processing CSV Request")
	resultArray, csvFile, err := processCSVFile()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "fails to process csv file")
	}

	err = sendGoMail(resultArray, csvFile, body)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "fails to send email ")
	}

	return echo.NewHTTPError(http.StatusOK, "request processed")
}

func processCSVFile() ([]string, []byte, error) {

	csvFilePath := "/app/data/txns.csv"
	csvFilePath = strings.TrimSpace(csvFilePath)
	csvContent, err := ioutil.ReadFile(csvFilePath)
	if err != nil {
		fmt.Println("Error al leer el archivo CSV:", err)
		return nil, nil, err
	}

	reader := csv.NewReader(strings.NewReader(string(csvContent)))
	reader.Comma = ','
	reader.FieldsPerRecord = -1

	_, err = reader.Read()
	if err != nil {
		fmt.Println("Error al leer el encabezado:", err)
		return nil, nil, err
	}

	dataByMonth := make(map[string]*models.TransactionData)

	var totalBalance float64
	var totalCredit float64
	var totalDebit float64
	var totalCreditCount int
	var totalDebitCount int

	transactionsByMonth := make(map[string]int)

	months := map[time.Month]string{
		time.January:   "January",
		time.February:  "February",
		time.March:     "March",
		time.April:     "April",
		time.May:       "May",
		time.June:      "June",
		time.July:      "July",
		time.August:    "August",
		time.September: "September",
		time.October:   "October",
		time.November:  "November",
		time.December:  "December",
	}

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Println("Error al leer el archivo CSV:", err)
			return nil, nil, err
		}

		if len(record) < 3 {
			fmt.Println("Registro con campos insuficientes:", record)
			continue
		}

		idStr := strings.TrimSpace(record[0])
		_, err = strconv.Atoi(idStr)
		if err != nil {
			fmt.Printf("Error al convertir Id '%s' a entero: %v\n", idStr, err)
			continue
		}

		dateStr := strings.TrimSpace(record[1])
		date, err := time.Parse("01/02/2006", fmt.Sprintf("%s/2022", dateStr))
		if err != nil {
			fmt.Printf("Error al parsear la fecha '%s': %v\n", dateStr, err)
			continue
		}

		monthName := months[date.Month()]
		transactionStr := strings.TrimSpace(record[2])
		transaction, err := strconv.ParseFloat(transactionStr, 64)
		if err != nil {
			fmt.Printf("Error al convertir Transaction '%s' a float64: %v\n", transactionStr, err)
			continue
		}

		transactionObject := models.TransactionRow{
			Id:          idStr,
			Date:        primitive.NewDateTimeFromTime(date),
			Transaction: transaction,
			FileName:    "txns.csv",
		}
		repository.SaveData("transaction_data", "transactions", transactionObject)

		if _, exists := dataByMonth[monthName]; !exists {
			dataByMonth[monthName] = &models.TransactionData{}
		}

		monthData := dataByMonth[monthName]

		if transaction > 0 {
			monthData.CreditTotal += transaction
			monthData.CreditCount++
			totalCredit += transaction
			totalCreditCount++
		} else if transaction < 0 {
			monthData.DebitTotal += transaction
			monthData.DebitCount++
			totalDebit += transaction
			totalDebitCount++
		}
		totalBalance += transaction
		transactionsByMonth[monthName]++

	}

	var arrayResult []string

	fmt.Printf("\nTotal balance is %.2f\n", totalBalance)

	arrayResult = append(arrayResult, fmt.Sprintf("Total balance is %.2f", totalBalance))

	monthOrder := []string{"January", "February", "March", "April", "May", "June",
		"July", "August", "September", "October", "November", "December"}

	for _, month := range monthOrder {
		if count, exists := transactionsByMonth[month]; exists {
			fmt.Printf("Number of transactions in %s: %d\n", month, count)
			arrayResult = append(arrayResult, fmt.Sprintf("Number of transactions in %s: %d", month, count))
		}
	}

	var averageCredit float64
	var averageDebit float64

	if totalCreditCount > 0 {
		averageCredit = totalCredit / float64(totalCreditCount)
	}

	if totalDebitCount > 0 {
		averageDebit = totalDebit / float64(totalDebitCount)
	}

	fmt.Printf("Average debit amount: %.2f\n", averageDebit)
	fmt.Printf("Average credit amount: %.2f\n", averageCredit)
	arrayResult = append(arrayResult, fmt.Sprintf("Average debit amount: %.2f\n", averageDebit))
	arrayResult = append(arrayResult, fmt.Sprintf("Average credit amount: %.2f\n", averageCredit))

	return arrayResult, csvContent, nil
}

func sendGoMail(resultArray []string, csvContent []byte, request *models.Request) error {

	smtpHost := "smtp.gmail.com"
	smtpPort := 587
	sender := os.Getenv("USER_EMAIL")
	password := os.Getenv("APPLICATION_KEY")

	to := []string{request.Email}
	subject := "Summary transaction"

	imageData, err := base64.StdEncoding.DecodeString(base64Image)
	if err != nil {
		fmt.Println("Error al decodificar la imagen:", err)
		return err
	}

	m := mail.NewMessage()

	m.SetHeader("From", sender)
	m.SetHeader("To", to...)
	m.SetHeader("Subject", subject)

	var tagArray []string

	for _, str := range resultArray {
		tagArray = append(tagArray, "<p>"+str+"</p>")
	}

	resultado := strings.Join(tagArray, "\n")

	htmlBody := `
        <html>
            <body>
                <p>Summary of file txns.csv transactions .</p>
               ` + resultado + `
                <br>
                <img src="cid:footer-image" alt="Image">
            </body>
        </html>
    `
	m.SetBody("text/html", htmlBody)

	m.EmbedReader("footer.png", bytes.NewReader(imageData), mail.SetHeader(map[string][]string{
		"Content-ID": {"<footer-image>"},
	}))

	m.AttachReader("txns.csv", bytes.NewReader(csvContent))
	d := mail.NewDialer(smtpHost, smtpPort, sender, password)

	if err := d.DialAndSend(m); err != nil {
		fmt.Println("Error al enviar el correo:", err)
		return err
	}

	fmt.Println("Correo enviado exitosamente!")
	return nil
}

const base64Image = "iVBORw0KGgoAAAANSUhEUgAAASwAAABvCAMAAABhGA0xAAAC9FBMVEVHcEwAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAAOkAeHArvAAAA/HRSTlMAIuP/0xTiKuhV+RK4HZz8ZmxrbY39BnignprxCsL4wQUTn7u5vJUMGtrMELJN1+7t5wEcMTAEDhsnMjpAQj02LCEVCdkDI0RMUUU3JBHf3uC3VIWtydvy9fb0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"
