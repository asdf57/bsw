package controller

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/asdf57/bsw/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Controller struct {
	db *gorm.DB
}

func fetchExchangeRate(fromCurrency string, toCurrency string) (float64, error) {
	if fromCurrency == toCurrency {
		return 1.0, nil
	}

	// Make a GET request to the exchange rate API
	url := fmt.Sprintf("https://api.frankfurter.dev/v1/latest?base=%s", fromCurrency)
	resp, err := http.Get(url)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch exchange rates: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("exchange rate API returned non-200 status: %d", resp.StatusCode)
	}

	var data struct {
		Rates map[string]float64 `json:"rates"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return 0, fmt.Errorf("failed to decode exchange rate response: %w", err)
	}

	rate, ok := data.Rates[toCurrency]
	if !ok {
		return 0, fmt.Errorf("exchange rate not found for currency: %s", toCurrency)
	}

	return rate, nil
}

func NewController(db *gorm.DB) *Controller {
	return &Controller{db: db}
}

func (ctrl *Controller) resolveUserIdFromName(name string) (*uint, error) {
	var dbEntry models.UserDBEntry

	if err := ctrl.db.Where("name = ?", name).First(&dbEntry).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("user was not found: %v", err)
			return nil, err
		}

		log.Printf("unhandled error while searching for user record: %v", err)
		return nil, err
	}

	return &dbEntry.ID, nil
}

func (ctrl *Controller) resolveUserNameFromId(id uint) (*string, error) {
	var dbEntry models.UserDBEntry

	if err := ctrl.db.First(&dbEntry, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("user was not found: %v", err)
			return nil, err
		}

		log.Printf("unhandled error while searching for user record: %v", err)
		return nil, err
	}

	return &dbEntry.Name, nil
}

func (ctrl *Controller) commitBalanceUpdates(dbEntries []models.BalanceDBEntry) error {
	tx := ctrl.db.Begin()
	if tx.Error != nil {
		log.Printf("failed to start transaction for updating balances: %v", tx.Error)
		return tx.Error
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var existingBalances []models.BalanceDBEntry
	if err := tx.Find(&existingBalances).Error; err != nil {
		log.Printf("failed to fetch existing balances: %v", err)
		tx.Rollback()
		return err
	}

	// Delete the balances that are no longer relevant
	for _, existingEntry := range existingBalances {
		found := false
		for _, desiredEntry := range dbEntries {
			if existingEntry.FromUserID == desiredEntry.FromUserID && existingEntry.ToUserID == desiredEntry.ToUserID {
				found = true
				break
			}
		}

		if !found {
			log.Printf("Deleting balance entry from DB: from user id %d to user id %d: $%.2f\n", existingEntry.FromUserID, existingEntry.ToUserID, existingEntry.Amount)
			if err := tx.Unscoped().Delete(&existingEntry).Error; err != nil {
				log.Printf("failed to delete balance entry: %v", err)
				tx.Rollback()
				return err
			}
		}
	}

	for _, entry := range dbEntries {
		log.Printf("Commiting balance update to DB: from user id %d to user id %d: $%.2f\n", entry.FromUserID, entry.ToUserID, entry.Amount)
		// if the entry DNE, create it. Otherwise update the existing row's amount!
		var existingEntry models.BalanceDBEntry
		err := tx.Where("from_user_id = ? AND to_user_id = ?", entry.FromUserID, entry.ToUserID).First(&existingEntry).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("Creating new balance entry in DB: from user id %d to user id %d: $%.2f\n", entry.FromUserID, entry.ToUserID, entry.Amount)
			if err := tx.Create(&entry).Error; err != nil {
				log.Printf("failed to create balance entry: %v", err)
				tx.Rollback()
				return err
			}
		} else if err != nil {
			log.Printf("db error while searching for existing balance entry: %v", err)
			tx.Rollback()
			return err
		} else {
			log.Printf("Updating existing balance entry in DB (id %d): from user id %d to user id %d: $%.2f\n", existingEntry.ID, entry.FromUserID, entry.ToUserID, entry.Amount)
			existingEntry.Amount = entry.Amount
			if err := tx.Save(&existingEntry).Error; err != nil {
				log.Printf("failed to update balance entry: %v", err)
				tx.Rollback()
				return err
			}
		}
	}

	return tx.Commit().Error
}

// GetPayment godoc
// @Summary Get a payment by ID
// @Param id path int true "Payment ID"
// @Router /api/v1/payment/{id} [get]
func (ctrl *Controller) GetPayment(c *gin.Context) {
	id := c.Param("id")

	var p models.PaymentDBEntry
	if err := ctrl.db.Preload("Owers").First(&p, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "payment not found"})
			return
		}

		log.Printf("db error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.JSON(http.StatusOK, p)
}

// GetPayments godoc
// @Summary Get all payments
// @Router /api/v1/payment/all [get]
func (ctrl *Controller) GetPayments(c *gin.Context) {
	var req []models.PaymentDBEntry

	if err := ctrl.db.Preload("Owers").Find(&req).Error; err != nil {
		log.Printf("db error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.JSON(http.StatusOK, req)
}

// PostPayment godoc
// @Summary      Create a new payment
// @Description  save a payment in the database
// @Accept       json
// @Produce      json
// @Param        payment  body      models.Payment  true  "payment data"
// @Success      200      {object}  models.Payment
// @Failure      400      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /api/v1/payment [post]
func (ctrl *Controller) PostPayment(c *gin.Context) {
	var req models.Payment
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	exchangeRate, err := fetchExchangeRate(req.FromExchangeRate, req.ToExchangeRate)
	if err != nil {
		log.Printf("failed to fetch exchange rate: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch exchange rate"})
		return
	}

	// check if the payer exists and get their id
	var payer models.UserDBEntry
	if err := ctrl.db.Where("name = ?", req.Payer).First(&payer).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "payer not found"})
			return
		}

		log.Printf("db error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	// check if the owers exist and load their user rows for many2many association
	var owers []models.UserDBEntry
	for _, ower := range req.Owers {
		var owerEntry models.UserDBEntry

		// if ower == payer, then you've done something wrong
		if ower == payer.Name {
			c.JSON(http.StatusBadRequest, gin.H{"error": "payer cannot owe itself"})
			return
		}

		if err := ctrl.db.Where("name = ?", ower).First(&owerEntry).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusBadRequest, gin.H{"error": "ower not found"})
				return
			}
			log.Printf("db error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			return
		}

		owers = append(owers, owerEntry)
	}

	// build the db req
	dbReq := models.PaymentDBEntry{
		Amount:           req.Amount,
		Description:      req.Description,
		Date:             time.Now(),
		PayerID:          payer.ID,
		PayerName:        payer.Name,
		FromExchangeRate: req.FromExchangeRate,
		ToExchangeRate:   req.ToExchangeRate,
		ExchangeRate:     exchangeRate,
		Owers:            owers,
	}

	// save the payment to the database
	if err := ctrl.db.Create(&dbReq).Error; err != nil {
		log.Printf("failed to create payment: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not save payment"})
		return
	}

	log.Printf("Received payment request: %+v\n", dbReq)
	c.JSON(http.StatusOK, req)
}

// DeletePayment godoc
// @Summary Delete a payment by ID
// @Param id path int true "Payment ID"
// @Router /api/v1/payment/{id} [delete]
func (ctrl *Controller) DeletePayment(c *gin.Context) {
	id := c.Param("id")

	var p models.PaymentDBEntry
	if err := ctrl.db.First(&p, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "payment not found"})
			return
		}

		log.Printf("db error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	if err := ctrl.db.Unscoped().Delete(&p).Error; err != nil {
		log.Printf("failed to delete payment: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not delete payment"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"info": "payment deleted"})
}

// GetDbHealth godoc
// @Summary Health check
// @Router /api/v1/health [get]
func (ctrl *Controller) GetDbHealth(c *gin.Context) {
	// Run simple query to verify connectivity
	var result int
	if err := ctrl.db.Raw("SELECT 1").Scan(&result).Error; err != nil {
		log.Printf("health check failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok", "value": result})
}

// BackupDB godoc
// @Summary      Trigger a database backup
// @Description  Starts a Postgres backup job (pg_dump)
// @Tags         admin
// @Produce      json
// @Success      202  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /api/v1/admin/backup [post]
func (ctrl *Controller) BackupDB(c *gin.Context) {
	filename := fmt.Sprintf("backup_%s.dump", time.Now().UTC().Format("20060102T150405Z"))
	path := filepath.Join("/backups", filename)

	if err := os.MkdirAll("/backups", os.ModePerm); err != nil {
		log.Printf("failed to create backup directory: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create backup directory"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Minute)
	defer cancel()

	cmd := exec.CommandContext(ctx,
		"pg_dump",
		"-Fc",
		"-h", os.Getenv("PGHOST"),
		"-p", os.Getenv("PGPORT"),
		"-U", os.Getenv("PGUSER"),
		"-d", os.Getenv("PGDATABASE"),
		"-f", path,
	)
	cmd.Env = append(os.Environ(), "PGPASSWORD="+os.Getenv("PGPASSWORD"))

	out, err := cmd.CombinedOutput()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to run backup command", "output": string(out), "errorInfo": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"backup": filename})
}

// GetBackup godoc
// @Summary      Download a database backup
// @Description  Downloads a Postgres backup file
// @Tags         admin
// @Produce      application/octet-stream
// @Param        filename path string true "Backup filename"
// @Success      200  {file}  string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /api/v1/admin/backup/{filename} [get]
func (ctrl *Controller) GetBackup(c *gin.Context) {
	filename := c.Param("filename")

	fmt.Printf("Exporting backup for %s", filename)

	path := filepath.Join("/backups", filename)

	if _, err := os.Stat(path); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "backup not found"})
		return
	}

	c.FileAttachment(path, filename)
}

// PostUser godoc
// @Summary      Create a new user
// @Description  save a user in the database
// @Accept       json
// @Produce      json
// @Param        user  body      models.User  true  "user data"
// @Success      200      {object}  models.User
// @Failure      400      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /api/v1/user [post]
func (ctrl *Controller) PostUser(c *gin.Context) {
	var req models.User

	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	dbReq := models.UserDBEntry{
		Name: req.Name,
	}

	if err := ctrl.db.Create(&dbReq).Error; err != nil {
		log.Printf("failed to create user: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create user"})
		return
	}

	log.Printf("Received user request: %+v\n", dbReq)
	c.JSON(http.StatusOK, req)
}

// DeleteUser godoc
// @Summary Delete a user by user by name
// @Param name path string true "user name"
// @Router /api/v1/user/{name} [delete]
func (ctrl *Controller) DeleteUser(c *gin.Context) {
	var payments []models.PaymentDBEntry

	name := c.Param("name")

	// if the user is a part of any payments, we should not delete them

	// First resolve the ID of the user
	userId, err := ctrl.resolveUserIdFromName(name)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}

		log.Printf("failed to find user to delete: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not find user to delete"})
		return
	}

	paymentQueryErr := ctrl.db.
		Preload("Owers").
		Where("payer_id = ?", *userId).
		Or("id IN (SELECT payment_db_entry_id FROM payment_owers WHERE user_db_entry_id = ?)", *userId).
		Find(&payments).Error
	if paymentQueryErr != nil {
		log.Printf("failed to find payments that the user is a part of: %v", paymentQueryErr)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not find payments for this user"})
		return
	}

	if len(payments) != 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "cannot delete user because it is a part of active payments"})
		return
	}

	log.Printf("payments: %+v\n", payments)

	// delete the user
	if err := ctrl.db.Unscoped().Delete(&models.UserDBEntry{}, *userId).Error; err != nil {
		log.Printf("failed to delete user: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not delete user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"info": "all gone"})
}

// GetUser godoc
// @Summary Get a user by name
// @Param name path string true "User name"
// @Router /api/v1/user/{name} [get]
func (ctrl *Controller) GetUser(c *gin.Context) {
	name := c.Param("name")

	var p models.UserDBEntry
	if err := ctrl.db.Where("name = ?", name).First(&p).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}

		log.Printf("db error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error while fetching user"})
		return
	}

	c.JSON(http.StatusOK, p)
}

// GetUser godoc
// @Summary Get all user's balances
// @Router /api/v1/user/balances/all [get]
func (ctrl *Controller) GetBalancesByUser(c *gin.Context) {
	var users []models.UserDBEntry
	if err := ctrl.db.Find(&users).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "no users found"})
			return
		}

		log.Printf("db error while fetching users: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error while fetching users"})
		return
	}

	var balances []models.BalanceDBEntry
	userBalances := make(map[string][]models.BalanceDBEntry)

	for _, user := range users {
		if err := ctrl.db.Where("from_user_id = ?", user.ID).Find(&balances).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "no balances found for user " + user.Name})
				return
			}

			log.Printf("db error while fetching balances for user %s: %v", user.Name, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error while fetching balances for user " + user.Name})
			return
		}

		userName, err := ctrl.resolveUserNameFromId(user.ID)
		if err != nil {
			log.Printf("failed to resolve user name from id: %s", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error while fetching user name from id"})
			return
		}

		userBalances[*userName] = balances
	}

	c.JSON(http.StatusOK, userBalances)
}

// GetUser godoc
// @Summary Get all user's balances
// @Router /api/v1/user/debts/all [get]
func (ctrl *Controller) GetUsersDebts(c *gin.Context) {
	var users []models.UserDBEntry
	if err := ctrl.db.Find(&users).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "no users found"})
			return
		}

		log.Printf("db error while fetching users: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error while fetching users"})
		return
	}

	var balances []models.BalanceDBEntry
	userBalances := make(map[string][]models.BalanceDBEntry)

	for _, user := range users {
		if err := ctrl.db.Where("from_user_id = ?", user.ID).Find(&balances).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "no balances found for user " + user.Name})
				return
			}

			log.Printf("db error while fetching balances for user %s: %v", user.Name, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error while fetching balances for user " + user.Name})
			return
		}

		userName, err := ctrl.resolveUserNameFromId(user.ID)
		if err != nil {
			log.Printf("failed to resolve user name from id: %s", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error while fetching user name from id"})
			return
		}

		userBalances[*userName] = balances
	}

	// Now we need to sum up the total amount owed to each user...

	results := make(map[string]map[string]float64)

	for user, debts := range userBalances {
		userOwesWhoMap := make(map[string]float64)
		for _, entry := range debts {
			userWeOweId := entry.FromUserID
			userWeOweName, err := ctrl.resolveUserNameFromId(userWeOweId)
			if err != nil {
				log.Printf("failed to resolve user name from id: %s", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error while fetching user name from id"})
				return
			}

			if _, ok := userOwesWhoMap[*userWeOweName]; !ok {
				userOwesWhoMap[*userWeOweName] = 0.0
			} else {
				userOwesWhoMap[*userWeOweName] += entry.Amount
			}
		}

		results[user] = userOwesWhoMap
	}

	c.JSON(http.StatusOK, results)
}

// GetUsers godoc
// @Summary Get all users
// @Router /api/v1/user/all [get]
func (ctrl *Controller) GetUsers(c *gin.Context) {
	var users []models.UserDBEntry

	if err := ctrl.db.Find(&users).Error; err != nil {
		log.Printf("db error while fetching users: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error while fetching users"})
		return
	}

	c.JSON(http.StatusOK, users)
}

// UpdateBalances godoc
// @Summary Update all balances
// @Router /api/v1/balance/all [post]
func (ctrl *Controller) UpdateBalances(c *gin.Context) {
	type debtKey struct {
		From     uint
		To       uint
		Currency string
	}

	var payments []models.PaymentDBEntry
	if err := ctrl.db.Preload("Owers").Find(&payments).Error; err != nil {
		log.Printf("db error while fetching payments: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error while fetching payments"})
		return
	}

	raw := map[debtKey]float64{}

	for _, paymentDbEntry := range payments {
		payerId := paymentDbEntry.PayerID
		owers := paymentDbEntry.Owers

		amountPerOwer := paymentDbEntry.Amount / float64(len(owers))

		for _, ower := range owers {
			//log.Printf("Adding payment %d: %s owes %s $%.2f in currency %s\n", paymentDbEntry.ID, ower.Name, paymentDbEntry.PayerName, amountPerOwer, paymentDbEntry.FromExchangeRate)
			raw[debtKey{From: ower.ID, To: payerId, Currency: paymentDbEntry.FromExchangeRate}] += amountPerOwer
		}
	}

	// Now that we've gone through each payment and calculated the raw debts, we need to
	// calculate the ACTUAL debts. If A owes B $30 and B owes A $10, then you can rewrite
	// this as A owing B $20 and B owing A $0
	log.Printf("Raw before calculating actual debts: %+v\n", raw)
	for debtEntry, amount := range raw {
		log.Printf("Evaluating debt entry: %d owes %d: $%.2f\n", debtEntry.From, debtEntry.To, amount)
		fromUser := debtEntry.From // person A
		toUser := debtEntry.To     // person B

		inverseDebt := raw[debtKey{From: toUser, To: fromUser, Currency: debtEntry.Currency}]

		log.Printf("Discovered inverse debt: %d owes %d: $%.2f\n", toUser, fromUser, inverseDebt)

		if amount >= inverseDebt {
			raw[debtEntry] = amount - inverseDebt

			// Remove the inverse debt since it's now accounted for rather than zeroing it
			delete(raw, debtKey{From: toUser, To: fromUser, Currency: debtEntry.Currency})
		}
	}

	log.Printf("Raw after calculating actual debts: %+v\n", raw)

	// When outputting to the API, we should use human-friendly names, not ids
	var response []models.Balance
	var dbEntries []models.BalanceDBEntry

	// Print the raw balances for debugging with the user names instead of ids
	for debtEntry, amount := range raw {
		// if amount == 0 {
		// 	continue // skip zero balances
		// }

		fromUserName, err := ctrl.resolveUserNameFromId(debtEntry.From)
		if err != nil {
			log.Printf("db error while fetching user name from id for balance: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "db error while fetching user name from id for balance"})
			return
		}

		toUserName, err := ctrl.resolveUserNameFromId(debtEntry.To)
		if err != nil {
			log.Printf("db error while fetching user name from id for balance: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "db error while fetching user name from id for balance"})
			return
		}

		response = append(response, models.Balance{
			FromUser: *fromUserName,
			ToUser:   *toUserName,
			Amount:   amount,
			Currency: debtEntry.Currency,
		})

		dbEntries = append(dbEntries, models.BalanceDBEntry{
			FromUserID: debtEntry.From,
			ToUserID:   debtEntry.To,
			Amount:     amount,
			Currency:   debtEntry.Currency,
		})

		log.Printf("Balance: %s owes %s: $%.2f\n", *fromUserName, *toUserName, amount)
	}

	log.Printf("db entries to commit: %+v", dbEntries)

	// recommit the balances to the database
	// we can do this in a transaction to ensure that the balances are always consistent with the payments
	if err := ctrl.commitBalanceUpdates(dbEntries); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetBalances godoc
// @Summary Get all balances
// @Router /api/v1/balance/all [get]
func (ctrl *Controller) GetBalances(c *gin.Context) {
	var users []models.UserDBEntry
	var dbResults []models.BalanceDBEntry
	var results []models.Balance
	if err := ctrl.db.Order("from_user_id ASC").Order("to_user_id ASC").Order("id ASC").Find(&dbResults).Error; err != nil {
		log.Printf("db error while fetching debts: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error while fetching debts"})
		return
	}

	log.Printf("I found these balances: %+v", dbResults)

	if err := ctrl.db.Find(&users).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "no users found"})
			return
		}

		log.Printf("db error while fetching users: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error while fetching users"})
		return
	}

	for _, user := range users {
		for _, entry := range dbResults {
			var err error
			var fromUser, toUser *string

			if fromUser, err = ctrl.resolveUserNameFromId(entry.FromUserID); err != nil {
				log.Printf("db error while fetching from user: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error while from user"})
				return
			}

			if *fromUser != user.Name {
				continue
			}

			if toUser, err = ctrl.resolveUserNameFromId(entry.ToUserID); err != nil {
				log.Printf("db error while fetching to user: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error while to user"})
				return
			}
			results = append(results, models.Balance{
				FromUser: *fromUser,
				ToUser:   *toUser,
				Amount:   entry.Amount,
				Currency: entry.Currency,
			})
		}
	}

	/*
		Want to return something like :
		{
			"alice": {
				"bob": {
					"USD": 20,
					"YEN": 3000,
				}
			}
		}

	*/

	c.JSON(http.StatusOK, results)
}

// GetExchangeRate godoc
// @Tags         admin
// @Summary Get exchange rate between two currencies
// @Param from query string true "Base currency code (e.g. USD)"
// @Param to query string true "Target currency code (e.g. EUR)"
// @Router /api/v1/admin/exchange-rate [get]
func (ctrl *Controller) GetExchangeRate(c *gin.Context) {
	fromCurrency := c.Query("from")
	toCurrency := c.Query("to")

	if fromCurrency == "" || toCurrency == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing from or to query parameter"})
		return
	}

	rate, err := fetchExchangeRate(strings.ToUpper(fromCurrency), strings.ToUpper(toCurrency))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch the desired exchange rate"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"rate": rate})
}
